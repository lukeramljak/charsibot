FROM oven/bun:1 AS base
WORKDIR /usr/src/app

FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock /temp/dev/
COPY shared/package.json /temp/dev/shared/
COPY twitch/package.json /temp/dev/twitch/
COPY twitch-overlay/package.json /temp/dev/twitch-overlay/
RUN cd /temp/dev && bun install --frozen-lockfile

# install with --production (exclude devDependencies)
RUN mkdir -p /temp/prod
COPY package.json bun.lock /temp/prod/
COPY shared/package.json /temp/prod/shared/
COPY twitch/package.json /temp/prod/twitch/
COPY twitch-overlay/package.json /temp/prod/twitch-overlay/
RUN cd /temp/prod && bun install --frozen-lockfile --production

# copy node_modules from temp directory
# then copy all (non-ignored) project files into the image
FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules
COPY --from=install /temp/dev/twitch-overlay/node_modules twitch-overlay/node_modules
COPY package.json .
COPY shared ./shared
COPY twitch-overlay ./twitch-overlay

# build the SvelteKit app
ENV NODE_ENV=production
WORKDIR /usr/src/app/twitch-overlay
RUN bunx svelte-kit sync
RUN bun run build

# copy production dependencies and built app into final image
FROM base AS release
COPY --from=install /temp/prod/node_modules node_modules
COPY --from=install /temp/prod/twitch-overlay/node_modules twitch-overlay/node_modules
COPY --from=prerelease /usr/src/app/package.json .
COPY --from=prerelease /usr/src/app/twitch-overlay/build ./twitch-overlay/build
COPY --from=prerelease /usr/src/app/twitch-overlay/package.json ./twitch-overlay/

# run the app
USER bun
EXPOSE 3000
ENV PORT=3000
CMD ["bun", "run", "twitch-overlay/build/index.js"]

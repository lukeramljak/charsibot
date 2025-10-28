FROM oven/bun:1 AS base
WORKDIR /usr/src/app

FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock /temp/dev/
COPY shared/package.json /temp/dev/shared/
COPY discord/package.json /temp/dev/discord/
RUN cd /temp/dev && bun install --frozen-lockfile

# install with --production (exclude devDependencies)
RUN mkdir -p /temp/prod
COPY package.json bun.lock /temp/prod/
COPY shared/package.json /temp/prod/shared/
COPY discord/package.json /temp/prod/discord/
RUN cd /temp/prod && bun install --frozen-lockfile --production

# copy node_modules from temp directory
# then copy all (non-ignored) project files into the image
FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules
COPY --from=install /temp/dev/discord/node_modules discord/node_modules
COPY package.json .
COPY shared ./shared
COPY discord ./discord

# copy production files
ENV NODE_ENV=production

# copy production dependencies and source code into final image
FROM base AS release
COPY --from=install /temp/prod/node_modules node_modules
COPY --from=install /temp/prod/discord/node_modules discord/node_modules
COPY --from=prerelease /usr/src/app/package.json .
COPY --from=prerelease /usr/src/app/shared ./shared
COPY --from=prerelease /usr/src/app/discord ./discord

# create data directory and set permissions
RUN mkdir -p /data && chown -R bun:bun /data

# run the app
USER bun
ENTRYPOINT [ "bun", "run", "discord/src/index.ts" ]

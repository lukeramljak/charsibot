FROM node:22-bookworm-slim AS build
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build
RUN pnpm prune --prod

FROM gcr.io/distroless/nodejs22-debian12
WORKDIR /app
COPY --from=build /app/build ./build
COPY --from=build /app/node_modules node_modules/

EXPOSE 3000
ENV NODE_ENV=production
ENV PORT=3000
CMD ["./build/index.js"]

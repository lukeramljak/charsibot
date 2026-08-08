FROM node:22-bookworm-slim AS build
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm run build
RUN pnpm prune --prod

FROM node:22-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends wget && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/web/build ./build
COPY --from=build /app/web/node_modules ./node_modules
EXPOSE 8081
CMD ["node", "build/server.js"]

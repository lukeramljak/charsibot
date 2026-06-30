FROM node:22-bookworm-slim AS web-build
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm run build

FROM golang:1.25-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-build /app/web/build ./server/web
RUN CGO_ENABLED=0 GOOS=linux go build -o /charsibot ./cmd/charsibot

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends wget && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /charsibot /charsibot
EXPOSE 8081
CMD ["/charsibot"]

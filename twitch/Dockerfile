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

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /charsibot /charsibot
EXPOSE 8081
CMD ["/charsibot"]

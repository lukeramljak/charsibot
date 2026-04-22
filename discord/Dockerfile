FROM oven/bun:1 AS build

WORKDIR /app
COPY . .

RUN bun install --frozen-lockfile
RUN bun build src/index.ts --compile --outfile charsibot-discord
RUN /app/charsibot-discord commands register

FROM gcr.io/distroless/base-debian13
COPY --from=build /app/charsibot-discord /charsibot-discord
ENTRYPOINT ["/charsibot-discord"]

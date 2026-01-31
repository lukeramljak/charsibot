FROM golang:1.25 as build
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /charsibot ./cmd/charsibot

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /charsibot /charsibot
EXPOSE 8081
CMD ["/charsibot"]

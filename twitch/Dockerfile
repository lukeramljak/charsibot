FROM golang:alpine
WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/charsibot ./cmd/charsibot

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY /bin/charsibot /app/charsibot

EXPOSE 8081

CMD ["/app/charsibot"]

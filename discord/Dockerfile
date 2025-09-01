FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o charsibot-discord ./cmd/bot

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/charsibot-discord .

CMD ["./charsibot-discord"]

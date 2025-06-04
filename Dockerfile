FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o charsibot ./cmd/bot

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/charsibot .

COPY .env .

CMD ["./charsibot"]

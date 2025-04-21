# Stage 1: Build the Go binary statically
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -ldflags="-s -w" -o charsibot

# Stage 2: Minimal runtime image
FROM scratch

WORKDIR /root/

COPY --from=builder /app/charsibot .
COPY --from=builder /app/.env .

CMD ["./charsibot"]

# Builder stage
FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /bin/lb \
    ./cmd/server/main.go

# Runtime stage
FROM alpine:3.20

LABEL maintainer="0x0FACED"
LABEL description="High-performance (maybe) Load Balancer with Redis-backed rate limiting"

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/lb /app/lb

COPY config/config.yaml /app/config/config.yaml

RUN mkdir -p /app/logs && \
    chmod -R 755 /app/logs

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

USER nobody:nobody

CMD ["./lb"]

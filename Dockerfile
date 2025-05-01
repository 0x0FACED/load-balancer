FROM golang:1.24-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/app ./cmd/app

# --- Runtime ---
FROM alpine:3.20

LABEL maintainer="0x0FACED"

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/app /app/app
COPY config/config.json /app/config/config.json

RUN mkdir -p /app/logs

EXPOSE 8080 8081 8082 8083 8084 8085 8086

CMD ["./app"]

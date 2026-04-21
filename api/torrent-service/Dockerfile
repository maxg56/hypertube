FROM golang:1.26-alpine AS development
WORKDIR /app
RUN apk --no-cache add gcc musl-dev
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8004
CMD ["air", "-c", ".air.toml"]

FROM golang:1.26-alpine AS builder
WORKDIR /app
RUN apk --no-cache add gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY src/ ./src/
RUN CGO_ENABLED=1 GOOS=linux go build -o torrent-service ./src

FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates
RUN addgroup -g 1001 -S torrentservice && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G torrentservice torrentservice
WORKDIR /app
COPY --from=builder --chown=torrentservice:torrentservice /app/torrent-service .
USER torrentservice
EXPOSE 8004
CMD ["./torrent-service"]

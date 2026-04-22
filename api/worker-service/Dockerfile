FROM golang:1.26-alpine AS development
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8006
CMD ["air", "-c", ".air.toml"]

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src/ ./src/
RUN CGO_ENABLED=0 GOOS=linux go build -o worker-service ./src

FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates
RUN addgroup -g 1001 -S workerservice && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G workerservice workerservice
WORKDIR /app
COPY --from=builder --chown=workerservice:workerservice /app/worker-service .
USER workerservice
EXPOSE 8006
CMD ["./worker-service"]

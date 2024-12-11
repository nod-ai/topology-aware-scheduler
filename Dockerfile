# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /build
COPY . .

# Install build dependencies
RUN apk add --no-cache git make

# Build the scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -a -o topology-scheduler cmd/scheduler/main.go

# Final stage
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /build/topology-scheduler .
COPY --from=builder /build/config /app/config

RUN adduser -D -u 1000 scheduler && \
    chown -R scheduler:scheduler /app

USER scheduler

ENTRYPOINT ["/app/topology-scheduler"]

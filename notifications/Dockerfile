FROM golang:1.27.1 AS builder

WORKDIR /app

COPY go.work ./
COPY notifications/go.mod notifications/go.sum ./notifications/
COPY pkg/go.mod pkg/go.sum ./pkg/

WORKDIR /app/notifications
RUN --mount=type=cache,target=/go/pkg/mod GOWORK=off go mod download

WORKDIR /app
COPY . .

WORKDIR /app/notifications
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -o /bin/notifications ./cmd/app

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /bin/notifications /app/notifications

CMD ["./notifications"]

FROM golang:1.27.1 AS builder

WORKDIR /app

COPY go.work ./
COPY cart/go.mod cart/go.sum ./cart/
COPY pkg/go.mod pkg/go.sum ./pkg/

WORKDIR /app/cart
RUN --mount=type=cache,target=/go/pkg/mod GOWORK=off go mod download

WORKDIR /app
COPY . .

WORKDIR /app/cart
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/cart ./cmd/app

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/cart /app/cart

CMD ["./cart"]

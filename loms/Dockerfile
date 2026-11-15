FROM golang:1.27.1 AS builder

WORKDIR /app

COPY go.work ./
COPY loms/go.mod loms/go.sum ./loms/
COPY pkg/go.mod pkg/go.sum ./pkg/

WORKDIR /app/loms
RUN --mount=type=cache,target=/go/pkg/mod GOWORK=off go mod download

WORKDIR /app
COPY . .

WORKDIR /app/loms
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/loms ./cmd/app

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/loms /app/loms

CMD ["./loms"]

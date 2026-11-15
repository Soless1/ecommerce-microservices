module github.com/igoroutine-courses/microservices.ecommerce.loms

go 1.26.2

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0
	github.com/igoroutine-courses/microservices.ecommerce.pkg v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.12.0
	github.com/pressly/goose/v3 v3.30.1
	github.com/redis/go-redis/v9 v9.23.0
	github.com/segmentio/kafka-go v0.4.51
	go.uber.org/mock v0.6.0
	go.uber.org/zap v1.27.1
	golang.org/x/sync v0.25.0
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/klauspost/compress v1.20.4 // indirect
	github.com/mfridman/interpolate v0.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/sethvargo/go-retry v0.5.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261110000943-09e21b237f2c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261110000943-09e21b237f2c // indirect
)

replace github.com/igoroutine-courses/microservices.ecommerce.pkg => ../pkg

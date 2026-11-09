package config

import "github.com/caarlos0/env/v10"

type (
	Config struct {
		Clients struct {
			LOMSGrpcAddr string `env:"LOMS_GRPC_ADDR" envDefault:"localhost:50052"`
		}

		GRPC struct {
			Port        string `env:"GRPC_PORT" envDefault:"50051"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8080"`

			Endpoint string `env:"GRPC_ENDPOINT" envDefault:"localhost:50051"`
		}

		Postgres struct {
			DSN string `env:"POSTGRES_DSN" envDefault:"postgres://ecommerce_user:12345@localhost:5432/ecommerce?sslmode=disable"`
		}
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

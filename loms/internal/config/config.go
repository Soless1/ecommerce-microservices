package config

import "github.com/caarlos0/env/v10"

type (
	Config struct {
		GRPC struct {
			Port        string `env:"GRPC_PORT" envDefault:"50052"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8081"`
		}

		Postgres struct {
			DSN string `env:"POSTGRES_DSN" envDefault:"postgres://ecommerce_user:12345@localhost:5432/ecommerce?sslmode=disable"`
		}

		Clients struct {
			NotificationsGrpcAddr string `env:"NOTIFICATIONS_GRPC_ADDR" envDefault:"localhost:50053"`
		}
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

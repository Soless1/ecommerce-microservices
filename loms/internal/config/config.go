package config

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		GRPC struct {
			Port        string `env:"GRPC_PORT" envDefault:"50052"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8081"`
		}

		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
			User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
		}

		Outbox struct {
			Workers     int           `env:"OUTBOX_WORKERS" envDefault:"3"`
			BatchSize   int           `env:"OUTBOX_BATCH_SIZE" envDefault:"5"`
			FetchPeriod time.Duration `env:"OUTBOX_FETCH_PERIOD" envDefault:"5s"`
			TTL         time.Duration `env:"OUTBOX_IN_PROGRESS_TTL" envDefault:"60s"`
		}

		Kafka struct {
			Brokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
			Topic   string `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
		}
	}
)

func (c *Config) KafkaBrokerAddrs() []string {
	parts := strings.Split(c.Kafka.Brokers, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

func (c *Config) ConstructPostgresURL() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.PG.User, c.PG.Password),
		Host:     net.JoinHostPort(c.PG.Host, c.PG.Port),
		Path:     "/" + c.PG.DB,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

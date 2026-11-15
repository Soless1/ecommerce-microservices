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
			Port                   string        `env:"GRPC_PORT" envDefault:"50052"`
			GatewayPort            string        `env:"GRPC_GATEWAY_PORT" envDefault:"8081"`
			GatewayShutdownTimeout time.Duration `env:"GRPC_GATEWAY_SHUTDOWN_TIMEOUT" envDefault:"5s"`
		}

		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
			User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`

			HealthCheckPeriod time.Duration `env:"POSTGRES_HEALTH_CHECK_PERIOD" envDefault:"30s"`
			MaxConnIdleTime   time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"5m"`
			MaxConnLifetime   time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"0s"`
			MaxConns          int32         `env:"POSTGRES_MAX_CONNS" envDefault:"8"`
			MinConns          int32         `env:"POSTGRES_MIN_CONNS" envDefault:"1"`
		}

		Outbox struct {
			Workers     int           `env:"OUTBOX_WORKERS" envDefault:"3"`
			BatchSize   int           `env:"OUTBOX_BATCH_SIZE" envDefault:"5"`
			FetchPeriod time.Duration `env:"OUTBOX_FETCH_PERIOD" envDefault:"5s"`
			TTL         time.Duration `env:"OUTBOX_IN_PROGRESS_TTL" envDefault:"60s"`
		}

		Redis struct {
			Host                string        `env:"REDIS_HOST" envDefault:"localhost"`
			Port                string        `env:"REDIS_PORT" envDefault:"6379"`
			ProductCacheTTL     time.Duration `env:"REDIS_PRODUCT_CACHE_TTL" envDefault:"1m"`
			ProductCacheTimeout time.Duration `env:"REDIS_PRODUCT_CACHE_TIMEOUT" envDefault:"50ms"`
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

func (c *Config) ConstructRedisURL() string {
	u := url.URL{
		Scheme: "redis",
		Host:   net.JoinHostPort(c.Redis.Host, c.Redis.Port),
	}
	return u.String()
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

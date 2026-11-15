package config

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Kafka struct {
		Brokers         string        `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
		Topic           string        `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
		ConsumerGroup   string        `env:"KAFKA_CONSUMER_GROUP" envDefault:"notifications"`
		DeliveryTimeout time.Duration `env:"KAFKA_DELIVERY_TIMEOUT" envDefault:"30s"`
		CommitTimeout   time.Duration `env:"KAFKA_COMMIT_TIMEOUT" envDefault:"10s"`
	}

	TelegramNotifier struct {
		Token    string `env:"TELEGRAM_TOKEN" envDefault:""`
		ClientID int64  `env:"TELEGRAM_CHAT_ID" envDefault:""`
	}
	Clients struct {
		CallbackAddr string        `env:"CALLBACK_ADDR" envDefault:""`
		HTTPTimeout  time.Duration `env:"NOTIFICATIONS_HTTP_TIMEOUT" envDefault:"5s"`
	}
}

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

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}

package config

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Kafka struct {
		Brokers       string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
		Topic         string `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
		ConsumerGroup string `env:"KAFKA_CONSUMER_GROUP" envDefault:"notifications"`
	}

	TelegramNotifier struct {
		Token    string `env:"TELEGRAM_TOKEN" envDefault:""`
		ClientID int64  `env:"TELEGRAM_CHAT_ID" envDefault:""`
	}
	Clients struct {
		CallbackAddr string `env:"CALLBACK_ADDR" envDefault:""`
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

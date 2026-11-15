package app

import (
	"context"
	"os/signal"
	"strings"
	"syscall"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	kafkaconsumer "github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/consumer/kafka"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/usecase/notifier"
	"go.uber.org/zap"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(cfg.KafkaBrokerAddrs()) == 0 {
		logger.Fatal("no kafka brokers configured (KAFKA_BROKERS)")
	}

	var statusNotifier interface {
		NotifyOrderStatusChanged(ctx context.Context, userID, orderID int64, status string) error
	}
	var err error
	if strings.TrimSpace(cfg.Clients.CallbackAddr) != "" {
		statusNotifier, err = notifier.NewOrderStatusRestNotifier(logger, cfg)
	} else {
		statusNotifier, err = notifier.NewOrderStatusTelegramNotifier(logger, cfg)
	}

	if err != nil {
		logger.Fatal("notifier init error", zap.Error(err))
	}

	if err := kafkaconsumer.RunConsumer(ctx, logger, cfg, statusNotifier); err != nil {
		logger.Fatal("kafka consumer stopped", zap.Error(err))
	}
}

package app

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	kafkaconsumer "github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/consumer/kafka"
	tgnotifier "github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/usecase/notifier"
	"go.uber.org/zap"
)

func Run(logger *zap.Logger, cfg *config.Config) {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if len(cfg.KafkaBrokerAddrs()) == 0 {
		logger.Fatal("no kafka brokers configured (KAFKA_BROKERS)")
	}

	notifier, err := tgnotifier.NewOrderStatusTelegramNotifier(logger, cfg)
	//notifier, err := tgnotifier.NewOrderStatusRestNotifier(logger, cfg)

	if err != nil {
		logger.Fatal("tg notifier init error", zap.Error(err))
	}

	if err := kafkaconsumer.RunConsumer(ctx, logger, cfg, notifier); err != nil {
		logger.Fatal("kafka consumer stopped", zap.Error(err))
	}
}

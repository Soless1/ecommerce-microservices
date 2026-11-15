package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	"go.uber.org/zap"
)

const (
	retryInterval  = 500 * time.Millisecond
	sessionTimeout = 10 * time.Second
	pollTimeout    = time.Second
)

type (
	notifier interface {
		NotifyOrderStatusChanged(
			ctx context.Context,
			userID,
			orderID int64,
			status string,
		) error
	}
)

type eventPayload struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

func retry(ctx context.Context, action func() error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := action()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w (last attempt: %v)", ctx.Err(), err)
		case <-time.After(retryInterval):
		}
	}
}

func RunConsumer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	notifier notifier,
) error {
	brokers := cfg.KafkaBrokerAddrs()
	if len(brokers) == 0 {
		return errors.New("empty kafka brokers")
	}

	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers":        strings.Join(brokers, ","),
		"group.id":                 cfg.Kafka.ConsumerGroup,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
		"session.timeout.ms":       int(sessionTimeout.Milliseconds()),
	})

	if err != nil {
		return err
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("close kafka consumer", zap.Error(err))
		}
	}()

	if err := consumer.SubscribeTopics([]string{cfg.Kafka.Topic}, nil); err != nil {
		return err
	}

	logger.Info("kafka consumer started",
		zap.Strings("brokers", brokers),
		zap.String("topic", cfg.Kafka.Topic),
		zap.String("group", cfg.Kafka.ConsumerGroup),
	)

	for {
		select {
		case <-ctx.Done():
			logger.Info("kafka consumer stopped", zap.Error(ctx.Err()))
			return nil
		default:
		}

		// TODO: inbox + batch
		ev := consumer.Poll(int(pollTimeout.Milliseconds()))
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *ckafka.Message:
			logger.Info("message fetched",
				zap.String("topic", *e.TopicPartition.Topic),
				zap.Int32("partition", e.TopicPartition.Partition),
				zap.Int64("offset", int64(e.TopicPartition.Offset)),
				zap.ByteString("key", e.Key),
				zap.ByteString("value", e.Value),
			)

			var payload eventPayload
			if err := json.Unmarshal(e.Value, &payload); err != nil {
				logger.Error("kafka message unmarshal",
					zap.Error(err),
					zap.ByteString("raw", e.Value),
				)

				// TODO: DLQ
				continue
			}

			eventLogger := logger.With(
				zap.String("topic", *e.TopicPartition.Topic),
				zap.Int32("partition", e.TopicPartition.Partition),
				zap.Int64("offset", int64(e.TopicPartition.Offset)),
				zap.Int64("user_id", payload.UserID),
				zap.Int64("order_id", payload.OrderID),
				zap.String("status", payload.Status),
			)

			deliveryCtx, cancelDelivery := context.WithTimeout(ctx, cfg.Kafka.DeliveryTimeout)
			err := retry(deliveryCtx, func() error {
				err := notifier.NotifyOrderStatusChanged(
					deliveryCtx, payload.UserID, payload.OrderID, payload.Status,
				)
				if err != nil && ctx.Err() == nil {
					eventLogger.Warn("notification delivery failed", zap.Error(err))
				}
				return err
			})
			cancelDelivery()

			if err != nil {
				if ctx.Err() != nil {
					return nil // штатная остановка
				}
				return fmt.Errorf("deliver notification for order %d: %w", payload.OrderID, err)
			}

			commitCtx, cancelCommit := context.WithTimeout(ctx, cfg.Kafka.CommitTimeout)
			err = retry(commitCtx, func() error {
				_, commitErr := consumer.CommitMessage(e)
				if commitErr != nil && ctx.Err() == nil {
					eventLogger.Warn("kafka offset commit failed", zap.Error(commitErr))
				}
				return commitErr
			})
			cancelCommit()

			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("commit offset for order %d: %w", payload.OrderID, err)
			}

			eventLogger.Info("message committed")

		case ckafka.Error:
			logger.Warn("kafka poll error",
				zap.String("code", e.Code().String()),
				zap.Error(e),
			)

			time.Sleep(1 * time.Second)

		case ckafka.AssignedPartitions:
			logger.Info("kafka partitions assigned",
				zap.Any("partitions", e.Partitions),
			)

			if err := consumer.Assign(e.Partitions); err != nil {
				logger.Error("kafka assign partitions", zap.Error(err))
			}

		case ckafka.RevokedPartitions:
			logger.Info("kafka partitions revoked",
				zap.Any("partitions", e.Partitions),
			)

			if err := consumer.Unassign(); err != nil {
				logger.Error("kafka unassign partitions", zap.Error(err))
			}
		}
	}
}

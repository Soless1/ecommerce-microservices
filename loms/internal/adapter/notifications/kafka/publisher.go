package kafka

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
			RequiredAcks:           kafka.RequireAll},
	}
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}

func (p *Publisher) SendOrderStatusChangedNotification(
	ctx context.Context,
	body port.OrderStatusChangedNotification,
) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(body.OrderID, 10)),
		Value: raw,
	})
}

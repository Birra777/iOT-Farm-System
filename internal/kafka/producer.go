package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer wraps a kafka-go writer with structured logging.
type Producer struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// NewProducer creates a new Kafka producer writing to the given brokers.
// balancer uses field_id as the partition key for ordered per-field processing.
func NewProducer(brokers []string, logger *slog.Logger) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
	}
	return &Producer{writer: w, logger: logger}
}

// Publish serialises v to JSON and sends it to topic, keyed by key.
func (p *Producer) Publish(ctx context.Context, topic, key string, v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("kafka publish failed", "topic", topic, "key", key, "error", err)
		return fmt.Errorf("write message: %w", err)
	}

	p.logger.Debug("published message", "topic", topic, "key", key, "bytes", len(payload))
	return nil
}

// Close flushes pending messages and closes the underlying writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

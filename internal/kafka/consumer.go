package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps a kafka-go reader with structured logging and offset management.
type Consumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

// NewConsumer creates a Consumer subscribed to the given topic and group.
func NewConsumer(brokers []string, topic, groupID string, logger *slog.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10 MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 0, // manual commit only
		StartOffset:    kafka.FirstOffset,
		Logger:         kafka.LoggerFunc(func(msg string, args ...any) {}), // silence default logs
	})
	return &Consumer{reader: r, logger: logger}
}

// Message is the raw Kafka message returned to callers.
type Message = kafka.Message

// ReadMessage blocks until a message is available or ctx is cancelled.
func (c *Consumer) ReadMessage(ctx context.Context) (Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return Message{}, fmt.Errorf("fetch message: %w", err)
	}
	return msg, nil
}

// Commit marks the message as processed, advancing the consumer offset.
func (c *Consumer) Commit(ctx context.Context, msg Message) error {
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		c.logger.Error("kafka commit failed", "topic", msg.Topic, "offset", msg.Offset, "error", err)
		return fmt.Errorf("commit message: %w", err)
	}
	return nil
}

// Close closes the reader cleanly.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

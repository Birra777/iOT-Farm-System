package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/agristream/agristream/internal/config"
	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/kafka"
	"github.com/agristream/agristream/internal/models"
	"github.com/agristream/agristream/internal/window"
)

const (
	consumerGroup  = "stream-processor-group"
	windowDuration = time.Minute
	aggInterval    = time.Minute
	maxRetries     = 3
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "processor")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		logger.Info("shutdown signal received", "signal", sig.String())
		cancel()
	}()

	pool, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	readingsRepo := db.NewReadingsRepo(pool)
	brokers := strings.Split(cfg.KafkaBrokers, ",")

	soilConsumer := kafka.NewConsumer(brokers, cfg.KafkaTopicSoilRaw, consumerGroup, logger)
	weatherConsumer := kafka.NewConsumer(brokers, cfg.KafkaTopicWeatherRaw, consumerGroup, logger)
	producer := kafka.NewProducer(brokers, logger)

	defer func() {
		soilConsumer.Close()
		weatherConsumer.Close()
		if err := producer.Close(); err != nil {
			logger.Error("producer close error", "error", err)
		}
	}()

	registry := window.NewRegistry(windowDuration)

	logger.Info("stream processor starting",
		"topics", []string{cfg.KafkaTopicSoilRaw, cfg.KafkaTopicWeatherRaw},
		"group", consumerGroup,
	)

	var wg sync.WaitGroup

	wg.Go(func() { consumeLoop(ctx, soilConsumer, producer, readingsRepo, registry, cfg, logger) })
	wg.Go(func() { consumeLoop(ctx, weatherConsumer, producer, readingsRepo, registry, cfg, logger) })
	wg.Go(func() { aggregationLoop(ctx, registry, producer, cfg, logger) })

	wg.Wait()
	logger.Info("processor stopped cleanly")
}

// consumeLoop reads messages from a consumer, validates, writes to DB, updates windows, commits offset.
func consumeLoop(
	ctx context.Context,
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	repo *db.ReadingsRepo,
	registry *window.Registry,
	cfg *config.Config,
	logger *slog.Logger,
) {
	for {
		if ctx.Err() != nil {
			return
		}

		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error("read message failed", "error", err)
			continue
		}

		var reading models.SensorReading
		if err := json.Unmarshal(msg.Value, &reading); err != nil {
			logger.Warn("malformed message — routing to DLQ",
				"topic", msg.Topic,
				"offset", msg.Offset,
				"error", err,
			)
			sendToDLQ(ctx, producer, cfg.KafkaTopicDLQ, msg, "unmarshal_error", err, logger)
			// Commit so we don't reprocess the bad message.
			_ = consumer.Commit(ctx, msg)
			continue
		}

		if err := validate(reading); err != nil {
			logger.Warn("invalid reading — routing to DLQ",
				"sensor_id", reading.SensorID,
				"metric", reading.Metric,
				"error", err,
			)
			sendToDLQ(ctx, producer, cfg.KafkaTopicDLQ, msg, "validation_error", err, logger)
			_ = consumer.Commit(ctx, msg)
			continue
		}

		// Write to DB with retries. Offset is committed only on success.
		if err := writeWithRetry(ctx, repo, reading, maxRetries, logger); err != nil {
			// After maxRetries, send to DLQ but don't block the pipeline.
			sendToDLQ(ctx, producer, cfg.KafkaTopicDLQ, msg, "db_write_error", err, logger)
		}

		// Update the sliding window regardless of DB outcome so aggregations continue.
		key := windowKey(reading.FieldID, reading.Metric)
		registry.Add(key, reading.Value, reading.Timestamp)

		// Commit offset after DB write attempt — at-least-once semantics.
		if err := consumer.Commit(ctx, msg); err != nil {
			logger.Error("offset commit failed", "offset", msg.Offset, "error", err)
		}

		logger.Info("reading processed",
			"sensor_id", reading.SensorID,
			"field_id", reading.FieldID,
			"metric", reading.Metric,
			"value", reading.Value,
		)
	}
}

// aggregationLoop fires every aggInterval, snapshots all windows, and publishes aggregated readings.
func aggregationLoop(
	ctx context.Context,
	registry *window.Registry,
	producer *kafka.Producer,
	cfg *config.Config,
	logger *slog.Logger,
) {
	ticker := time.NewTicker(aggInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			snapshot := registry.Snapshot()
			for key, stats := range snapshot {
				fieldID, sensorType, metric, err := parseWindowKey(key)
				if err != nil {
					logger.Warn("bad window key", "key", key)
					continue
				}

				agg := models.AggregatedReading{
					FieldID:     fieldID,
					SensorType:  sensorType,
					Metric:      metric,
					Avg:         stats.Avg,
					Min:         stats.Min,
					Max:         stats.Max,
					Count:       stats.Count,
					WindowStart: stats.WindowStart,
					WindowEnd:   stats.WindowEnd,
				}

				if err := producer.Publish(ctx, cfg.KafkaTopicAggregated, fieldID, agg); err != nil {
					if ctx.Err() != nil {
						return
					}
					logger.Error("aggregated publish failed", "key", key, "error", err)
				} else {
					logger.Info("aggregated reading published",
						"field_id", fieldID,
						"metric", metric,
						"avg", stats.Avg,
						"count", stats.Count,
					)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// writeWithRetry attempts to insert the reading into PostgreSQL up to maxRetries times.
func writeWithRetry(ctx context.Context, repo *db.ReadingsRepo, reading models.SensorReading, maxRetries int, logger *slog.Logger) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := repo.Insert(ctx, reading); err != nil {
			lastErr = err
			logger.Warn("db insert failed, retrying",
				"attempt", attempt,
				"sensor_id", reading.SensorID,
				"error", err,
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt*100) * time.Millisecond):
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("db insert failed after %d attempts: %w", maxRetries, lastErr)
}

// dlqEnvelope wraps the original payload with failure metadata for the DLQ.
type dlqEnvelope struct {
	OriginalTopic string `json:"original_topic"`
	Offset        int64  `json:"offset"`
	FailureReason string `json:"failure_reason"`
	ErrorMessage  string `json:"error_message"`
	FailedAt      string `json:"failed_at"`
	Payload       []byte `json:"payload"`
}

func sendToDLQ(ctx context.Context, producer *kafka.Producer, dlqTopic string, msg kafka.Message, reason string, cause error, logger *slog.Logger) {
	env := dlqEnvelope{
		OriginalTopic: msg.Topic,
		Offset:        msg.Offset,
		FailureReason: reason,
		ErrorMessage:  cause.Error(),
		FailedAt:      time.Now().UTC().Format(time.RFC3339),
		Payload:       msg.Value,
	}
	if err := producer.Publish(ctx, dlqTopic, string(msg.Key), env); err != nil {
		logger.Error("DLQ publish failed", "reason", reason, "error", err)
	}
}

// validate checks that a SensorReading has all required fields populated.
func validate(r models.SensorReading) error {
	if r.SensorID == "" {
		return fmt.Errorf("missing sensor_id")
	}
	if r.FieldID == "" {
		return fmt.Errorf("missing field_id")
	}
	if r.Metric == "" {
		return fmt.Errorf("missing metric")
	}
	if r.Timestamp.IsZero() {
		return fmt.Errorf("missing timestamp")
	}
	return nil
}

// windowKey builds a map key from field + metric.
// Format: "<fieldID>|<sensorType>|<metric>"
func windowKey(fieldID, metric string) string {
	return fieldID + "|" + metric
}

// parseWindowKey splits a window key back into its components.
func parseWindowKey(key string) (fieldID string, sensorType models.SensorType, metric string, err error) {
	parts := strings.SplitN(key, "|", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid window key: %q", key)
	}
	fieldID = parts[0]
	metric = parts[1]

	// Derive sensor type from metric prefix.
	if strings.HasPrefix(metric, "soil.") {
		sensorType = models.SensorTypeSoil
	} else if strings.HasPrefix(metric, "weather.") {
		sensorType = models.SensorTypeWeather
	} else {
		sensorType = models.SensorTypeSoil // fallback
	}

	return fieldID, sensorType, metric, nil
}

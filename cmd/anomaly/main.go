package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/agristream/agristream/internal/config"
	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/email"
	"github.com/agristream/agristream/internal/kafka"
	"github.com/agristream/agristream/internal/models"
	"github.com/agristream/agristream/internal/notifications"
	"github.com/agristream/agristream/internal/rules"
)

const (
	consumerGroup  = "anomaly-detector-group"
	cooldownWindow = 10 * time.Minute
)

// cooldownKey uniquely identifies a (field, metric) pair for cooldown tracking.
func cooldownKey(fieldID, metric string) string {
	return fieldID + "|" + metric
}

// cooldownCache is a TTL map that prevents alert flooding.
// A (field_id, metric) pair is suppressed for cooldownWindow after firing.
type cooldownCache struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func newCooldownCache() *cooldownCache {
	c := &cooldownCache{entries: make(map[string]time.Time)}
	return c
}

// allowed returns true if the key is not on cooldown, and arms the cooldown if so.
func (c *cooldownCache) allowed(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if expiry, exists := c.entries[key]; exists && time.Now().Before(expiry) {
		return false
	}
	c.entries[key] = time.Now().Add(cooldownWindow)
	return true
}

// ruleStore holds the active rule set and supports hot-reload from the DB.
type ruleStore struct {
	mu    sync.RWMutex
	rules []rules.Rule
}

func (s *ruleStore) get() []rules.Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rules
}

func (s *ruleStore) reload(ctx context.Context, repo *db.ThresholdsRepo, logger *slog.Logger) {
	r, err := rules.LoadRules(ctx, repo)
	if err != nil {
		logger.Error("rule reload failed", "error", err)
		return
	}
	s.mu.Lock()
	s.rules = r
	s.mu.Unlock()
	logger.Info("alert rules reloaded from DB", "count", len(r))
}

func (s *ruleStore) reloadLoop(ctx context.Context, repo *db.ThresholdsRepo, logger *slog.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.reload(ctx, repo, logger)
		case <-ctx.Done():
			return
		}
	}
}

// gcLoop periodically removes expired entries to prevent unbounded growth.
func (c *cooldownCache) gcLoop(ctx context.Context) {
	ticker := time.NewTicker(cooldownWindow)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, expiry := range c.entries {
				if now.After(expiry) {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "anomaly-detector")

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

	alertsRepo := db.NewAlertsRepo(pool)
	notifsRepo := db.NewNotificationsRepo(pool)

	// Load field names once at startup for use in notification messages.
	fieldNames, err := loadFieldNames(ctx, db.NewFieldsRepo(pool))
	if err != nil {
		logger.Error("failed to load field names", "error", err)
		os.Exit(1)
	}
	logger.Info("field names loaded", "count", len(fieldNames))

	mailer := email.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	if mailer.Enabled() {
		logger.Info("email alerts enabled", "to", cfg.AlertEmailTo)
	} else {
		logger.Info("email alerts disabled — set SMTP_HOST to enable")
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	thresholdsRepo := db.NewThresholdsRepo(pool)
	initialRules, err := rules.LoadRules(ctx, thresholdsRepo)
	if err != nil {
		logger.Warn("could not load rules from DB, using defaults", "error", err)
		initialRules = rules.DefaultRules()
	}
	store := &ruleStore{rules: initialRules}
	go store.reloadLoop(ctx, thresholdsRepo, logger)
	cooldown := newCooldownCache()

	soilConsumer := kafka.NewConsumer(brokers, cfg.KafkaTopicSoilRaw, consumerGroup, logger)
	weatherConsumer := kafka.NewConsumer(brokers, cfg.KafkaTopicWeatherRaw, consumerGroup, logger)
	aggConsumer := kafka.NewConsumer(brokers, cfg.KafkaTopicAggregated, consumerGroup, logger)
	producer := kafka.NewProducer(brokers, logger)

	defer func() {
		soilConsumer.Close()
		weatherConsumer.Close()
		aggConsumer.Close()
		if err := producer.Close(); err != nil {
			logger.Error("producer close error", "error", err)
		}
	}()

	logger.Info("anomaly detector starting",
		"rules", len(store.get()),
		"cooldown_minutes", cooldownWindow.Minutes(),
	)

	go cooldown.gcLoop(ctx)

	var wg sync.WaitGroup
	wg.Go(func() {
		rawLoop(ctx, soilConsumer, producer, alertsRepo, notifsRepo, fieldNames, store, cooldown, cfg, mailer, logger)
	})
	wg.Go(func() {
		rawLoop(ctx, weatherConsumer, producer, alertsRepo, notifsRepo, fieldNames, store, cooldown, cfg, mailer, logger)
	})
	wg.Go(func() {
		aggLoop(ctx, aggConsumer, producer, alertsRepo, notifsRepo, fieldNames, store, cooldown, cfg, mailer, logger)
	})

	wg.Wait()
	logger.Info("anomaly detector stopped cleanly")
}

// loadFieldNames returns a map of field UUID → field name.
func loadFieldNames(ctx context.Context, repo *db.FieldsRepo) (map[string]string, error) {
	fields, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(fields))
	for _, f := range fields {
		m[f.ID] = f.Name
	}
	return m, nil
}

// rawLoop evaluates raw sensor readings from a single topic.
func rawLoop(
	ctx context.Context,
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	repo *db.AlertsRepo,
	notifsRepo *db.NotificationsRepo,
	fieldNames map[string]string,
	store *ruleStore,
	cooldown *cooldownCache,
	cfg *config.Config,
	mailer *email.Mailer,
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
			logger.Error("read failed", "error", err)
			continue
		}

		var reading models.SensorReading
		if err := json.Unmarshal(msg.Value, &reading); err != nil {
			logger.Warn("unmarshal failed — skipping", "offset", msg.Offset, "error", err)
			_ = consumer.Commit(ctx, msg)
			continue
		}

		processReading(ctx, reading, producer, repo, notifsRepo, fieldNames, store, cooldown, cfg, mailer, logger)
		_ = consumer.Commit(ctx, msg)
	}
}

// aggLoop evaluates aggregated readings for trend-based anomaly detection.
func aggLoop(
	ctx context.Context,
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	repo *db.AlertsRepo,
	notifsRepo *db.NotificationsRepo,
	fieldNames map[string]string,
	store *ruleStore,
	cooldown *cooldownCache,
	cfg *config.Config,
	mailer *email.Mailer,
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
			logger.Error("agg read failed", "error", err)
			continue
		}

		var agg models.AggregatedReading
		if err := json.Unmarshal(msg.Value, &agg); err != nil {
			logger.Warn("agg unmarshal failed — skipping", "offset", msg.Offset, "error", err)
			_ = consumer.Commit(ctx, msg)
			continue
		}

		// Convert aggregated avg into a synthetic reading for rule evaluation.
		synthetic := models.SensorReading{
			SensorID:   "aggregated",
			FieldID:    agg.FieldID,
			SensorType: agg.SensorType,
			Metric:     agg.Metric,
			Value:      agg.Avg,
			Unit:       "",
			Timestamp:  agg.WindowEnd,
		}

		processReading(ctx, synthetic, producer, repo, notifsRepo, fieldNames, store, cooldown, cfg, mailer, logger)
		_ = consumer.Commit(ctx, msg)
	}
}

// processReading applies all rules to a reading and persists + publishes any alerts,
// then creates a plain-English notification for each triggered alert.
func processReading(
	ctx context.Context,
	reading models.SensorReading,
	producer *kafka.Producer,
	repo *db.AlertsRepo,
	notifsRepo *db.NotificationsRepo,
	fieldNames map[string]string,
	store *ruleStore,
	cooldown *cooldownCache,
	cfg *config.Config,
	mailer *email.Mailer,
	logger *slog.Logger,
) {
	triggered := rules.ApplyRules(reading, store.get())
	for _, alert := range triggered {
		key := cooldownKey(alert.FieldID, alert.Metric)
		if !cooldown.allowed(key) {
			logger.Debug("alert suppressed by cooldown",
				"field_id", alert.FieldID,
				"metric", alert.Metric,
			)
			continue
		}

		// Persist alert to DB.
		id, err := repo.Insert(ctx, alert)
		if err != nil {
			logger.Error("alert insert failed",
				"field_id", alert.FieldID,
				"metric", alert.Metric,
				"error", err,
			)
			// Release cooldown so the alert can be retried.
			cooldown.mu.Lock()
			delete(cooldown.entries, key)
			cooldown.mu.Unlock()
			continue
		}
		alert.ID = id

		// Publish to alerts Kafka topic.
		if err := producer.Publish(ctx, cfg.KafkaTopicAlerts, alert.FieldID, alert); err != nil {
			logger.Error("alert publish failed", "id", id, "error", err)
		}

		logger.Info("alert triggered",
			"id", id,
			"field_id", alert.FieldID,
			"metric", alert.Metric,
			"value", alert.Value,
			"severity", alert.Severity,
			"message", alert.Message,
		)

		// Compose and persist a plain-English notification.
		fieldName := fieldNames[alert.FieldID]
		if fieldName == "" {
			fieldName = "Unknown Field"
		}
		notif := notifications.Compose(alert, fieldName)
		notif.AlertID = &id
		if _, err := notifsRepo.Insert(ctx, notif); err != nil {
			logger.Error("notification insert failed", "alert_id", id, "error", err)
		}

		// Send email alert if configured.
		if mailer.Enabled() && cfg.AlertEmailTo != "" {
			subject := "[AgriStream] " + notif.Title
			body := notif.Title + "\n\n" + notif.Body + "\n\nField: " + fieldName + "\nSeverity: " + string(alert.Severity)
			if err := mailer.Send(cfg.AlertEmailTo, subject, body); err != nil {
				logger.Error("email send failed", "alert_id", id, "error", err)
			} else {
				logger.Info("email alert sent", "alert_id", id, "to", cfg.AlertEmailTo)
			}
		}
	}
}

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agristream/agristream/internal/config"
	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/models"
	"github.com/agristream/agristream/internal/notifications"
	"github.com/agristream/agristream/internal/rules"
	"github.com/agristream/agristream/internal/stats"
)

const (
	lookbackWindow  = 90 * time.Minute
	forecastHorizon = 3 * time.Hour    // how far ahead we project
	runInterval     = 10 * time.Minute // how often we run
)

// predictiveMetrics are the slowly-changing metrics worth forecasting.
var predictiveMetrics = []string{
	models.MetricSoilMoisture,
	models.MetricSoilNitrogen,
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "predictor")

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

	fieldsRepo := db.NewFieldsRepo(pool)
	readingsRepo := db.NewReadingsRepo(pool)
	alertsRepo := db.NewAlertsRepo(pool)
	notifsRepo := db.NewNotificationsRepo(pool)
	thresholdsRepo := db.NewThresholdsRepo(pool)

	fieldNames, err := loadFieldNames(ctx, fieldsRepo)
	if err != nil {
		logger.Error("load field names failed", "error", err)
		os.Exit(1)
	}

	logger.Info("predictor starting", "interval_min", runInterval.Minutes())

	// Run immediately then on a ticker.
	runPrediction(ctx, fieldsRepo, readingsRepo, alertsRepo, notifsRepo, thresholdsRepo, fieldNames, logger)

	ticker := time.NewTicker(runInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			runPrediction(ctx, fieldsRepo, readingsRepo, alertsRepo, notifsRepo, thresholdsRepo, fieldNames, logger)
		case <-ctx.Done():
			logger.Info("predictor stopped cleanly")
			return
		}
	}
}

func runPrediction(
	ctx context.Context,
	fieldsRepo *db.FieldsRepo,
	readingsRepo *db.ReadingsRepo,
	alertsRepo *db.AlertsRepo,
	notifsRepo *db.NotificationsRepo,
	thresholdsRepo *db.ThresholdsRepo,
	fieldNames map[string]string,
	logger *slog.Logger,
) {
	fields, err := fieldsRepo.List(ctx)
	if err != nil {
		logger.Error("list fields failed", "error", err)
		return
	}

	activeRules, err := rules.LoadRules(ctx, thresholdsRepo)
	if err != nil {
		logger.Warn("load rules failed, using defaults", "error", err)
		activeRules = rules.DefaultRules()
	}

	// Build metric → critical low threshold map from the loaded rules.
	critLow := map[string]float64{}
	for _, r := range activeRules {
		tr, ok := r.(rules.ThresholdRule)
		if !ok {
			continue
		}
		critLow[tr.Metric] = tr.LowCritical
	}

	now := time.Now()
	from := now.Add(-lookbackWindow)

	for _, field := range fields {
		for _, metric := range predictiveMetrics {
			readings, err := readingsRepo.ListByFieldAndMetric(ctx, field.ID, metric, from, now)
			if err != nil {
				logger.Error("list readings failed", "field", field.ID, "metric", metric, "error", err)
				continue
			}
			if len(readings) < 5 {
				continue // not enough data to fit a meaningful regression
			}

			// Build xs (seconds from epoch) and ys (values) for regression.
			xs := make([]float64, len(readings))
			ys := make([]float64, len(readings))
			for i, rd := range readings {
				xs[i] = float64(rd.Timestamp.Unix())
				ys[i] = rd.Value
			}

			slope := stats.LinearSlope(xs, ys)
			currentVal := readings[0].Value // most recent first (DESC order)
			projected := currentVal + slope*forecastHorizon.Seconds()

			threshold, ok := critLow[metric]
			if !ok {
				continue
			}

			// Only alert if projected is below critical and current is not already critical.
			if projected >= threshold || currentVal <= threshold {
				continue
			}

			hoursUntil := estimateHoursUntilBreach(currentVal, threshold, slope)

			alert := models.Alert{
				FieldID:     field.ID,
				SensorType:  models.SensorTypeSoil,
				Metric:      metric,
				Value:       currentVal,
				Threshold:   threshold,
				Severity:    models.SeverityWarning,
				Message:     fmt.Sprintf("[PREDICTED in ~%.1fh] %s trending towards critical threshold (%.2f → projected %.2f)", hoursUntil, metric, currentVal, projected),
				Status:      models.AlertStatusActive,
				TriggeredAt: now,
			}

			id, err := alertsRepo.Insert(ctx, alert)
			if err != nil {
				logger.Error("predictive alert insert failed", "field", field.ID, "metric", metric, "error", err)
				continue
			}
			alert.ID = id

			fieldName := fieldNames[field.ID]
			if fieldName == "" {
				fieldName = "Unknown Field"
			}
			notif := notifications.Compose(alert, fieldName)
			notif.AlertID = &id
			// Prefix the notification body to make it clearly predictive.
			notif.Title = fmt.Sprintf("📈 %s", notif.Title)
			notif.Body = fmt.Sprintf("Trending towards critical — act now to prevent damage.\n\n%s", notif.Body)

			if _, err := notifsRepo.Insert(ctx, notif); err != nil {
				logger.Error("predictive notification insert failed", "alert_id", id, "error", err)
			}

			logger.Info("predictive alert triggered",
				"field", field.Name,
				"metric", metric,
				"current", currentVal,
				"projected", projected,
				"hours_until_breach", hoursUntil,
			)
		}
	}
}

// estimateHoursUntilBreach returns hours until value reaches threshold at given slope.
// Returns forecastHorizon.Hours() if slope is not negative enough.
func estimateHoursUntilBreach(current, threshold, slope float64) float64 {
	if slope >= 0 {
		return forecastHorizon.Hours()
	}
	seconds := (threshold - current) / slope
	if seconds < 0 {
		return forecastHorizon.Hours()
	}
	return seconds / 3600
}

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

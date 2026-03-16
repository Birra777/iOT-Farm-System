package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strings"

	"github.com/agristream/agristream/internal/advisor"
	"github.com/agristream/agristream/internal/api"
	"github.com/agristream/agristream/internal/config"
	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/kafka"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "api")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	adv := advisor.New(cfg.AnthropicAPIKey, cfg.AnthropicModel)
	if adv.Enabled() {
		logger.Info("AI advisor enabled", "model", cfg.AnthropicModel)
	} else {
		logger.Info("AI advisor disabled — set ANTHROPIC_API_KEY to enable")
	}

	handlers := api.NewHandlers(
		db.NewFieldsRepo(pool),
		db.NewReadingsRepo(pool),
		db.NewAlertsRepo(pool),
		db.NewNotificationsRepo(pool),
		db.NewThresholdsRepo(pool),
		adv,
		logger,
	)

	// Start SSE Kafka consumer — broadcasts aggregated readings to browser clients.
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	sseConsumer := kafka.NewConsumer(brokers, cfg.KafkaTopicAggregated, "dashboard-sse-group", logger)
	defer sseConsumer.Close()
	go handlers.RunSSEConsumer(ctx, sseConsumer)
	logger.Info("SSE consumer started", "topic", cfg.KafkaTopicAggregated)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.APIPort),
		Handler:      api.NewRouter(handlers, logger),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background.
	go func() {
		logger.Info("API server listening", "port", cfg.APIPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case <-ctx.Done():
	}

	// Graceful shutdown — give in-flight requests 15s to complete.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	} else {
		logger.Info("API server stopped cleanly")
	}
}

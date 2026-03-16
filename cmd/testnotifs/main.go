// cmd/testnotifs — manual smoke-test for the notification system.
// Inserts synthetic alerts + notifications directly into the DB,
// then prints them back, so you can verify without restarting services.
//
// Usage:
//   go run ./cmd/testnotifs/...
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/models"
	"github.com/agristream/agristream/internal/notifications"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://agristream:agristream@localhost:5433/agristream?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("connect:", err)
	}
	defer pool.Close()

	fieldsRepo := db.NewFieldsRepo(pool)
	alertsRepo := db.NewAlertsRepo(pool)
	notifsRepo := db.NewNotificationsRepo(pool)

	// Load all fields so we can pick real UUIDs.
	fields, err := fieldsRepo.List(ctx)
	if err != nil {
		log.Fatal("list fields:", err)
	}
	if len(fields) == 0 {
		log.Fatal("no fields found in DB — run migrations first")
	}

	fmt.Printf("Found %d fields:\n", len(fields))
	for _, f := range fields {
		fmt.Printf("  [%s] %s (%s)\n", f.ZoneCode, f.Name, f.CropType)
	}
	fmt.Println()

	// Build a set of synthetic test alerts covering every notification template.
	testCases := []struct {
		field    models.Field
		metric   string
		value    float64
		unit     string
		severity models.Severity
	}{
		{fields[0], models.MetricSoilMoisture, 14.5, "%", models.SeverityCritical},
		{fields[1], models.MetricSoilMoisture, 26.2, "%", models.SeverityWarning},
		{fields[0], models.MetricSoilNitrogen, 61.0, "mg/kg", models.SeverityCritical},
		{fields[2], models.MetricSoilNitrogen, 88.0, "mg/kg", models.SeverityWarning},
		{fields[1], models.MetricSoilPH, 4.8, "", models.SeverityWarning},
		{fields[3], models.MetricSoilPH, 8.1, "", models.SeverityWarning},
		{fields[2], models.MetricWeatherTemperature, 39.6, "°C", models.SeverityCritical},
		{fields[0], models.MetricWeatherTemperature, 36.1, "°C", models.SeverityWarning},
		{fields[3], models.MetricWeatherHumidity, 92.0, "%", models.SeverityCritical},
		{fields[1], models.MetricWeatherHumidity, 87.5, "%", models.SeverityWarning},
	}

	fmt.Println("=== Inserting test alerts + notifications ===")
	fmt.Println()

	for _, tc := range testCases {
		alert := models.Alert{
			FieldID:     tc.field.ID,
			SensorType:  models.SensorTypeSoil,
			Metric:      tc.metric,
			Value:       tc.value,
			Threshold:   0,
			Severity:    tc.severity,
			Message:     fmt.Sprintf("test: %s=%.2f", tc.metric, tc.value),
			Status:      models.AlertStatusActive,
			TriggeredAt: time.Now().UTC(),
		}

		alertID, err := alertsRepo.Insert(ctx, alert)
		if err != nil {
			log.Printf("  SKIP %s/%s: %v\n", tc.field.ZoneCode, tc.metric, err)
			continue
		}
		alert.ID = alertID

		notif := notifications.Compose(alert, tc.field.Name)
		notif.AlertID = &alertID
		notifID, err := notifsRepo.Insert(ctx, notif)
		if err != nil {
			log.Printf("  SKIP notif for alert %d: %v\n", alertID, err)
			continue
		}

		fmt.Printf("  [%s] alert=%d  notif=%d\n", tc.field.ZoneCode, alertID, notifID)
		fmt.Printf("  Title: %s\n", notif.Title)
		fmt.Printf("  Body:  %s\n", notif.Body)
		fmt.Println()
	}

	// Read them back via the repo.
	fmt.Println("=== Latest 10 notifications from DB ===")
	fmt.Println()
	list, err := notifsRepo.List(ctx, false, 10)
	if err != nil {
		log.Fatal("list notifications:", err)
	}
	unread := 0
	for _, n := range list {
		mark := "  "
		if !n.IsRead {
			mark = "• "
			unread++
		}
		fmt.Printf("%s[%s] %s\n", mark, n.Severity, n.Title)
		fmt.Printf("    %s\n", n.Body)
		fmt.Printf("    id=%d  alert_id=%v  created=%s\n\n",
			n.ID, n.AlertID, n.CreatedAt.Format("15:04:05"))
	}
	fmt.Printf("%d unread out of %d shown\n", unread, len(list))
}

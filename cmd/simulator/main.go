package main

import (
	"context"
	"log/slog"
	"math/rand"
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
)

// sensorDef describes a single simulated sensor.
type sensorDef struct {
	sensorID   string
	fieldID    string // matches zone_code for seeded fields; resolved at runtime
	fieldName  string
	sensorType models.SensorType
	metric     string
	unit       string
	baseline   float64 // realistic centre value for this field/metric
	noise      float64 // ± standard deviation around baseline
	spikeMin   float64 // anomalous low spike value (used ~5% of the time)
	spikeMax   float64 // anomalous high spike value
}


// buildSensors returns all 25 sensor definitions.
// Baselines are tuned for Kavango East, Namibia conditions.
func buildSensors() []sensorDef {
	var sensors []sensorDef

	type fieldProfile struct {
		zoneCode string
		name     string
		// soil
		moisture float64
		ph       float64
		nitrogen float64
		// weather
		temp     float64
		humidity float64
		rainfall float64
		wind     float64
	}

	profiles := []fieldProfile{
		// North Block — irrigated maize, moderate conditions
		{"NB", "North Block", 58, 6.4, 145, 28, 65, 2.1, 12},
		// River Basin — near water, high humidity, good moisture
		{"RB", "River Basin", 72, 6.8, 160, 26, 80, 4.5, 8},
		// Dry Ridge — drought-prone, low moisture, low nitrogen
		{"DR", "Dry Ridge", 22, 5.8, 88, 33, 45, 0.3, 18},
		// South Paddock — groundnuts, slightly acidic
		{"SP", "South Paddock", 48, 5.9, 130, 30, 60, 1.8, 14},
	}

	soilCount := map[string]int{"NB": 0, "RB": 0, "DR": 0, "SP": 0}
	weatherCount := map[string]int{"NB": 0, "RB": 0, "DR": 0, "SP": 0}

	for _, p := range profiles {
		// Each field has multiple soil sensors — North Block 4, River Basin 3, Dry Ridge 2, South Paddock 3
		soilCounts := map[string]int{"NB": 4, "RB": 3, "DR": 2, "SP": 3}
		for i := 0; i < soilCounts[p.zoneCode]; i++ {
			soilCount[p.zoneCode]++
			n := soilCount[p.zoneCode]

			// soil.moisture
			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-SOIL-" + itoa(n) + "-MOIST",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeSoil,
				metric:     models.MetricSoilMoisture,
				unit:       "%",
				baseline:   p.moisture,
				noise:      4.0,
				spikeMin:   15.0,
				spikeMax:   95.0,
			})
			// soil.ph
			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-SOIL-" + itoa(n) + "-PH",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeSoil,
				metric:     models.MetricSoilPH,
				unit:       "pH",
				baseline:   p.ph,
				noise:      0.3,
				spikeMin:   4.5,
				spikeMax:   8.5,
			})
			// soil.nitrogen
			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-SOIL-" + itoa(n) + "-N",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeSoil,
				metric:     models.MetricSoilNitrogen,
				unit:       "mg/kg",
				baseline:   p.nitrogen,
				noise:      12.0,
				spikeMin:   60.0,
				spikeMax:   220.0,
			})
		}

		// Weather sensors — North Block 4, River Basin 3, Dry Ridge 2, South Paddock 4
		wxCounts := map[string]int{"NB": 4, "RB": 3, "DR": 2, "SP": 4}
		for i := 0; i < wxCounts[p.zoneCode]; i++ {
			weatherCount[p.zoneCode]++
			n := weatherCount[p.zoneCode]

			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-WX-" + itoa(n) + "-TEMP",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeWeather,
				metric:     models.MetricWeatherTemperature,
				unit:       "°C",
				baseline:   p.temp,
				noise:      2.5,
				spikeMin:   18.0,
				spikeMax:   42.0,
			})
			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-WX-" + itoa(n) + "-HUM",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeWeather,
				metric:     models.MetricWeatherHumidity,
				unit:       "%",
				baseline:   p.humidity,
				noise:      6.0,
				spikeMin:   20.0,
				spikeMax:   98.0,
			})
			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-WX-" + itoa(n) + "-RAIN",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeWeather,
				metric:     models.MetricWeatherRainfall,
				unit:       "mm",
				baseline:   p.rainfall,
				noise:      1.5,
				spikeMin:   0.0,
				spikeMax:   30.0,
			})
			sensors = append(sensors, sensorDef{
				sensorID:   p.zoneCode + "-WX-" + itoa(n) + "-WIND",
				fieldID:    p.zoneCode,
				fieldName:  p.name,
				sensorType: models.SensorTypeWeather,
				metric:     models.MetricWeatherWindSpeed,
				unit:       "km/h",
				baseline:   p.wind,
				noise:      3.0,
				spikeMin:   0.0,
				spikeMax:   60.0,
			})
		}
	}

	return sensors
}

// generateValue produces a single reading for a sensor.
// 5% chance of an anomalous spike to exercise the anomaly detector.
func generateValue(s sensorDef, rng *rand.Rand) float64 {
	if rng.Float64() < 0.05 {
		// spike: pick a value outside the normal range
		if rng.Float64() < 0.5 {
			return s.spikeMin + rng.Float64()*5
		}
		return s.spikeMax - rng.Float64()*5
	}
	// gaussian noise around baseline
	v := s.baseline + rng.NormFloat64()*s.noise
	if v < 0 {
		v = 0
	}
	return round2(v)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "simulator")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err)
		os.Exit(1)
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	producer := kafka.NewProducer(brokers, logger)
	defer func() {
		if err := producer.Close(); err != nil {
			logger.Error("producer close error", "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGTERM / SIGINT.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		logger.Info("shutdown signal received", "signal", sig.String())
		cancel()
	}()

	// Resolve zone codes (NB, RB, DR, SP) to real UUIDs from the fields table.
	pool, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	zoneToUUID, err := db.NewFieldsRepo(pool).ZoneCodeToUUID(ctx)
	if err != nil {
		logger.Error("field lookup failed", "error", err)
		os.Exit(1)
	}

	sensors := buildSensors()

	// Replace zone codes with real UUIDs so DB foreign keys resolve correctly.
	for i := range sensors {
		if uuid, ok := zoneToUUID[sensors[i].fieldID]; ok {
			sensors[i].fieldID = uuid
		} else {
			logger.Warn("no UUID found for zone code", "zone", sensors[i].fieldID)
		}
	}

	interval := time.Duration(cfg.SimulatorPublishIntervalSeconds) * time.Second

	logger.Info("sensor simulator starting",
		"sensors", len(sensors),
		"interval_seconds", cfg.SimulatorPublishIntervalSeconds,
		"fields_resolved", len(zoneToUUID),
	)

	var wg sync.WaitGroup
	for _, s := range sensors {
		wg.Add(1)
		go runSensor(ctx, &wg, s, producer, cfg, interval, logger)
	}

	wg.Wait()
	logger.Info("simulator stopped cleanly")
}

// runSensor is the per-sensor goroutine. It ticks every interval and publishes a reading.
func runSensor(
	ctx context.Context,
	wg *sync.WaitGroup,
	s sensorDef,
	producer *kafka.Producer,
	cfg *config.Config,
	interval time.Duration,
	logger *slog.Logger,
) {
	defer wg.Done()

	// Each sensor gets its own RNG seeded with a hash of the sensor ID
	// so readings are reproducible per sensor but independent across sensors.
	rng := rand.New(rand.NewSource(int64(hashString(s.sensorID))))

	topic := cfg.KafkaTopicSoilRaw
	if s.sensorType == models.SensorTypeWeather {
		topic = cfg.KafkaTopicWeatherRaw
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Publish one reading immediately on startup, then tick.
	publish(ctx, s, producer, topic, rng, logger)

	for {
		select {
		case <-ticker.C:
			publish(ctx, s, producer, topic, rng, logger)
		case <-ctx.Done():
			return
		}
	}
}

func publish(
	ctx context.Context,
	s sensorDef,
	producer *kafka.Producer,
	topic string,
	rng *rand.Rand,
	logger *slog.Logger,
) {
	reading := models.SensorReading{
		SensorID:   s.sensorID,
		FieldID:    s.fieldID,
		SensorType: s.sensorType,
		Metric:     s.metric,
		Value:      generateValue(s, rng),
		Unit:       s.unit,
		Timestamp:  time.Now().UTC(),
	}

	if err := producer.Publish(ctx, topic, s.fieldID, reading); err != nil {
		if ctx.Err() != nil {
			return // shutting down — not an error
		}
		logger.Error("publish failed",
			"sensor_id", s.sensorID,
			"metric", s.metric,
			"error", err,
		)
		return
	}

	logger.Info("reading published",
		"sensor_id", s.sensorID,
		"field", s.fieldID,
		"metric", s.metric,
		"value", reading.Value,
		"unit", s.unit,
	)
}

// hashString returns a simple deterministic hash for seeding per-sensor RNGs.
func hashString(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func itoa(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

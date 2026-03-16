package models

import (
	"time"
)

// SensorType identifies the category of a sensor.
type SensorType string

const (
	SensorTypeSoil    SensorType = "soil"
	SensorTypeWeather SensorType = "weather"
)

// Metric names for soil sensors.
const (
	MetricSoilMoisture  = "soil.moisture"
	MetricSoilPH        = "soil.ph"
	MetricSoilNitrogen  = "soil.nitrogen"
)

// Metric names for weather sensors.
const (
	MetricWeatherTemperature = "weather.temperature"
	MetricWeatherHumidity    = "weather.humidity"
	MetricWeatherRainfall    = "weather.rainfall"
	MetricWeatherWindSpeed   = "weather.wind_speed"
)

// SensorReading is a single measurement published by a sensor.
type SensorReading struct {
	SensorID   string     `json:"sensor_id"`
	FieldID    string     `json:"field_id"`
	SensorType SensorType `json:"sensor_type"`
	Metric     string     `json:"metric"`
	Value      float64    `json:"value"`
	Unit       string     `json:"unit"`
	Timestamp  time.Time  `json:"timestamp"`
}

// AggregatedReading is a 1-minute windowed summary published by the stream processor.
type AggregatedReading struct {
	FieldID    string     `json:"field_id"`
	SensorType SensorType `json:"sensor_type"`
	Metric     string     `json:"metric"`
	Avg        float64    `json:"avg"`
	Min        float64    `json:"min"`
	Max        float64    `json:"max"`
	Count      int        `json:"count"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
}

// Field represents a farm zone.
type Field struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CropType  string    `json:"crop_type"`
	Hectares  float64   `json:"hectares"`
	ZoneCode  string    `json:"zone_code"`
	CreatedAt time.Time `json:"created_at"`
}

// PipelineStats holds aggregate metrics about the data pipeline, served by GET /api/stats.
type PipelineStats struct {
	TotalReadingsArchived int64 `json:"total_readings_archived"`
	ReadingsLastMinute    int64 `json:"readings_last_minute"`
	ReadingsLastHour      int64 `json:"readings_last_hour"`
	ActiveAlerts          int64 `json:"active_alerts"`
}

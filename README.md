# AgriStream - IoT Farm Monitoring System

A real-time farm monitoring system built for Kavango East, Namibia. Sensors on four fields publish soil and weather readings every few seconds. The data flows through a streaming pipeline, gets stored in a database, triggers alerts when thresholds are breached, and is displayed on a live 3D dashboard.

---

## What It Does

- Collects soil moisture, soil pH, soil nitrogen, air temperature, humidity, rainfall, and wind speed from simulated IoT sensors across four fields.
- Streams all readings through Apache Kafka in real time.
- Detects anomalies (too dry, too hot, low nitrogen, etc.) and fires alerts.
- Sends email notifications when alerts trigger (optional, requires SMTP config).
- Projects trends 3 hours ahead and fires early warning alerts before damage occurs.
- Displays everything on a React dashboard with a live 3D farm view.
- Pushes live updates to the browser using Server-Sent Events instead of constant polling.
- Lets operators adjust alert thresholds through the dashboard UI without touching code.
- Provides plain-English farm advice on demand using the Claude AI API.

---

## Architecture

```
Sensors (simulator)
       |
       v
   Apache Kafka
  /      |      \
  |   Stream    |
  |  Processor  |
  |      |      |
  |   Kafka     |
  |  (aggregated|
  |   topic)    |
  |             |
  v             v
PostgreSQL    API Server -----> React Dashboard
  ^               |
  |           SSE stream
Anomaly        (live push)
Detector
  |
Predictor
(trend alerts)
```

Six services run as separate processes:

| Service | What it does |
|---|---|
| `cmd/simulator` | Generates fake sensor readings and publishes them to Kafka |
| `cmd/processor` | Consumes raw readings, aggregates them into 1-minute windows, writes to the database |
| `cmd/anomaly` | Reads from Kafka, checks readings against thresholds, creates alerts |
| `cmd/predictor` | Runs every 10 minutes, fits a trend line to recent data, fires early warnings |
| `cmd/api` | REST + SSE API server — serves the dashboard and handles all reads/writes |
| `cmd/migrate` | Applies database schema migrations |

---

## Requirements

- Go 1.22 or later
- Docker and Docker Compose (for Kafka and PostgreSQL)
- Node.js 18 or later (for the dashboard)
- Make (optional, but makes running commands easier)

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/Birra777/iOT-Farm-System.git
cd iOT-Farm-System
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

The defaults in `.env.example` work out of the box with the Docker Compose setup. You only need to change values if you want email alerts or AI advice (see Optional Features below).

### 3. Start the infrastructure

```bash
make up
```

This starts PostgreSQL (on port 5433) and Kafka (on port 9092) using Docker.

### 4. Run database migrations

```bash
make migrate
```

Creates all tables including `sensor_readings`, `alerts`, `notifications`, and `global_thresholds`.

### 5. Install dashboard dependencies

```bash
cd dashboard && npm install && cd ..
```

### 6. Start all services

Open a separate terminal for each service, or use a terminal multiplexer:

```bash
make run-simulator    # terminal 1
make run-processor    # terminal 2
make run-anomaly      # terminal 3
make run-api          # terminal 4
make run-dashboard    # terminal 5
```

Open http://localhost:5173 in your browser.

---

## Optional Features

### Email Alerts

Set these values in `.env`:

```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
ALERT_EMAIL_TO=recipient@example.com
```

For Gmail, use an App Password (not your main account password). Leave `SMTP_HOST` empty to disable.

### AI Farm Advisor

The dashboard has a "Get AI Farm Advice" button that generates plain-English recommendations based on current field conditions. It uses the Claude API.

Set this in `.env`:

```
ANTHROPIC_API_KEY=sk-ant-...
```

Leave it empty to disable the feature. The API key is never sent to the browser.

### Predictive Alerts

Run the predictor service alongside the others:

```bash
make run-predictor
```

It checks moisture and nitrogen trends every 10 minutes. If a field is trending toward a critical level and will breach it within 3 hours, it creates a warning alert prefixed with `[PREDICTED]`.

---

## Dashboard Features

- **Live sensor readings** - Updates in real time via Server-Sent Events
- **30-minute history charts** - Per field, per metric
- **3D farm view** - Interactive three-dimensional representation of the four fields
- **Active alerts panel** - Shows current threshold breaches with severity
- **Notification bell** - Tracks unread farm notifications
- **Alert thresholds settings** (gear icon in header) - Adjust warning and critical levels for any metric without restarting services. Changes take effect within 5 minutes.
- **AI farm advisor** - Click the button in the right panel to get contextual advice based on current conditions
- **Pipeline stats** - Shows total readings, readings in the last hour, and active alert count
- **Event log** - Scrollable history of all alerts

---

## Project Structure

```
.
├── cmd/
│   ├── api/          - REST + SSE API server
│   ├── anomaly/      - Real-time anomaly detector
│   ├── migrate/      - Database migration runner
│   ├── predictor/    - Trend-based predictive alert service
│   ├── processor/    - Kafka stream aggregator
│   └── simulator/    - Sensor data simulator
├── dashboard/        - React frontend (Vite + Three.js)
│   └── src/
│       ├── components/
│       └── hooks/
├── internal/
│   ├── advisor/      - Claude API client for farm advice
│   ├── api/          - HTTP handlers, SSE hub, routing
│   ├── config/       - Environment variable loading
│   ├── db/           - PostgreSQL repository layer
│   ├── email/        - SMTP email sender
│   ├── kafka/        - Kafka producer and consumer wrappers
│   ├── models/       - Shared data types
│   ├── notifications/- Notification message composer
│   ├── rules/        - Alert threshold rules and DB loader
│   ├── stats/        - Linear regression for trend analysis
│   └── window/       - Time-window aggregation logic
├── migrations/       - SQL migration files (001 through 005)
├── docker-compose.yml
├── Makefile
└── .env.example
```

---

## API Reference

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/api/fields` | List all fields |
| GET | `/api/fields/{id}/summary` | Latest reading per metric for a field |
| GET | `/api/fields/{id}/history` | Time-series readings (query params: metric, from, to) |
| GET | `/api/alerts` | List alerts (query param: status=active or resolved) |
| POST | `/api/alerts/{id}/resolve` | Mark an alert as resolved |
| GET | `/api/notifications` | List notifications (query params: unread, limit) |
| POST | `/api/notifications/{id}/read` | Mark a notification as read |
| POST | `/api/notifications/read-all` | Mark all notifications as read |
| GET | `/api/stats` | Pipeline statistics |
| GET | `/api/thresholds` | List all global alert thresholds |
| PUT | `/api/thresholds/{metric}` | Update thresholds for a metric |
| POST | `/api/advisor` | Get AI-generated farm advice |
| GET | `/api/events` | Server-Sent Events stream for live updates |

---

## Running Tests

```bash
make test
```

Tests cover the data models, alert rule evaluation, and time-window aggregation logic.

---

## Technology Stack

- **Go** - All backend services
- **Apache Kafka** - Message streaming between services
- **PostgreSQL** - Persistent storage
- **React + Vite** - Dashboard frontend
- **Three.js / React Three Fiber** - 3D farm visualisation
- **Recharts** - Sensor history charts
- **Docker Compose** - Local infrastructure

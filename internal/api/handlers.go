package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/agristream/agristream/internal/advisor"
	"github.com/agristream/agristream/internal/db"
	"github.com/agristream/agristream/internal/models"
)

// Handlers holds all HTTP handler dependencies.
type Handlers struct {
	fields        *db.FieldsRepo
	readings      *db.ReadingsRepo
	alerts        *db.AlertsRepo
	notifications *db.NotificationsRepo
	thresholds    *db.ThresholdsRepo
	advisor       *advisor.Advisor
	sseHub        *SSEHub
	logger        *slog.Logger

	// advisor rate-limit: one call per 30s
	advisorMu     sync.Mutex
	advisorLastAt time.Time
	advisorCached string
}

// NewHandlers constructs Handlers.
func NewHandlers(
	fields *db.FieldsRepo,
	readings *db.ReadingsRepo,
	alerts *db.AlertsRepo,
	notifications *db.NotificationsRepo,
	thresholds *db.ThresholdsRepo,
	adv *advisor.Advisor,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		fields:        fields,
		readings:      readings,
		alerts:        alerts,
		notifications: notifications,
		thresholds:    thresholds,
		advisor:       adv,
		sseHub:        newSSEHub(logger),
		logger:        logger,
	}
}

// GetFields handles GET /api/fields
func (h *Handlers) GetFields(w http.ResponseWriter, r *http.Request) {
	fields, err := h.fields.List(r.Context())
	if err != nil {
		h.internalError(w, "list fields", err)
		return
	}
	writeJSON(w, http.StatusOK, fields)
}

// GetFieldSummary handles GET /api/fields/:id/summary
// Returns the latest reading per metric for the given field.
func (h *Handlers) GetFieldSummary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	field, err := h.fields.Get(r.Context(), id)
	if err != nil {
		h.internalError(w, "get field", err)
		return
	}

	readings, err := h.readings.LatestPerMetric(r.Context(), id)
	if err != nil {
		h.internalError(w, "latest readings", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"field":    field,
		"readings": readings,
	})
}

// GetFieldHistory handles GET /api/fields/:id/history
// Query params: metric (required), from (RFC3339, optional), to (RFC3339, optional)
func (h *Handlers) GetFieldHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "metric query param required"})
		return
	}

	to := time.Now()
	from := to.Add(-30 * time.Minute)

	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from timestamp (use RFC3339)"})
			return
		}
		from = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to timestamp (use RFC3339)"})
			return
		}
		to = t
	}

	readings, err := h.readings.ListByFieldAndMetric(r.Context(), id, metric, from, to)
	if err != nil {
		h.internalError(w, "list history", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"field_id": id,
		"metric":   metric,
		"from":     from,
		"to":       to,
		"readings": readings,
	})
}

// GetAlerts handles GET /api/alerts?status=active|resolved
func (h *Handlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	alerts, err := h.alerts.List(r.Context(), status)
	if err != nil {
		h.internalError(w, "list alerts", err)
		return
	}
	if alerts == nil {
		alerts = []models.Alert{}
	}
	writeJSON(w, http.StatusOK, alerts)
}

// ResolveAlert handles POST /api/alerts/:id/resolve
func (h *Handlers) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid alert id"})
		return
	}

	if err := h.alerts.Resolve(r.Context(), id); err != nil {
		h.internalError(w, "resolve alert", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "resolved"})
}

// GetStats handles GET /api/stats
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.fields.Stats(r.Context())
	if err != nil {
		h.internalError(w, "get stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// Health handles GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "api"})
}

// internalError logs and returns a 500.
func (h *Handlers) internalError(w http.ResponseWriter, op string, err error) {
	h.logger.Error(op+" failed", "error", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

// GetNotifications handles GET /api/notifications?unread=true&limit=50
func (h *Handlers) GetNotifications(w http.ResponseWriter, r *http.Request) {
	onlyUnread := r.URL.Query().Get("unread") == "true"
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	notifs, err := h.notifications.List(r.Context(), onlyUnread, limit)
	if err != nil {
		h.internalError(w, "list notifications", err)
		return
	}
	if notifs == nil {
		notifs = []models.Notification{}
	}

	count, err := h.notifications.UnreadCount(r.Context())
	if err != nil {
		h.internalError(w, "count unread", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"notifications": notifs,
		"unread_count":  count,
	})
}

// MarkNotificationRead handles POST /api/notifications/:id/read
func (h *Handlers) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid notification id"})
		return
	}
	if err := h.notifications.MarkRead(r.Context(), id); err != nil {
		h.internalError(w, "mark notification read", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "is_read": true})
}

// MarkAllNotificationsRead handles POST /api/notifications/read-all
func (h *Handlers) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if err := h.notifications.MarkAllRead(r.Context()); err != nil {
		h.internalError(w, "mark all notifications read", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetThresholds handles GET /api/thresholds
func (h *Handlers) GetThresholds(w http.ResponseWriter, r *http.Request) {
	ts, err := h.thresholds.List(r.Context())
	if err != nil {
		h.internalError(w, "list thresholds", err)
		return
	}
	if ts == nil {
		ts = []models.GlobalThreshold{}
	}
	writeJSON(w, http.StatusOK, ts)
}

// UpdateThreshold handles PUT /api/thresholds/{metric}
func (h *Handlers) UpdateThreshold(w http.ResponseWriter, r *http.Request) {
	metric := chi.URLParam(r, "metric")

	var body models.GlobalThreshold
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	body.Metric = metric

	updated, err := h.thresholds.Upsert(r.Context(), body)
	if err != nil {
		h.internalError(w, "upsert threshold", err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// GetAIAdvice handles POST /api/advisor
// Rate-limited to one Claude API call per 30s; returns cached advice in between.
func (h *Handlers) GetAIAdvice(w http.ResponseWriter, r *http.Request) {
	if !h.advisor.Enabled() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "AI advisor not configured — set ANTHROPIC_API_KEY",
		})
		return
	}

	h.advisorMu.Lock()
	if h.advisorCached != "" && time.Since(h.advisorLastAt) < 30*time.Second {
		cached := h.advisorCached
		h.advisorMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]string{"advice": cached})
		return
	}
	h.advisorMu.Unlock()

	// Fetch live context.
	fields, err := h.fields.List(r.Context())
	if err != nil {
		h.internalError(w, "advisor: list fields", err)
		return
	}
	activeAlerts, err := h.alerts.List(r.Context(), "active")
	if err != nil {
		h.internalError(w, "advisor: list alerts", err)
		return
	}

	summaries := make(map[string][]models.SensorReading, len(fields))
	for _, f := range fields {
		readings, err := h.readings.LatestPerMetric(r.Context(), f.ID)
		if err == nil {
			summaries[f.ID] = readings
		}
	}

	advice, err := h.advisor.Advise(r.Context(), fields, summaries, activeAlerts)
	if err != nil {
		h.logger.Error("advisor call failed", "error", err)
		h.internalError(w, "advisor", err)
		return
	}

	h.advisorMu.Lock()
	h.advisorCached = advice
	h.advisorLastAt = time.Now()
	h.advisorMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"advice": advice})
}

// writeJSON serialises v to JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Headers already sent — nothing more we can do.
		return
	}
}

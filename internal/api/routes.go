package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// NewRouter wires all routes and middleware onto a chi router.
func NewRouter(h *Handlers, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(CORS)
	r.Use(Logger(logger))
	r.Use(chimw.Recoverer)

	r.Get("/health", h.Health)

	r.Route("/api", func(r chi.Router) {
		r.Get("/fields", h.GetFields)
		r.Get("/fields/{id}/summary", h.GetFieldSummary)
		r.Get("/fields/{id}/history", h.GetFieldHistory)

		r.Get("/alerts", h.GetAlerts)
		r.Post("/alerts/{id}/resolve", h.ResolveAlert)

		r.Get("/notifications", h.GetNotifications)
		r.Post("/notifications/{id}/read", h.MarkNotificationRead)
		r.Post("/notifications/read-all", h.MarkAllNotificationsRead)

		r.Get("/stats", h.GetStats)

		r.Get("/thresholds", h.GetThresholds)
		r.Put("/thresholds/{metric}", h.UpdateThreshold)

		r.Post("/advisor", h.GetAIAdvice)

		r.Get("/events", h.ServeSSE)
	})

	return r
}

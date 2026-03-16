package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/agristream/agristream/internal/kafka"
	"github.com/agristream/agristream/internal/models"
)

// SSEHub fans out server-sent events to all connected browser clients.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
	logger  *slog.Logger
}

func newSSEHub(logger *slog.Logger) *SSEHub {
	return &SSEHub{
		clients: make(map[chan string]struct{}),
		logger:  logger,
	}
}

func (h *SSEHub) subscribe() chan string {
	ch := make(chan string, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *SSEHub) unsubscribe(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *SSEHub) broadcast(event, data string) {
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
			// Slow client — drop rather than block.
		}
	}
}

// ServeSSE handles GET /api/events — streams real-time events to the browser.
func (h *Handlers) ServeSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := h.sseHub.subscribe()
	defer h.sseHub.unsubscribe(ch)

	// Send an initial heartbeat so the browser knows it's connected.
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-ticker.C:
			// Keepalive comment to prevent proxy/browser timeouts.
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// RunSSEConsumer subscribes to the aggregated Kafka topic and broadcasts
// each reading as an SSE event. Runs until ctx is cancelled.
func (h *Handlers) RunSSEConsumer(ctx context.Context, consumer *kafka.Consumer) {
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			h.logger.Error("SSE consumer read failed", "error", err)
			continue
		}

		var reading models.AggregatedReading
		if err := json.Unmarshal(msg.Value, &reading); err != nil {
			_ = consumer.Commit(ctx, msg)
			continue
		}

		data, err := json.Marshal(reading)
		if err == nil {
			h.sseHub.broadcast("reading", string(data))
		}
		_ = consumer.Commit(ctx, msg)
	}
}

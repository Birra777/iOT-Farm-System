package window

import (
	"math"
	"sync"
	"time"
)

// entry is a single timestamped value in the window.
type entry struct {
	value float64
	ts    time.Time
}

// Window is a thread-safe sliding time window for a single metric stream.
// It evicts entries older than the configured duration on every Add call.
type Window struct {
	mu       sync.Mutex
	entries  []entry
	duration time.Duration
}

// New creates a Window with the given sliding duration (e.g. 1 minute).
func New(duration time.Duration) *Window {
	return &Window{
		duration: duration,
		entries:  make([]entry, 0, 64),
	}
}

// Add appends a new value and evicts stale entries outside the window.
func (w *Window) Add(value float64, ts time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entries = append(w.entries, entry{value: value, ts: ts})
	w.evict(ts)
}

// Stats returns aggregate statistics for all entries currently in the window.
// Returns ok=false when the window is empty.
func (w *Window) Stats() (avg, min, max float64, count int, windowStart, windowEnd time.Time, ok bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.entries) == 0 {
		return 0, 0, 0, 0, time.Time{}, time.Time{}, false
	}

	sum := 0.0
	minV := math.MaxFloat64
	maxV := -math.MaxFloat64
	windowStart = w.entries[0].ts
	windowEnd = w.entries[len(w.entries)-1].ts

	for _, e := range w.entries {
		sum += e.value
		if e.value < minV {
			minV = e.value
		}
		if e.value > maxV {
			maxV = e.value
		}
	}

	count = len(w.entries)
	avg = round2(sum / float64(count))
	return avg, round2(minV), round2(maxV), count, windowStart, windowEnd, true
}

// evict removes entries older than (now - duration). Must be called with mu held.
func (w *Window) evict(now time.Time) {
	cutoff := now.Add(-w.duration)
	i := 0
	for i < len(w.entries) && w.entries[i].ts.Before(cutoff) {
		i++
	}
	if i > 0 {
		w.entries = w.entries[i:]
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Registry manages a sliding window per (sensorID, metric) key.
type Registry struct {
	mu       sync.RWMutex
	windows  map[string]*Window
	duration time.Duration
}

// NewRegistry creates a Registry with windows of the given duration.
func NewRegistry(duration time.Duration) *Registry {
	return &Registry{
		windows:  make(map[string]*Window),
		duration: duration,
	}
}

// Add records a value for the given key, creating a window if needed.
func (r *Registry) Add(key string, value float64, ts time.Time) {
	r.mu.RLock()
	w, exists := r.windows[key]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()
		// Double-check after acquiring write lock.
		if _, exists = r.windows[key]; !exists {
			w = New(r.duration)
			r.windows[key] = w
		} else {
			w = r.windows[key]
		}
		r.mu.Unlock()
	}

	w.Add(value, ts)
}

// Snapshot returns stats for every key currently in the registry.
// Keys with empty windows are skipped.
func (r *Registry) Snapshot() map[string]WindowStats {
	r.mu.RLock()
	keys := make([]string, 0, len(r.windows))
	for k := range r.windows {
		keys = append(keys, k)
	}
	r.mu.RUnlock()

	out := make(map[string]WindowStats, len(keys))
	for _, k := range keys {
		r.mu.RLock()
		w := r.windows[k]
		r.mu.RUnlock()

		avg, min, max, count, start, end, ok := w.Stats()
		if !ok {
			continue
		}
		out[k] = WindowStats{
			Avg: avg, Min: min, Max: max,
			Count: count, WindowStart: start, WindowEnd: end,
		}
	}
	return out
}

// WindowStats holds the computed statistics for one window snapshot.
type WindowStats struct {
	Avg         float64
	Min         float64
	Max         float64
	Count       int
	WindowStart time.Time
	WindowEnd   time.Time
}

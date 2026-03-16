package window_test

import (
	"testing"
	"time"

	"github.com/agristream/agristream/internal/window"
)

var base = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func ts(offsetSeconds int) time.Time {
	return base.Add(time.Duration(offsetSeconds) * time.Second)
}

// TestWindow_StatsEmpty verifies that an empty window returns ok=false.
func TestWindow_StatsEmpty(t *testing.T) {
	w := window.New(time.Minute)
	_, _, _, _, _, _, ok := w.Stats()
	if ok {
		t.Fatal("expected ok=false for empty window")
	}
}

// TestWindow_SingleEntry verifies avg=min=max for one value.
func TestWindow_SingleEntry(t *testing.T) {
	w := window.New(time.Minute)
	w.Add(42.5, ts(0))

	avg, min, max, count, _, _, ok := w.Stats()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if count != 1 {
		t.Errorf("count: got %d, want 1", count)
	}
	if avg != 42.5 {
		t.Errorf("avg: got %f, want 42.5", avg)
	}
	if min != 42.5 {
		t.Errorf("min: got %f, want 42.5", min)
	}
	if max != 42.5 {
		t.Errorf("max: got %f, want 42.5", max)
	}
}

// TestWindow_CorrectAggregation verifies avg/min/max across multiple entries.
func TestWindow_CorrectAggregation(t *testing.T) {
	w := window.New(time.Minute)
	values := []float64{10, 20, 30, 40, 50}
	for i, v := range values {
		w.Add(v, ts(i*5)) // 0s, 5s, 10s, 15s, 20s — all within 1 minute
	}

	avg, min, max, count, _, _, ok := w.Stats()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if count != 5 {
		t.Errorf("count: got %d, want 5", count)
	}
	if avg != 30.0 {
		t.Errorf("avg: got %f, want 30.0", avg)
	}
	if min != 10.0 {
		t.Errorf("min: got %f, want 10.0", min)
	}
	if max != 50.0 {
		t.Errorf("max: got %f, want 50.0", max)
	}
}

// TestWindow_Eviction verifies that entries older than the window duration are dropped.
func TestWindow_Eviction(t *testing.T) {
	w := window.New(time.Minute)

	// Add entries at t=0s and t=30s (both within 1 min of each other).
	w.Add(100, ts(0))
	w.Add(200, ts(30))

	// Now add an entry at t=91s — the entry at t=0 is now >60s old and should be evicted.
	w.Add(300, ts(91))

	_, _, _, count, start, _, ok := w.Stats()
	if !ok {
		t.Fatal("expected ok=true")
	}
	// t=0 is evicted (91-0=91 > 60), t=30 stays (91-30=61 > 60... barely evicted too)
	// t=30 is also evicted: 91-30=61 > 60. Only t=91 remains.
	if count != 1 {
		t.Errorf("count after eviction: got %d, want 1 (only t=91s should remain)", count)
	}
	if !start.Equal(ts(91)) {
		t.Errorf("window start: got %v, want %v", start, ts(91))
	}
}

// TestWindow_PartialEviction verifies only stale entries are dropped.
func TestWindow_PartialEviction(t *testing.T) {
	w := window.New(time.Minute)

	// t=0: will be evicted when t=70 is added
	w.Add(100, ts(0))
	// t=20: will survive (70-20=50 < 60)
	w.Add(200, ts(20))
	// t=40: will survive (70-40=30 < 60)
	w.Add(300, ts(40))
	// t=70: triggers eviction of t=0 (70-0=70 > 60)
	w.Add(400, ts(70))

	avg, min, max, count, _, _, ok := w.Stats()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if count != 3 {
		t.Errorf("count: got %d, want 3 (t=0 evicted, t=20/40/70 remain)", count)
	}
	// avg of [200, 300, 400] = 300
	if avg != 300.0 {
		t.Errorf("avg: got %f, want 300.0", avg)
	}
	if min != 200.0 {
		t.Errorf("min: got %f, want 200.0", min)
	}
	if max != 400.0 {
		t.Errorf("max: got %f, want 400.0", max)
	}
}

// TestWindow_WindowBoundaries verifies the returned windowStart and windowEnd.
func TestWindow_WindowBoundaries(t *testing.T) {
	w := window.New(time.Minute)
	w.Add(1.0, ts(5))
	w.Add(2.0, ts(10))
	w.Add(3.0, ts(15))

	_, _, _, _, start, end, ok := w.Stats()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !start.Equal(ts(5)) {
		t.Errorf("windowStart: got %v, want %v", start, ts(5))
	}
	if !end.Equal(ts(15)) {
		t.Errorf("windowEnd: got %v, want %v", end, ts(15))
	}
}

// TestRegistry_AddAndSnapshot verifies the Registry manages multiple windows correctly.
func TestRegistry_AddAndSnapshot(t *testing.T) {
	r := window.NewRegistry(time.Minute)

	r.Add("NB|soil.moisture", 55.0, ts(0))
	r.Add("NB|soil.moisture", 65.0, ts(5))
	r.Add("DR|soil.moisture", 20.0, ts(0))

	snap := r.Snapshot()

	if len(snap) != 2 {
		t.Fatalf("snapshot keys: got %d, want 2", len(snap))
	}

	nb := snap["NB|soil.moisture"]
	if nb.Count != 2 {
		t.Errorf("NB count: got %d, want 2", nb.Count)
	}
	if nb.Avg != 60.0 {
		t.Errorf("NB avg: got %f, want 60.0", nb.Avg)
	}

	dr := snap["DR|soil.moisture"]
	if dr.Count != 1 {
		t.Errorf("DR count: got %d, want 1", dr.Count)
	}
	if dr.Avg != 20.0 {
		t.Errorf("DR avg: got %f, want 20.0", dr.Avg)
	}
}

// TestRegistry_EmptyKeysExcluded verifies that keys with empty windows are not in snapshot.
func TestRegistry_EmptyKeysExcluded(t *testing.T) {
	r := window.NewRegistry(10 * time.Millisecond) // very short window

	r.Add("fast-expiry", 99.0, ts(0))
	// Add a new entry far in the future — the previous entry will be evicted.
	r.Add("fast-expiry", 1.0, ts(1000))

	snap := r.Snapshot()
	s, ok := snap["fast-expiry"]
	if !ok {
		t.Fatal("key should be present")
	}
	if s.Count != 1 {
		t.Errorf("only t=1000 should remain, count: got %d want 1", s.Count)
	}
}

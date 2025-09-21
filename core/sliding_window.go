package suricata

import (
	"fmt"
	"time"
)

func NewSlidingWindow(duration time.Duration) *SlidingWindow {
	return &SlidingWindow{
		Duration: duration,
		Events:   make([]Alert, 0),
	}
}

// Dodanie alertu do okna
func (w *SlidingWindow) Add(alert Alert) {
	// dodajemy nowy alert
	w.Events = append(w.Events, alert)

	// cutoff = granica czasu (np. 10 sekund wstecz)
	cutoff := time.Now().Add(-w.Duration)

	filtered := make([]Alert, 0)
	for _, e := range w.Events {
		if e.ParsedTime.After(cutoff) {
			filtered = append(filtered, e)
		}
	}
	w.Events = filtered
}

// DEBUG
func (w *SlidingWindow) Print() {
	fmt.Println("Sliding window zawiera", len(w.Events), "alertów:")
	for i, e := range w.Events {
		fmt.Printf("  %d. %s -> %s:%d (%s)\n",
			i+1, e.SrcIP, e.DstIP, e.DstPort, e.Alert.Signature)
	}
	fmt.Println("--------------------------------")
}

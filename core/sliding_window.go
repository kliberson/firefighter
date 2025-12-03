package suricata

import (
	"container/list"
	"time"
)

func NewSlidingWindow(duration time.Duration) *SlidingWindow {
	return &SlidingWindow{
		Duration: duration,
		Events:   list.New(),
	}
}

// Add new alert and remove old ones
func (w *SlidingWindow) Add(alert Alert) {
	// add new alert
	w.Events.PushBack(alert)

	// cutoff = current time - window duration
	cutoff := time.Now().Add(-w.Duration)

	// remove old alerts
	for w.Events.Len() > 0 {
		front := w.Events.Front().Value.(Alert)
		if front.ParsedTime.After(cutoff) {
			break
		}
		w.Events.Remove(w.Events.Front())
	}
}

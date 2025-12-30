package suricata

import (
	"container/list"
	"time"
)

type SlidingWindow struct {
	Duration time.Duration
	Events   *list.List
}

func NewSlidingWindow(duration time.Duration) *SlidingWindow {
	return &SlidingWindow{
		Duration: duration,
		Events:   list.New(),
	}
}

func (w *SlidingWindow) Add(alert Alert) {
	w.Events.PushBack(alert)

	if w.Events.Len() > 200 {
		w.Events.Remove(w.Events.Front())
	}
	cutoff := time.Now().Add(-w.Duration)

	for w.Events.Len() > 0 {
		front := w.Events.Front().Value.(Alert)
		if front.ParsedTime.After(cutoff) {
			break
		}
		w.Events.Remove(w.Events.Front())
	}
}

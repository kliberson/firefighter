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

// // DEBUG
// func (w *SlidingWindow) Print() {
// 	fmt.Println("Sliding window zawiera", len(w.Events), "alertów:")
// 	for i, e := range w.Events {
// 		fmt.Printf("  %d. %s -> %s:%d (%s)\n",
// 			i+1, e.SrcIP, e.DstIP, e.DstPort, e.Alert.Signature)
// 	}
// 	fmt.Println("--------------------------------")
// }

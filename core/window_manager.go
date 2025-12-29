package suricata

import (
	"fmt"
	"time"
)

// Manages multiple SlidingWindows by source IP
type WindowManager struct {
	Duration time.Duration
	Windows  map[string]*SlidingWindow
}

// Creating new WindowManager
func NewWindowManager(duration time.Duration) *WindowManager {
	return &WindowManager{
		Duration: duration,
		Windows:  make(map[string]*SlidingWindow),
	}
}

// Adding alert to WindowManager by src IP
func (wm *WindowManager) Add(alert Alert) {
	ip := alert.SrcIP
	if _, exists := wm.Windows[ip]; !exists {
		wm.Windows[ip] = NewSlidingWindow(wm.Duration)
	}
	wm.Windows[ip].Add(alert)
}

func (wm *WindowManager) RemoveIP(ip string) {
	delete(wm.Windows, ip)
	fmt.Printf("🧹 Wyczyszczono sliding window dla %s\n", ip)
}

// Print all windows (for debugging)
func (wm *WindowManager) PrintAll() {
	fmt.Println("Debug Sliding Window: ")
	for ip, window := range wm.Windows {
		fmt.Printf(" IP %s (%d alertów)\n", ip, window.Events.Len())
		i := 1
		for e := window.Events.Front(); e != nil; e = e.Next() {
			a := e.Value.(Alert)
			fmt.Printf("   %d. %s -> %s:%d (%s)\n",
				i, a.SrcIP, a.DstIP, a.DstPort, a.Alert.Signature)
		}
		fmt.Println("--------------------------------")
	}
}

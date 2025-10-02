package suricata

import (
	"fmt"
)

func (wm *WindowManager) AnalyzeAlerts() {
	for ip, window := range wm.Windows {
		if window.Events.Len() >= 5 {
			fmt.Printf("Blocking IP: %s (Alerts number: %d)\n", ip, window.Events.Len())
			if err := BlockIP(ip); err != nil {
				fmt.Println("Blocking IP failed:", err)
			} else {

				// Clear the window after blocking
				window.Events.Init()
			}
		}
	}
}

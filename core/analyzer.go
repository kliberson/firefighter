// analyzer.go
package suricata

import (
	"firefighter/data"
	"fmt"
	"log"
)

type BlockDecision struct {
	IP     string
	Reason string
	Score  int
}

// AnalyzeAlerts analizuje wszystkie okna i zwraca listę IP do zablokowania

func (wm *WindowManager) AnalyzeAlerts(db *data.DbManager) []BlockDecision {
	var decisions []BlockDecision

	for ip, window := range wm.Windows {

		isWhitelisted, err := db.IsWhitelisted(ip)

		if err != nil {
			log.Printf("Warning: Failed to check whitelist for %s: %v", ip, err)
		}

		if isWhitelisted {
			log.Printf("⚪ IP %s is whitelisted, skipping analysis", ip)
			continue
		}

		// W przyszłości zamień na: decision := wm.advancedAnalysis(ip, window)
		if window.Events.Len() >= 5 {

			fmt.Printf("⚠️  IP %s triggered blocking threshold (Alerts: %d)\n", ip, window.Events.Len())
			decisions = append(decisions, BlockDecision{
				IP:     ip,
				Reason: fmt.Sprintf("Alert threshold exceeded: %d alerts", window.Events.Len()),
				Score:  window.Events.Len(), // for future analysis
			})

			// Wyczyść okno po decyzji
			window.Events.Init()
		}
	}

	return decisions
}

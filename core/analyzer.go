// analyzer.go
package suricata

import (
	"fmt"
)

// BlockDecision zawiera decyzję o zablokowaniu IP
type BlockDecision struct {
	IP     string
	Reason string
	Score  int // Opcjonalnie - dla przyszłej rozbudowy
}

// AnalyzeAlerts analizuje wszystkie okna i zwraca listę IP do zablokowania

func (wm *WindowManager) AnalyzeAlerts() []BlockDecision {
	var decisions []BlockDecision

	for ip, window := range wm.Windows {
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

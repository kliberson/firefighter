package suricata

import (
	"firefighter/data"
	"fmt"
	"log"
)

type BlockDecision struct {
	IP            string
	Reason        string
	Score         int
	Details       string
	AlertCount    int
	SeverityScore int
	UniquePorts   int
	UniqueProtos  int
	UniqueFlows   int
	Categories    map[string]int
}

func (wm *WindowManager) AnalyzeAlerts(db *data.DbManager) []BlockDecision {
	var decisions []BlockDecision
	baseThreshold := 30

	for ip, window := range wm.Windows {
		// Cleanup pustych okien
		if window.Events.Len() == 0 {
			delete(wm.Windows, ip)
			continue
		}

		stats := struct {
			Count         int
			SeverityScore int
			Categories    map[string]int
			UniquePorts   map[int]bool
			UniqueProtos  map[string]bool
			UniqueSIDs    map[int]bool
			UniqueFlows   map[uint64]bool
		}{
			Categories:   make(map[string]int),
			UniquePorts:  make(map[int]bool),
			UniqueProtos: make(map[string]bool),
			UniqueSIDs:   make(map[int]bool),
			UniqueFlows:  make(map[uint64]bool),
		}

		// Pętla przez wszystkie alerty w sliding window
		for e := window.Events.Front(); e != nil; e = e.Next() {
			a := e.Value.(Alert)
			stats.Count++

			// Severity scoring
			switch a.Alert.Severity {
			case 1:
				stats.SeverityScore += 10
			case 2:
				stats.SeverityScore += 5
			case 3:
				stats.SeverityScore += 2
			}

			// Agregacja statystyk
			stats.Categories[a.Alert.Category]++
			stats.UniquePorts[a.DstPort] = true
			stats.UniqueProtos[a.Proto] = true
			stats.UniqueSIDs[a.Alert.SignatureID] = true

			// Flow tracking
			if a.FlowID != 0 {
				stats.UniqueFlows[a.FlowID] = true
			}
		}

		// Obliczanie końcowego scoringu
		score := stats.SeverityScore
		score += len(stats.Categories) * 5
		score += len(stats.UniquePorts) * 3
		score += len(stats.UniqueProtos) * 4
		score += len(stats.UniqueSIDs)

		// Flow scoring - wiele flow z jednego IP = podejrzane
		flowCount := len(stats.UniqueFlows)
		if flowCount >= 5 {
			score += flowCount * 4
			log.Printf("⚠️  IP %s ma %d różnych flow - podejrzane skanowanie!", ip, flowCount)
		}

		// Tworzenie szczegółowego raportu
		reason := fmt.Sprintf(
			"Score:%d, Severity:%d, Categories:%v, Ports:%d, Protos:%d, SIDs:%d, Flows:%d, Count:%d",
			score, stats.SeverityScore, stats.Categories, len(stats.UniquePorts),
			len(stats.UniqueProtos), len(stats.UniqueSIDs), flowCount, stats.Count,
		)

		// Sprawdzanie warunków blokowania
		isWhitelisted, err := db.IsWhitelisted(ip)
		if err != nil {
			log.Printf("Błąd sprawdzania whitelisty dla %s: %v", ip, err)
		}

		isBlocked, err := db.IsBlocked(ip)
		if err != nil {
			log.Printf("Błąd sprawdzania statusu blokady dla %s: %v", ip, err)
		}

		// Decyzja o blokowaniu
		if !isWhitelisted && !isBlocked && score >= baseThreshold {
			log.Printf("🚨 IP %s przekroczył threshold scoringu (%d), blokada!", ip, score)

			dynamicReason := generateBlockReason(
				stats.Categories,
				stats.Count,
				len(stats.UniquePorts),
				len(stats.UniqueFlows),
			)

			decisions = append(decisions, BlockDecision{
				IP:            ip,
				Reason:        dynamicReason, // ← ZMIANA
				Score:         score,
				Details:       reason,
				AlertCount:    stats.Count,
				SeverityScore: stats.SeverityScore,
				UniquePorts:   len(stats.UniquePorts),
				UniqueProtos:  len(stats.UniqueProtos),
				UniqueFlows:   len(stats.UniqueFlows),
				Categories:    stats.Categories,
			})
			window.Events.Init()
		}
	}

	return decisions
}

func generateBlockReason(categories map[string]int, count, ports, flows int) string {
	// Znajdź top kategorię
	topCategory := "Multiple attacks"
	maxCount := 0
	for cat, cnt := range categories {
		if cnt > maxCount {
			maxCount = cnt
			topCategory = cat
		}
	}

	// Dynamiczny opis
	if flows >= 5 {
		return fmt.Sprintf("%s (%d flows, %d alerts)", topCategory, flows, count)
	}

	if ports >= 5 {
		return fmt.Sprintf("%s (%d ports)", topCategory, ports)
	}

	if count >= 10 {
		return fmt.Sprintf("%s (%d alerts)", topCategory, count)
	}

	return fmt.Sprintf("%s (%d alerts)", topCategory, count)
}

package suricata

import (
	"sync"
	"time"
)

// SlidingWindow przechowuje alerty w oknie czasowym
type SlidingWindow struct {
	alerts    []time.Time // Lista czasów alertów
	windowSec int         // Rozmiar okna w sekundach
	threshold int         // Próg alertów
	mutex     sync.Mutex  // Zabezpieczenie przed race conditions
}

// NewSlidingWindow tworzy nowe okno
func NewSlidingWindow(windowSeconds, threshold int) *SlidingWindow {
	return &SlidingWindow{
		alerts:    make([]time.Time, 0),
		windowSec: windowSeconds,
		threshold: threshold,
	}
}

// AddAlert dodaje nowy alert do okna
func (sw *SlidingWindow) AddAlert() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	now := time.Now()

	// 1. Dodaj nowy alert
	sw.alerts = append(sw.alerts, now)

	// 2. Usuń stare alerty (starsze niż okno czasowe)
	sw.removeOldAlerts(now)

	// 3. Sprawdź czy przekroczono próg
	return len(sw.alerts) >= sw.threshold
}

// removeOldAlerts usuwa alerty starsze niż okno czasowe
func (sw *SlidingWindow) removeOldAlerts(currentTime time.Time) {
	cutoffTime := currentTime.Add(-time.Duration(sw.windowSec) * time.Second)

	// Znajdź pierwszy alert który jest wystarczająco nowy
	validIndex := 0
	for i, alertTime := range sw.alerts {
		if alertTime.After(cutoffTime) {
			validIndex = i
			break
		}
		validIndex = len(sw.alerts) // Wszystkie są stare
	}

	// Usuń stare alerty
	if validIndex > 0 {
		sw.alerts = sw.alerts[validIndex:]
	}
}

// GetCount zwraca aktualną liczbę alertów w oknie
func (sw *SlidingWindow) GetCount() int {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	// Usuń stare alerty przed zwróceniem liczby
	sw.removeOldAlerts(time.Now())
	return len(sw.alerts)
}

// Clear czyści wszystkie alerty
func (sw *SlidingWindow) Clear() {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	sw.alerts = sw.alerts[:0] // Wyczyść slice
}

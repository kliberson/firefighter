// main.go
package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	api "firefighter/api"
	suricata "firefighter/core"
	"firefighter/data"
)

func main() {
	// ← DODANE: Setup loggera (tekstowy)
	os.MkdirAll("/var/log/firefighter", 0755)
	logFile, err := os.OpenFile("/var/log/firefighter/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Cannot open log file:", err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	slog.SetDefault(logger)

	slog.Info("System starting", "version", "1.0.0") // ← DODANE

	db, err := data.New("/home/lucas/firefighter/data/firefighter.db")
	if err != nil {
		slog.Error("Database connection failed", "error", err) // ← DODANE
		log.Fatal("Unable to open database:", err)
	}
	defer db.Close()
	slog.Info("Database connected", "path", "/home/lucas/firefighter/data/firefighter.db") // ← DODANE

	go api.StartHub()

	wm := suricata.NewWindowManager(600 * time.Second)

	// HTTP server
	r := api.SetupRouter(db, wm)

	go func() {
		slog.Info("HTTP server starting", "port", 8080) // ← DODANE
		if err := r.Run(":8080"); err != nil {
			slog.Error("HTTP server failed", "error", err) // ← DODANE
			log.Fatal("Failed to run server:", err)
		}
	}()

	// Suricata setup
	suricataCmd := exec.Command("sudo", "suricata",
		"-c", "/etc/suricata/suricata.yaml",
		"-i", "enp0s3",
		"-v")

	if err := suricataCmd.Start(); err != nil {
		slog.Error("Suricata start failed", "error", err) // ← DODANE
		log.Fatal("Failed to start Suricata:", err)
	}
	defer suricataCmd.Process.Kill()

	alertChan := make(chan suricata.Alert, 1000)

	go func() {
		if err := suricata.StartServer(suricata.SuricataSocketPath, alertChan); err != nil {
			slog.Error("Unix socket server failed", "error", err) // ← DODANE
			log.Fatal(err)
		}
	}()

	// ← DODANE: Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		slog.Info("Shutdown signal received, stopping gracefully") // ← DODANE
		db.Close()
		slog.Info("Database closed") // ← DODANE
		suricataCmd.Process.Kill()
		slog.Info("Firefighter stopped") // ← DODANE
		os.Exit(0)
	}()

	slog.Info("Firefighter started successfully") // ← DODANE
	fmt.Println("Firefighter started! Waiting for alerts...")

	// Main loop
	for alert := range alertChan {
		suricata.HandleAlert(alert)

		// Zapisz alert do bazy
		if err := db.AddAlert(alert.SrcIP, alert.Alert.SignatureID, alert.Alert.Signature); err != nil {
			slog.Error("Failed to save alert to database", "error", err) // ← DODANE
			log.Printf("Database error: %v", err)
		}

		api.BroadcastAlert(
			alert.SrcIP,
			alert.Alert.Signature,
			alert.Alert.SignatureID,
			alert.Alert.Severity,
			alert.SrcPort,
			alert.DstPort,
			alert.Proto,
			alert.Alert.Category,
		)

		wm.Add(alert)
		wm.PrintAll()

		// === ANALIZA I BLOKOWANIE ===
		decisions := wm.AnalyzeAlerts(db)

		for _, decision := range decisions {
			// 1. Sprawdź whitelist
			if whitelisted, _ := db.IsWhitelisted(decision.IP); whitelisted {
				slog.Info("IP whitelisted, skipping block", "ip", decision.IP) // ← DODANE
				fmt.Printf("⚪ IP %s is whitelisted - skipping\n", decision.IP)
				continue
			}

			// 2. Sprawdź czy już zablokowany
			if blocked, _ := db.IsBlocked(decision.IP); blocked {
				slog.Warn("IP already blocked, skipping", "ip", decision.IP) // ← DODANE
				fmt.Printf("⚠️  IP %s already blocked - skipping\n", decision.IP)
				continue
			}

			// 3. Blokuj w firewall
			if err := suricata.BlockIP(decision.IP); err != nil {
				slog.Error("Firewall block failed", "ip", decision.IP, "error", err) // ← DODANE
				log.Printf("❌ Firewall block failed for %s: %v", decision.IP, err)
				continue
			}

			// 4. Zapisz do bazy z pełnymi danymi
			categoriesStr := formatCategories(decision.Categories)

			if err := db.AddBlocked(
				decision.IP,
				decision.Reason,
				decision.Score,
				decision.AlertCount,
				decision.SeverityScore,
				decision.UniquePorts,
				decision.UniqueProtos,
				decision.UniqueFlows,
				categoriesStr,
				decision.Details,
			); err != nil {
				slog.Error("Failed to save block to database", "ip", decision.IP, "error", err) // ← DODANE
				log.Printf("Database save failed for %s: %v", decision.IP, err)
			} else {
				slog.Info("IP blocked successfully", "ip", decision.IP, "score", decision.Score, "reason", decision.Reason) // ← DODANE
				fmt.Printf("🚫 BLOCKED: %s - %s (Score: %d)\n", decision.IP, decision.Reason, decision.Score)

				api.BroadcastBlockWithScore(
					decision.IP,
					decision.Reason,
					decision.Score,
					decision.AlertCount,
					decision.SeverityScore,
					decision.UniquePorts,
					decision.UniqueProtos,
					decision.UniqueFlows,
					formatCategories(decision.Categories),
					decision.Details,
				)
			}
		}
	}
}

// Helper: konwertuje map[string]int do stringa "Category1:5, Category2:3"
type catPair struct {
	name  string
	count int
}

func formatCategories(cats map[string]int) string {
	var pairs []catPair
	for name, count := range cats {
		pairs = append(pairs, catPair{name, count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	var parts []string
	for _, p := range pairs {
		parts = append(parts, fmt.Sprintf("%s:%d", p.name, p.count))
	}
	return strings.Join(parts, ", ")
}

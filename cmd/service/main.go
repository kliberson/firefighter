// main.go
package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	api "firefighter/api"
	suricata "firefighter/core"
	"firefighter/data"
)

func main() {

	db, err := data.New("/home/lucas/firefighter/data/firefighter.db")
	if err != nil {
		log.Fatal("Unable to open database:", err)
	}
	defer db.Close()

	go api.StartHub()

	//WSTĘPNIE SERVER HTTP
	r := api.SetupRouter(db)

	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Failed to run server:", err)
		}
	}()

	// Suricata setup
	suricataCmd := exec.Command("sudo", "suricata",
		"-c", "/etc/suricata/suricata.yaml",
		"-i", "enp0s3",
		"-v")

	if err := suricataCmd.Start(); err != nil {
		log.Fatal("Failed to start Suricata:", err)
	}
	defer suricataCmd.Process.Kill()

	alertChan := make(chan suricata.Alert, 100)

	go func() {
		if err := suricata.StartServer(suricata.SuricataSocketPath, alertChan); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("🔥 Firefighter started! Waiting for alerts...")

	wm := suricata.NewWindowManager(25 * time.Second)

	// Main loop
	for alert := range alertChan {
		suricata.HandleAlert(alert)

		// Zapisz alert do bazy
		if err := db.AddAlert(alert.SrcIP, alert.Alert.SignatureID, alert.Alert.Signature); err != nil {
			log.Printf("Database error: %v", err)
		}

		wm.Add(alert)
		wm.PrintAll()

		// === ANALIZA I BLOKOWANIE ===
		decisions := wm.AnalyzeAlerts(db)

		for _, decision := range decisions {
			// 1. Sprawdź whitelist
			if whitelisted, _ := db.IsWhitelisted(decision.IP); whitelisted {
				fmt.Printf("⚪ IP %s is whitelisted - skipping\n", decision.IP)
				continue
			}

			// 2. Sprawdź czy już zablokowany
			if blocked, _ := db.IsBlocked(decision.IP); blocked {
				fmt.Printf("⚠️  IP %s already blocked - skipping\n", decision.IP)
				continue
			}

			// 3. Blokuj w firewall
			if err := suricata.BlockIP(decision.IP); err != nil {
				log.Printf("❌ Firewall block failed for %s: %v", decision.IP, err)
				continue
			}

			// 4. Zapisz do bazy
			if err := db.AddBlocked(decision.IP, decision.Reason); err != nil {
				log.Printf("Database save failed for %s: %v", decision.IP, err)
			} else {
				fmt.Printf("BLOCKED: %s - %s\n", decision.IP, decision.Reason)

				//  Wysyłanie alertu do klientów WebSocket
				api.BroadcastBlock(decision.IP, decision.Reason)
			}
		}
	}
}

package main

import (
	"fmt"
	"log"
	"os/exec"

	suricata "firefighter/core" // lub twoja ścieżka
)

func main() {

	suricataCmd := exec.Command("sudo", "suricata", "-c", "/etc/suricata/suricata.yaml", "-i", "enp0s3")
	if err := suricataCmd.Start(); err != nil {
		log.Fatal("Nie można uruchomić Suricata:", err)
	}
	defer suricataCmd.Process.Kill()

	alertChan := make(chan suricata.Alert, 100)

	// Uruchom serwer w goroutine
	go func() {
		if err := suricata.StartServer(suricata.SuricataSocketPath, alertChan); err != nil {
			log.Fatal(err)
		}
	}()

	// Odbieraj alerty
	for alert := range alertChan {
		fmt.Printf("🚨 ALERT: %s -> %s:%d | %s\n",
			alert.SrcIP, alert.DstIP, alert.DstPort, alert.Alert.Signature)
	}
}

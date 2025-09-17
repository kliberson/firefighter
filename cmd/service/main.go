package main

import (
	"fmt"
	"log"
	"os/exec"

	suricata "firefighter/core"
)

func main() {

	suricataCmd := exec.Command("sudo", "suricata",
		"-c", "/etc/suricata/suricata.yaml",
		"-i", "enp0s3",
		"-v")

	if err := suricataCmd.Start(); err != nil {
		log.Fatal("Nie można uruchomić Suricata:", err)
	}
	defer suricataCmd.Process.Kill()

	// Channel for receiving alerts
	alertChan := make(chan suricata.Alert, 100)

	// Starting unix socket server
	go func() {
		if err := suricata.StartServer(suricata.SuricataSocketPath, alertChan); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("🔥 Firefighter uruchomiony! Oczekuję na alerty...")

	for alert := range alertChan {
		suricata.HandleAlert(alert)
	}
}

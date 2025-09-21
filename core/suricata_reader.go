package suricata

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const SuricataSocketPath = "/var/run/suricata/eve.sock"

func StartServer(socketPath string, out chan<- Alert) error {
	// Usuń stary socket
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("nie mogę utworzyć Unix socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Printf("Ostrzeżenie: Nie można ustawić uprawnień socketu: %v", err)
	}

	fmt.Printf("[Suricata] Serwer nasłuchuje na %s\n", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("błąd akceptowania połączenia: %w", err)
		}
		fmt.Println("[Suricata] Suricata połączona!")
		go handleConnection(conn, out)
	}
}

// Parsing alerts and sending to unix socket channel
func handleConnection(conn net.Conn, out chan<- Alert) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	alertCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var alert Alert
		if err := json.Unmarshal([]byte(line), &alert); err != nil {
			log.Printf("[Suricata] Błąd parsowania JSON: %v\nJSON: %s", err, line)
			continue
		}

		if alert.Timestamp != "" {
			parsed, err := time.Parse(time.RFC3339Nano, alert.Timestamp)
			if err == nil {
				alert.ParsedTime = parsed
			} else {
				// fallback: jeśli parsing się nie uda, ustaw teraz
				alert.ParsedTime = time.Now()
			}
		} else {
			alert.ParsedTime = time.Now()
		}

		// filtering SID == 0
		if alert.Alert.SignatureID == 0 {
			continue
		}

		alertCount++
		out <- alert
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[Suricata] Błąd czytania ze socketu: %v", err)
	}
}

// Writing alert to console
func HandleAlert(alert Alert) {
	text := alert.Alert.Signature
	if text == "" {
		text = fmt.Sprintf("%s -> %s:%d", alert.SrcIP, alert.DstIP, alert.DstPort)
	}

	fmt.Printf("[ALERT] %s (SID: %d)\n", text, alert.Alert.SignatureID)
	fmt.Printf("  └─ %s:%d -> %s:%d (%s)\n",
		alert.SrcIP, alert.SrcPort, alert.DstIP, alert.DstPort, alert.Proto)
	fmt.Println(strings.Repeat("-", 50))
}

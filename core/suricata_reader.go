package suricata

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

const SuricataSocketPath = "/var/run/suricata/eve.sock"

func StartServer(socketPath string, out chan<- Alert) error {

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("Cannot create Unix Socket server: %w", err)
	}

	defer listener.Close()
	defer os.Remove(socketPath)

	if err := os.Chmod(socketPath, 0666); err != nil {
		slog.Warn("Cannot set socket permissions", "error", err)
	}

	slog.Info("Listening on Unix socket", "path", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Socket accept failed", "error", err)
		}
		slog.Info("Suricata connected")

		go handleConnection(conn, out)
	}
}

// Parsing alerts and sending to unix socket channel
func handleConnection(conn net.Conn, out chan<- Alert) {
	defer conn.Close()
	defer slog.Warn("Suricata disconnected") // ← DODANE

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
		slog.Error("Socket read error", "error", err) // ← DODANE
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

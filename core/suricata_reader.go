package suricata

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type Alert struct {
	Timestamp string `json:"timestamp"`
	SrcIP     string `json:"src_ip"`
	SrcPort   int    `json:"src_port"`
	DstIP     string `json:"dest_ip"`
	DstPort   int    `json:"dest_port"`
	Proto     string `json:"proto"`
	Alert     struct {
		Signature string `json:"signature"`
		Category  string `json:"category"`
		Severity  int    `json:"severity"`
	} `json:"alert"`
}

const SuricataSocketPath = "/var/run/suricata/eve.sock"

// StartServer tworzy Unix socket server i czeka na połączenie od Suricata
func StartServer(socketPath string, out chan<- Alert) error {
	// Usuń socket jeśli już istnieje
	os.Remove(socketPath)

	// Utwórz Unix socket listener (SERVER)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("nie mogę utworzyć Unix socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	// Ustaw uprawnienia żeby Suricata mogła się połączyć
	err = os.Chmod(socketPath, 0666)
	if err != nil {
		log.Printf("Ostrzeżenie: Nie można ustawić uprawnień socketu: %v", err)
	}

	fmt.Printf("[Suricata] Serwer nasłuchuje na %s\n", socketPath)
	fmt.Println("[Suricata] Czekam na połączenie od Suricata...")

	for {
		// Akceptuj połączenie od Suricata
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("błąd akceptowania połączenia: %w", err)
		}

		fmt.Println("[Suricata] Suricata połączona!")

		// Obsługuj połączenie w osobnej goroutine
		go handleSuricataConnection(conn, out)
	}
}

func handleSuricataConnection(conn net.Conn, out chan<- Alert) {
	defer conn.Close()
	defer fmt.Println("[Suricata] Suricata rozłączona")

	scanner := bufio.NewScanner(conn)
	alertCount := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		var alert Alert

		if err := json.Unmarshal(line, &alert); err != nil {
			// Jeśli to nie jest alert (np. log HTTP, DNS itp.), pomijamy
			continue
		}

		alertCount++
		fmt.Printf("[Suricata] Alert #%d: %s\n", alertCount, alert.Alert.Signature)

		// Wysyłamy do kanału, żeby inne części systemu mogły to odebrać
		select {
		case out <- alert:
		default:
			log.Println("[Suricata] Ostrzeżenie: Kanał pełny, pomijam alert")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[Suricata] Błąd czytania z socketu: %v", err)
	}
}

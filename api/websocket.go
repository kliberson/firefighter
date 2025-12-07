package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type      string `json:"type"`
	IP        string `json:"ip"`
	Reason    string `json:"reason"`
	Timestamp int64  `json:"timestamp"`

	// Alerty
	SID      int    `json:"sid,omitempty"`
	Category string `json:"category,omitempty"`
	Severity int    `json:"severity,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	SrcPort  int    `json:"src_port,omitempty"`
	DstPort  int    `json:"dst_port,omitempty"`

	// Blokady
	Score         int    `json:"score"`
	Details       string `json:"details"`
	AlertCount    int    `json:"alert_count"`    // ← DODAJ
	SeverityScore int    `json:"severity_score"` // ← DODAJ
	UniquePorts   int    `json:"unique_ports"`   // ← DODAJ
	UniqueProtos  int    `json:"unique_protos"`  // ← DODAJ
	UniqueFlows   int    `json:"unique_flows"`   // ← DODAJ
	Categories    string `json:"categories"`     // ← DODAJ
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WebSocketMessage // ← Zmienione z WebSocketEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

var hub *Hub

func init() {
	hub = &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WebSocketMessage, 100), // ← Zmienione
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartHub() {
	go hub.Run()
	log.Println("✅ WebSocket Hub started")
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			log.Println("✅ New WebSocket client connected. Total:", len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
				log.Println("🔌 Client disconnected. Total:", len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast: // ← Zmienione z event
			h.mu.RLock()
			log.Printf("📡 Broadcasting %s: %s to %d clients", message.Type, message.IP, len(h.clients))
			for conn := range h.clients {
				err := conn.WriteJSON(message)
				if err != nil {
					log.Printf("❌ Write error: %v", err)
					h.mu.RUnlock()
					h.unregister <- conn
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func BroadcastAlert(ip, signature string, sid, severity, srcPort, dstPort int, protocol, category string) {
	hub.broadcast <- WebSocketMessage{
		Type:      "alert",
		IP:        ip,
		Reason:    signature,
		Timestamp: time.Now().Unix(),
		SID:       sid,
		Category:  category,
		Severity:  severity,
		Protocol:  protocol,
		SrcPort:   srcPort,
		DstPort:   dstPort,
	}
}

// BroadcastBlock - wysyła blokadę (prostsze, bez dodatkowych danych)
func BroadcastBlockWithScore(ip, reason string, score, alertCount, severityScore, uniquePorts, uniqueProtos, uniqueFlows int, categories, details string) {
	hub.broadcast <- WebSocketMessage{
		Type:      "block",
		IP:        ip,
		Reason:    reason,
		Score:     score,
		Details:   details,
		Timestamp: time.Now().Unix(),
		// ← DODAJ te pola do struktury WebSocketMessage:
		AlertCount:    alertCount,
		SeverityScore: severityScore,
		UniquePorts:   uniquePorts,
		UniqueProtos:  uniqueProtos,
		UniqueFlows:   uniqueFlows,
		Categories:    categories,
	}
	log.Printf("📤 Queued block event for %s (score: %d)", ip, score)
}

// BroadcastUnblock - wysyła odblokowanie
func BroadcastUnblock(ip string) {
	hub.broadcast <- WebSocketMessage{
		Type:      "unblock",
		IP:        ip,
		Reason:    "Manually unblocked",
		Timestamp: time.Now().Unix(),
	}
	log.Printf("📤 Queued unblock event for %s", ip)
}

func handleWebSocket(c *gin.Context) {
	log.Println("🔌 WebSocket upgrade request from:", c.ClientIP())

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade FAILED: %v", err)
		return
	}

	log.Println("✅ WebSocket connection established")
	hub.register <- conn

	// Keep-alive ping
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	defer func() {
		hub.unregister <- conn
	}()

	// Read loop (keep connection open)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket unexpected close: %v", err)
			}
			break
		}
	}
}

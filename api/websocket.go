package api

import (
	"fmt"
	"log"
	"log/slog"
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
	SID      string `json:"sid,omitempty"`
	Category string `json:"category,omitempty"`
	Severity string `json:"severity,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	SrcPort  string `json:"src_port,omitempty"`
	DstPort  string `json:"dst_port,omitempty"`

	// Blokady
	Score         string `json:"score"`
	Details       string `json:"details"`
	AlertCount    string `json:"alert_count"`
	SeverityScore string `json:"severity_score"`
	UniquePorts   string `json:"unique_ports"`
	UniqueProtos  string `json:"unique_protos"`
	UniqueFlows   string `json:"unique_flows"`
	Categories    string `json:"categories"`
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WebSocketMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

var hub *Hub

func init() {
	hub = &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WebSocketMessage, 100),
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
	log.Println("WebSocket Hub started")
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			clientCount := len(h.clients)
			h.mu.Unlock()
			slog.Info("WebSocket client connected", "total_clients", clientCount) // ← DODANE
			log.Println("New WebSocket client connected. Total:", clientCount)

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
				clientCount := len(h.clients)
				h.mu.Unlock()
				slog.Info("WebSocket client disconnected", "total_clients", clientCount) // ← DODANE
				log.Println("Client disconnected. Total:", clientCount)
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.clients {
				err := conn.WriteJSON(message)
				if err != nil {
					slog.Warn("WebSocket write failed, closing connection", "error", err) // ← DODANE
					log.Printf("Write error: %v", err)
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
		SID:       fmt.Sprintf("%d", sid),
		Category:  category,
		Severity:  fmt.Sprintf("%d", severity),
		Protocol:  protocol,
		SrcPort:   fmt.Sprintf("%d", srcPort),
		DstPort:   fmt.Sprintf("%d", dstPort),
	}
}

func BroadcastBlockWithScore(ip, reason string, score, alertCount, severityScore, uniquePorts, uniqueProtos, uniqueFlows int, categories, details string) {
	hub.broadcast <- WebSocketMessage{
		Type:          "block",
		IP:            ip,
		Reason:        reason,
		Score:         fmt.Sprintf("%d", score),
		Details:       details,
		Timestamp:     time.Now().Unix(),
		AlertCount:    fmt.Sprintf("%d", alertCount),
		SeverityScore: fmt.Sprintf("%d", severityScore),
		UniquePorts:   fmt.Sprintf("%d", uniquePorts),
		UniqueProtos:  fmt.Sprintf("%d", uniqueProtos),
		UniqueFlows:   fmt.Sprintf("%d", uniqueFlows),
		Categories:    categories,
	}
}

func BroadcastUnblock(ip string) {
	hub.broadcast <- WebSocketMessage{
		Type:      "unblock",
		IP:        ip,
		Reason:    "Manually unblocked",
		Timestamp: time.Now().Unix(),
	}
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err) // ← DODANE
		log.Printf("WebSocket upgrade FAILED: %v", err)
		return
	}

	hub.register <- conn

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

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("WebSocket unexpected close", "error", err) // ← DODANE
				log.Printf("WebSocket unexpected close: %v", err)
			}
			break
		}
	}
}

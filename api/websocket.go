// api/websocket.go
package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketEvent struct {
	Type      string    `json:"type"`
	IP        string    `json:"ip"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WebSocketEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

var hub *Hub

func init() {
	hub = &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WebSocketEvent, 100),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Development: allow all
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

		case event := <-h.broadcast:
			h.mu.RLock()
			log.Printf("📡 Broadcasting event: %s to %d clients", event.Type, len(h.clients))
			for conn := range h.clients {
				err := conn.WriteJSON(event)
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

func BroadcastBlock(ip, reason string) {
	hub.broadcast <- WebSocketEvent{
		Type:      "block",
		IP:        ip,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	log.Printf("📤 Queued block event for %s", ip)
}

func BroadcastUnblock(ip string) {
	hub.broadcast <- WebSocketEvent{
		Type:      "unblock",
		IP:        ip,
		Reason:    "Manually unblocked",
		Timestamp: time.Now(),
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

	// Keep connection alive with pings
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
				log.Printf("❌ WebSocket unexpected close: %v", err)
			}
			break
		}
	}
}

package authorityapi

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

var liveUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
}

// EventBroadcaster manages live WebSocket event broadcasting
type EventBroadcaster struct {
	clients map[*websocket.Conn]bool
	mutex   sync.RWMutex
}

// Global broadcaster instance
var broadcaster = &EventBroadcaster{
	clients: make(map[*websocket.Conn]bool),
}

// HandleWebSocketEvents handles live WebSocket connections for authority events
func HandleWebSocketEvents(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := liveUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	broadcaster.AddClient(conn)
	defer broadcaster.RemoveClient(conn)

	log.Printf("WebSocket client connected from %s", r.RemoteAddr)

	// Send initial event burst with recent decisions
	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if ok {
		sendInitialEvents(conn, db, r.Context())
	}

	// Keep connection alive and handle client messages
	for {
		// Read message from client (ping/pong, etc.)
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Send pong response to keep connection alive
		conn.WriteMessage(websocket.PongMessage, []byte("pong"))
	}
}

// AddClient adds a new WebSocket client
func (eb *EventBroadcaster) AddClient(conn *websocket.Conn) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	eb.clients[conn] = true
}

// RemoveClient removes a WebSocket client
func (eb *EventBroadcaster) RemoveClient(conn *websocket.Conn) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	delete(eb.clients, conn)
}

// BroadcastEvent sends an event to all connected clients
func (eb *EventBroadcaster) BroadcastEvent(event Event) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	for conn := range eb.clients {
		err := conn.WriteJSON(event)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			// Remove broken connection
			delete(eb.clients, conn)
			conn.Close()
		}
	}
}

// BroadcastDecisionEvent broadcasts a decision event to all clients
func BroadcastDecisionEvent(decision map[string]interface{}) {
	event := Event{
		ID:        "evt-live-" + time.Now().Format("150405"),
		Type:      "authority.decision.v1",
		Domain:    decision["domain"].(string),
		Actor:     decision["actor"].(string),
		CreatedAt: time.Now(),
		Phase:     "9B-live",
		Payload:   decision,
	}

	broadcaster.BroadcastEvent(event)
}

// sendInitialEvents sends recent events to newly connected client
func sendInitialEvents(conn *websocket.Conn, db *pgxpool.Pool, ctx context.Context) {
	events, err := queryRecentEvents(ctx, db)
	if err != nil {
		log.Printf("Failed to query initial events: %v", err)
		return
	}

	for _, event := range events {
		err := conn.WriteJSON(event)
		if err != nil {
			log.Printf("Failed to send initial event: %v", err)
			break
		}
		time.Sleep(100 * time.Millisecond) // Stagger events
	}
}

// SendHeartbeatEvents periodically sends heartbeat events (optional)
func SendHeartbeatEvents() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		heartbeat := Event{
			ID:        "heartbeat-" + time.Now().Format("150405"),
			Type:      "system.heartbeat.v1",
			Domain:    "system",
			Actor:     "dis-core",
			CreatedAt: time.Now(),
			Phase:     "9B-live",
		}

		broadcaster.BroadcastEvent(heartbeat)
	}
}

package authorityapi

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for mock implementation
		return true
	},
}

// MockEvent represents a mock event for WebSocket streaming
type MockEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Domain    string    `json:"domain"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
	Phase     string    `json:"phase"`
}

var eventTypes = []string{
	"domain.freeze.v1",
	"domain.unfreeze.v1",
	"ci.call.v1",
	"policy.evaluate.v1",
	"schema.validate.v1",
	"domain.create.v1",
	"domain.update.v1",
}

var domains = []string{
	"domain.user.rick",
	"domain.user.alice",
	"domain.terra",
	"domain.null",
	"domain.test.new",
}

var actors = []string{
	"id-rick",
	"id-alice",
	"id-bob",
	"system",
	"ci-agent",
}

// generateMockEvent creates a random mock event
func generateMockEvent() MockEvent {
	eventID := generateEventID()

	return MockEvent{
		ID:        eventID,
		Type:      eventTypes[rand.Intn(len(eventTypes))],
		Domain:    domains[rand.Intn(len(domains))],
		Actor:     actors[rand.Intn(len(actors))],
		CreatedAt: time.Now(),
		Phase:     "9A-mock",
	}
}

// generateEventID creates a simple event ID
func generateEventID() string {
	return "evt-" + time.Now().Format("20060102-150405") + "-" + generateRandomSuffix()
}

// generateRandomSuffix creates a random 3-digit suffix
func generateRandomSuffix() string {
	return string(rune('A'+rand.Intn(26))) + string(rune('A'+rand.Intn(26))) + string(rune('A'+rand.Intn(26)))
}

// HandleEventsWebSocketMock handles WebSocket connections for /ws/authority/events
func HandleEventsWebSocketMock(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("New WebSocket client connected to /ws/authority/events")

	// Create a ticker that sends events every 3-5 seconds
	ticker := time.NewTicker(time.Duration(3+rand.Intn(3)) * time.Second)
	defer ticker.Stop()

	// Send initial event immediately
	initialEvent := generateMockEvent()
	if err := conn.WriteJSON(initialEvent); err != nil {
		log.Printf("Error sending initial event: %v", err)
		return
	}

	// Main event loop
	for {
		select {
		case <-ticker.C:
			// Generate and send new mock event
			event := generateMockEvent()
			if err := conn.WriteJSON(event); err != nil {
				log.Printf("Error sending event: %v", err)
				return
			}
			log.Printf("Sent mock event: %s (%s)", event.ID, event.Type)

			// Reset ticker with new random interval (3-5 seconds)
			ticker.Reset(time.Duration(3+rand.Intn(3)) * time.Second)

		default:
			// Check if client is still connected
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Client disconnected")
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}

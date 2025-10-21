package net

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Manager oversees peer connections, health checks, and listener lifecycle.
type Manager struct {
	mu      sync.RWMutex
	peers   map[string]*Peer
	ticker  *time.Ticker
	stop    chan struct{}
	ln      net.Listener
	running bool
	db      *sql.DB
}

// NewManager constructs a new network manager with periodic health checks.
// Optionally accepts a *sql.DB for persistence; pass nil to disable DB ops.
func NewManager(db *sql.DB) *Manager {
	return &Manager{
		peers:  make(map[string]*Peer),
		ticker: time.NewTicker(30 * time.Second),
		stop:   make(chan struct{}),
		db:     db,
	}
}

// Listen starts a TCP listener for inbound peer connections on the given port.
// This replaces the deprecated net.Start() approach.
func (m *Manager) Listen(port int) error {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	m.ln = ln
	m.running = true

	log.Printf("✅ DIS-Network listening on %s", addr)

	go m.acceptLoop()
	m.StartHealthChecks()
	return nil
}

// acceptLoop continuously accepts incoming connections and adds them as peers.
func (m *Manager) acceptLoop() {
	for {
		conn, err := m.ln.Accept()
		if err != nil {
			select {
			case <-m.stop:
				log.Println("🛑 Listener closed, stopping accept loop.")
				return
			default:
				log.Printf("⚠️ accept error: %v", err)
				continue
			}
		}
		go m.handleConn(conn)
	}
}

// handleConn processes a single inbound connection.
func (m *Manager) handleConn(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()

	m.mu.Lock()
	if _, ok := m.peers[addr]; !ok {
		m.peers[addr] = &Peer{Address: addr, Healthy: true}
		log.Printf("🔗 Peer connected: %s", addr)
	}
	m.mu.Unlock()
}

// AddPeer manually adds a peer by address.
func (m *Manager) AddPeer(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.peers[addr]; !ok {
		m.peers[addr] = &Peer{Address: addr, Healthy: false}
		log.Printf("➕ Added peer: %s", addr)
		// Save to DB if available
		if m.db != nil {
			_ = m.SavePeerToDB(m.db, addr)
		}
	}
}

// ListPeers returns a snapshot of current peers.
func (m *Manager) ListPeers() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Peer, 0, len(m.peers))
	for _, p := range m.peers {
		out = append(out, p)
	}
	return out
}

// StartHealthChecks runs a background goroutine to ping peers every 30 seconds.
func (m *Manager) StartHealthChecks() {
	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.mu.RLock()
				addrs := make([]string, 0, len(m.peers))
				for addr := range m.peers {
					addrs = append(addrs, addr)
				}
				m.mu.RUnlock()

				for _, addr := range addrs {
					peer, _ := PingPeer(addr)
					m.mu.Lock()
					m.peers[addr] = peer
					m.mu.Unlock()
				}
			case <-m.stop:
				return
			}
		}
	}()
}

// Close stops all network activity and closes the listener.
func (m *Manager) Close() {
	if !m.running {
		return
	}
	close(m.stop)
	m.ticker.Stop()
	if m.ln != nil {
		_ = m.ln.Close()
	}
	m.running = false
	log.Println("🛑 DIS-Network manager stopped.")
}

// Stop is an alias for Close (for backward compatibility).
func (m *Manager) Stop() {
	m.Close()
}

func (m *Manager) LoadPeersFromDB(db *sql.DB) error {
	if db == nil {
		return nil
	}
	rows, err := db.Query(`SELECT address FROM peers`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return err
		}
		m.AddPeer(addr)
	}
	return nil
}

func (m *Manager) SavePeerToDB(db *sql.DB, addr string) error {
	if db == nil {
		return nil
	}
	_, err := db.Exec(`INSERT INTO peers(address) VALUES($1)
	       ON CONFLICT (address) DO NOTHING`, addr)
	return err
}

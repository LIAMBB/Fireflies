package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for this example
	},
}

const gridSize = 30 // Size of the grid (30x30) - Changed from 50 to 30

// Firefly represents a single firefly in the grid
type Firefly struct {
	state         int           // Current state: -1 (inactive), 0 (dim), 1 (flashing)
	cycleRate     time.Duration // Duration for the full cycle (dim + flash)
	flashDuration time.Duration // Duration for the 'flashing' state
	x, y          int           // Position in the grid
	nextFlashTime time.Time     // Time for the next 'flashing' state
	nextDimTime   time.Time     // Time to return to dim state if flashing
}

// Client represents a connected WebSocket client
type Client struct {
	conn *websocket.Conn
}

// Server manages the firefly simulation and WebSocket connections
type Server struct {
	clients   map[*Client]bool // Connected clients
	fireflies [][]*Firefly     // 2D grid of fireflies
	mutex     sync.RWMutex     // Mutex for thread-safe operations
	broadcast chan bool        // Channel to signal state updates
}

// newServer creates and initializes a new Server instance
func newServer() *Server {
	s := &Server{
		clients:   make(map[*Client]bool),
		fireflies: make([][]*Firefly, gridSize),
		broadcast: make(chan bool),
	}
	// Initialize the grid with inactive fireflies
	for i := range s.fireflies {
		s.fireflies[i] = make([]*Firefly, gridSize)
		for j := range s.fireflies[i] {
			s.fireflies[i][j] = &Firefly{state: -1, x: i, y: j}
		}
	}
	s.initializeState()
	return s
}

// initializeState sets up the initial state of active fireflies
func (s *Server) initializeState() {
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if rand.Float32() < 0.12 { // Approximately 12% chance of a firefly being active (40% reduction)
				firefly := s.fireflies[i][j]
				firefly.cycleRate = time.Duration(rand.Float64()*2000+4000) * time.Millisecond
				firefly.flashDuration = time.Duration(rand.Float64()*133+600) * time.Millisecond
				firefly.nextFlashTime = time.Now().Add(time.Duration(rand.Float64()) * firefly.cycleRate)
				firefly.state = 0
			}
		}
	}
	go s.updateFireflies() // Start the firefly update goroutine
}

// updateFireflies continuously updates the state of all fireflies
func (s *Server) updateFireflies() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		updated := false
		now := time.Now()
		for i := 0; i < gridSize; i++ {
			for j := 0; j < gridSize; j++ {
				firefly := s.fireflies[i][j]
				if firefly.state == -1 {
					continue // Skip inactive fireflies
				}

				if now.After(firefly.nextFlashTime) {
					firefly.state = 1
					firefly.nextDimTime = now.Add(firefly.flashDuration)
					firefly.nextFlashTime = firefly.nextFlashTime.Add(firefly.cycleRate)
					updated = true
				} else if firefly.state == 1 && now.After(firefly.nextDimTime) {
					firefly.state = 0
					updated = true
				} else if firefly.state == 0 && s.checkNeighbors(firefly) {
					// Accelerate next flash time if neighbors are flashing
					firefly.nextFlashTime = firefly.nextFlashTime.Add(-250 * time.Millisecond)
					if now.After(firefly.nextFlashTime) {
						firefly.state = 1
						firefly.nextDimTime = now.Add(firefly.flashDuration)
						firefly.nextFlashTime = now.Add(firefly.cycleRate)
						updated = true
					}
				}
			}
		}
		if updated {
			s.broadcast <- true // Signal that the state has been updated
		}
	}
}

// checkNeighbors checks if any firefly within a 10-cell radius is in the 'flashing' state
func (s *Server) checkNeighbors(firefly *Firefly) bool {
	for i := -10; i <= 10; i++ {
		for j := -10; j <= 10; j++ {
			if i == 0 && j == 0 {
				continue // Skip the firefly itself
			}
			x := (firefly.x + i + gridSize) % gridSize
			y := (firefly.y + j + gridSize) % gridSize
			if s.fireflies[x][y].state == 1 {
				return true // Found a flashing neighbor
			}
		}
	}
	return false
}

// handleConnections manages WebSocket connections
func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{conn: conn}
	defer conn.Close()

	s.mutex.Lock()
	s.clients[client] = true
	s.mutex.Unlock()

	log.Printf("New client connected. Total clients: %d", len(s.clients))

	s.sendFullState(client)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			s.mutex.Lock()
			delete(s.clients, client)
			s.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(s.clients))
			break
		}

		if string(message) == "restart" {
			s.restartSimulation()
		}
	}
}

// restartSimulation reinitializes the simulation
func (s *Server) restartSimulation() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Reset all fireflies to inactive
	for i := range s.fireflies {
		for j := range s.fireflies[i] {
			s.fireflies[i][j] = &Firefly{state: -1, x: i, y: j}
		}
	}

	// Reinitialize active fireflies
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			if rand.Float32() < 0.12 { // Approximately 12% chance of a firefly being active
				firefly := s.fireflies[i][j]
				firefly.cycleRate = time.Duration(rand.Float64()*2000+4000) * time.Millisecond
				firefly.flashDuration = time.Duration(rand.Float64()*133+600) * time.Millisecond
				firefly.nextFlashTime = time.Now().Add(time.Duration(rand.Float64()) * firefly.cycleRate)
				firefly.state = 0
			}
		}
	}

	// Broadcast the new state to all clients
	s.broadcast <- true
}

// broadcastState sends updates to all connected clients
func (s *Server) broadcastState() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.broadcast:
			// State has been updated, but we don't need to do anything here
		case <-ticker.C:
			// Send updates every 100ms
			flatState := s.flattenState()
			s.mutex.RLock()
			for client := range s.clients {
				go s.sendState(client, flatState)
			}
			s.mutex.RUnlock()
		}
	}
}

// sendFullState sends the entire grid state to a single client
func (s *Server) sendFullState(client *Client) {
	flatState := s.flattenState()
	s.sendState(client, flatState)
}

// sendState sends the current state to a single client
func (s *Server) sendState(client *Client, state []int) {
	err := client.conn.WriteJSON(state)
	if err != nil {
		log.Printf("Error broadcasting to client: %v", err)
		s.mutex.Lock()
		delete(s.clients, client)
		s.mutex.Unlock()
	}
}

// flattenState converts the 2D grid into a 1D array for transmission
func (s *Server) flattenState() []int {
	flatState := make([]int, gridSize*gridSize)
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			flatState[i*gridSize+j] = s.fireflies[i][j].state
		}
	}
	return flatState
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Initialize random number generator
	server := newServer()

	http.HandleFunc("/ws", server.handleConnections)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	go server.broadcastState() // Start broadcasting goroutine

	// Use TLS with domain-based certificate
	fmt.Println("Server starting on :443 (WSS)")
	err := http.ListenAndServeTLS(":8443", "/etc/letsencrypt/live/firefly-server.liambarter.me/fullchain.pem", "/etc/letsencrypt/live/firefly-server.liambarter.me/privkey.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

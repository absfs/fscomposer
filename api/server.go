package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/absfs/absfs"
	"github.com/absfs/fscomposer/engine"
	"github.com/absfs/fscomposer/registry"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server represents the API server
type Server struct {
	router    *mux.Router
	store     *CompositionStore
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	broadcast chan Message
}

// CompositionStore manages compositions
type CompositionStore struct {
	compositions map[string]*engine.CompositionSpec
	mu           sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewServer creates a new API server
func NewServer() *Server {
	store := &CompositionStore{
		compositions: make(map[string]*engine.CompositionSpec),
	}

	s := &Server{
		router: mux.NewRouter(),
		store:  store,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Message, 256),
	}

	s.setupRoutes()
	go s.handleBroadcasts()

	return s
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Enable CORS
	s.router.Use(corsMiddleware)

	// Node type endpoints
	s.router.HandleFunc("/api/nodes", s.handleListNodes).Methods("GET")
	s.router.HandleFunc("/api/nodes/{type}", s.handleGetNode).Methods("GET")

	// Composition endpoints
	s.router.HandleFunc("/api/compositions", s.handleListCompositions).Methods("GET")
	s.router.HandleFunc("/api/compositions", s.handleCreateComposition).Methods("POST")
	s.router.HandleFunc("/api/compositions/{id}", s.handleGetComposition).Methods("GET")
	s.router.HandleFunc("/api/compositions/{id}", s.handleUpdateComposition).Methods("PUT")
	s.router.HandleFunc("/api/compositions/{id}", s.handleDeleteComposition).Methods("DELETE")
	s.router.HandleFunc("/api/compositions/{id}/validate", s.handleValidateComposition).Methods("POST")
	s.router.HandleFunc("/api/compositions/{id}/build", s.handleBuildComposition).Methods("POST")

	// WebSocket endpoint
	s.router.HandleFunc("/api/ws", s.handleWebSocket)

	// Serve static files
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// handleListNodes returns all available node types
func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	types := registry.ListTypes()
	nodes := make([]map[string]interface{}, 0, len(types))

	for _, t := range types {
		schema, err := registry.GetSchema(t)
		if err != nil {
			continue
		}

		nodes = append(nodes, map[string]interface{}{
			"type":        schema.Type,
			"description": schema.Description,
			"category":    getNodeCategory(t),
			"fields":      schema.Fields,
		})
	}

	respondJSON(w, http.StatusOK, nodes)
}

// handleGetNode returns details for a specific node type
func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeType := vars["type"]

	schema, err := registry.GetSchema(nodeType)
	if err != nil {
		respondError(w, http.StatusNotFound, "Node type not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"type":        schema.Type,
		"description": schema.Description,
		"category":    getNodeCategory(nodeType),
		"fields":      schema.Fields,
	})
}

// handleListCompositions returns all compositions
func (s *Server) handleListCompositions(w http.ResponseWriter, r *http.Request) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	compositions := make([]map[string]interface{}, 0, len(s.store.compositions))
	for id, spec := range s.store.compositions {
		compositions = append(compositions, map[string]interface{}{
			"id":          id,
			"name":        spec.Name,
			"description": spec.Description,
			"version":     spec.Version,
			"nodeCount":   len(spec.Nodes),
		})
	}

	respondJSON(w, http.StatusOK, compositions)
}

// handleCreateComposition creates a new composition
func (s *Server) handleCreateComposition(w http.ResponseWriter, r *http.Request) {
	var spec engine.CompositionSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate the composition
	validator := engine.NewValidator(&spec)
	if err := validator.ValidateAll(); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Generate ID from name
	id := spec.Name

	s.store.mu.Lock()
	s.store.compositions[id] = &spec
	s.store.mu.Unlock()

	// Broadcast update
	s.broadcast <- Message{
		Type: "composition_created",
		Data: map[string]interface{}{"id": id, "spec": spec},
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   id,
		"spec": spec,
	})
}

// handleGetComposition returns a specific composition
func (s *Server) handleGetComposition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.store.mu.RLock()
	spec, exists := s.store.compositions[id]
	s.store.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Composition not found")
		return
	}

	respondJSON(w, http.StatusOK, spec)
}

// handleUpdateComposition updates a composition
func (s *Server) handleUpdateComposition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var spec engine.CompositionSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate the composition
	validator := engine.NewValidator(&spec)
	if err := validator.ValidateAll(); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	s.store.mu.Lock()
	s.store.compositions[id] = &spec
	s.store.mu.Unlock()

	// Broadcast update
	s.broadcast <- Message{
		Type: "composition_updated",
		Data: map[string]interface{}{"id": id, "spec": spec},
	}

	respondJSON(w, http.StatusOK, spec)
}

// handleDeleteComposition deletes a composition
func (s *Server) handleDeleteComposition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.store.mu.Lock()
	delete(s.store.compositions, id)
	s.store.mu.Unlock()

	// Broadcast update
	s.broadcast <- Message{
		Type: "composition_deleted",
		Data: map[string]interface{}{"id": id},
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// handleValidateComposition validates a composition
func (s *Server) handleValidateComposition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.store.mu.RLock()
	spec, exists := s.store.compositions[id]
	s.store.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Composition not found")
		return
	}

	validator := engine.NewValidator(spec)
	err := validator.ValidateAll()

	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
	})
}

// handleBuildComposition builds and tests a composition
func (s *Server) handleBuildComposition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s.store.mu.RLock()
	spec, exists := s.store.compositions[id]
	s.store.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, "Composition not found")
		return
	}

	builder := engine.NewBuilder(spec)
	fs, err := builder.Build()

	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Test basic operations
	testResult := testFilesystem(fs)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": testResult.Success,
		"tests":   testResult.Tests,
		"error":   testResult.Error,
	})
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// handleBroadcasts sends messages to all connected clients
func (s *Server) handleBroadcasts() {
	for msg := range s.broadcast {
		s.clientsMu.RLock()
		for client := range s.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(s.clients, client)
			}
		}
		s.clientsMu.RUnlock()
	}
}

// TestResult represents filesystem test results
type TestResult struct {
	Success bool                   `json:"success"`
	Tests   []string               `json:"tests"`
	Error   string                 `json:"error,omitempty"`
}

// testFilesystem performs basic filesystem tests
func testFilesystem(fs absfs.FileSystem) TestResult {
	tests := []string{}

	// Test: Create file
	f, err := fs.Create("test.txt")
	if err != nil {
		return TestResult{Success: false, Error: err.Error()}
	}
	tests = append(tests, "✓ Create file")

	// Test: Write data
	data := []byte("Hello from fscomposer!")
	_, err = f.Write(data)
	if err != nil {
		f.Close()
		return TestResult{Success: false, Error: err.Error()}
	}
	f.Close()
	tests = append(tests, "✓ Write data")

	// Test: Read data
	f, err = fs.Open("test.txt")
	if err != nil {
		return TestResult{Success: false, Error: err.Error()}
	}
	readData := make([]byte, len(data))
	_, err = f.Read(readData)
	f.Close()
	if err != nil {
		return TestResult{Success: false, Error: err.Error()}
	}
	tests = append(tests, "✓ Read data")

	// Test: Verify data
	if string(readData) != string(data) {
		return TestResult{Success: false, Error: "data mismatch"}
	}
	tests = append(tests, "✓ Verify data")

	return TestResult{Success: true, Tests: tests}
}

// Helper functions

func getNodeCategory(nodeType string) string {
	if engine.IsBackendNode(nodeType) {
		return "backend"
	}
	return "wrapper"
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

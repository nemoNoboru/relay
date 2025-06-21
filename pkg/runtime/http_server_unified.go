package runtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// UnifiedHTTPServer provides HTTP endpoints using the unified message routing architecture
type UnifiedHTTPServer struct {
	config           *HTTPServerConfig
	evaluator        *Evaluator
	messageRouter    *MessageRouter
	transportAdapter *TransportAdapter
	server           *http.Server
	running          bool
	nodeID           string
}

// NewUnifiedHTTPServer creates a new unified HTTP server instance
func NewUnifiedHTTPServer(evaluator *Evaluator, config *HTTPServerConfig) *UnifiedHTTPServer {
	if config == nil {
		config = DefaultHTTPServerConfig()
	}

	// Generate node ID if not provided
	if config.NodeID == "" {
		config.NodeID = generateUnifiedNodeID()
	}

	// Create message router
	messageRouter := NewMessageRouter()

	// Create transport adapter
	transportAdapter := NewTransportAdapter(messageRouter, config.NodeID)

	// Register all servers from evaluator
	servers := evaluator.GetAllServers()
	for name, server := range servers {
		messageRouter.RegisterServer(name, server)
	}

	return &UnifiedHTTPServer{
		config:           config,
		evaluator:        evaluator,
		messageRouter:    messageRouter,
		transportAdapter: transportAdapter,
		running:          false,
		nodeID:           config.NodeID,
	}
}

// generateUnifiedNodeID creates a unique node identifier
func generateUnifiedNodeID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Start starts the unified HTTP server
func (h *UnifiedHTTPServer) Start() error {
	if h.running {
		return fmt.Errorf("HTTP server is already running")
	}

	// Start message router
	h.messageRouter.Start()

	mux := http.NewServeMux()

	// JSON-RPC 2.0 endpoint - uses transport adapter
	mux.HandleFunc("/rpc", h.transportAdapter.HandleHTTPJSONRPC)

	// WebSocket P2P endpoint - uses transport adapter
	mux.HandleFunc("/ws/p2p", h.transportAdapter.HandleWebSocket)

	// Health check endpoint
	mux.HandleFunc("/health", h.handleHealth)

	// Server info endpoint
	mux.HandleFunc("/info", h.handleInfo)

	// Registry endpoints
	mux.HandleFunc("/registry", h.handleRegistry)
	mux.HandleFunc("/registry/servers", h.handleRegistryServers)
	mux.HandleFunc("/registry/peers", h.handleRegistryPeers)

	// Apply middleware
	handler := h.applyMiddleware(mux)

	h.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", h.config.Host, h.config.Port),
		Handler:      handler,
		ReadTimeout:  h.config.ReadTimeout,
		WriteTimeout: h.config.WriteTimeout,
	}

	h.running = true

	log.Printf("Starting Unified Relay HTTP server on %s:%d", h.config.Host, h.config.Port)
	log.Printf("JSON-RPC 2.0 endpoint: http://%s:%d/rpc", h.config.Host, h.config.Port)
	log.Printf("WebSocket P2P endpoint: ws://%s:%d/ws/p2p", h.config.Host, h.config.Port)
	log.Printf("Server registry: http://%s:%d/registry", h.config.Host, h.config.Port)
	log.Printf("Node ID: %s", h.nodeID)

	return h.server.ListenAndServe()
}

// Stop stops the HTTP server gracefully
func (h *UnifiedHTTPServer) Stop() error {
	if !h.running {
		return nil
	}

	h.running = false

	// Stop message router
	h.messageRouter.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return h.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (h *UnifiedHTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"node_id":   h.nodeID,
		"servers":   len(h.messageRouter.GetAllServers()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleInfo handles server info requests
func (h *UnifiedHTTPServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]interface{}{
		"name":         "Relay HTTP Server (Unified)",
		"version":      "0.3.0-dev",
		"node_id":      h.nodeID,
		"architecture": "unified_message_router",
		"endpoints": map[string]string{
			"jsonrpc":   "/rpc",
			"websocket": "/ws/p2p",
			"health":    "/health",
			"info":      "/info",
			"registry":  "/registry",
		},
		"capabilities": []string{
			"jsonrpc-2.0",
			"websocket-p2p",
			"server-registry",
			"remote-calls",
			"transparent-routing",
		},
		"server_count": len(h.messageRouter.GetAllServers()),
		"timestamp":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleRegistry handles server registry requests
func (h *UnifiedHTTPServer) handleRegistry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	servers := h.messageRouter.GetAllServers()
	serverList := make([]map[string]interface{}, 0, len(servers))

	for name, server := range servers {
		serverInfo := map[string]interface{}{
			"name":    name,
			"type":    "relay_server",
			"node_id": h.nodeID,
			"methods": h.getServerMethods(server),
			"status":  "active",
		}
		serverList = append(serverList, serverInfo)
	}

	registry := map[string]interface{}{
		"node_id":   h.nodeID,
		"servers":   serverList,
		"peers":     []interface{}{}, // TODO: Implement peer tracking
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registry)
}

// handleRegistryServers handles server list requests
func (h *UnifiedHTTPServer) handleRegistryServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	servers := h.messageRouter.GetAllServers()
	serverList := make([]map[string]interface{}, 0, len(servers))

	for name, server := range servers {
		serverInfo := map[string]interface{}{
			"name":    name,
			"type":    "relay_server",
			"node_id": h.nodeID,
			"methods": h.getServerMethods(server),
			"status":  "active",
		}
		serverList = append(serverList, serverInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serverList)
}

// handleRegistryPeers handles peer list requests
func (h *UnifiedHTTPServer) handleRegistryPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement peer tracking in message router
	peers := []interface{}{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// getServerMethods extracts available methods from a server
func (h *UnifiedHTTPServer) getServerMethods(server *Value) []string {
	if server == nil || server.Type != ValueTypeServer || server.Server == nil {
		return []string{}
	}

	methods := make([]string, 0, len(server.Server.Receivers))
	for methodName := range server.Server.Receivers {
		methods = append(methods, methodName)
	}

	// Add built-in methods
	methods = append(methods, "get_state", "set_state")

	return methods
}

// applyMiddleware applies all middleware to the handler
func (h *UnifiedHTTPServer) applyMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied is executed first)
	handler = h.headersMiddleware(handler)
	handler = h.loggingMiddleware(handler)

	if h.config.EnableCORS {
		handler = h.corsMiddleware(handler)
	}

	return handler
}

// corsMiddleware adds CORS headers
func (h *UnifiedHTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (h *UnifiedHTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// headersMiddleware adds custom headers
func (h *UnifiedHTTPServer) headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, value := range h.config.Headers {
			w.Header().Set(key, value)
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterServer registers a new server with the message router
func (h *UnifiedHTTPServer) RegisterServer(name string, server *Value) {
	h.messageRouter.RegisterServer(name, server)
}

// CallServer calls any server (local or remote) through the unified router
func (h *UnifiedHTTPServer) CallServer(nodeID, serverName, method string, args []*Value, timeout time.Duration) (*Value, error) {
	return h.messageRouter.CallServer(nodeID, serverName, method, args, timeout)
}

// AddPeer adds a P2P peer node
func (h *UnifiedHTTPServer) AddPeer(nodeID, address string) {
	h.messageRouter.AddP2PNode(nodeID, address)
}

// RemovePeer removes a P2P peer node
func (h *UnifiedHTTPServer) RemovePeer(nodeID string) {
	h.messageRouter.RemoveP2PNode(nodeID)
}

// GetNodeID returns the node ID
func (h *UnifiedHTTPServer) GetNodeID() string {
	return h.nodeID
}

// GetMessageRouter returns the message router (for advanced usage)
func (h *UnifiedHTTPServer) GetMessageRouter() *MessageRouter {
	return h.messageRouter
}

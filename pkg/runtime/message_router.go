package runtime

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// MessageRouter is the central actor that routes all messages to servers
// regardless of transport mechanism (HTTP, WebSocket, local calls)
type MessageRouter struct {
	// Core routing
	servers    map[string]*Value // Local servers
	serversMux sync.RWMutex      // Temporary until we make this fully actor-based

	// Message channels
	routeChannel    chan *RouteRequest
	registerChannel chan *RegisterRequest
	running         bool

	// P2P integration
	p2pNodes map[string]*P2PNode
	nodesMux sync.RWMutex // Temporary until we make this fully actor-based

	// Response tracking
	pendingResponses map[string]chan *RouteResponse
	responsesMux     sync.RWMutex
}

// RouteRequest represents a unified request to call any server method
type RouteRequest struct {
	// Request identification
	ID   string `json:"id"`
	From string `json:"from"` // Source node/client ID

	// Target information
	NodeID     string `json:"node_id"`     // Target node (empty for local)
	ServerName string `json:"server_name"` // Target server
	Method     string `json:"method"`      // Method to call

	// Request data
	Args    []*Value      `json:"args"`
	Timeout time.Duration `json:"timeout"`

	// Response channel
	ResponseChan chan *RouteResponse `json:"-"`
}

// RouteResponse represents a unified response from any server call
type RouteResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Result  *Value `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
	NodeID  string `json:"node_id"`
}

// RegisterRequest represents a request to register a server
type RegisterRequest struct {
	ServerName string `json:"server_name"`
	Server     *Value `json:"server"`
	NodeID     string `json:"node_id"`
}

// P2PNode represents a remote node connection
type P2PNode struct {
	NodeID      string
	Address     string
	Connected   bool
	LastSeen    time.Time
	SendChannel chan *RouteRequest
}

// NewMessageRouter creates a new message router actor
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		servers:          make(map[string]*Value),
		routeChannel:     make(chan *RouteRequest, 1000),
		registerChannel:  make(chan *RegisterRequest, 100),
		p2pNodes:         make(map[string]*P2PNode),
		pendingResponses: make(map[string]chan *RouteResponse),
		running:          false,
	}
}

// Start starts the message router actor
func (mr *MessageRouter) Start() {
	if mr.running {
		return
	}

	mr.running = true
	go mr.routerLoop()
	log.Printf("Message router started")
}

// Stop stops the message router actor
func (mr *MessageRouter) Stop() {
	if !mr.running {
		return
	}

	mr.running = false
	close(mr.routeChannel)
	close(mr.registerChannel)
	log.Printf("Message router stopped")
}

// RegisterServer registers a server with the router (actor-safe)
func (mr *MessageRouter) RegisterServer(serverName string, server *Value) {
	if !mr.running {
		// Fallback for non-actor mode
		mr.serversMux.Lock()
		mr.servers[serverName] = server
		mr.serversMux.Unlock()
		return
	}

	mr.registerChannel <- &RegisterRequest{
		ServerName: serverName,
		Server:     server,
		NodeID:     "local",
	}
}

// RouteMessage routes a message to the appropriate server (local or remote)
func (mr *MessageRouter) RouteMessage(req *RouteRequest) *RouteResponse {
	if !mr.running {
		// Fallback for non-actor mode
		return mr.routeMessageDirect(req)
	}

	// Create response channel
	responseChan := make(chan *RouteResponse, 1)
	req.ResponseChan = responseChan

	// Send to router actor
	mr.routeChannel <- req

	// Wait for response with timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	select {
	case response := <-responseChan:
		return response
	case <-time.After(timeout):
		return &RouteResponse{
			ID:      req.ID,
			Success: false,
			Error:   "timeout waiting for response",
			NodeID:  "router",
		}
	}
}

// GetServer gets a server by name (thread-safe)
func (mr *MessageRouter) GetServer(serverName string) (*Value, bool) {
	mr.serversMux.RLock()
	defer mr.serversMux.RUnlock()

	server, exists := mr.servers[serverName]
	return server, exists
}

// GetAllServers returns all registered servers (thread-safe)
func (mr *MessageRouter) GetAllServers() map[string]*Value {
	mr.serversMux.RLock()
	defer mr.serversMux.RUnlock()

	result := make(map[string]*Value)
	for name, server := range mr.servers {
		result[name] = server
	}
	return result
}

// AddP2PNode adds a P2P node connection
func (mr *MessageRouter) AddP2PNode(nodeID, address string) {
	mr.nodesMux.Lock()
	defer mr.nodesMux.Unlock()

	mr.p2pNodes[nodeID] = &P2PNode{
		NodeID:      nodeID,
		Address:     address,
		Connected:   true,
		LastSeen:    time.Now(),
		SendChannel: make(chan *RouteRequest, 100),
	}

	log.Printf("Added P2P node: %s at %s", nodeID, address)
}

// RemoveP2PNode removes a P2P node connection
func (mr *MessageRouter) RemoveP2PNode(nodeID string) {
	mr.nodesMux.Lock()
	defer mr.nodesMux.Unlock()

	if node, exists := mr.p2pNodes[nodeID]; exists {
		close(node.SendChannel)
		delete(mr.p2pNodes, nodeID)
		log.Printf("Removed P2P node: %s", nodeID)
	}
}

// routerLoop is the main actor loop that processes all routing requests
func (mr *MessageRouter) routerLoop() {
	for mr.running {
		select {
		case req := <-mr.routeChannel:
			mr.handleRouteRequest(req)

		case regReq := <-mr.registerChannel:
			mr.handleRegisterRequest(regReq)
		}
	}
}

// handleRouteRequest processes a route request
func (mr *MessageRouter) handleRouteRequest(req *RouteRequest) {
	if req == nil {
		log.Printf("Warning: Received nil route request")
		return
	}

	var response *RouteResponse

	// Determine if this is a local or remote call
	if req.NodeID == "" || req.NodeID == "local" {
		// Local server call
		response = mr.routeToLocalServer(req)
	} else {
		// Remote server call
		response = mr.routeToRemoteServer(req)
	}

	// Send response back
	if req.ResponseChan != nil {
		select {
		case req.ResponseChan <- response:
		case <-time.After(1 * time.Second):
			log.Printf("Warning: Failed to send response for request %s", req.ID)
		}
	}
}

// handleRegisterRequest processes a server registration
func (mr *MessageRouter) handleRegisterRequest(regReq *RegisterRequest) {
	mr.servers[regReq.ServerName] = regReq.Server
	log.Printf("Registered server: %s", regReq.ServerName)
}

// routeToLocalServer routes a request to a local server
func (mr *MessageRouter) routeToLocalServer(req *RouteRequest) *RouteResponse {
	// Get the server
	server, exists := mr.servers[req.ServerName]
	if !exists {
		return &RouteResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("Server '%s' not found", req.ServerName),
			NodeID:  "local",
		}
	}

	// Call the server method
	result, err := server.Server.SendMessage(req.Method, req.Args, true)
	if err != nil {
		return &RouteResponse{
			ID:      req.ID,
			Success: false,
			Error:   err.Error(),
			NodeID:  "local",
		}
	}

	return &RouteResponse{
		ID:      req.ID,
		Success: true,
		Result:  result,
		NodeID:  "local",
	}
}

// routeToRemoteServer routes a request to a remote server
func (mr *MessageRouter) routeToRemoteServer(req *RouteRequest) *RouteResponse {
	mr.nodesMux.RLock()
	node, exists := mr.p2pNodes[req.NodeID]
	mr.nodesMux.RUnlock()

	if !exists || !node.Connected {
		return &RouteResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("Node '%s' not found or not connected", req.NodeID),
			NodeID:  "router",
		}
	}

	// TODO: Implement actual P2P message sending
	// For now, return an error indicating P2P is not fully implemented
	return &RouteResponse{
		ID:      req.ID,
		Success: false,
		Error:   "P2P routing not yet implemented in unified router",
		NodeID:  "router",
	}
}

// routeMessageDirect handles direct routing when actor is not running
func (mr *MessageRouter) routeMessageDirect(req *RouteRequest) *RouteResponse {
	if req.NodeID != "" && req.NodeID != "local" {
		return &RouteResponse{
			ID:      req.ID,
			Success: false,
			Error:   "Remote routing not available in direct mode",
			NodeID:  "router",
		}
	}

	return mr.routeToLocalServer(req)
}

// CallServer is a convenience method for calling any server (local or remote)
func (mr *MessageRouter) CallServer(nodeID, serverName, method string, args []*Value, timeout time.Duration) (*Value, error) {
	req := &RouteRequest{
		ID:         generateRequestID(),
		From:       "local",
		NodeID:     nodeID,
		ServerName: serverName,
		Method:     method,
		Args:       args,
		Timeout:    timeout,
	}

	response := mr.RouteMessage(req)
	if !response.Success {
		return nil, fmt.Errorf(response.Error)
	}

	return response.Result, nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

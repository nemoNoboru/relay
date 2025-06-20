package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketP2P handles WebSocket-based peer-to-peer communication
type WebSocketP2P struct {
	registry         *ExposableServerRegistry
	nodeID           string
	connections      map[string]*PeerConnection
	connMutex        sync.RWMutex
	upgrader         websocket.Upgrader
	messageQueue     chan *P2PMessage
	handlers         map[string]P2PMessageHandler
	handlerMutex     sync.RWMutex
	running          bool
	responseChannels map[string]chan *RemoteServerResponse // For tracking responses
	responseMutex    sync.RWMutex
}

// PeerConnection represents a WebSocket connection to a peer node
type PeerConnection struct {
	NodeID    string
	Address   string
	Conn      *websocket.Conn
	LastSeen  time.Time
	IsHealthy bool
	SendChan  chan *P2PMessage
	CloseChan chan bool
	mutex     sync.RWMutex
}

// P2PMessage represents a message sent between peer nodes
type P2PMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Route     []string               `json:"route,omitempty"`    // For multistep routing
	TTL       int                    `json:"ttl,omitempty"`      // Time to live
	ReplyTo   string                 `json:"reply_to,omitempty"` // For request/response
}

// P2PMessageHandler handles different types of P2P messages
type P2PMessageHandler func(conn *PeerConnection, msg *P2PMessage) error

// RemoteServerCall represents a call to a remote server
type RemoteServerCall struct {
	ServerName string        `json:"server_name"`
	Method     string        `json:"method"`
	Args       []*Value      `json:"args"`
	Timeout    time.Duration `json:"timeout"`
}

// RemoteServerResponse represents a response from a remote server
type RemoteServerResponse struct {
	Success bool   `json:"success"`
	Result  *Value `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
	NodeID  string `json:"node_id"`
}

// NewWebSocketP2P creates a new WebSocket P2P communication system
func NewWebSocketP2P(registry *ExposableServerRegistry, nodeID string) *WebSocketP2P {
	p2p := &WebSocketP2P{
		registry:    registry,
		nodeID:      nodeID,
		connections: make(map[string]*PeerConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		messageQueue:     make(chan *P2PMessage, 1000),
		handlers:         make(map[string]P2PMessageHandler),
		responseChannels: make(map[string]chan *RemoteServerResponse),
	}

	// Register default message handlers
	p2p.RegisterHandler("ping", p2p.handlePing)
	p2p.RegisterHandler("pong", p2p.handlePong)
	p2p.RegisterHandler("server_call", p2p.handleServerCall)
	p2p.RegisterHandler("server_response", p2p.handleServerResponse)
	p2p.RegisterHandler("registry_sync", p2p.handleRegistrySync)
	p2p.RegisterHandler("route_message", p2p.handleRouteMessage)

	return p2p
}

// Start starts the WebSocket P2P system
func (p *WebSocketP2P) Start() {
	if p.running {
		return
	}

	p.running = true

	// Start message processing goroutine
	go p.processMessages()

	// Start periodic health checks
	go p.healthCheckLoop()

	log.Printf("WebSocket P2P system started for node %s", p.nodeID)
}

// Stop stops the WebSocket P2P system
func (p *WebSocketP2P) Stop() {
	if !p.running {
		return
	}

	p.running = false

	// Close all connections
	p.connMutex.Lock()
	for _, conn := range p.connections {
		conn.Close()
	}
	p.connections = make(map[string]*PeerConnection)
	p.connMutex.Unlock()

	close(p.messageQueue)

	log.Printf("WebSocket P2P system stopped for node %s", p.nodeID)
}

// RegisterHandler registers a message handler for a specific message type
func (p *WebSocketP2P) RegisterHandler(msgType string, handler P2PMessageHandler) {
	p.handlerMutex.Lock()
	defer p.handlerMutex.Unlock()
	p.handlers[msgType] = handler
}

// SetupWebSocketEndpoint sets up the WebSocket endpoint for peer connections
func (p *WebSocketP2P) SetupWebSocketEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/ws/p2p", p.handleWebSocketConnection)
}

// handleWebSocketConnection handles incoming WebSocket connections
func (p *WebSocketP2P) handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Get peer node ID from query parameter
	peerNodeID := r.URL.Query().Get("node_id")
	if peerNodeID == "" {
		log.Printf("WebSocket connection rejected: no node_id provided")
		conn.Close()
		return
	}

	peerConn := &PeerConnection{
		NodeID:    peerNodeID,
		Address:   r.RemoteAddr,
		Conn:      conn,
		LastSeen:  time.Now(),
		IsHealthy: true,
		SendChan:  make(chan *P2PMessage, 100),
		CloseChan: make(chan bool, 1),
	}

	// Register connection
	p.connMutex.Lock()
	p.connections[peerNodeID] = peerConn
	p.connMutex.Unlock()

	log.Printf("New WebSocket P2P connection from node %s", peerNodeID)

	// Start connection handlers
	go p.handlePeerConnection(peerConn)
	go p.handlePeerSender(peerConn)
}

// ConnectToPeer connects to a peer node via WebSocket
func (p *WebSocketP2P) ConnectToPeer(nodeID, address string) error {
	// Check if already connected
	p.connMutex.RLock()
	if _, exists := p.connections[nodeID]; exists {
		p.connMutex.RUnlock()
		return fmt.Errorf("already connected to peer %s", nodeID)
	}
	p.connMutex.RUnlock()

	// Create WebSocket connection
	url := fmt.Sprintf("ws://%s/ws/p2p?node_id=%s", address, p.nodeID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %v", nodeID, err)
	}

	peerConn := &PeerConnection{
		NodeID:    nodeID,
		Address:   address,
		Conn:      conn,
		LastSeen:  time.Now(),
		IsHealthy: true,
		SendChan:  make(chan *P2PMessage, 100),
		CloseChan: make(chan bool, 1),
	}

	// Register connection
	p.connMutex.Lock()
	p.connections[nodeID] = peerConn
	p.connMutex.Unlock()

	log.Printf("Connected to peer node %s at %s", nodeID, address)

	// Start connection handlers
	go p.handlePeerConnection(peerConn)
	go p.handlePeerSender(peerConn)

	return nil
}

// SendMessage sends a message to a specific peer
func (p *WebSocketP2P) SendMessage(to string, msgType string, data map[string]interface{}) error {
	msg := &P2PMessage{
		Type:      msgType,
		ID:        generateMessageID(),
		From:      p.nodeID,
		To:        to,
		Data:      data,
		Timestamp: time.Now(),
		TTL:       10, // Default TTL
	}

	return p.routeMessage(msg)
}

// BroadcastMessage broadcasts a message to all connected peers
func (p *WebSocketP2P) BroadcastMessage(msgType string, data map[string]interface{}) {
	p.connMutex.RLock()
	defer p.connMutex.RUnlock()

	for nodeID := range p.connections {
		go p.SendMessage(nodeID, msgType, data)
	}
}

// CallRemoteServer calls a method on a server running on a remote node
func (p *WebSocketP2P) CallRemoteServer(nodeID, serverName, method string, args []*Value, timeout time.Duration) (*Value, error) {
	// Create remote server call
	call := RemoteServerCall{
		ServerName: serverName,
		Method:     method,
		Args:       args,
		Timeout:    timeout,
	}

	// Convert to JSON-serializable format
	data := map[string]interface{}{
		"server_name": call.ServerName,
		"method":      call.Method,
		"args":        convertArgsToJSON(call.Args),
		"timeout":     call.Timeout.Seconds(),
	}

	// Create response channel and tracking
	responseID := generateMessageID()
	responseChan := make(chan *RemoteServerResponse, 1)

	// Store response channel for tracking
	p.responseMutex.Lock()
	p.responseChannels[responseID] = responseChan
	p.responseMutex.Unlock()

	// Cleanup response channel after completion
	defer func() {
		p.responseMutex.Lock()
		delete(p.responseChannels, responseID)
		p.responseMutex.Unlock()
		close(responseChan)
	}()

	// Send message
	err := p.SendMessage(nodeID, "server_call", map[string]interface{}{
		"call":        data,
		"response_id": responseID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send remote server call: %v", err)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		if !response.Success {
			return nil, fmt.Errorf("remote server call failed: %s", response.Error)
		}
		return response.Result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("remote server call timed out")
	}
}

// routeMessage routes a message to its destination, potentially through multiple hops
func (p *WebSocketP2P) routeMessage(msg *P2PMessage) error {
	// Decrement TTL
	msg.TTL--
	if msg.TTL <= 0 {
		return fmt.Errorf("message TTL exceeded")
	}

	// Add this node to route
	msg.Route = append(msg.Route, p.nodeID)

	// Check if destination is directly connected
	p.connMutex.RLock()
	if conn, exists := p.connections[msg.To]; exists {
		p.connMutex.RUnlock()

		select {
		case conn.SendChan <- msg:
			return nil
		case <-time.After(1 * time.Second):
			return fmt.Errorf("failed to send message to %s: send channel full", msg.To)
		}
	}
	p.connMutex.RUnlock()

	// Destination not directly connected, try multistep routing
	return p.routeViaNeighbors(msg)
}

// routeViaNeighbors attempts to route a message via neighboring nodes
func (p *WebSocketP2P) routeViaNeighbors(msg *P2PMessage) error {
	p.connMutex.RLock()
	defer p.connMutex.RUnlock()

	// Check if we've already been through this node (avoid loops)
	for _, nodeID := range msg.Route {
		if nodeID == p.nodeID {
			return fmt.Errorf("routing loop detected")
		}
	}

	// Try to route through each connected peer
	for nodeID, conn := range p.connections {
		// Don't route back through nodes already in the route
		inRoute := false
		for _, routeNode := range msg.Route {
			if routeNode == nodeID {
				inRoute = true
				break
			}
		}

		if !inRoute && conn.IsHealthy {
			// Wrap in route message
			routeMsg := &P2PMessage{
				Type:      "route_message",
				ID:        generateMessageID(),
				From:      p.nodeID,
				To:        nodeID,
				Data:      map[string]interface{}{"original_message": msg},
				Timestamp: time.Now(),
				TTL:       msg.TTL,
			}

			select {
			case conn.SendChan <- routeMsg:
				return nil
			case <-time.After(1 * time.Second):
				continue // Try next peer
			}
		}
	}

	return fmt.Errorf("no route found to destination %s", msg.To)
}

// processMessages processes incoming messages from the message queue
func (p *WebSocketP2P) processMessages() {
	for msg := range p.messageQueue {
		p.handlerMutex.RLock()
		handler, exists := p.handlers[msg.Type]
		p.handlerMutex.RUnlock()

		if exists {
			// Find the connection this message came from
			p.connMutex.RLock()
			conn := p.connections[msg.From]
			p.connMutex.RUnlock()

			if conn != nil {
				if err := handler(conn, msg); err != nil {
					log.Printf("Error handling message type %s from %s: %v", msg.Type, msg.From, err)
				}
			}
		} else {
			log.Printf("No handler for message type: %s", msg.Type)
		}
	}
}

// handlePeerConnection handles incoming messages from a peer connection
func (p *WebSocketP2P) handlePeerConnection(conn *PeerConnection) {
	defer conn.Close()

	for {
		var msg P2PMessage
		err := conn.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message from peer %s: %v", conn.NodeID, err)
			break
		}

		conn.mutex.Lock()
		conn.LastSeen = time.Now()
		conn.mutex.Unlock()

		// Add to message queue for processing
		select {
		case p.messageQueue <- &msg:
		case <-time.After(1 * time.Second):
			log.Printf("Message queue full, dropping message from %s", conn.NodeID)
		}
	}

	// Remove connection
	p.connMutex.Lock()
	delete(p.connections, conn.NodeID)
	p.connMutex.Unlock()

	log.Printf("Peer connection closed: %s", conn.NodeID)
}

// handlePeerSender handles outgoing messages to a peer connection
func (p *WebSocketP2P) handlePeerSender(conn *PeerConnection) {
	for {
		select {
		case msg := <-conn.SendChan:
			err := conn.Conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error sending message to peer %s: %v", conn.NodeID, err)
				return
			}
		case <-conn.CloseChan:
			return
		}
	}
}

// Close closes a peer connection
func (conn *PeerConnection) Close() {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	if conn.Conn != nil {
		conn.Conn.Close()
		conn.Conn = nil
	}

	select {
	case conn.CloseChan <- true:
	default:
	}

	close(conn.SendChan)
}

// healthCheckLoop performs periodic health checks on peer connections
func (p *WebSocketP2P) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !p.running {
			return
		}

		p.connMutex.RLock()
		for nodeID, conn := range p.connections {
			conn.mutex.RLock()
			timeSinceLastSeen := time.Since(conn.LastSeen)
			conn.mutex.RUnlock()

			if timeSinceLastSeen > 60*time.Second {
				log.Printf("Peer %s appears unhealthy, last seen %v ago", nodeID, timeSinceLastSeen)
				conn.mutex.Lock()
				conn.IsHealthy = false
				conn.mutex.Unlock()
			} else {
				// Send ping to check health
				p.SendMessage(nodeID, "ping", map[string]interface{}{
					"timestamp": time.Now().Unix(),
				})
			}
		}
		p.connMutex.RUnlock()
	}
}

// Message handlers

func (p *WebSocketP2P) handlePing(conn *PeerConnection, msg *P2PMessage) error {
	// Respond with pong
	return p.SendMessage(msg.From, "pong", map[string]interface{}{
		"timestamp": msg.Data["timestamp"],
		"pong_time": time.Now().Unix(),
	})
}

func (p *WebSocketP2P) handlePong(conn *PeerConnection, msg *P2PMessage) error {
	// Update connection health
	conn.mutex.Lock()
	conn.IsHealthy = true
	conn.LastSeen = time.Now()
	conn.mutex.Unlock()

	return nil
}

func (p *WebSocketP2P) handleServerCall(conn *PeerConnection, msg *P2PMessage) error {
	// Extract server call data
	callData, ok := msg.Data["call"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid server call data")
	}

	serverName, ok := callData["server_name"].(string)
	if !ok {
		return fmt.Errorf("invalid server name")
	}

	method, ok := callData["method"].(string)
	if !ok {
		return fmt.Errorf("invalid method name")
	}

	// Get server from local registry
	server, exists := p.registry.GetServer(serverName)
	if !exists {
		// Send error response
		return p.SendMessage(msg.From, "server_response", map[string]interface{}{
			"success":     false,
			"error":       fmt.Sprintf("server %s not found", serverName),
			"response_id": msg.Data["response_id"],
		})
	}

	// Convert arguments
	args := convertJSONToArgs(callData["args"])

	// Call the server method
	result, err := server.Server.SendMessage(method, args, true)

	// Send response
	responseData := map[string]interface{}{
		"response_id": msg.Data["response_id"],
		"node_id":     p.nodeID,
	}

	if err != nil {
		responseData["success"] = false
		responseData["error"] = err.Error()
	} else {
		responseData["success"] = true
		responseData["result"] = convertValueToJSON(result)
	}

	return p.SendMessage(msg.From, "server_response", responseData)
}

func (p *WebSocketP2P) handleServerResponse(conn *PeerConnection, msg *P2PMessage) error {
	// Extract response ID
	responseID, ok := msg.Data["response_id"].(string)
	if !ok {
		log.Printf("Server response missing response_id from %s", msg.From)
		return nil
	}

	// Find the response channel
	p.responseMutex.RLock()
	responseChan, exists := p.responseChannels[responseID]
	p.responseMutex.RUnlock()

	if !exists {
		log.Printf("No response channel found for response_id %s from %s", responseID, msg.From)
		return nil
	}

	// Convert response data
	success, _ := msg.Data["success"].(bool)
	errorMsg, _ := msg.Data["error"].(string)
	nodeID, _ := msg.Data["node_id"].(string)

	response := &RemoteServerResponse{
		Success: success,
		Error:   errorMsg,
		NodeID:  nodeID,
	}

	// Convert result if present
	if resultData, exists := msg.Data["result"]; exists && success {
		response.Result = convertJSONToValue(resultData)
	}

	// Send response to waiting channel
	select {
	case responseChan <- response:
		log.Printf("Delivered server response for %s from %s", responseID, msg.From)
	case <-time.After(1 * time.Second):
		log.Printf("Failed to deliver server response for %s from %s: channel full", responseID, msg.From)
	}

	return nil
}

func (p *WebSocketP2P) handleRegistrySync(conn *PeerConnection, msg *P2PMessage) error {
	// Handle registry synchronization
	// This could be used to share server registry information between peers
	log.Printf("Registry sync from %s: %v", msg.From, msg.Data)
	return nil
}

func (p *WebSocketP2P) handleRouteMessage(conn *PeerConnection, msg *P2PMessage) error {
	// Extract original message
	originalData, ok := msg.Data["original_message"]
	if !ok {
		return fmt.Errorf("invalid route message: no original_message")
	}

	// Convert back to P2PMessage
	originalJSON, err := json.Marshal(originalData)
	if err != nil {
		return fmt.Errorf("failed to marshal original message: %v", err)
	}

	var originalMsg P2PMessage
	if err := json.Unmarshal(originalJSON, &originalMsg); err != nil {
		return fmt.Errorf("failed to unmarshal original message: %v", err)
	}

	// Route the original message
	return p.routeMessage(&originalMsg)
}

// Utility functions

func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func convertArgsToJSON(args []*Value) []interface{} {
	result := make([]interface{}, len(args))
	for i, arg := range args {
		result[i] = convertValueToJSON(arg)
	}
	return result
}

func convertJSONToArgs(data interface{}) []*Value {
	if argsArray, ok := data.([]interface{}); ok {
		result := make([]*Value, len(argsArray))
		for i, arg := range argsArray {
			result[i] = convertJSONToValue(arg)
		}
		return result
	}
	return []*Value{}
}

func convertValueToJSON(value *Value) interface{} {
	switch value.Type {
	case ValueTypeNil:
		return nil
	case ValueTypeBool:
		return value.Bool
	case ValueTypeNumber:
		return value.Number
	case ValueTypeString:
		return value.Str
	case ValueTypeArray:
		result := make([]interface{}, len(value.Array))
		for i, elem := range value.Array {
			result[i] = convertValueToJSON(elem)
		}
		return result
	case ValueTypeObject:
		result := make(map[string]interface{})
		for key, val := range value.Object {
			result[key] = convertValueToJSON(val)
		}
		return result
	default:
		return value.String()
	}
}

func convertJSONToValue(data interface{}) *Value {
	switch v := data.(type) {
	case nil:
		return NewNil()
	case bool:
		return NewBool(v)
	case float64:
		return NewNumber(v)
	case string:
		return NewString(v)
	case []interface{}:
		elements := make([]*Value, len(v))
		for i, elem := range v {
			elements[i] = convertJSONToValue(elem)
		}
		return NewArray(elements)
	case map[string]interface{}:
		fields := make(map[string]*Value)
		for key, val := range v {
			fields[key] = convertJSONToValue(val)
		}
		return NewObject(fields)
	default:
		return NewString(fmt.Sprintf("%v", v))
	}
}

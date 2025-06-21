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

// NodeType defines the role of a node in the federation architecture.
type NodeType string

const (
	// NodeTypeMain indicates a publicly accessible server node.
	NodeTypeMain NodeType = "main"
	// NodeTypeHome indicates a client node, typically behind a NAT.
	NodeTypeHome NodeType = "home"
)

// FederationRouter handles WebSocket-based peer-to-peer communication
type FederationRouter struct {
	nodeType         NodeType
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

// NewFederationRouter creates a new WebSocket P2P communication system
func NewFederationRouter(nodeType NodeType, registry *ExposableServerRegistry, nodeID string) *FederationRouter {
	router := &FederationRouter{
		nodeType:    nodeType,
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
	router.RegisterHandler("ping", router.handlePing)
	router.RegisterHandler("pong", router.handlePong)
	router.RegisterHandler("server_call", router.handleServerCall)
	router.RegisterHandler("server_response", router.handleServerResponse)
	router.RegisterHandler("registry_sync", router.handleRegistrySync)
	router.RegisterHandler("route_message", router.handleRouteMessage)

	return router
}

// Start starts the WebSocket P2P system
func (fr *FederationRouter) Start() {
	if fr.running {
		return
	}

	fr.running = true

	// Start message processing goroutine
	go fr.processMessages()

	// Start periodic health checks
	go fr.healthCheckLoop()

	log.Printf("WebSocket P2P system started for node %s", fr.nodeID)
}

// Stop stops the WebSocket P2P system
func (fr *FederationRouter) Stop() {
	if !fr.running {
		return
	}

	fr.running = false

	// Close all connections
	fr.connMutex.Lock()
	for _, conn := range fr.connections {
		conn.Close()
	}
	fr.connections = make(map[string]*PeerConnection)
	fr.connMutex.Unlock()

	close(fr.messageQueue)

	log.Printf("WebSocket P2P system stopped for node %s", fr.nodeID)
}

// RegisterHandler registers a message handler for a specific message type
func (fr *FederationRouter) RegisterHandler(msgType string, handler P2PMessageHandler) {
	fr.handlerMutex.Lock()
	defer fr.handlerMutex.Unlock()
	fr.handlers[msgType] = handler
}

// SetupWebSocketEndpoint sets up the WebSocket endpoint for peer connections
func (fr *FederationRouter) SetupWebSocketEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/ws/p2p", fr.handleWebSocketConnection)
}

// handleWebSocketConnection handles incoming WebSocket connections
func (fr *FederationRouter) handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := fr.upgrader.Upgrade(w, r, nil)
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
	fr.connMutex.Lock()
	fr.connections[peerNodeID] = peerConn
	fr.connMutex.Unlock()

	log.Printf("New WebSocket P2P connection from node %s", peerNodeID)

	// Start connection handlers
	go fr.handlePeerConnection(peerConn)
	go fr.handlePeerSender(peerConn)
}

// ConnectToPeer connects to a peer node via WebSocket
func (fr *FederationRouter) ConnectToPeer(nodeID, address string) error {
	// Check if already connected
	fr.connMutex.RLock()
	if _, exists := fr.connections[nodeID]; exists {
		fr.connMutex.RUnlock()
		return fmt.Errorf("already connected to peer %s", nodeID)
	}
	fr.connMutex.RUnlock()

	// Create WebSocket connection
	url := fmt.Sprintf("ws://%s/ws/p2p?node_id=%s", address, fr.nodeID)
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
	fr.connMutex.Lock()
	fr.connections[nodeID] = peerConn
	fr.connMutex.Unlock()

	log.Printf("Connected to peer node %s at %s", nodeID, address)

	// Start connection handlers
	go fr.handlePeerConnection(peerConn)
	go fr.handlePeerSender(peerConn)

	return nil
}

// SendMessage sends a message to a specific peer
func (fr *FederationRouter) SendMessage(to string, msgType string, data map[string]interface{}) error {
	msg := &P2PMessage{
		Type:      msgType,
		ID:        generateMessageID(),
		From:      fr.nodeID,
		To:        to,
		Data:      data,
		Timestamp: time.Now(),
		TTL:       10, // Default TTL
	}

	return fr.routeMessage(msg)
}

// BroadcastMessage broadcasts a message to all connected peers
func (fr *FederationRouter) BroadcastMessage(msgType string, data map[string]interface{}) {
	fr.connMutex.RLock()
	defer fr.connMutex.RUnlock()

	for nodeID := range fr.connections {
		go fr.SendMessage(nodeID, msgType, data)
	}
}

// CallRemoteServer invokes a method on a server running on a remote peer
func (fr *FederationRouter) CallRemoteServer(nodeID, serverName, method string, args []*Value, timeout time.Duration) (*Value, error) {
	callData, err := json.Marshal(RemoteServerCall{
		ServerName: serverName,
		Method:     method,
		Args:       args,
		Timeout:    timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal remote call: %v", err)
	}

	// Create a response channel
	responseChan := make(chan *RemoteServerResponse, 1)
	requestID := generateMessageID()

	fr.responseMutex.Lock()
	fr.responseChannels[requestID] = responseChan
	fr.responseMutex.Unlock()

	defer func() {
		fr.responseMutex.Lock()
		delete(fr.responseChannels, requestID)
		fr.responseMutex.Unlock()
		close(responseChan)
	}()

	// Send the server call message
	msg := &P2PMessage{
		Type:    "server_call",
		ID:      generateMessageID(),
		From:    fr.nodeID,
		To:      nodeID,
		ReplyTo: requestID,
		Data: map[string]interface{}{
			"call": string(callData),
		},
		Timestamp: time.Now(),
		TTL:       10,
	}

	if err := fr.routeMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to route server call: %v", err)
	}

	// Wait for the response or timeout
	select {
	case response := <-responseChan:
		if !response.Success {
			return nil, fmt.Errorf("remote server call failed: %s", response.Error)
		}
		return response.Result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("remote server call timed out after %v", timeout)
	}
}

// routeMessage routes a message to the correct peer
func (fr *FederationRouter) routeMessage(msg *P2PMessage) error {
	// Add self to route
	msg.Route = append(msg.Route, fr.nodeID)

	// Check if the destination is a direct peer
	fr.connMutex.RLock()
	conn, directPeer := fr.connections[msg.To]
	fr.connMutex.RUnlock()

	if directPeer {
		select {
		case conn.SendChan <- msg:
			return nil
		case <-time.After(2 * time.Second):
			return fmt.Errorf("send channel for peer %s is full or blocked", msg.To)
		}
	}

	// If not a direct peer, try routing via neighbors
	return fr.routeViaNeighbors(msg)
}

// routeViaNeighbors broadcasts the message to all neighbors except the sender and those already in the route
func (fr *FederationRouter) routeViaNeighbors(msg *P2PMessage) error {
	fr.connMutex.RLock()
	defer fr.connMutex.RUnlock()

	if len(fr.connections) == 0 {
		return fmt.Errorf("no peers connected to route message to %s", msg.To)
	}

	routed := false
	for nodeID, conn := range fr.connections {
		// Avoid sending back to the original sender or along a path already taken
		if nodeID == msg.From || contains(msg.Route, nodeID) {
			continue
		}

		// Decrement TTL
		msg.TTL--
		if msg.TTL <= 0 {
			log.Printf("Message TTL expired, dropping message for %s", msg.To)
			return fmt.Errorf("message TTL expired")
		}

		cloneMsg := *msg
		cloneMsg.Route = append([]string{}, msg.Route...) // Create a copy of the route

		select {
		case conn.SendChan <- &cloneMsg:
			routed = true
		case <-time.After(1 * time.Second):
			log.Printf("Failed to send routed message to %s: send channel full", nodeID)
		}
	}

	if !routed {
		return fmt.Errorf("failed to route message to %s: no available peers", msg.To)
	}

	return nil
}

// processMessages processes incoming messages from the message queue
func (fr *FederationRouter) processMessages() {
	for msg := range fr.messageQueue {
		fr.handlerMutex.RLock()
		handler, ok := fr.handlers[msg.Type]
		fr.handlerMutex.RUnlock()

		if ok {
			// Find the connection associated with the 'from' field
			fr.connMutex.RLock()
			conn, connOk := fr.connections[msg.From]
			fr.connMutex.RUnlock()

			if connOk {
				if err := handler(conn, msg); err != nil {
					log.Printf("Error handling message type %s from %s: %v", msg.Type, msg.From, err)
				}
			} else {
				log.Printf("No active connection found for message source: %s. Message dropped.", msg.From)
			}
		} else {
			log.Printf("No handler registered for message type: %s", msg.Type)
		}
	}
}

// handlePeerConnection reads messages from a peer connection and adds them to the queue
func (fr *FederationRouter) handlePeerConnection(conn *PeerConnection) {
	defer func() {
		conn.Close()
		fr.connMutex.Lock()
		delete(fr.connections, conn.NodeID)
		fr.connMutex.Unlock()
		log.Printf("P2P connection with node %s closed", conn.NodeID)
	}()

	for {
		_, message, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading from peer %s: %v", conn.NodeID, err)
			}
			break
		}

		var msg P2PMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message from %s: %v", conn.NodeID, err)
			continue
		}

		// Update peer's last seen time on any message
		conn.mutex.Lock()
		conn.LastSeen = time.Now()
		conn.mutex.Unlock()

		fr.messageQueue <- &msg
	}
}

// handlePeerSender sends messages from the send channel to the peer
func (fr *FederationRouter) handlePeerSender(conn *PeerConnection) {
	for {
		select {
		case msg := <-conn.SendChan:
			if err := conn.Conn.WriteJSON(msg); err != nil {
				log.Printf("Error writing to peer %s: %v", conn.NodeID, err)
				return // Stop sender on error
			}
		case <-conn.CloseChan:
			return // Stop sender on close signal
		}
	}
}

// Close closes the peer connection
func (conn *PeerConnection) Close() {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	select {
	case conn.CloseChan <- true:
		// Signal sent
	default:
		// Already signaled
	}

	conn.Conn.Close()
}

// healthCheckLoop periodically checks the health of connected peers
func (fr *FederationRouter) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !fr.running {
				return
			}
			fr.connMutex.RLock()
			for _, conn := range fr.connections {
				// Send a ping message
				pingMsg := &P2PMessage{
					Type: "ping",
					From: fr.nodeID,
					To:   conn.NodeID,
				}
				if err := conn.Conn.WriteJSON(pingMsg); err != nil {
					log.Printf("Health check failed for peer %s: %v", conn.NodeID, err)
					conn.mutex.Lock()
					conn.IsHealthy = false
					conn.mutex.Unlock()
				}
			}
			fr.connMutex.RUnlock()
		}
	}
}

// handlePing responds to a ping message with a pong
func (fr *FederationRouter) handlePing(conn *PeerConnection, msg *P2PMessage) error {
	pongMsg := &P2PMessage{
		Type: "pong",
		From: fr.nodeID,
		To:   msg.From,
	}
	conn.SendChan <- pongMsg
	return nil
}

// handlePong updates the health status of a peer
func (fr *FederationRouter) handlePong(conn *PeerConnection, msg *P2PMessage) error {
	conn.mutex.Lock()
	conn.IsHealthy = true
	conn.LastSeen = time.Now()
	conn.mutex.Unlock()
	log.Printf("Received pong from %s, connection is healthy", msg.From)
	return nil
}

// handleServerCall handles a remote server call from a peer
func (fr *FederationRouter) handleServerCall(conn *PeerConnection, msg *P2PMessage) error {
	var callData string
	if data, ok := msg.Data["call"].(string); ok {
		callData = data
	} else {
		return fmt.Errorf("invalid server call data format")
	}

	var call RemoteServerCall
	if err := json.Unmarshal([]byte(callData), &call); err != nil {
		return fmt.Errorf("failed to unmarshal server call: %v", err)
	}

	response := &RemoteServerResponse{NodeID: fr.nodeID}

	server, serverExists := fr.registry.GetServer(call.ServerName)
	if !serverExists {
		response.Success = false
		response.Error = fmt.Sprintf("server '%s' not found", call.ServerName)
	} else {
		result, callErr := server.Server.SendMessage(call.Method, call.Args, true)
		if callErr != nil {
			response.Success = false
			response.Error = callErr.Error()
		} else {
			response.Success = true
			response.Result = result
		}
	}

	// Send the response back
	responseData, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal server response: %v", marshalErr)
	}

	replyMsg := &P2PMessage{
		Type: "server_response",
		ID:   generateMessageID(),
		From: fr.nodeID,
		To:   msg.From,
		Data: map[string]interface{}{
			"response": string(responseData),
		},
		ReplyTo:   msg.ID,
		Timestamp: time.Now(),
	}

	return fr.routeMessage(replyMsg)
}

// handleServerResponse handles a response from a remote server call
func (fr *FederationRouter) handleServerResponse(conn *PeerConnection, msg *P2PMessage) error {
	if msg.ReplyTo == "" {
		return fmt.Errorf("received server response without a reply_to ID")
	}

	fr.responseMutex.RLock()
	responseChan, ok := fr.responseChannels[msg.ReplyTo]
	fr.responseMutex.RUnlock()

	if !ok {
		// It's possible the request timed out and the channel was deleted.
		// Log this for debugging purposes.
		log.Printf("Received response for unknown or timed-out request ID: %s", msg.ReplyTo)
		return nil
	}

	var responseData string
	if data, ok := msg.Data["response"].(string); ok {
		responseData = data
	} else {
		return fmt.Errorf("invalid server response data format")
	}

	var response RemoteServerResponse
	if err := json.Unmarshal([]byte(responseData), &response); err != nil {
		return fmt.Errorf("failed to unmarshal server response: %v", err)
	}

	// Send response to the waiting channel
	select {
	case responseChan <- &response:
		// Response sent successfully
	case <-time.After(1 * time.Second):
		log.Printf("Failed to send response to channel for request ID: %s. Channel may be blocked.", msg.ReplyTo)
	}

	return nil
}

// handleRegistrySync handles requests for registry synchronization
func (fr *FederationRouter) handleRegistrySync(conn *PeerConnection, msg *P2PMessage) error {
	// For now, this is a placeholder. A real implementation would involve
	// more complex logic for merging and updating registries.
	log.Printf("Registry sync request from %s - feature not fully implemented.", msg.From)
	return nil
}

// handleRouteMessage handles messages that need to be routed to other peers
func (fr *FederationRouter) handleRouteMessage(conn *PeerConnection, msg *P2PMessage) error {
	// This handler is for messages that are explicitly of type 'route_message'
	// which contain another message to be routed.
	if routedMsgData, ok := msg.Data["routed_message"]; ok {
		var routedMsg P2PMessage
		// Assuming routedMsgData is a map[string]interface{} that needs to be remarshaled
		jsonBytes, err := json.Marshal(routedMsgData)
		if err != nil {
			return fmt.Errorf("failed to remarshal routed message: %v", err)
		}
		if err := json.Unmarshal(jsonBytes, &routedMsg); err != nil {
			return fmt.Errorf("failed to unmarshal routed message: %v", err)
		}

		// Prevent routing loops
		if contains(routedMsg.Route, fr.nodeID) {
			return fmt.Errorf("routing loop detected for message %s", routedMsg.ID)
		}

		return fr.routeMessage(&routedMsg)
	}
	return fmt.Errorf("invalid route_message format")
}

// generateMessageID generates a new unique message ID
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
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

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

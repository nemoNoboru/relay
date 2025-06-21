package runtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// HTTPServerConfig contains configuration for the HTTP server
type HTTPServerConfig struct {
	Host         string
	Port         int
	EnableCORS   bool
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Headers      map[string]string
	// New fields for P2P functionality
	NodeID            string
	EnableRegistry    bool
	DiscoveryInterval time.Duration
}

// HTTPServer represents the HTTP server that exposes Relay servers via JSON-RPC 2.0
type HTTPServer struct {
	config            *HTTPServerConfig
	evaluator         *Evaluator
	server            *http.Server
	running           bool
	exposableRegistry *ExposableServerRegistry // New field for P2P registry
	websocketP2P      *FederationRouter        // WebSocket P2P communication
}

// DefaultHTTPServerConfig returns default configuration
func DefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              8080,
		EnableCORS:        true,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		Headers:           make(map[string]string),
		NodeID:            generateNodeID(),
		EnableRegistry:    true,
		DiscoveryInterval: 30 * time.Second,
	}
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(evaluator *Evaluator, config *HTTPServerConfig) *HTTPServer {
	if config == nil {
		config = DefaultHTTPServerConfig()
	}

	// Generate node ID if not provided
	if config.NodeID == "" {
		config.NodeID = generateNodeID()
	}

	httpServer := &HTTPServer{
		config:    config,
		evaluator: evaluator,
		running:   false,
	}

	// Create exposable registry if enabled
	if config.EnableRegistry {
		nodeAddress := fmt.Sprintf("%s:%d", config.Host, config.Port)
		baseRegistry := NewEvaluatorServerRegistry(evaluator)
		httpServer.exposableRegistry = NewExposableServerRegistry(baseRegistry, config.NodeID, nodeAddress)

		// Create WebSocket P2P system
		httpServer.websocketP2P = NewFederationRouter(NodeTypeMain, httpServer.exposableRegistry, config.NodeID)
	}

	return httpServer
}

// generateNodeID creates a unique node identifier
func generateNodeID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	if h.running {
		return fmt.Errorf("HTTP server is already running")
	}

	mux := http.NewServeMux()

	// JSON-RPC 2.0 endpoint
	mux.HandleFunc("/rpc", h.handleJSONRPC)

	// Health check endpoint
	mux.HandleFunc("/health", h.handleHealth)

	// Server info endpoint
	mux.HandleFunc("/info", h.handleInfo)

	// Setup registry endpoints if enabled
	if h.config.EnableRegistry && h.exposableRegistry != nil {
		h.exposableRegistry.SetupHTTPEndpoints(mux)

		// Setup WebSocket P2P endpoints
		if h.websocketP2P != nil {
			h.websocketP2P.SetupWebSocketEndpoint(mux)
			h.websocketP2P.Start()
		}

		// Start periodic peer discovery
		if h.config.DiscoveryInterval > 0 {
			h.exposableRegistry.StartPeriodicDiscovery(h.config.DiscoveryInterval)
		}
	}

	// Apply middleware
	handler := h.applyMiddleware(mux)

	h.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", h.config.Host, h.config.Port),
		Handler:      handler,
		ReadTimeout:  h.config.ReadTimeout,
		WriteTimeout: h.config.WriteTimeout,
	}

	h.running = true

	log.Printf("Starting Relay HTTP server on %s:%d", h.config.Host, h.config.Port)
	log.Printf("JSON-RPC 2.0 endpoint: http://%s:%d/rpc", h.config.Host, h.config.Port)

	if h.config.EnableRegistry {
		log.Printf("Server registry: http://%s:%d/registry", h.config.Host, h.config.Port)
		log.Printf("WebSocket P2P endpoint: ws://%s:%d/ws/p2p", h.config.Host, h.config.Port)
		log.Printf("Node ID: %s", h.config.NodeID)
	}

	return h.server.ListenAndServe()
}

// Stop stops the HTTP server gracefully
func (h *HTTPServer) Stop() error {
	if !h.running {
		return nil
	}

	h.running = false

	// Stop WebSocket P2P system
	if h.websocketP2P != nil {
		h.websocketP2P.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return h.server.Shutdown(ctx)
}

// handleJSONRPC handles JSON-RPC 2.0 requests
func (h *HTTPServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSONRPCError(w, nil, -32600, "Invalid Request", "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeJSONRPCError(w, nil, -32700, "Parse error", err.Error(), http.StatusBadRequest)
		return
	}

	var request JSONRPCRequest
	if err := json.Unmarshal(body, &request); err != nil {
		h.writeJSONRPCError(w, nil, -32700, "Parse error", err.Error(), http.StatusBadRequest)
		return
	}

	// Validate JSON-RPC 2.0 format
	if request.JSONRPC != "2.0" {
		h.writeJSONRPCError(w, request.ID, -32600, "Invalid Request", "Missing or invalid jsonrpc version", http.StatusBadRequest)
		return
	}

	if request.Method == "" {
		h.writeJSONRPCError(w, request.ID, -32600, "Invalid Request", "Missing method", http.StatusBadRequest)
		return
	}

	// Process the RPC call
	result, err := h.processRPCCall(request)
	if err != nil {
		if rpcErr, ok := err.(*JSONRPCError); ok {
			h.writeJSONRPCError(w, request.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data, http.StatusOK)
		} else {
			h.writeJSONRPCError(w, request.ID, -32603, "Internal error", err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Write successful response
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      request.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// processRPCCall processes a JSON-RPC method call
func (h *HTTPServer) processRPCCall(request JSONRPCRequest) (interface{}, error) {
	// Check if this is a remote server call
	if request.Method == "remote_call" {
		return h.processRemoteCall(request)
	}

	// Parse method - format: "server_name.method_name" or just "method_name" for default server
	serverName, methodName, err := h.parseMethod(request.Method)
	if err != nil {
		return nil, &JSONRPCError{Code: -32601, Message: "Method not found", Data: err.Error()}
	}

	// Get the server (use exposable registry if available, otherwise fallback to evaluator)
	var server *Value
	var exists bool

	if h.exposableRegistry != nil {
		server, exists = h.exposableRegistry.GetServer(serverName)
	} else {
		server, exists = h.evaluator.GetServer(serverName)
	}

	if !exists {
		return nil, &JSONRPCError{Code: -32601, Message: "Method not found", Data: fmt.Sprintf("Server '%s' not found", serverName)}
	}

	// Convert params to Value arguments
	args, err := h.convertParams(request.Params)
	if err != nil {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params", Data: err.Error()}
	}

	// Call the method on the server
	result, err := server.Server.SendMessage(methodName, args, true)
	if err != nil {
		return nil, &JSONRPCError{Code: -32603, Message: "Internal error", Data: err.Error()}
	}

	// Convert result back to JSON-serializable format
	return h.convertValueToJSON(result), nil
}

// processRemoteCall processes a remote server call via WebSocket P2P
func (h *HTTPServer) processRemoteCall(request JSONRPCRequest) (interface{}, error) {
	if h.websocketP2P == nil {
		return nil, &JSONRPCError{Code: -32601, Message: "Method not found", Data: "WebSocket P2P not enabled"}
	}

	// Parse remote call parameters
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params", Data: "remote_call requires object parameters"}
	}

	nodeID, ok := params["node_id"].(string)
	if !ok {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params", Data: "node_id is required"}
	}

	serverName, ok := params["server_name"].(string)
	if !ok {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params", Data: "server_name is required"}
	}

	method, ok := params["method"].(string)
	if !ok {
		return nil, &JSONRPCError{Code: -32602, Message: "Invalid params", Data: "method is required"}
	}

	// Convert arguments
	var args []*Value
	if argsParam, exists := params["args"]; exists {
		convertedArgs, err := h.convertParams(argsParam)
		if err != nil {
			return nil, &JSONRPCError{Code: -32602, Message: "Invalid params", Data: fmt.Sprintf("invalid args: %v", err)}
		}
		args = convertedArgs
	}

	// Get timeout (default to 30 seconds)
	timeout := 30 * time.Second
	if timeoutParam, exists := params["timeout"]; exists {
		if timeoutFloat, ok := timeoutParam.(float64); ok {
			timeout = time.Duration(timeoutFloat) * time.Second
		}
	}

	// Call remote server
	result, err := h.websocketP2P.CallRemoteServer(nodeID, serverName, method, args, timeout)
	if err != nil {
		return nil, &JSONRPCError{Code: -32603, Message: "Internal error", Data: err.Error()}
	}

	return h.convertValueToJSON(result), nil
}

// parseMethod parses the method string to extract server and method names
func (h *HTTPServer) parseMethod(method string) (string, string, error) {
	// For now, assume format is "server_name.method_name"
	// Later we can support just "method_name" for a default server

	for i, char := range method {
		if char == '.' {
			serverName := method[:i]
			methodName := method[i+1:]
			if serverName == "" || methodName == "" {
				return "", "", fmt.Errorf("invalid method format: %s", method)
			}
			return serverName, methodName, nil
		}
	}

	// No dot found - treat as method on default server (if we implement that)
	return "", "", fmt.Errorf("method must be in format 'server_name.method_name': %s", method)
}

// convertParams converts JSON-RPC params to Relay Value arguments
func (h *HTTPServer) convertParams(params interface{}) ([]*Value, error) {
	if params == nil {
		return []*Value{}, nil
	}

	switch p := params.(type) {
	case map[string]interface{}:
		// Named parameters - convert to object
		fields := make(map[string]*Value)
		for key, value := range p {
			fields[key] = h.convertJSONToValue(value)
		}
		return []*Value{NewObject(fields)}, nil

	case []interface{}:
		// Positional parameters
		args := make([]*Value, len(p))
		for i, value := range p {
			args[i] = h.convertJSONToValue(value)
		}
		return args, nil

	default:
		// Single parameter
		return []*Value{h.convertJSONToValue(params)}, nil
	}
}

// convertJSONToValue converts JSON data to Relay Value
func (h *HTTPServer) convertJSONToValue(data interface{}) *Value {
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
			elements[i] = h.convertJSONToValue(elem)
		}
		return NewArray(elements)
	case map[string]interface{}:
		fields := make(map[string]*Value)
		for key, value := range v {
			fields[key] = h.convertJSONToValue(value)
		}
		return NewObject(fields)
	default:
		// Fallback to string representation
		return NewString(fmt.Sprintf("%v", v))
	}
}

// convertValueToJSON converts Relay Value to JSON-serializable data
func (h *HTTPServer) convertValueToJSON(value *Value) interface{} {
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
			result[i] = h.convertValueToJSON(elem)
		}
		return result
	case ValueTypeObject:
		result := make(map[string]interface{})
		for key, val := range value.Object {
			result[key] = h.convertValueToJSON(val)
		}
		return result
	case ValueTypeStruct:
		result := make(map[string]interface{})
		result["_type"] = value.Struct.Name
		for key, val := range value.Struct.Fields {
			result[key] = h.convertValueToJSON(val)
		}
		return result
	case ValueTypeFunction:
		return fmt.Sprintf("<function: %s>", value.Function.Name)
	case ValueTypeServer:
		return fmt.Sprintf("<server: %s>", value.Server.Name)
	case ValueTypeServerState:
		result := make(map[string]interface{})
		for key, val := range *value.ServerState.State {
			result[key] = h.convertValueToJSON(val)
		}
		return result
	default:
		return value.String()
	}
}

// writeJSONRPCError writes a JSON-RPC error response
func (h *HTTPServer) writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}, httpStatus int) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles health check requests
func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"servers":   len(h.evaluator.servers),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// handleInfo handles server info requests
func (h *HTTPServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	serverNames := make([]string, 0, len(h.evaluator.servers))
	for name := range h.evaluator.servers {
		serverNames = append(serverNames, name)
	}

	info := map[string]interface{}{
		"relay_version": "0.3.0-dev",
		"servers":       serverNames,
		"endpoints": map[string]string{
			"rpc":    "/rpc",
			"health": "/health",
			"info":   "/info",
		},
		"jsonrpc_version": "2.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

// applyMiddleware applies HTTP middleware
func (h *HTTPServer) applyMiddleware(handler http.Handler) http.Handler {
	// CORS middleware
	if h.config.EnableCORS {
		handler = h.corsMiddleware(handler)
	}

	// Logging middleware
	handler = h.loggingMiddleware(handler)

	// Custom headers middleware
	handler = h.headersMiddleware(handler)

	return handler
}

// corsMiddleware adds CORS headers
func (h *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (h *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// headersMiddleware adds custom headers
func (h *HTTPServer) headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, value := range h.config.Headers {
			w.Header().Set(key, value)
		}
		next.ServeHTTP(w, r)
	})
}

// GetExposableRegistry returns the exposable registry if enabled
func (h *HTTPServer) GetExposableRegistry() *ExposableServerRegistry {
	return h.exposableRegistry
}

// GetNodeID returns the node ID for this server
func (h *HTTPServer) GetNodeID() string {
	return h.config.NodeID
}

// AddPeer adds a peer node to the registry
func (h *HTTPServer) AddPeer(nodeID, address string) {
	if h.exposableRegistry != nil {
		h.exposableRegistry.AddPeer(nodeID, address)
	}
}

// RemovePeer removes a peer node from the registry
func (h *HTTPServer) RemovePeer(nodeID string) {
	// This will need to be adapted for the new FederationRouter
}

// GetWebSocketP2P returns the WebSocket P2P system
func (h *HTTPServer) GetWebSocketP2P() *FederationRouter {
	return h.websocketP2P
}

// ConnectToPeer connects to a peer node via WebSocket
func (h *HTTPServer) ConnectToPeer(nodeID, address string) error {
	if h.websocketP2P == nil {
		return fmt.Errorf("WebSocket P2P not enabled")
	}
	return h.websocketP2P.ConnectToPeer(nodeID, address)
}

// SendP2PMessage sends a message to a peer node
func (h *HTTPServer) SendP2PMessage(to, msgType string, data map[string]interface{}) error {
	if h.websocketP2P == nil {
		return fmt.Errorf("WebSocket P2P not enabled")
	}
	return h.websocketP2P.SendMessage(to, msgType, data)
}

// CallRemoteServer calls a method on a server running on a remote node
func (h *HTTPServer) CallRemoteServer(nodeID, serverName, method string, args []*Value, timeout time.Duration) (*Value, error) {
	if h.websocketP2P == nil {
		return nil, fmt.Errorf("WebSocket P2P not enabled")
	}
	return h.websocketP2P.CallRemoteServer(nodeID, serverName, method, args, timeout)
}

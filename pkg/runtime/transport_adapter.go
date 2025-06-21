package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// TransportAdapter handles all transport mechanisms (HTTP, WebSocket)
// and converts them to unified RouteRequest messages
type TransportAdapter struct {
	router   *MessageRouter
	upgrader websocket.Upgrader
	nodeID   string
}

// NewTransportAdapter creates a new transport adapter
func NewTransportAdapter(router *MessageRouter, nodeID string) *TransportAdapter {
	return &TransportAdapter{
		router: router,
		nodeID: nodeID,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// HandleHTTPJSONRPC handles HTTP JSON-RPC requests and converts them to RouteRequest
func (ta *TransportAdapter) HandleHTTPJSONRPC(w http.ResponseWriter, r *http.Request) {
	// Parse JSON-RPC request
	var jsonRPCReq JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&jsonRPCReq); err != nil {
		ta.writeJSONRPCError(w, nil, -32700, "Parse error", err.Error(), http.StatusBadRequest)
		return
	}

	// Convert to RouteRequest
	routeReq, err := ta.jsonRPCToRouteRequest(&jsonRPCReq, "http_client")
	if err != nil {
		ta.writeJSONRPCError(w, jsonRPCReq.ID, -32602, "Invalid params", err.Error(), http.StatusOK)
		return
	}

	// Route the request
	response := ta.router.RouteMessage(routeReq)

	// Convert back to JSON-RPC response
	jsonRPCResp := ta.routeResponseToJSONRPC(response, jsonRPCReq.ID)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonRPCResp)
}

// HandleWebSocket handles WebSocket connections and converts messages to RouteRequest
func (ta *TransportAdapter) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get peer node ID from query parameter
	peerNodeID := r.URL.Query().Get("node_id")
	if peerNodeID == "" {
		http.Error(w, "node_id parameter required", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket
	conn, err := ta.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Add peer to router
	ta.router.AddP2PNode(peerNodeID, r.RemoteAddr)
	defer ta.router.RemoveP2PNode(peerNodeID)

	// Handle WebSocket messages
	ta.handleWebSocketConnection(conn, peerNodeID)
}

// handleWebSocketConnection handles a single WebSocket connection
func (ta *TransportAdapter) handleWebSocketConnection(conn *websocket.Conn, peerNodeID string) {
	for {
		// Read message
		var wsMsg WebSocketMessage
		if err := conn.ReadJSON(&wsMsg); err != nil {
			break // Connection closed or error
		}

		// Convert to RouteRequest based on message type
		switch wsMsg.Type {
		case "server_call":
			ta.handleWebSocketServerCall(conn, &wsMsg, peerNodeID)
		case "ping":
			ta.handleWebSocketPing(conn, &wsMsg, peerNodeID)
		default:
			// Unknown message type, ignore or log
		}
	}
}

// handleWebSocketServerCall handles a server call via WebSocket
func (ta *TransportAdapter) handleWebSocketServerCall(conn *websocket.Conn, wsMsg *WebSocketMessage, peerNodeID string) {
	// Extract server call data
	callData, ok := wsMsg.Data["call"].(map[string]interface{})
	if !ok {
		ta.sendWebSocketError(conn, wsMsg.ID, "Invalid server call data")
		return
	}

	// Convert to RouteRequest
	routeReq := &RouteRequest{
		ID:         wsMsg.ID,
		From:       peerNodeID,
		NodeID:     "", // Local call
		ServerName: callData["server_name"].(string),
		Method:     callData["method"].(string),
		Args:       ta.convertWebSocketArgs(callData["args"]),
		Timeout:    30 * time.Second,
	}

	// Route the request
	response := ta.router.RouteMessage(routeReq)

	// Send WebSocket response
	ta.sendWebSocketResponse(conn, response)
}

// handleWebSocketPing handles ping messages
func (ta *TransportAdapter) handleWebSocketPing(conn *websocket.Conn, wsMsg *WebSocketMessage, peerNodeID string) {
	// Send pong response
	pongMsg := WebSocketMessage{
		Type: "pong",
		ID:   wsMsg.ID,
		From: ta.nodeID,
		To:   peerNodeID,
		Data: map[string]interface{}{
			"timestamp": time.Now(),
		},
	}
	conn.WriteJSON(pongMsg)
}

// jsonRPCToRouteRequest converts a JSON-RPC request to a RouteRequest
func (ta *TransportAdapter) jsonRPCToRouteRequest(jsonReq *JSONRPCRequest, from string) (*RouteRequest, error) {
	// Handle remote_call specially
	if jsonReq.Method == "remote_call" {
		return ta.parseRemoteCall(jsonReq, from)
	}

	// Parse local server call: "server_name.method_name"
	serverName, methodName, err := ta.parseMethod(jsonReq.Method)
	if err != nil {
		return nil, err
	}

	// Convert params to Value arguments
	args, err := ta.convertParams(jsonReq.Params)
	if err != nil {
		return nil, err
	}

	return &RouteRequest{
		ID:         fmt.Sprintf("%v", jsonReq.ID),
		From:       from,
		NodeID:     "", // Local call
		ServerName: serverName,
		Method:     methodName,
		Args:       args,
		Timeout:    30 * time.Second,
	}, nil
}

// parseRemoteCall parses a remote_call JSON-RPC request
func (ta *TransportAdapter) parseRemoteCall(jsonReq *JSONRPCRequest, from string) (*RouteRequest, error) {
	params, ok := jsonReq.Params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("remote_call requires object parameters")
	}

	nodeID, ok := params["node_id"].(string)
	if !ok {
		return nil, fmt.Errorf("node_id is required")
	}

	serverName, ok := params["server_name"].(string)
	if !ok {
		return nil, fmt.Errorf("server_name is required")
	}

	method, ok := params["method"].(string)
	if !ok {
		return nil, fmt.Errorf("method is required")
	}

	// Convert arguments
	var args []*Value
	if argsParam, exists := params["args"]; exists {
		convertedArgs, err := ta.convertParams(argsParam)
		if err != nil {
			return nil, fmt.Errorf("invalid args: %v", err)
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

	return &RouteRequest{
		ID:         fmt.Sprintf("%v", jsonReq.ID),
		From:       from,
		NodeID:     nodeID,
		ServerName: serverName,
		Method:     method,
		Args:       args,
		Timeout:    timeout,
	}, nil
}

// routeResponseToJSONRPC converts a RouteResponse to a JSON-RPC response
func (ta *TransportAdapter) routeResponseToJSONRPC(routeResp *RouteResponse, requestID interface{}) JSONRPCResponse {
	if routeResp.Success {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  ta.convertValueToJSON(routeResp.Result),
			ID:      requestID,
		}
	} else {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    routeResp.Error,
			},
			ID: requestID,
		}
	}
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// sendWebSocketResponse sends a response via WebSocket
func (ta *TransportAdapter) sendWebSocketResponse(conn *websocket.Conn, response *RouteResponse) {
	wsMsg := WebSocketMessage{
		Type: "server_response",
		ID:   response.ID,
		From: ta.nodeID,
		Data: map[string]interface{}{
			"success": response.Success,
			"result":  ta.convertValueToJSON(response.Result),
			"error":   response.Error,
		},
		Timestamp: time.Now(),
	}
	conn.WriteJSON(wsMsg)
}

// sendWebSocketError sends an error via WebSocket
func (ta *TransportAdapter) sendWebSocketError(conn *websocket.Conn, requestID, errorMsg string) {
	wsMsg := WebSocketMessage{
		Type: "error",
		ID:   requestID,
		From: ta.nodeID,
		Data: map[string]interface{}{
			"error": errorMsg,
		},
		Timestamp: time.Now(),
	}
	conn.WriteJSON(wsMsg)
}

// Helper methods (reuse existing implementations)

func (ta *TransportAdapter) parseMethod(method string) (string, string, error) {
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
	return "", "", fmt.Errorf("method must be in format 'server_name.method_name': %s", method)
}

func (ta *TransportAdapter) convertParams(params interface{}) ([]*Value, error) {
	if params == nil {
		return []*Value{}, nil
	}

	switch p := params.(type) {
	case map[string]interface{}:
		// Named parameters - convert to object
		fields := make(map[string]*Value)
		for key, value := range p {
			fields[key] = ta.convertJSONToValue(value)
		}
		return []*Value{NewObject(fields)}, nil

	case []interface{}:
		// Positional parameters
		args := make([]*Value, len(p))
		for i, value := range p {
			args[i] = ta.convertJSONToValue(value)
		}
		return args, nil

	default:
		// Single parameter
		return []*Value{ta.convertJSONToValue(params)}, nil
	}
}

func (ta *TransportAdapter) convertJSONToValue(data interface{}) *Value {
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
			elements[i] = ta.convertJSONToValue(elem)
		}
		return NewArray(elements)
	case map[string]interface{}:
		fields := make(map[string]*Value)
		for key, value := range v {
			fields[key] = ta.convertJSONToValue(value)
		}
		return NewObject(fields)
	default:
		return NewString(fmt.Sprintf("%v", v))
	}
}

func (ta *TransportAdapter) convertValueToJSON(value *Value) interface{} {
	if value == nil {
		return nil
	}

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
			result[i] = ta.convertValueToJSON(elem)
		}
		return result
	case ValueTypeObject:
		result := make(map[string]interface{})
		for key, val := range value.Object {
			result[key] = ta.convertValueToJSON(val)
		}
		return result
	default:
		return value.String()
	}
}

func (ta *TransportAdapter) convertWebSocketArgs(args interface{}) []*Value {
	if args == nil {
		return []*Value{}
	}

	if argsArray, ok := args.([]interface{}); ok {
		result := make([]*Value, len(argsArray))
		for i, arg := range argsArray {
			result[i] = ta.convertJSONToValue(arg)
		}
		return result
	}

	return []*Value{}
}

func (ta *TransportAdapter) writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}, httpStatus int) {
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

package runtime

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransportAdapter_HTTPJSONRPCBasic(t *testing.T) {
	// Create router and adapter
	router := NewMessageRouter()
	defer router.Stop()

	adapter := NewTransportAdapter(router, "test_node")

	// Register test server
	server := createTestServerWithMethods("test_server", map[string]func([]*Value) *Value{
		"hello": func(args []*Value) *Value {
			return NewString("Hello, World!")
		},
	})
	router.RegisterServer("test_server", server)

	// Create HTTP request
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "test_server.hello",
		"id":      1,
	}

	jsonData, _ := json.Marshal(rpcRequest)
	req := httptest.NewRequest("POST", "/rpc", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Handle request
	adapter.HandleHTTPJSONRPC(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["jsonrpc"] != "2.0" {
		t.Fatal("Expected JSON-RPC 2.0 response")
	}

	if response["result"] != "Hello, World!" {
		t.Fatalf("Expected 'Hello, World!', got %v", response["result"])
	}
}

func TestTransportAdapter_HTTPJSONRPCWithParams(t *testing.T) {
	// Create router and adapter
	router := NewMessageRouter()
	defer router.Stop()

	adapter := NewTransportAdapter(router, "test_node")

	// Register test server with parameters
	server := createTestServerWithMethods("calc_server", map[string]func([]*Value) *Value{
		"add": func(args []*Value) *Value {
			if len(args) >= 2 {
				return NewNumber(args[0].Number + args[1].Number)
			}
			return NewNumber(0)
		},
	})
	router.RegisterServer("calc_server", server)

	// Test with array parameters
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "calc_server.add",
		"params":  []interface{}{5, 3},
		"id":      1,
	}

	jsonData, _ := json.Marshal(rpcRequest)
	req := httptest.NewRequest("POST", "/rpc", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	adapter.HandleHTTPJSONRPC(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["result"] != float64(8) {
		t.Fatalf("Expected 8, got %v", response["result"])
	}
}

func TestTransportAdapter_HTTPJSONRPCObjectParams(t *testing.T) {
	// Create router and adapter
	router := NewMessageRouter()
	defer router.Stop()

	adapter := NewTransportAdapter(router, "test_node")

	// Register test server that expects object parameter
	server := createTestServerWithMethods("user_server", map[string]func([]*Value) *Value{
		"create_user": func(args []*Value) *Value {
			if len(args) > 0 && args[0].Type == ValueTypeObject {
				if name, exists := args[0].Object["name"]; exists {
					return NewString("Created user: " + name.Str)
				}
			}
			return NewString("Invalid user data")
		},
	})
	router.RegisterServer("user_server", server)

	// Test with object parameters
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user_server.create_user",
		"params": map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"id": 1,
	}

	jsonData, _ := json.Marshal(rpcRequest)
	req := httptest.NewRequest("POST", "/rpc", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	adapter.HandleHTTPJSONRPC(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["result"] != "Created user: Alice" {
		t.Fatalf("Expected 'Created user: Alice', got %v", response["result"])
	}
}

func TestTransportAdapter_HTTPJSONRPCRemoteCall(t *testing.T) {
	// Create router and adapter
	router := NewMessageRouter()
	defer router.Stop()

	adapter := NewTransportAdapter(router, "test_node")

	// Test remote_call method (should fail since P2P routing not implemented)
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "remote_call",
		"params": map[string]interface{}{
			"node_id":     "remote_node",
			"server_name": "remote_server",
			"method":      "test",
			"args":        []interface{}{},
		},
		"id": 1,
	}

	jsonData, _ := json.Marshal(rpcRequest)
	req := httptest.NewRequest("POST", "/rpc", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	adapter.HandleHTTPJSONRPC(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should return an error since P2P routing is not fully implemented
	if response["error"] == nil {
		t.Fatal("Expected error for remote call")
	}
}

func TestTransportAdapter_HTTPJSONRPCErrors(t *testing.T) {
	// Create router and adapter
	router := NewMessageRouter()
	defer router.Stop()

	adapter := NewTransportAdapter(router, "test_node")

	tests := []struct {
		name           string
		request        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Invalid JSON",
			request:        `{"invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Non-existent server",
			request: `{
				"jsonrpc": "2.0",
				"method": "nonexistent_server.test",
				"id": 1
			}`,
			expectedStatus: http.StatusOK,
			expectError:    true,
		},
		{
			name: "Invalid method format",
			request: `{
				"jsonrpc": "2.0",
				"method": "invalid_method_format",
				"id": 1
			}`,
			expectedStatus: http.StatusOK,
			expectError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/rpc", strings.NewReader(test.request))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			adapter.HandleHTTPJSONRPC(w, req)

			if w.Code != test.expectedStatus {
				t.Fatalf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			if test.expectError {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				if response["error"] == nil {
					t.Fatal("Expected error in response")
				}
			}
		})
	}
}

func TestTransportAdapter_HTTPMethodNotAllowed(t *testing.T) {
	// Create router and adapter
	router := NewMessageRouter()
	defer router.Stop()

	adapter := NewTransportAdapter(router, "test_node")

	// Test GET request (should be rejected)
	req := httptest.NewRequest("GET", "/rpc", nil)
	w := httptest.NewRecorder()

	adapter.HandleHTTPJSONRPC(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for GET request, got %d", w.Code)
	}
}

func TestTransportAdapter_ConvertJSONToValue(t *testing.T) {
	router := NewMessageRouter()
	adapter := NewTransportAdapter(router, "test_node")

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"Nil", nil, "nil"},
		{"Bool true", true, "true"},
		{"Bool false", false, "false"},
		{"Number", 42.5, "42.5"},
		{"String", "hello", `"hello"`},
		{"Array", []interface{}{1.0, 2.0, 3.0}, "[1, 2, 3]"},
		{"Object", map[string]interface{}{"key": "value"}, `{key: "value"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := adapter.convertJSONToValue(test.input)
			if result.String() != test.expected {
				t.Fatalf("Expected %s, got %s", test.expected, result.String())
			}
		})
	}
}

func TestTransportAdapter_ConvertValueToJSON(t *testing.T) {
	router := NewMessageRouter()
	adapter := NewTransportAdapter(router, "test_node")

	tests := []struct {
		name     string
		input    *Value
		expected interface{}
	}{
		{"Nil", NewNil(), nil},
		{"Bool", NewBool(true), true},
		{"Number", NewNumber(42.5), 42.5},
		{"String", NewString("hello"), "hello"},
		{"Array", NewArray([]*Value{NewNumber(1), NewNumber(2)}), []interface{}{1.0, 2.0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := adapter.convertValueToJSON(test.input)

			// Use JSON marshaling for comparison to handle type differences
			expectedJSON, _ := json.Marshal(test.expected)
			resultJSON, _ := json.Marshal(result)

			if string(expectedJSON) != string(resultJSON) {
				t.Fatalf("Expected %s, got %s", string(expectedJSON), string(resultJSON))
			}
		})
	}
}

func TestTransportAdapter_ParseMethod(t *testing.T) {
	router := NewMessageRouter()
	adapter := NewTransportAdapter(router, "test_node")

	tests := []struct {
		name           string
		method         string
		expectedServer string
		expectedMethod string
		expectError    bool
	}{
		{"Valid method", "server.method", "server", "method", false},
		{"Complex method", "my_server.complex_method_name", "my_server", "complex_method_name", false},
		{"Invalid no dot", "invalid", "", "", true},
		{"Invalid empty server", ".method", "", "", true},
		{"Invalid empty method", "server.", "", "", true},
		{"Multiple dots", "server.method.extra", "server", "method.extra", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, method, err := adapter.parseMethod(test.method)

			if test.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if server != test.expectedServer {
					t.Fatalf("Expected server %s, got %s", test.expectedServer, server)
				}
				if method != test.expectedMethod {
					t.Fatalf("Expected method %s, got %s", test.expectedMethod, method)
				}
			}
		})
	}
}

func TestTransportAdapter_WebSocketHandleRequiresNodeID(t *testing.T) {
	router := NewMessageRouter()
	adapter := NewTransportAdapter(router, "test_node")

	// Test WebSocket handler without node_id parameter
	req := httptest.NewRequest("GET", "/ws/p2p", nil)
	w := httptest.NewRecorder()

	adapter.HandleWebSocket(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for WebSocket without node_id, got %d", w.Code)
	}
}

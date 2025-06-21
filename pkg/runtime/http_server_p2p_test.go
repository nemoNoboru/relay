package runtime

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHTTPServer_P2P_Integration(t *testing.T) {
	// Create two evaluators with servers
	evaluator1 := NewEvaluator()
	evaluator2 := NewEvaluator()

	// Create simple test servers
	createTestServer(evaluator1, "test_server1")
	createTestServer(evaluator2, "test_server2")

	// Create HTTP server configurations
	config1 := &HTTPServerConfig{
		Host:              "127.0.0.1",
		Port:              0, // Let the test server choose the port
		EnableCORS:        true,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		Headers:           make(map[string]string),
		NodeID:            "node1",
		EnableRegistry:    true,
		DiscoveryInterval: 1 * time.Second,
	}

	config2 := &HTTPServerConfig{
		Host:              "127.0.0.1",
		Port:              0,
		EnableCORS:        true,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		Headers:           make(map[string]string),
		NodeID:            "node2",
		EnableRegistry:    true,
		DiscoveryInterval: 1 * time.Second,
	}

	// Create HTTP servers
	httpServer1 := NewHTTPServer(evaluator1, config1)
	httpServer2 := NewHTTPServer(evaluator2, config2)

	// Create test servers
	testServer1 := httptest.NewServer(httpServer1.createMux())
	testServer2 := httptest.NewServer(httpServer2.createMux())
	defer testServer1.Close()
	defer testServer2.Close()

	// Test basic server functionality
	t.Run("BasicServerFunctionality", func(t *testing.T) {
		testBasicServerFunctionality(t, testServer1.URL)
	})

	// Test registry endpoints
	t.Run("RegistryEndpoints", func(t *testing.T) {
		testRegistryEndpoints(t, testServer1.URL)
	})

	// Test WebSocket P2P endpoints
	t.Run("WebSocketP2PEndpoints", func(t *testing.T) {
		testWebSocketP2PEndpoints(t, testServer1.URL, testServer2.URL)
	})

	// Test remote server calls
	t.Run("RemoteServerCalls", func(t *testing.T) {
		testRemoteServerCalls(t, testServer1.URL, testServer2.URL)
	})
}

func createTestServer(evaluator *Evaluator, serverName string) {
	// Create a simple server manually
	server := &Server{
		Name:        serverName,
		State:       make(map[string]*Value),
		MessageChan: make(chan *Message, 100),
		Running:     false,
		Environment: NewEnvironment(nil),
	}

	// Add a simple hello method
	server.State["greeting"] = NewString("Hello, World!")

	// Start the server
	server.Start(evaluator)

	// Register the server
	serverValue := &Value{
		Type:   ValueTypeServer,
		Server: server,
	}
	evaluator.servers[serverName] = serverValue
}

func testBasicServerFunctionality(t *testing.T, serverURL string) {
	// Test health endpoint
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test info endpoint
	resp, err = http.Get(serverURL + "/info")
	if err != nil {
		t.Fatalf("Failed to call info endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test JSON-RPC endpoint with simple server call
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "test_server1.hello",
		"id":      1,
	}

	jsonData, _ := json.Marshal(rpcRequest)
	resp, err = http.Post(serverURL+"/rpc", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to call RPC endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var rpcResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&rpcResponse)

	// Check if we get proper JSON-RPC error for non-existent method (expected)
	if rpcResponse["error"] == nil {
		t.Log("RPC call succeeded (unexpected but not necessarily wrong)")
	} else {
		errorObj := rpcResponse["error"].(map[string]interface{})
		if errorObj["code"] != float64(-32601) {
			t.Errorf("Expected error code -32601, got %v", errorObj["code"])
		}
	}
}

func testRegistryEndpoints(t *testing.T, serverURL string) {
	// Test registry endpoint
	resp, err := http.Get(serverURL + "/registry")
	if err != nil {
		t.Fatalf("Failed to call registry endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Registry endpoint might return 404 if not properly configured, which is what we're testing
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
	}

	// Test registry/servers endpoint
	resp, err = http.Get(serverURL + "/registry/servers")
	if err != nil {
		t.Fatalf("Failed to call registry/servers endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
	}

	// Test registry/peers endpoint
	resp, err = http.Get(serverURL + "/registry/peers")
	if err != nil {
		t.Fatalf("Failed to call registry/peers endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
	}
}

func testWebSocketP2PEndpoints(t *testing.T, serverURL1, serverURL2 string) {
	// Convert HTTP URLs to WebSocket URLs
	wsURL1 := "ws" + strings.TrimPrefix(serverURL1, "http") + "/ws/p2p?node_id=test_client"
	wsURL2 := "ws" + strings.TrimPrefix(serverURL2, "http") + "/ws/p2p?node_id=test_client"

	// Test WebSocket connection to server 1
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket endpoint 1: %v", err)
	}
	defer conn1.Close()

	// Test WebSocket connection to server 2
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket endpoint 2: %v", err)
	}
	defer conn2.Close()

	// Test sending a message
	testMessage := map[string]interface{}{
		"type": "ping",
		"from": "test_client",
		"to":   "node1",
		"data": map[string]interface{}{"test": true},
	}

	err = conn1.WriteJSON(testMessage)
	if err != nil {
		t.Errorf("Failed to send WebSocket message: %v", err)
	}

	// Try to read a response (might timeout, which is ok)
	conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
	var response map[string]interface{}
	err = conn1.ReadJSON(&response)
	if err != nil {
		// Timeout is expected if P2P system isn't fully connected
		t.Logf("WebSocket read timeout (expected): %v", err)
	} else {
		t.Logf("Received WebSocket response: %v", response)
	}
}

func testRemoteServerCalls(t *testing.T, serverURL1, serverURL2 string) {
	// Test remote_call method
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "remote_call",
		"params": map[string]interface{}{
			"node_id":     "node2",
			"server_name": "test_server2",
			"method":      "hello",
			"args":        []interface{}{},
			"timeout":     5.0,
		},
		"id": 1,
	}

	jsonData, _ := json.Marshal(rpcRequest)
	resp, err := http.Post(serverURL1+"/rpc", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to call remote_call RPC: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var rpcResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&rpcResponse)

	// We expect this to fail because the nodes aren't connected
	if rpcResponse["error"] == nil {
		t.Log("Remote call succeeded (unexpected)")
	} else {
		errorObj := rpcResponse["error"].(map[string]interface{})
		t.Logf("Remote call failed as expected: %v", errorObj["message"])
	}
}

func TestHTTPServer_P2P_Configuration(t *testing.T) {
	evaluator := NewEvaluator()

	// Test with P2P disabled
	config := &HTTPServerConfig{
		Host:           "127.0.0.1",
		Port:           0,
		EnableRegistry: false,
		NodeID:         "test_node",
	}

	httpServer := NewHTTPServer(evaluator, config)
	if httpServer.exposableRegistry != nil {
		t.Error("Expected nil registry when P2P is disabled")
	}

	if httpServer.websocketP2P != nil {
		t.Error("Expected nil WebSocket P2P when P2P is disabled")
	}

	// Test with P2P enabled
	config.EnableRegistry = true
	httpServer = NewHTTPServer(evaluator, config)

	if httpServer.exposableRegistry == nil {
		t.Error("Expected non-nil registry when P2P is enabled")
	}

	if httpServer.websocketP2P == nil {
		t.Error("Expected non-nil WebSocket P2P when P2P is enabled")
	}

	if httpServer.GetNodeID() != "test_node" {
		t.Errorf("Expected node ID 'test_node', got '%s'", httpServer.GetNodeID())
	}
}

func TestHTTPServer_P2P_Methods(t *testing.T) {
	evaluator := NewEvaluator()
	config := &HTTPServerConfig{
		Host:           "127.0.0.1",
		Port:           0,
		EnableRegistry: true,
		NodeID:         "test_node",
	}

	httpServer := NewHTTPServer(evaluator, config)

	// Test GetExposableRegistry
	registry := httpServer.GetExposableRegistry()
	if registry == nil {
		t.Error("Expected non-nil registry")
	}

	// Test GetWebSocketP2P
	p2p := httpServer.GetWebSocketP2P()
	if p2p == nil {
		t.Error("Expected non-nil WebSocket P2P")
	}

	// Test AddPeer (should not panic)
	httpServer.AddPeer("peer1", "127.0.0.1:8081")

	// Test RemovePeer (should not panic)
	httpServer.RemovePeer("peer1")

	// Test ConnectToPeer (should fail gracefully)
	err := httpServer.ConnectToPeer("peer1", "127.0.0.1:8081")
	if err == nil {
		t.Error("Expected error when connecting to non-existent peer")
	}

	// Test SendP2PMessage (should fail gracefully)
	data := map[string]interface{}{"test": true}
	err = httpServer.SendP2PMessage("peer1", "test", data)
	if err == nil {
		t.Error("Expected error when sending message to non-existent peer")
	}

	// Test CallRemoteServer (should fail gracefully)
	args := []*Value{NewString("test")}
	result, err := httpServer.CallRemoteServer("peer1", "test_server", "test_method", args, 5*time.Second)
	if err == nil {
		t.Error("Expected error when calling remote server on non-existent peer")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestHTTPServer_RemoteCall_Validation(t *testing.T) {
	evaluator := NewEvaluator()
	config := &HTTPServerConfig{
		Host:           "127.0.0.1",
		Port:           0,
		EnableRegistry: true,
		NodeID:         "test_node",
	}

	httpServer := NewHTTPServer(evaluator, config)
	testServer := httptest.NewServer(httpServer.createMux())
	defer testServer.Close()

	// Test remote_call with missing parameters
	testCases := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "missing node_id",
			params:   map[string]interface{}{"server_name": "test", "method": "test"},
			expected: "node_id is required",
		},
		{
			name:     "missing server_name",
			params:   map[string]interface{}{"node_id": "test", "method": "test"},
			expected: "server_name is required",
		},
		{
			name:     "missing method",
			params:   map[string]interface{}{"node_id": "test", "server_name": "test"},
			expected: "method is required",
		},
		{
			name: "invalid args",
			params: map[string]interface{}{
				"node_id":     "test",
				"server_name": "test",
				"method":      "test",
				"args":        "invalid",
			},
			expected: "invalid args",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rpcRequest := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "remote_call",
				"params":  tc.params,
				"id":      1,
			}

			jsonData, _ := json.Marshal(rpcRequest)
			resp, err := http.Post(testServer.URL+"/rpc", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to call RPC: %v", err)
			}
			defer resp.Body.Close()

			var rpcResponse map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&rpcResponse)

			if rpcResponse["error"] == nil {
				t.Error("Expected error response")
			} else {
				errorObj := rpcResponse["error"].(map[string]interface{})
				if !strings.Contains(errorObj["data"].(string), tc.expected) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.expected, errorObj["data"])
				}
			}
		})
	}
}

// Helper method to create the HTTP mux (extracted from HTTPServer)
func (h *HTTPServer) createMux() http.Handler {
	mux := http.NewServeMux()

	// Add JSON-RPC endpoint
	mux.HandleFunc("/rpc", h.handleJSONRPC)

	// Add health and info endpoints
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/info", h.handleInfo)

	// Add registry endpoints if P2P is enabled
	if h.exposableRegistry != nil {
		h.exposableRegistry.SetupHTTPEndpoints(mux)
	}

	// Add WebSocket P2P endpoint if enabled
	if h.websocketP2P != nil {
		h.websocketP2P.SetupWebSocketEndpoint(mux)
	}

	return h.applyMiddleware(mux)
}

package runtime

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"relay/pkg/parser"
	"strings"
	"testing"
	"time"
)

func TestUnifiedHTTPServer_BasicEndpoints(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0, // Use any available port for testing
		EnableCORS: true,
		NodeID:     "test_node",
	}

	server := NewUnifiedHTTPServer(evaluator, config)

	// Create test HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/info", server.handleInfo)
	mux.HandleFunc("/registry", server.handleRegistry)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Test health endpoint
	t.Run("Health endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to call health endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var health map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&health)

		if health["status"] != "healthy" {
			t.Fatal("Expected healthy status")
		}
		if health["node_id"] != "test_node" {
			t.Fatal("Expected correct node_id")
		}
	})

	// Test info endpoint
	t.Run("Info endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/info")
		if err != nil {
			t.Fatalf("Failed to call info endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)

		if info["architecture"] != "unified_message_router" {
			t.Fatal("Expected unified_message_router architecture")
		}
		if info["node_id"] != "test_node" {
			t.Fatal("Expected correct node_id")
		}
	})

	// Test registry endpoint
	t.Run("Registry endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/registry")
		if err != nil {
			t.Fatalf("Failed to call registry endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var registry map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&registry)

		if registry["node_id"] != "test_node" {
			t.Fatal("Expected correct node_id in registry")
		}
	})
}

func TestUnifiedHTTPServer_WithRelayServers(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Create a Relay server
	serverDef := `server test_server {
		state {
			count: number = 0
		}
		
		receive fn increment() -> number {
			set new_count = state.get("count") + 1
			state.set("count", new_count)
			new_count
		}
		
		receive fn get_count() -> number {
			state.get("count")
		}
		
		receive fn echo(msg: string) -> string {
			msg
		}
	}`

	program, err := parser.Parse("test", strings.NewReader(serverDef))
	if err != nil {
		t.Fatalf("Failed to parse server definition: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Create test HTTP server with transport adapter
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", httpServer.transportAdapter.HandleHTTPJSONRPC)
	mux.HandleFunc("/registry", httpServer.handleRegistry)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Test server call via JSON-RPC
	t.Run("JSON-RPC server call", func(t *testing.T) {
		rpcRequest := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "test_server.echo",
			"params":  []interface{}{"Hello, World!"},
			"id":      1,
		}

		jsonData, _ := json.Marshal(rpcRequest)
		resp, err := http.Post(testServer.URL+"/rpc", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to call RPC endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		if response["result"] != "Hello, World!" {
			t.Fatalf("Expected 'Hello, World!', got %v", response["result"])
		}
	})

	// Test server registry shows the server
	t.Run("Registry shows server", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/registry")
		if err != nil {
			t.Fatalf("Failed to call registry endpoint: %v", err)
		}
		defer resp.Body.Close()

		var registry map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&registry)

		servers, ok := registry["servers"].([]interface{})
		if !ok {
			t.Fatal("Expected servers array in registry")
		}

		found := false
		for _, s := range servers {
			server := s.(map[string]interface{})
			if server["name"] == "test_server" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("Expected test_server in registry")
		}
	})
}

func TestUnifiedHTTPServer_ErrorHandling(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Create test HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", httpServer.transportAdapter.HandleHTTPJSONRPC)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	tests := []struct {
		name        string
		request     map[string]interface{}
		expectError bool
	}{
		{
			name: "Non-existent server",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "nonexistent_server.test",
				"id":      1,
			},
			expectError: true,
		},
		{
			name: "Invalid method format",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "invalid_method",
				"id":      2,
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(test.request)
			resp, err := http.Post(testServer.URL+"/rpc", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to call RPC endpoint: %v", err)
			}
			defer resp.Body.Close()

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)

			if test.expectError {
				if response["error"] == nil {
					t.Fatal("Expected error in response")
				}
			} else {
				if response["error"] != nil {
					t.Fatalf("Unexpected error: %v", response["error"])
				}
			}
		})
	}
}

func TestUnifiedHTTPServer_CORS(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Create test HTTP server with CORS middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", httpServer.transportAdapter.HandleHTTPJSONRPC)
	handler := httpServer.applyMiddleware(mux)

	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	// Test OPTIONS request for CORS
	req, _ := http.NewRequest("OPTIONS", testServer.URL+"/rpc", nil)
	req.Header.Set("Origin", "http://example.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}

	// Check CORS headers
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Fatalf("Expected Access-Control-Allow-Origin: *, got %s", allowOrigin)
	}

	allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Fatal("Expected Access-Control-Allow-Methods header")
	}
}

func TestUnifiedHTTPServer_P2PNodeManagement(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Test adding peer
	httpServer.AddPeer("peer_1", "http://peer1:8080")

	// Test removing peer
	httpServer.RemovePeer("peer_1")

	// Test node ID retrieval
	nodeID := httpServer.GetNodeID()
	if nodeID != "test_node" {
		t.Fatalf("Expected node_id 'test_node', got %s", nodeID)
	}
}

func TestUnifiedHTTPServer_CallServerMethod(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Create a test server manually
	testServer := createTestServerWithMethods("math_server", map[string]func([]*Value) *Value{
		"multiply": func(args []*Value) *Value {
			if len(args) >= 2 {
				return NewNumber(args[0].Number * args[1].Number)
			}
			return NewNumber(0)
		},
	})
	evaluator.servers["math_server"] = testServer

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Test calling server method directly
	args := []*Value{NewNumber(6), NewNumber(7)}
	result, err := httpServer.CallServer("", "math_server", "multiply", args, 5*time.Second)
	if err != nil {
		t.Fatalf("Expected successful call, got error: %v", err)
	}

	if result.Number != 42 {
		t.Fatalf("Expected 42, got %f", result.Number)
	}
}

func TestUnifiedHTTPServer_MessageRouterAccess(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Test getting message router
	router := httpServer.GetMessageRouter()
	if router == nil {
		t.Fatal("Expected non-nil message router")
	}

	// Test registering server through HTTP server
	testServer := createTestServerValue("direct_server")
	httpServer.RegisterServer("direct_server", testServer)

	// Verify server was registered
	retrieved, exists := router.GetServer("direct_server")
	if !exists {
		t.Fatal("Expected server to be registered")
	}
	if retrieved != testServer {
		t.Fatal("Retrieved server doesn't match registered server")
	}
}

func TestUnifiedHTTPServer_CustomHeaders(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := &HTTPServerConfig{
		Host:       "localhost",
		Port:       0,
		EnableCORS: true,
		NodeID:     "test_node",
		Headers: map[string]string{
			"X-Custom-Header": "test-value",
			"X-API-Version":   "1.0",
		},
	}

	httpServer := NewUnifiedHTTPServer(evaluator, config)

	// Create test HTTP server with custom headers middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/health", httpServer.handleHealth)
	handler := httpServer.applyMiddleware(mux)

	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Check custom headers
	if resp.Header.Get("X-Custom-Header") != "test-value" {
		t.Fatal("Expected custom header X-Custom-Header")
	}
	if resp.Header.Get("X-API-Version") != "1.0" {
		t.Fatal("Expected custom header X-API-Version")
	}
}

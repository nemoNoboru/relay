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

func TestHTTPServerBasics(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Create a test server
	serverDef := `server test_server {
		state {
			count: number = 0,
			message: string = "Hello"
		}
		
		receive fn get_count() -> number {
			state.get("count")
		}
		
		receive fn increment() -> number {
			state.set("count", state.get("count") + 1)
			state.get("count")
		}
		
		receive fn get_message() -> string {
			state.get("message")
		}
		
		receive fn echo(msg: string) -> string {
			msg
		}
		
		receive fn add(a: number, b: number) -> number {
			a + b
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

	// Create HTTP server
	httpServer := NewHTTPServer(evaluator, DefaultHTTPServerConfig())

	tests := []struct {
		name           string
		method         string
		params         interface{}
		expectedResult interface{}
		expectError    bool
	}{
		{
			name:           "Get initial count",
			method:         "test_server.get_count",
			params:         nil,
			expectedResult: float64(0),
			expectError:    false,
		},
		{
			name:           "Increment count",
			method:         "test_server.increment",
			params:         nil,
			expectedResult: float64(1),
			expectError:    false,
		},
		{
			name:           "Get message",
			method:         "test_server.get_message",
			params:         nil,
			expectedResult: "Hello",
			expectError:    false,
		},
		{
			name:           "Echo with string parameter",
			method:         "test_server.echo",
			params:         []interface{}{"test message"},
			expectedResult: "test message",
			expectError:    false,
		},
		{
			name:           "Add two numbers",
			method:         "test_server.add",
			params:         []interface{}{10, 20},
			expectedResult: float64(30),
			expectError:    false,
		},
		{
			name:        "Invalid server name",
			method:      "nonexistent_server.method",
			params:      nil,
			expectError: true,
		},
		{
			name:           "Invalid method name",
			method:         "test_server.nonexistent_method",
			params:         nil,
			expectError:    false,
			expectedResult: nil,
		},
		{
			name:        "Invalid method format",
			method:      "invalid_format",
			params:      nil,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  test.method,
				Params:  test.params,
				ID:      1,
			}

			body, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			httpServer.handleJSONRPC(recorder, req)

			var response JSONRPCResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if test.expectError {
				if response.Error == nil {
					t.Fatalf("Expected error but got result: %v", response.Result)
				}
			} else {
				if response.Error != nil {
					t.Fatalf("Expected result but got error: %v", response.Error)
				}
				if response.Result != test.expectedResult {
					t.Fatalf("Expected result %v but got %v", test.expectedResult, response.Result)
				}
			}
		})
	}
}

func TestHTTPServerParameterTypes(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Create a test server that handles different parameter types
	serverDef := `server param_server {
		receive fn handle_object(data: object) -> object {
			data
		}
		
		receive fn handle_array(items: [number]) -> [number] {
			items
		}
		
		receive fn simple_array(items) {
			items
		}
		
		receive fn handle_mixed(str: string, num: number, bool: bool) -> object {
			{message: str, value: num, flag: bool}
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

	time.Sleep(10 * time.Millisecond)

	httpServer := NewHTTPServer(evaluator, DefaultHTTPServerConfig())

	tests := []struct {
		name           string
		method         string
		params         interface{}
		expectedResult interface{}
	}{
		{
			name:   "Object parameter",
			method: "param_server.handle_object",
			params: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expectedResult: map[string]interface{}{
				"name": "John",
				"age":  float64(30),
			},
		},
		{
			name:           "Array parameter",
			method:         "param_server.simple_array",
			params:         []interface{}{[]interface{}{1, 2, 3, 4, 5}},
			expectedResult: []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
		},
		{
			name:   "Multiple parameters",
			method: "param_server.handle_mixed",
			params: []interface{}{"hello", 42, true},
			expectedResult: map[string]interface{}{
				"message": "hello",
				"value":   float64(42),
				"flag":    true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  test.method,
				Params:  test.params,
				ID:      1,
			}

			body, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			httpServer.handleJSONRPC(recorder, req)

			var response JSONRPCResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Error != nil {
				t.Fatalf("Expected result but got error: %v", response.Error)
			}

			// Deep comparison for complex types
			expectedJSON, _ := json.Marshal(test.expectedResult)
			resultJSON, _ := json.Marshal(response.Result)

			if string(expectedJSON) != string(resultJSON) {
				t.Fatalf("Expected result %v but got %v", test.expectedResult, response.Result)
			}
		})
	}
}

func TestHTTPServerHealthAndInfo(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	httpServer := NewHTTPServer(evaluator, DefaultHTTPServerConfig())

	t.Run("Health endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		recorder := httptest.NewRecorder()

		httpServer.handleHealth(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Expected status 200 but got %d", recorder.Code)
		}

		var health map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &health)
		if err != nil {
			t.Fatalf("Failed to unmarshal health response: %v", err)
		}

		if health["status"] != "healthy" {
			t.Fatalf("Expected status 'healthy' but got %v", health["status"])
		}
	})

	t.Run("Info endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/info", nil)
		recorder := httptest.NewRecorder()

		httpServer.handleInfo(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Expected status 200 but got %d", recorder.Code)
		}

		var info map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &info)
		if err != nil {
			t.Fatalf("Failed to unmarshal info response: %v", err)
		}

		if info["relay_version"] != "0.3.0-dev" {
			t.Fatalf("Expected relay_version '0.3.0-dev' but got %v", info["relay_version"])
		}

		if info["jsonrpc_version"] != "2.0" {
			t.Fatalf("Expected jsonrpc_version '2.0' but got %v", info["jsonrpc_version"])
		}
	})
}

func TestJSONRPCErrorHandling(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	httpServer := NewHTTPServer(evaluator, DefaultHTTPServerConfig())

	tests := []struct {
		name         string
		method       string
		body         string
		expectedCode int
		httpStatus   int
	}{
		{
			name:         "Invalid JSON",
			method:       "POST",
			body:         `{invalid json}`,
			expectedCode: -32700,
			httpStatus:   http.StatusBadRequest,
		},
		{
			name:         "Missing jsonrpc version",
			method:       "POST",
			body:         `{"method": "test", "id": 1}`,
			expectedCode: -32600,
			httpStatus:   http.StatusBadRequest,
		},
		{
			name:         "Missing method",
			method:       "POST",
			body:         `{"jsonrpc": "2.0", "id": 1}`,
			expectedCode: -32600,
			httpStatus:   http.StatusBadRequest,
		},
		{
			name:         "GET method not allowed",
			method:       "GET",
			body:         "",
			expectedCode: -32600,
			httpStatus:   http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, "/rpc", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			httpServer.handleJSONRPC(recorder, req)

			if recorder.Code != test.httpStatus {
				t.Fatalf("Expected HTTP status %d but got %d", test.httpStatus, recorder.Code)
			}

			var response JSONRPCResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if response.Error == nil {
				t.Fatalf("Expected error but got none")
			}

			if response.Error.Code != test.expectedCode {
				t.Fatalf("Expected error code %d but got %d", test.expectedCode, response.Error.Code)
			}
		})
	}
}

func TestHTTPServerCORS(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	config := DefaultHTTPServerConfig()
	config.EnableCORS = true

	httpServer := NewHTTPServer(evaluator, config)
	handler := httpServer.applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("CORS headers are set", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/rpc", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		headers := recorder.Header()

		if headers.Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("Expected CORS origin header '*' but got %v", headers.Get("Access-Control-Allow-Origin"))
		}

		if headers.Get("Access-Control-Allow-Methods") != "POST, GET, OPTIONS" {
			t.Fatalf("Expected CORS methods header but got %v", headers.Get("Access-Control-Allow-Methods"))
		}
	})

	t.Run("OPTIONS request handling", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/rpc", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Expected status 200 for OPTIONS but got %d", recorder.Code)
		}
	})
}

func TestHTTPServerConfiguration(t *testing.T) {
	config := &HTTPServerConfig{
		Port:         9090,
		Host:         "127.0.0.1",
		EnableCORS:   false,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Headers: map[string]string{
			"X-Custom-Header": "test-value",
		},
	}

	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	httpServer := NewHTTPServer(evaluator, config)

	if httpServer.config.Port != 9090 {
		t.Fatalf("Expected port 9090 but got %d", httpServer.config.Port)
	}

	if httpServer.config.Host != "127.0.0.1" {
		t.Fatalf("Expected host '127.0.0.1' but got %s", httpServer.config.Host)
	}

	if httpServer.config.EnableCORS != false {
		t.Fatalf("Expected CORS disabled")
	}

	// Test custom headers middleware
	handler := httpServer.headersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Header().Get("X-Custom-Header") != "test-value" {
		t.Fatalf("Expected custom header 'test-value' but got %v", recorder.Header().Get("X-Custom-Header"))
	}
}

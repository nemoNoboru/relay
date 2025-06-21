package runtime

import (
	"testing"
	"time"
)

func TestMessageRouter_BasicOperations(t *testing.T) {
	router := NewMessageRouter()
	defer router.Stop()

	// Test server registration without starting router (direct mode)
	server := createTestServerValue("test_server")
	router.RegisterServer("test_server", server)

	// Test server retrieval
	retrieved, exists := router.GetServer("test_server")
	if !exists {
		t.Fatal("Expected server to exist")
	}
	if retrieved != server {
		t.Fatal("Retrieved server doesn't match registered server")
	}

	// Test GetAllServers
	allServers := router.GetAllServers()
	if len(allServers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(allServers))
	}
	if allServers["test_server"] != server {
		t.Fatal("GetAllServers doesn't return correct server")
	}
}

func TestMessageRouter_ActorMode(t *testing.T) {
	router := NewMessageRouter()
	router.Start()
	defer router.Stop()

	// Give router time to start
	time.Sleep(10 * time.Millisecond)

	// Test server registration in actor mode
	server := createTestServerValue("actor_server")
	router.RegisterServer("actor_server", server)

	// Give time for registration to process
	time.Sleep(10 * time.Millisecond)

	// Test server retrieval
	retrieved, exists := router.GetServer("actor_server")
	if !exists {
		t.Fatal("Expected server to exist in actor mode")
	}
	if retrieved != server {
		t.Fatal("Retrieved server doesn't match registered server in actor mode")
	}
}

func TestMessageRouter_LocalServerCall(t *testing.T) {
	router := NewMessageRouter()
	defer router.Stop()

	// Create a test server with a simple method
	server := createTestServerWithMethods("echo_server", map[string]func([]*Value) *Value{
		"echo": func(args []*Value) *Value {
			if len(args) > 0 {
				return args[0]
			}
			return NewString("default")
		},
	})
	router.RegisterServer("echo_server", server)

	// Give some time for registration
	time.Sleep(10 * time.Millisecond)

	// Test local server call using CallServer method
	args := []*Value{NewString("hello")}
	result, err := router.CallServer("", "echo_server", "echo", args, 5*time.Second)
	if err != nil {
		t.Fatalf("Expected successful call, got error: %v", err)
	}
	if result.String() != `"hello"` {
		t.Fatalf("Expected 'hello', got %s", result.String())
	}
}

func TestMessageRouter_NonExistentServer(t *testing.T) {
	router := NewMessageRouter()
	defer router.Stop()

	// Test call to non-existent server
	req := &RouteRequest{
		ID:         "test_2",
		From:       "test_client",
		NodeID:     "", // Local call
		ServerName: "nonexistent_server",
		Method:     "test",
		Args:       []*Value{},
		Timeout:    5 * time.Second,
	}

	response := router.RouteMessage(req)
	if response.Success {
		t.Fatal("Expected error for non-existent server")
	}
	if response.Error == "" {
		t.Fatal("Expected error message for non-existent server")
	}
}

func TestMessageRouter_P2PNodeManagement(t *testing.T) {
	router := NewMessageRouter()
	defer router.Stop()

	// Add P2P node
	router.AddP2PNode("peer_1", "http://peer1:8080")

	// Test remote server call (should fail since P2P routing not fully implemented)
	req := &RouteRequest{
		ID:         "test_3",
		From:       "test_client",
		NodeID:     "peer_1", // Remote call
		ServerName: "remote_server",
		Method:     "test",
		Args:       []*Value{},
		Timeout:    5 * time.Second,
	}

	response := router.RouteMessage(req)
	if response.Success {
		t.Fatal("Expected error for remote call (not yet implemented)")
	}

	// Remove P2P node
	router.RemoveP2PNode("peer_1")
}

func TestMessageRouter_CallServerMethod(t *testing.T) {
	router := NewMessageRouter()
	defer router.Stop()

	// Create test server
	server := createTestServerWithMethods("calc_server", map[string]func([]*Value) *Value{
		"add": func(args []*Value) *Value {
			if len(args) >= 2 {
				a := args[0].Number
				b := args[1].Number
				return NewNumber(a + b)
			}
			return NewNumber(0)
		},
	})
	router.RegisterServer("calc_server", server)

	// Test CallServer convenience method
	args := []*Value{NewNumber(5), NewNumber(3)}
	result, err := router.CallServer("", "calc_server", "add", args, 5*time.Second)
	if err != nil {
		t.Fatalf("Expected successful call, got error: %v", err)
	}
	if result.Number != 8 {
		t.Fatalf("Expected 8, got %f", result.Number)
	}
}

func TestMessageRouter_Timeout(t *testing.T) {
	router := NewMessageRouter()
	router.Start()
	defer router.Stop()

	// Give router time to start
	time.Sleep(10 * time.Millisecond)

	// Create a slow server that doesn't respond
	server := createTestServerWithMethods("slow_server", map[string]func([]*Value) *Value{
		"slow": func(args []*Value) *Value {
			time.Sleep(2 * time.Second) // Longer than timeout
			return NewString("done")
		},
	})
	router.RegisterServer("slow_server", server)

	// Give time for registration
	time.Sleep(10 * time.Millisecond)

	// Test timeout
	args := []*Value{}
	result, err := router.CallServer("", "slow_server", "slow", args, 100*time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if result != nil {
		t.Fatal("Expected nil result on timeout")
	}
}

// Helper functions for tests

func createTestServerValue(name string) *Value {
	return &Value{
		Type: ValueTypeServer,
		Server: &Server{
			Name:        name,
			State:       make(map[string]*Value),
			Receivers:   make(map[string]*Function),
			MessageChan: make(chan *Message, 100),
			Running:     true,
			Environment: NewEnvironment(nil),
		},
	}
}

func createTestServerWithMethods(name string, methods map[string]func([]*Value) *Value) *Value {
	server := &Server{
		Name:        name,
		State:       make(map[string]*Value),
		Receivers:   make(map[string]*Function),
		MessageChan: make(chan *Message, 100),
		Running:     true,
		Environment: NewEnvironment(nil),
	}

	// Add methods as receivers
	for methodName, handler := range methods {
		// Capture the handler in the closure
		h := handler
		server.Receivers[methodName] = &Function{
			Name:      methodName,
			IsBuiltin: true,
			Builtin: func(args []*Value) (*Value, error) {
				return h(args), nil
			},
		}
	}

	// Start a simple message handler
	go func() {
		for server.Running {
			select {
			case message, ok := <-server.MessageChan:
				if !ok {
					return
				}
				if handler, exists := methods[message.Method]; exists {
					result := handler(message.Args)
					if message.Reply != nil {
						message.Reply <- result
					}
				} else {
					if message.Reply != nil {
						message.Reply <- NewNil()
					}
				}
			}
		}
	}()

	return &Value{
		Type:   ValueTypeServer,
		Server: server,
	}
}

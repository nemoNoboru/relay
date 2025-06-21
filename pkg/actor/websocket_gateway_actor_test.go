package actor

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketGatewayActor(t *testing.T) {
	router := NewRouter()
	defer router.StopAll()

	gateway := NewWebSocketGatewayActor("ws-gateway", "", router) // Use httptest.Server URL

	server := httptest.NewServer(gateway)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create a mock receiver actor
	receiver := NewActor("test-receiver", router, func(msg ActorMsg) {
		t.Logf("Test receiver got message: %v", msg)
	})
	receiver.Start()
	defer receiver.Stop()

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// Send a message
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "test-receiver",
		Params:  []interface{}{"hello from ws"},
		ID:      2,
	}
	if err := ws.WriteJSON(req); err != nil {
		t.Fatalf("Failed to write JSON to WebSocket: %v", err)
	}

	// Read the response
	var resp JSONRPCResponse
	if err := ws.ReadJSON(&resp); err != nil {
		t.Fatalf("Failed to read JSON from WebSocket: %v", err)
	}

	// Check the response
	if resp.Result != "message sent" {
		t.Errorf("unexpected result: got %v, want 'message sent'", resp.Result)
	}
	if resp.ID.(float64) != 2 {
		t.Errorf("unexpected ID: got %v, want 2", resp.ID)
	}
}

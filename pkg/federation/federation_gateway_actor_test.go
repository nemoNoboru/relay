package federation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"relay/pkg/actor"

	"github.com/gorilla/websocket"
)

func TestFederationGatewayActorLifecycle(t *testing.T) {
	router := actor.NewRouter()
	defer router.StopAll()

	gateway := NewFederationGatewayActor("test-gateway", router)
	gateway.Start()

	// Allow a moment for the actor to start
	time.Sleep(10 * time.Millisecond)

	// Send a dummy message to ensure it doesn't crash
	router.Send(actor.ActorMsg{
		To:   "test-gateway",
		From: "test",
		Type: "ping",
	})

	// Give it a moment to process the message
	time.Sleep(10 * time.Millisecond)

	// The StopAll call in defer will handle stopping the actor.
	// This test mainly ensures the actor can be initialized and started
	// without panicking.
}

func TestFederationGatewayActor_ConnectToPeer(t *testing.T) {
	// Create a mock server that will act as the "main" relay.
	// This server will upgrade the connection to a WebSocket.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Mock server failed to upgrade connection: %v", err)
		}
		// We don't need to do anything with the connection here, just establish it.
	}))
	defer server.Close()

	// 2. Create the gateway actor we are testing
	router := actor.NewRouter()
	defer router.StopAll()

	homeGateway := NewFederationGatewayActor("home-gateway", router)
	homeGateway.Start()

	// 3. Send a message to the gateway telling it to connect to the mock server
	// The URL needs to be converted from http to ws
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	connectMsg := actor.ActorMsg{
		To:   "home-gateway",
		From: "test",
		Type: "connect_to_peer",
		Data: wsURL,
	}
	router.Send(connectMsg)

	// 4. Assert that the gateway has established the connection
	time.Sleep(50 * time.Millisecond) // Give time for connection to happen

	if len(homeGateway.GetPeers()) != 1 {
		t.Fatalf("Expected 1 peer connection, got %d", len(homeGateway.GetPeers()))
	}

	peers := homeGateway.GetPeers()
	if _, ok := peers[wsURL]; !ok {
		t.Errorf("Expected peer with URL %s not found", wsURL)
	}
}

func TestFederationGatewayActor_AcceptsPeerConnection(t *testing.T) {
	// 1. Create the 'main' gateway and start it listening on a random port.
	router := actor.NewRouter()
	defer router.StopAll()

	mainGateway := NewFederationGatewayActor("main-gateway", router)
	// We need a way to tell the gateway to start listening.
	// This will fail until we implement the StartListening method.
	err := mainGateway.StartListening("127.0.0.1:0") // :0 means random port
	if err != nil {
		t.Fatalf("Main gateway failed to start listening: %v", err)
	}
	defer mainGateway.StopListening()

	// 2. Create the 'home' gateway that will connect to the main one.
	homeGateway := NewFederationGatewayActor("home-gateway", router)
	homeGateway.Start()

	// 3. Send a message to the home gateway telling it to connect.
	connectMsg := actor.ActorMsg{
		To:   "home-gateway",
		From: "test",
		Type: "connect_to_peer",
		Data: mainGateway.GetListenURL(), // This method will also need to be created.
	}
	router.Send(connectMsg)

	// 4. Assert that the main gateway has registered the incoming connection.
	time.Sleep(50 * time.Millisecond) // Give time for connection to happen

	if len(mainGateway.GetPeers()) != 1 {
		t.Fatalf("Expected main gateway to have 1 peer, got %d", len(mainGateway.GetPeers()))
	}
}

func TestFederationGatewayActor_ForwardsMessagesBetweenPeers(t *testing.T) {
	// 1. Setup: main gateway, home gateway, and a listener actor on the 'main' side.
	router := actor.NewRouter()
	defer router.StopAll()

	// The listener actor will just store any message it receives.
	var receivedMsg actor.ActorMsg
	var mu sync.Mutex
	listenerActor := actor.NewActor("listener", router, func(msg actor.ActorMsg) {
		mu.Lock()
		defer mu.Unlock()
		receivedMsg = msg
	})
	listenerActor.Start()

	// Main gateway listens for connections.
	mainGateway := NewFederationGatewayActor("main-gateway", router)
	mainGateway.Start()
	err := mainGateway.StartListening("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Main gateway failed to start listening: %v", err)
	}
	defer mainGateway.StopListening()

	// Home gateway connects to main.
	homeGateway := NewFederationGatewayActor("home-gateway", router)
	homeGateway.Start()

	connectMsg := actor.ActorMsg{
		To:   "home-gateway",
		From: "test",
		Type: "connect_to_peer",
		Data: mainGateway.GetListenURL(),
	}
	router.Send(connectMsg)
	time.Sleep(100 * time.Millisecond) // allow for connection

	// 2. Action: Send a message from the 'home' side to be forwarded to the 'listener' on the 'main' side.
	msgToSend := actor.ActorMsg{
		To:   "listener",
		From: "home-test",
		Type: "test_message",
		Data: "hello world",
	}

	forwardMsg := actor.ActorMsg{
		To:   "home-gateway",
		From: "test",
		Type: "forward_message", // This is a new type we'll have to handle
		Data: msgToSend,
	}
	router.Send(forwardMsg)

	// 3. Assert: The listener actor on the 'main' side should have received the message.
	time.Sleep(100 * time.Millisecond) // allow for forwarding

	mu.Lock()
	defer mu.Unlock()

	if receivedMsg.To != "listener" {
		t.Errorf("Listener expected message for 'listener', but got '%s'", receivedMsg.To)
	}

	if receivedMsg.Type != "test_message" {
		t.Errorf("Listener expected message type 'test_message', but got '%s'", receivedMsg.Type)
	}

	// Because ActorMsg.Data is interface{}, it gets deserialized into a map[string]interface{}
	// So we need to handle that to check the value.
	data, ok := receivedMsg.Data.(string)
	if !ok {
		t.Fatalf("Received message data is not a string, got %T", receivedMsg.Data)
	}

	if data != "hello world" {
		t.Errorf("Listener expected message data 'hello world', but got '%v'", data)
	}
}

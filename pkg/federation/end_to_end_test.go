package federation_test

import (
	"log"
	"sync"
	"testing"
	"time"

	"relay/pkg/actor"
	"relay/pkg/federation"
)

func TestTwoNodeCommunication(t *testing.T) {
	// NODE A ("main") SETUP
	// =========================================================================
	routerA := actor.NewRouter()
	defer routerA.StopAll()

	gatewayA := federation.NewFederationGatewayActor("gateway-a", routerA)
	gatewayA.Start()
	if err := gatewayA.StartListening("127.0.0.1:0"); err != nil {
		t.Fatalf("Node A gateway failed to start listening: %v", err)
	}
	defer gatewayA.StopListening()

	// The responder actor waits for a "ping" and sends back a "pong".
	responder := actor.NewActor("responder", routerA, func(msg actor.ActorMsg) {
		if msg.Type == "ping" {
			log.Printf("[Responder A] Received ping from %s. Sending pong.", msg.From)
			reply := actor.ActorMsg{
				To:   msg.From, // Send back to the original sender
				From: "responder",
				Type: "pong",
				Data: "hello from node A",
			}
			// To send it back, it must go through the gateway
			forwardMsg := actor.ActorMsg{To: "gateway-a", From: "responder", Type: "forward_message", Data: reply}
			routerA.Send(forwardMsg)
		}
	})
	responder.Start()

	// NODE B ("home") SETUP
	// =========================================================================
	routerB := actor.NewRouter()
	defer routerB.StopAll()

	gatewayB := federation.NewFederationGatewayActor("gateway-b", routerB)
	gatewayB.Start()

	// The requester actor sends the initial "ping" and waits for a "pong".
	// It uses a channel to signal the main test goroutine when it's done.
	var wg sync.WaitGroup
	wg.Add(1)
	var finalData interface{}

	requester := actor.NewActor("requester", routerB, func(msg actor.ActorMsg) {
		if msg.Type == "pong" {
			log.Printf("[Requester B] Received pong from %s!", msg.From)
			finalData = msg.Data
			wg.Done()
		}
	})
	requester.Start()

	// CONNECTION & COMMUNICATION
	// =========================================================================
	// 1. Tell Node B's gateway to connect to Node A's gateway.
	connectMsg := actor.ActorMsg{
		To:   "gateway-b",
		From: "test",
		Type: "connect_to_peer",
		Data: gatewayA.GetListenURL(),
	}
	routerB.Send(connectMsg)
	time.Sleep(100 * time.Millisecond) // Allow time for WebSocket handshake

	// 2. The requester sends the first message, addressed to the responder.
	pingMsg := actor.ActorMsg{
		To:   "responder",
		From: "requester",
		Type: "ping",
		Data: "hello from node B",
	}
	// It sends this message via its local gateway.
	forwardPing := actor.ActorMsg{To: "gateway-b", From: "requester", Type: "forward_message", Data: pingMsg}
	routerB.Send(forwardPing)

	// ASSERTION
	// =========================================================================
	// Wait for the pong to be received, with a timeout.
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// Success!
		if finalData != "hello from node A" {
			t.Errorf("Expected final data to be 'hello from node A', but got '%v'", finalData)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for pong message.")
	}
}

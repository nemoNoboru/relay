package actor

import (
	"relay/pkg/runtime"
	"testing"
	"time"
)

func TestRelayServerActorEvaluatesCode(t *testing.T) {
	router := NewRouter()
	defer router.StopAll()

	// 1. Create the actor under test
	relayServer := NewRelayServerActor("relay-server", "", "", router, nil)
	relayServer.Start()

	// 2. Send the message to evaluate code with a reply channel
	replyChan := make(chan ActorMsg, 1)
	evalMsg := ActorMsg{
		To:        "relay-server",
		From:      "test",
		Type:      "eval",
		Data:      "10 + 5",
		ReplyChan: replyChan,
	}
	router.Send(evalMsg)

	// 3. Wait for the result and assert
	select {
	case reply := <-replyChan:
		result, ok := reply.Data.(*runtime.Value)
		if !ok {
			t.Fatalf("Expected *runtime.Value, got %T", reply.Data)
		}

		if result.Type != runtime.ValueTypeNumber {
			t.Errorf("Expected result type Number, got %s", result.Type)
		}
		if result.Number != 15 {
			t.Errorf("Expected result to be 15, got %f", result.Number)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for eval result")
	}
}

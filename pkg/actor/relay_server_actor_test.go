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
	relayServer := NewRelayServerActor("relay-server", router)
	relayServer.Start()

	// 2. Create a "probe" actor to send the message and receive the reply
	resultChan := make(chan *runtime.Value, 1)
	probe := NewActor("test-probe", router, func(msg ActorMsg) {
		if msg.Type == "eval_result" {
			result, ok := msg.Data.(*runtime.Value)
			if !ok {
				t.Errorf("Expected *runtime.Value, got %T", msg.Data)
				resultChan <- nil
				return
			}
			resultChan <- result
		}
	})
	probe.Start()

	// 3. Send the message to evaluate code
	evalMsg := ActorMsg{
		To:   "relay-server",
		From: "test-probe",
		Type: "eval",
		Data: "2 + 2",
	}
	router.Send(evalMsg)

	// 4. Wait for the result and assert
	select {
	case result := <-resultChan:
		if result == nil {
			t.Fatal("Test failed to get a valid result.")
		}
		if result.Type != runtime.ValueTypeNumber {
			t.Errorf("Expected result type Number, got %s", result.Type)
		}
		if result.Number != 4 {
			t.Errorf("Expected result to be 4, got %f", result.Number)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for eval result")
	}
}

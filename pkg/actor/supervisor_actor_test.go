package actor

import (
	"testing"
	"time"
)

func TestSupervisorActor(t *testing.T) {
	router := NewRouter()
	supervisor := NewSupervisorActor("supervisor", router)

	supervisor.Start()
	defer supervisor.Stop()

	// Simple test to ensure the supervisor can start and stop.
	// We'll add more sophisticated tests for message routing and actor creation next.

	// Send a message and check for a log, for example
	testMsg := ActorMsg{
		To:   "supervisor",
		From: "test",
		Type: "test_message",
		Data: "hello",
	}
	router.Send(testMsg)

	// Allow some time for the message to be processed
	time.Sleep(10 * time.Millisecond)

	// In a real test, we would check for some side effect,
	// but for now, we're just ensuring it doesn't crash.
}

func TestSupervisorCreatesAndManagesActors(t *testing.T) {
	router := NewRouter()
	defer router.StopAll()

	supervisor := NewSupervisorActor("supervisor", router)
	supervisor.Start()

	// Use a probe actor to receive the reply from the supervisor
	replyChan := make(chan ActorMsg, 1)
	probe := NewActor("test-probe", router, func(msg ActorMsg) {
		if msg.Type == "child_created" {
			replyChan <- msg
		}
	})
	probe.Start()
	defer probe.Stop()

	// 1. Send a message to the supervisor to create a new actor
	createMsg := ActorMsg{
		To:   "supervisor",
		From: "test-probe",
		Type: "create_child",
		Data: "RelayServerActor",
	}
	router.Send(createMsg)

	// 2. Wait for the reply and assert
	var childName string
	select {
	case reply := <-replyChan:
		var ok bool
		childName, ok = reply.Data.(string)
		if !ok {
			t.Fatal("Supervisor reply did not contain a string for the child's name")
		}
		if childName == "" {
			t.Fatal("Supervisor returned an empty name for the child")
		}

		// Check that the actor exists in the supervisor's map
		supervisor.mu.RLock()
		_, exists := supervisor.actors[childName]
		supervisor.mu.RUnlock()
		if !exists {
			t.Errorf("Supervisor claims to have created '%s', but it's not in its map", childName)
		}

	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out waiting for supervisor to create child")
	}

	// 3. Tell the supervisor to stop the child
	stopMsg := ActorMsg{
		To:   "supervisor",
		From: "test-probe",
		Type: "stop_child",
		Data: childName,
	}
	router.Send(stopMsg)

	time.Sleep(50 * time.Millisecond) // Allow time for stop message to be processed

	// 4. Assert the child was removed
	supervisor.mu.RLock()
	_, exists := supervisor.actors[childName]
	supervisor.mu.RUnlock()
	if exists {
		t.Errorf("Supervisor did not remove stopped child '%s' from its map", childName)
	}
}

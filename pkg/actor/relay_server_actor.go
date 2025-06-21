package actor

import (
	"log"
	"relay/pkg/parser"
	"relay/pkg/runtime"
	"strings"
)

// RelayServerActor encapsulates a Relay language runtime evaluator in an actor.
type RelayServerActor struct {
	actor     *Actor
	evaluator *runtime.Evaluator
}

// NewRelayServerActor creates a new RelayServerActor.
func NewRelayServerActor(name string, router *Router) *RelayServerActor {
	rsa := &RelayServerActor{
		evaluator: runtime.NewEvaluator(),
	}
	rsa.actor = NewActor(name, router, rsa.Receive)
	return rsa
}

// Start begins the actor's message processing loop.
func (rsa *RelayServerActor) Start() {
	rsa.actor.Start()
}

// Stop terminates the actor's message processing loop.
func (rsa *RelayServerActor) Stop() {
	rsa.actor.Stop()
}

// Receive handles incoming messages for the RelayServerActor.
func (rsa *RelayServerActor) Receive(msg ActorMsg) {
	switch msg.Type {
	case "eval":
		code, ok := msg.Data.(string)
		if !ok {
			log.Printf("RelayServerActor %s: invalid data for 'eval', expected string", rsa.actor.Name)
			return
		}

		program, err := parser.Parse("eval", strings.NewReader(code))
		if err != nil {
			// In a real implementation, we'd send back a proper error message.
			log.Printf("RelayServerActor %s: parse error: %v", rsa.actor.Name, err)
			return
		}

		var lastResult *runtime.Value
		for _, expr := range program.Expressions {
			lastResult, err = rsa.evaluator.Evaluate(expr)
			if err != nil {
				log.Printf("RelayServerActor %s: evaluation error: %v", rsa.actor.Name, err)
				// Send error back
				return
			}
		}

		// Send the result of the last expression back to the caller.
		rsa.actor.router.Send(ActorMsg{
			To:   msg.From,
			From: rsa.actor.Name,
			Type: "eval_result",
			Data: lastResult,
		})

	default:
		// Default behavior, maybe log or send an "unknown_type" error.
		log.Printf("RelayServerActor %s received unhandled message type: %s", rsa.actor.Name, msg.Type)
	}
}

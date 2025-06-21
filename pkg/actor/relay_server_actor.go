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
			// If there's a reply channel, send an error back.
			if msg.ReplyChan != nil {
				msg.ReplyChan <- ActorMsg{Data: "invalid data for 'eval'"}
			}
			return
		}

		program, err := parser.Parse("eval", strings.NewReader(code))
		if err != nil {
			log.Printf("RelayServerActor %s: parse error: %v", rsa.actor.Name, err)
			if msg.ReplyChan != nil {
				msg.ReplyChan <- ActorMsg{Data: err}
			}
			return
		}

		var lastResult *runtime.Value
		for _, expr := range program.Expressions {
			lastResult, err = rsa.evaluator.Evaluate(expr)
			if err != nil {
				log.Printf("RelayServerActor %s: evaluation error: %v", rsa.actor.Name, err)
				if msg.ReplyChan != nil {
					msg.ReplyChan <- ActorMsg{Data: err}
				}
				return // Stop on the first error
			}
		}

		// Send the result of the last expression back to the caller if a reply
		// channel was provided.
		if msg.ReplyChan != nil {
			msg.ReplyChan <- ActorMsg{Data: lastResult}
		}

	default:
		// Default behavior, maybe log or send an "unknown_type" error.
		log.Printf("RelayServerActor %s received unhandled message type: %s", rsa.actor.Name, msg.Type)
	}
}

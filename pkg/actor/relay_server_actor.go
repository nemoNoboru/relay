package actor

import (
	"fmt"
	"log"
	"relay/pkg/parser"
	"relay/pkg/runtime"
	"strings"
)

// RelayServerActor encapsulates a Relay language runtime.
type RelayServerActor struct {
	*Actor
	eval           *runtime.Evaluator
	gatewayName    string                       // Name of the gateway for this actor's node
	supervisorName string                       // Name of this actor's supervisor
	receives       map[string]*runtime.Function // Relay functions to handle messages
}

// NewRelayServerActor creates a new RelayServerActor.
func NewRelayServerActor(name, gatewayName, supervisorName string, router *Router, initData *runtime.ServerInitData) *RelayServerActor {
	s := &RelayServerActor{
		gatewayName:    gatewayName,
		supervisorName: supervisorName,
		receives:       make(map[string]*runtime.Function),
	}
	s.Actor = NewActor(name, router, s.Receive)

	// The serverCreator callback allows the evaluator to ask us (the actor)
	// to create a new server, which we do by messaging our supervisor.
	serverCreator := func(data runtime.ServerInitData) {
		createMsg := NewCreateChildMsg(s.supervisorName, s.Name, "RelayServerActor", data, make(chan ActorMsg))
		s.router.Send(createMsg)
		// We could wait for the reply here if needed
	}

	s.eval = runtime.NewEvaluator(serverCreator)
	s.eval.SetGlobal("send", s.makeSendBuiltin())

	// If this actor is being created from a `server {}` block, initialize it.
	if initData != nil {
		s.receives = initData.Receives
		// TODO: Initialize state from initData.State into the evaluator's global env
	}

	return s
}

// Start begins the actor's message processing loop.
func (s *RelayServerActor) Start() {
	s.Actor.Start()
}

// Stop terminates the actor's message processing loop.
func (s *RelayServerActor) Stop() {
	s.Actor.Stop()
}

// makeSendBuiltin creates a closure that acts as the 'send' function in Relay.
// This function is aware of the actor system and can forward messages to the gateway.
func (s *RelayServerActor) makeSendBuiltin() *runtime.Value {
	return &runtime.Value{
		Type: runtime.ValueTypeFunction,
		Function: &runtime.Function{
			Name:      "send",
			IsBuiltin: true,
			Builtin: func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("send() requires 2 arguments: destination (string) and message (any)")
				}
				dest, ok := args[0].Str, (args[0].Type == runtime.ValueTypeString)
				if !ok {
					return nil, fmt.Errorf("send() first argument must be a string destination")
				}
				// The second argument is the message payload. We pass the *runtime.Value
				// directly. The actor message's Data field is interface{}, so it fits.
				// The JSON marshaller will handle serialization when it's sent over the network.
				messageData := args[1]
				parts := strings.Split(dest, ".")
				if len(parts) != 2 {
					// Handle error - not exactly 2 parts
					return nil, fmt.Errorf("expected format: to.function, got: %s", dest)
				}

				to, fun := parts[0], parts[1]
				replyChan := make(chan ActorMsg)

				msg := NewEvalMsg(to, s.Name, fun, replyChan)
				msg.Data = messageData

				if !s.router.HasActor(dest) && s.gatewayName != "" {
					log.Printf("[%s] Forwarding message to remote actor '%s' via gateway '%s'", s.Name, dest, s.gatewayName)
					forwardMsg := NewForwardMessageMsg(s.gatewayName, s.Name, msg)
					s.router.Send(forwardMsg)
				} else {
					s.router.Send(msg)
				}

				reply := <-replyChan
				if reply.Type == "receive_result" {
					return reply.Data.(*runtime.Value), nil
				}
				return nil, fmt.Errorf("received an unknown error from remote call: %v", reply)
			},
		},
	}
}

// Receive handles messages for the RelayServerActor.
func (s *RelayServerActor) Receive(msg ActorMsg) {
	log.Printf("RelayServerActor %s received message: %v", s.Name, msg)
	switch msg.Type {
	case "eval":
		s.handleEval(msg)
	case "stop":
		s.Stop()
	default:
		// This is not an internal message, so try to handle it with a `receive` block.
		s.handleReceive(msg)
	}
}

func (s *RelayServerActor) handleEval(msg ActorMsg) {
	code, ok := msg.Data.(string)
	if !ok {
		log.Printf("RelayServerActor %s: invalid data for 'eval', expected string", s.Name)
		if msg.ReplyChan != nil {
			reply := NewEvalErrorMsg(msg.From, s.Name, "Invalid eval data type")
			msg.ReplyChan <- reply
		}
		return
	}

	program, err := parser.Parse("eval", strings.NewReader(code))
	if err != nil {
		log.Printf("RelayServerActor %s: parse error: %v", s.Name, err)
		if msg.ReplyChan != nil {
			reply := NewEvalErrorMsg(msg.From, s.Name, err.Error())
			msg.ReplyChan <- reply
		}
		return
	}

	var lastResult *runtime.Value
	for _, expr := range program.Expressions {
		// Use the actor's main evaluator
		lastResult, err = s.eval.Evaluate(expr)
		if err != nil {
			log.Printf("RelayServerActor %s: evaluation error: %v", s.Name, err)
			if msg.ReplyChan != nil {
				reply := NewEvalErrorMsg(msg.From, s.Name, err.Error())
				msg.ReplyChan <- reply
			}
			return // Stop on the first error
		}
	}

	if msg.ReplyChan != nil {
		reply := NewEvalResultMsg(msg.From, s.Name, lastResult)
		msg.ReplyChan <- reply
	}
}

func (s *RelayServerActor) handleReceive(msg ActorMsg) {
	receiveFn, ok := s.receives[msg.Type]
	if !ok {
		log.Printf("RelayServerActor %s received unhandled message type: %s", s.Name, msg.Type)
		return
	}

	// Convert the message data into arguments for the Relay function.
	// We expect the arguments to be in a map where the key is the parameter name.
	argsObj, ok := msg.Data.(*runtime.Value)
	if !ok {
		log.Printf("RelayServerActor %s: expected object data for receive function, got %T", s.Name, msg.Data)
		return
	}

	if argsObj.Type != runtime.ValueTypeObject {
		log.Printf("RelayServerActor %s: expected object data for receive function, got %T", s.Name, argsObj.Type)
		return
	}

	// Match the map values to the function's declared parameters by name.

	args := make([]*runtime.Value, len(receiveFn.Parameters))
	log.Printf("receiveFn.Parameters: %v", receiveFn.Parameters)
	for i, paramName := range receiveFn.Parameters {
		if val, exists := argsObj.Object[paramName]; exists {
			args[i] = val
		} else {
			args[i] = runtime.NewNil()
		}
	}

	// Execute the Relay function.
	result, err := s.eval.ExecuteFunction(
		&runtime.Value{Type: runtime.ValueTypeFunction, Function: receiveFn},
		args,
	)

	if err != nil {
		log.Printf("RelayServerActor %s error executing receive function '%s': %v", s.Name, msg.Type, err)
		// Optionally, send an error reply
		if msg.ReplyChan != nil {
			reply := NewReceiveErrorMsg(msg.From, s.Name, err.Error())
			msg.ReplyChan <- reply
		}
		return
	}

	// If the Relay code returned a value and there's a reply channel, send it back.
	if msg.ReplyChan != nil {
		// The result of the receive function is the reply.
		reply := NewReceiveResultMsg(msg.From, s.Name, result)
		msg.ReplyChan <- reply
	}
}

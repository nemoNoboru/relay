package runtime

import (
	"fmt"
	"time"
)

// ServerCore represents the core server functionality, decoupled from parser/evaluator
type ServerCore struct {
	Name        string
	State       map[string]*Value
	Receivers   map[string]FunctionHandler
	MessageChan chan *Message
	Running     bool
	Environment *Environment
	Engine      ExecutionEngine
}

// NewServerCore creates a new server core instance
func NewServerCore(name string, state map[string]*Value, receivers map[string]FunctionHandler, env *Environment, engine ExecutionEngine) *ServerCore {
	return &ServerCore{
		Name:        name,
		State:       state,
		Receivers:   receivers,
		MessageChan: make(chan *Message, 100),
		Environment: env,
		Engine:      engine,
	}
}

// Start starts the server goroutine
func (s *ServerCore) Start() {
	if s.Running {
		return
	}

	s.Running = true
	go s.runServerLoop()
}

// Stop stops the server goroutine
func (s *ServerCore) Stop() {
	if !s.Running {
		return
	}

	s.Running = false
	close(s.MessageChan)
}

// SendMessage sends a message to the server and optionally waits for a reply
func (s *ServerCore) SendMessage(method string, args []*Value, waitForReply bool) (*Value, error) {
	if !s.Running {
		return nil, fmt.Errorf("server %s is not running", s.Name)
	}

	var replyChan chan *Value
	if waitForReply {
		replyChan = make(chan *Value, 1)
	}

	message := &Message{
		Method: method,
		Args:   args,
		Reply:  replyChan,
	}

	select {
	case s.MessageChan <- message:
		if waitForReply {
			select {
			case reply := <-replyChan:
				return reply, nil
			case <-time.After(5 * time.Second): // Timeout
				return nil, fmt.Errorf("timeout waiting for reply from server %s", s.Name)
			}
		}
		return NewNil(), nil
	case <-time.After(1 * time.Second):
		return nil, fmt.Errorf("failed to send message to server %s: channel full", s.Name)
	}
}

// runServerLoop runs the main server message handling loop
func (s *ServerCore) runServerLoop() {
	for s.Running {
		select {
		case message, ok := <-s.MessageChan:
			if !ok {
				return // Channel closed
			}
			s.handleMessage(message)
		}
	}
}

// handleMessage processes an incoming message using the abstract execution engine
func (s *ServerCore) handleMessage(message *Message) {
	// Find the function handler
	handler, exists := s.Receivers[message.Method]
	if !exists {
		if message.Reply != nil {
			message.Reply <- NewNil() // Send nil as error
		}
		return
	}

	// Create server environment with state access
	serverEnv := s.Engine.CreateEnvironment(s.Environment)

	// Add state variable to environment (no mutex needed - server processes messages sequentially)
	stateValue := NewServerStateActorSafe(&s.State)
	serverEnv.Define("state", stateValue)

	// Execute the function using the abstract execution engine
	result, err := s.Engine.ExecuteFunction(handler, message.Args, serverEnv)

	if err != nil {
		if message.Reply != nil {
			message.Reply <- NewNil() // Send nil on error
		}
		return
	}

	if message.Reply != nil {
		message.Reply <- result
	}
}

// GetState safely gets a state variable (no mutex needed - called from single goroutine)
func (s *ServerCore) GetState(key string) (*Value, bool) {
	value, exists := s.State[key]
	return value, exists
}

// SetState safely sets a state variable (no mutex needed - called from single goroutine)
func (s *ServerCore) SetState(key string, value *Value) {
	s.State[key] = value
}

// GetReceiverNames returns a list of receiver method names for debugging
func (s *ServerCore) GetReceiverNames() []string {
	names := make([]string, 0, len(s.Receivers))
	for name := range s.Receivers {
		names = append(names, name)
	}
	return names
}

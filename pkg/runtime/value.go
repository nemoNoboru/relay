package runtime

import (
	"fmt"
	"relay/pkg/parser"
	"strconv"
	"sync"
	"time"
)

// Environment represents a variable scope with optional parent chain
type Environment struct {
	variables map[string]*Value
	parent    *Environment
}

// NewEnvironment creates a new environment with an optional parent
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables: make(map[string]*Value),
		parent:    parent,
	}
}

// Get looks up a variable in the environment chain
func (e *Environment) Get(name string) (*Value, bool) {
	if value, exists := e.variables[name]; exists {
		return value, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set updates an existing variable in the environment chain
func (e *Environment) Set(name string, value *Value) {
	e.variables[name] = value
}

// Define creates a new variable in the current environment
func (e *Environment) Define(name string, value *Value) {
	e.variables[name] = value
}

// ValueType represents the type of a runtime value
type ValueType int

const (
	ValueTypeNil ValueType = iota
	ValueTypeNumber
	ValueTypeString
	ValueTypeBool
	ValueTypeArray
	ValueTypeObject
	ValueTypeFunction
	ValueTypeStruct
	ValueTypeServer
	ValueTypeServerState
)

// String returns a string representation of the ValueType
func (vt ValueType) String() string {
	switch vt {
	case ValueTypeNil:
		return "nil"
	case ValueTypeNumber:
		return "number"
	case ValueTypeString:
		return "string"
	case ValueTypeBool:
		return "bool"
	case ValueTypeArray:
		return "array"
	case ValueTypeObject:
		return "object"
	case ValueTypeFunction:
		return "function"
	case ValueTypeStruct:
		return "struct"
	case ValueTypeServer:
		return "server"
	case ValueTypeServerState:
		return "serverstate"
	default:
		return "unknown"
	}
}

// Value represents a runtime value in the Relay language
type Value struct {
	Type        ValueType
	Number      float64
	Str         string
	Bool        bool
	Array       []*Value
	Object      map[string]*Value
	Function    *Function    // For functions and lambdas
	Struct      *Struct      // For struct instances
	Server      *Server      // For server instances
	ServerState *ServerState // For mutable server state
}

// Function represents a callable function
type Function struct {
	Name       string
	Parameters []string
	Body       *parser.Block // AST block for user-defined functions
	IsBuiltin  bool
	Builtin    func(args []*Value) (*Value, error)
	ClosureEnv *Environment // Captured environment for closures
}

// Struct represents a struct instance
type Struct struct {
	Name   string            // Struct type name (e.g., "User")
	Fields map[string]*Value // Field values
}

// StructDefinition represents a struct type definition
type StructDefinition struct {
	Name   string            // Struct name
	Fields map[string]string // Field name -> type name mapping
}

// Message represents a message sent to a server
type Message struct {
	Method string      // The receive function to call
	Args   []*Value    // Arguments for the function
	Reply  chan *Value // Channel to send reply back (optional)
}

// Server represents a running server instance
type Server struct {
	Name        string               // Server name
	State       map[string]*Value    // Server state variables
	Receivers   map[string]*Function // Receive function handlers
	MessageChan chan *Message        // Channel for incoming messages
	StateMutex  sync.RWMutex         // Protects state access
	Running     bool                 // Whether server is running
	Environment *Environment         // Server's environment
}

// ServerState represents mutable server state that behaves like an object but allows in-place updates
type ServerState struct {
	State *map[string]*Value // Pointer to the actual server state map
	Mutex *sync.RWMutex      // Pointer to the server's state mutex
}

// NewNumber creates a new number value
func NewNumber(n float64) *Value {
	return &Value{
		Type:   ValueTypeNumber,
		Number: n,
	}
}

// NewString creates a new string value
func NewString(s string) *Value {
	return &Value{
		Type: ValueTypeString,
		Str:  s,
	}
}

// NewBool creates a new boolean value
func NewBool(b bool) *Value {
	return &Value{
		Type: ValueTypeBool,
		Bool: b,
	}
}

// NewNil creates a new nil value
func NewNil() *Value {
	return &Value{
		Type: ValueTypeNil,
	}
}

// NewArray creates a new array value
func NewArray(elements []*Value) *Value {
	return &Value{
		Type:  ValueTypeArray,
		Array: elements,
	}
}

// NewObject creates a new object value
func NewObject(fields map[string]*Value) *Value {
	return &Value{
		Type:   ValueTypeObject,
		Object: fields,
	}
}

// NewStruct creates a new struct instance
func NewStruct(name string, fields map[string]*Value) *Value {
	return &Value{
		Type: ValueTypeStruct,
		Struct: &Struct{
			Name:   name,
			Fields: fields,
		},
	}
}

// NewServer creates a new server instance
func NewServer(name string, state map[string]*Value, receivers map[string]*Function, env *Environment) *Value {
	server := &Server{
		Name:        name,
		State:       state,
		Receivers:   receivers,
		MessageChan: make(chan *Message, 100), // Buffered channel
		Running:     false,
		Environment: env,
	}

	return &Value{
		Type:   ValueTypeServer,
		Server: server,
	}
}

// NewServerState creates a new server state value that allows mutable operations
func NewServerState(state *map[string]*Value, mutex *sync.RWMutex) *Value {
	return &Value{
		Type: ValueTypeServerState,
		ServerState: &ServerState{
			State: state,
			Mutex: mutex,
		},
	}
}

// String returns a string representation of the value
func (v *Value) String() string {
	switch v.Type {
	case ValueTypeNil:
		return "nil"
	case ValueTypeNumber:
		// Format nicely - show integers without decimal point
		if v.Number == float64(int64(v.Number)) {
			return strconv.FormatInt(int64(v.Number), 10)
		}
		return strconv.FormatFloat(v.Number, 'g', -1, 64)
	case ValueTypeString:
		return fmt.Sprintf(`"%s"`, v.Str)
	case ValueTypeBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case ValueTypeArray:
		result := "["
		for i, elem := range v.Array {
			if i > 0 {
				result += ", "
			}
			result += elem.String()
		}
		result += "]"
		return result
	case ValueTypeObject:
		result := "{"
		first := true
		for key, value := range v.Object {
			if !first {
				result += ", "
			}
			result += fmt.Sprintf(`%s: %s`, key, value.String())
			first = false
		}
		result += "}"
		return result
	case ValueTypeFunction:
		if v.Function.IsBuiltin {
			return fmt.Sprintf("<builtin function: %s>", v.Function.Name)
		}
		return fmt.Sprintf("<function: %s>", v.Function.Name)
	case ValueTypeStruct:
		result := fmt.Sprintf("%s{", v.Struct.Name)
		first := true
		for key, value := range v.Struct.Fields {
			if !first {
				result += ", "
			}
			result += fmt.Sprintf(`%s: %s`, key, value.String())
			first = false
		}
		result += "}"
		return result
	case ValueTypeServer:
		status := "stopped"
		if v.Server.Running {
			status = "running"
		}
		return fmt.Sprintf("<server %s: %s>", v.Server.Name, status)
	case ValueTypeServerState:
		v.ServerState.Mutex.RLock()
		defer v.ServerState.Mutex.RUnlock()
		result := "{"
		first := true
		for key, value := range *v.ServerState.State {
			if !first {
				result += ", "
			}
			result += fmt.Sprintf(`%s: %s`, key, value.String())
			first = false
		}
		result += "}"
		return result
	default:
		return "<unknown>"
	}
}

// IsTruthy returns whether the value is considered truthy
func (v *Value) IsTruthy() bool {
	switch v.Type {
	case ValueTypeNil:
		return false
	case ValueTypeBool:
		return v.Bool
	case ValueTypeNumber:
		return v.Number != 0
	case ValueTypeString:
		return v.Str != ""
	case ValueTypeArray:
		return len(v.Array) > 0
	case ValueTypeObject:
		return len(v.Object) > 0
	case ValueTypeFunction:
		return true
	case ValueTypeStruct:
		return true // Structs are always truthy
	case ValueTypeServer:
		return true // Servers are always truthy
	case ValueTypeServerState:
		v.ServerState.Mutex.RLock()
		defer v.ServerState.Mutex.RUnlock()
		return len(*v.ServerState.State) > 0
	default:
		return false
	}
}

// IsEqual checks if two values are equal
func (v *Value) IsEqual(other *Value) bool {
	if v.Type != other.Type {
		return false
	}

	switch v.Type {
	case ValueTypeNil:
		return true
	case ValueTypeNumber:
		return v.Number == other.Number
	case ValueTypeString:
		return v.Str == other.Str
	case ValueTypeBool:
		return v.Bool == other.Bool
	case ValueTypeArray:
		if len(v.Array) != len(other.Array) {
			return false
		}
		for i, elem := range v.Array {
			if !elem.IsEqual(other.Array[i]) {
				return false
			}
		}
		return true
	case ValueTypeObject:
		if len(v.Object) != len(other.Object) {
			return false
		}
		for key, value := range v.Object {
			otherValue, exists := other.Object[key]
			if !exists || !value.IsEqual(otherValue) {
				return false
			}
		}
		return true
	case ValueTypeFunction:
		// Functions are equal if they're the same instance
		return v.Function == other.Function
	case ValueTypeStruct:
		// Structs are equal if they have the same name and equal fields
		if v.Struct.Name != other.Struct.Name {
			return false
		}
		if len(v.Struct.Fields) != len(other.Struct.Fields) {
			return false
		}
		for key, value := range v.Struct.Fields {
			otherValue, exists := other.Struct.Fields[key]
			if !exists || !value.IsEqual(otherValue) {
				return false
			}
		}
		return true
	case ValueTypeServer:
		// Servers are equal if they're the same instance
		return v.Server == other.Server
	case ValueTypeServerState:
		// ServerStates are equal if they point to the same state
		return v.ServerState.State == other.ServerState.State
	default:
		return false
	}
}

// Start starts the server goroutine
func (s *Server) Start(evaluator interface{}) {
	if s.Running {
		return
	}

	s.Running = true
	go s.runServerLoop(evaluator)
}

// Stop stops the server goroutine
func (s *Server) Stop() {
	if !s.Running {
		return
	}

	s.Running = false
	close(s.MessageChan)
}

// SendMessage sends a message to the server and optionally waits for a reply
func (s *Server) SendMessage(method string, args []*Value, waitForReply bool) (*Value, error) {
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
func (s *Server) runServerLoop(evaluator interface{}) {
	for s.Running {
		select {
		case message, ok := <-s.MessageChan:
			if !ok {
				return // Channel closed
			}
			s.handleMessage(message, evaluator)
		}
	}
}

// handleMessage processes an incoming message
func (s *Server) handleMessage(message *Message, evaluator interface{}) {
	// Find the receive function (no lock needed for read-only access to receivers)
	receiver, exists := s.Receivers[message.Method]
	if !exists {
		if message.Reply != nil {
			message.Reply <- NewNil() // Send nil as error
		}
		return
	}

	// Create server environment with state access
	serverEnv := NewEnvironment(s.Environment)

	// Add state variable to environment (as mutable server state for .get/.set access)
	stateValue := NewServerState(&s.State, &s.StateMutex)
	serverEnv.Define("state", stateValue)

	// Type assert evaluator and call receive function
	if eval, ok := evaluator.(interface {
		CallUserFunction(*Function, []*Value, *Environment) (*Value, error)
	}); ok {
		// Create a modified function with the correct environment for state access
		modifiedReceiver := &Function{
			Name:       receiver.Name,
			Parameters: receiver.Parameters,
			Body:       receiver.Body,
			IsBuiltin:  false,
			ClosureEnv: serverEnv, // Use server environment with state defined
		}

		result, err := eval.CallUserFunction(modifiedReceiver, message.Args, serverEnv)

		// No need to update server state manually - ServerState updates it directly

		if err != nil {
			if message.Reply != nil {
				message.Reply <- NewNil() // Send nil on error
			}
			return
		}

		if message.Reply != nil {
			message.Reply <- result
		}
	} else {
		// Fallback if evaluator doesn't have the right interface
		if message.Reply != nil {
			message.Reply <- NewNil()
		}
	}
}

// GetState safely gets a state variable
func (s *Server) GetState(key string) (*Value, bool) {
	s.StateMutex.RLock()
	defer s.StateMutex.RUnlock()

	value, exists := s.State[key]
	return value, exists
}

// SetState safely sets a state variable
func (s *Server) SetState(key string, value *Value) {
	s.StateMutex.Lock()
	defer s.StateMutex.Unlock()

	s.State[key] = value
}

// getReceiverNames returns a list of receiver method names for debugging
func (s *Server) getReceiverNames() []string {
	names := make([]string, 0, len(s.Receivers))
	for name := range s.Receivers {
		names = append(names, name)
	}
	return names
}

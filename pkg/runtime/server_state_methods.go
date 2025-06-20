package runtime

import "fmt"

// ServerStateMethodHandler handles all server state method calls
type ServerStateMethodHandler struct {
	*BasicTypeHandler
}

// NewServerStateMethodHandler creates a new server state method handler with all built-in methods
func NewServerStateMethodHandler() *ServerStateMethodHandler {
	handler := &ServerStateMethodHandler{
		BasicTypeHandler: NewBasicTypeHandler(),
	}

	// Register all built-in server state methods
	handler.AddMethod("get", handler.getMethod)
	handler.AddMethod("set", handler.setMethod)

	return handler
}

// getMethod implements serverState.get(key)
func (h *ServerStateMethodHandler) getMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeServerState {
		return nil, fmt.Errorf("expected server state")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("get method expects 1 argument")
	}

	if args[0].Type != ValueTypeString {
		return nil, fmt.Errorf("state key must be a string")
	}

	key := args[0].Str

	// In actor model, mutex is nil since server processes messages sequentially
	if target.ServerState.Mutex != nil {
		target.ServerState.Mutex.RLock()
		defer target.ServerState.Mutex.RUnlock()
	}

	if value, exists := (*target.ServerState.State)[key]; exists {
		return value, nil
	}
	return NewNil(), nil
}

// setMethod implements serverState.set(key, value) - mutates in place
func (h *ServerStateMethodHandler) setMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeServerState {
		return nil, fmt.Errorf("expected server state")
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("set method expects 2 arguments")
	}

	if args[0].Type != ValueTypeString {
		return nil, fmt.Errorf("state key must be a string")
	}

	key := args[0].Str
	value := args[1]

	// In actor model, mutex is nil since server processes messages sequentially
	if target.ServerState.Mutex != nil {
		target.ServerState.Mutex.Lock()
		defer target.ServerState.Mutex.Unlock()
	}

	// Mutable update - modify the original state map
	(*target.ServerState.State)[key] = value

	// Return the value for chaining
	return value, nil
}

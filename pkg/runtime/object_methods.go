package runtime

import "fmt"

// ObjectMethodHandler handles all object method calls
type ObjectMethodHandler struct {
	*BasicTypeHandler
}

// NewObjectMethodHandler creates a new object method handler with all built-in methods
func NewObjectMethodHandler() *ObjectMethodHandler {
	handler := &ObjectMethodHandler{
		BasicTypeHandler: NewBasicTypeHandler(),
	}

	// Register all built-in object methods
	handler.AddMethod("get", handler.getMethod)
	handler.AddMethod("set", handler.setMethod)

	return handler
}

// getMethod implements object.get(key)
func (h *ObjectMethodHandler) getMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeObject {
		return nil, fmt.Errorf("expected object")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("get method expects 1 argument")
	}

	if args[0].Type != ValueTypeString {
		return nil, fmt.Errorf("object key must be a string")
	}

	key := args[0].Str
	if value, exists := target.Object[key]; exists {
		return value, nil
	}
	return NewNil(), nil
}

// setMethod implements object.set(key, value) - returns new object (immutable)
func (h *ObjectMethodHandler) setMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeObject {
		return nil, fmt.Errorf("expected object")
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("set method expects 2 arguments")
	}

	if args[0].Type != ValueTypeString {
		return nil, fmt.Errorf("object key must be a string")
	}

	key := args[0].Str
	value := args[1]

	// Create a new object with the updated field (immutable semantics)
	newObject := make(map[string]*Value)
	for k, v := range target.Object {
		newObject[k] = v
	}
	newObject[key] = value

	return &Value{
		Type:   ValueTypeObject,
		Object: newObject,
	}, nil
}

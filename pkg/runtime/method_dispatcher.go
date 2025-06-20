package runtime

import "fmt"

// MethodDispatcher provides the main interface for calling methods on values
type MethodDispatcher interface {
	CallMethod(target *Value, methodName string, args []*Value) (*Value, error)
	RegisterHandler(valueType ValueType, handler TypeMethodHandler)
	GetHandler(valueType ValueType) TypeMethodHandler
}

// TypeMethodHandler handles method calls for a specific value type
type TypeMethodHandler interface {
	CallMethod(target *Value, methodName string, args []*Value) (*Value, error)
	AddMethod(methodName string, impl MethodImplementation)
	HasMethod(methodName string) bool
	ListMethods() []string
}

// MethodImplementation is a function that implements a specific method
type MethodImplementation func(target *Value, args []*Value) (*Value, error)

// BasicMethodDispatcher is the default implementation of MethodDispatcher
type BasicMethodDispatcher struct {
	handlers map[ValueType]TypeMethodHandler
}

// NewMethodDispatcher creates a new method dispatcher
func NewMethodDispatcher() *BasicMethodDispatcher {
	return &BasicMethodDispatcher{
		handlers: make(map[ValueType]TypeMethodHandler),
	}
}

// CallMethod implements MethodDispatcher
func (d *BasicMethodDispatcher) CallMethod(target *Value, methodName string, args []*Value) (*Value, error) {
	handler, exists := d.handlers[target.Type]
	if !exists {
		return nil, fmt.Errorf("method '%s' not supported for %v", methodName, target.Type)
	}

	return handler.CallMethod(target, methodName, args)
}

// RegisterHandler implements MethodDispatcher
func (d *BasicMethodDispatcher) RegisterHandler(valueType ValueType, handler TypeMethodHandler) {
	d.handlers[valueType] = handler
}

// GetHandler implements MethodDispatcher
func (d *BasicMethodDispatcher) GetHandler(valueType ValueType) TypeMethodHandler {
	return d.handlers[valueType]
}

// BasicTypeHandler provides a basic implementation of TypeMethodHandler
type BasicTypeHandler struct {
	methods map[string]MethodImplementation
}

// NewBasicTypeHandler creates a new basic type handler
func NewBasicTypeHandler() *BasicTypeHandler {
	return &BasicTypeHandler{
		methods: make(map[string]MethodImplementation),
	}
}

// CallMethod implements TypeMethodHandler
func (h *BasicTypeHandler) CallMethod(target *Value, methodName string, args []*Value) (*Value, error) {
	impl, exists := h.methods[methodName]
	if !exists {
		return nil, fmt.Errorf("unknown method: %s", methodName)
	}

	return impl(target, args)
}

// AddMethod implements TypeMethodHandler
func (h *BasicTypeHandler) AddMethod(methodName string, impl MethodImplementation) {
	h.methods[methodName] = impl
}

// HasMethod implements TypeMethodHandler
func (h *BasicTypeHandler) HasMethod(methodName string) bool {
	_, exists := h.methods[methodName]
	return exists
}

// ListMethods implements TypeMethodHandler
func (h *BasicTypeHandler) ListMethods() []string {
	methods := make([]string, 0, len(h.methods))
	for name := range h.methods {
		methods = append(methods, name)
	}
	return methods
}

package runtime

import "fmt"

// StringMethodHandler handles all string method calls
type StringMethodHandler struct {
	*BasicTypeHandler
}

// NewStringMethodHandler creates a new string method handler with all built-in methods
func NewStringMethodHandler() *StringMethodHandler {
	handler := &StringMethodHandler{
		BasicTypeHandler: NewBasicTypeHandler(),
	}

	// Register all built-in string methods
	handler.AddMethod("length", handler.lengthMethod)

	return handler
}

// lengthMethod implements string.length()
func (h *StringMethodHandler) lengthMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeString {
		return nil, fmt.Errorf("expected string")
	}

	return NewNumber(float64(len(target.Str))), nil
}

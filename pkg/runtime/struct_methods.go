package runtime

import "fmt"

// StructMethodHandler handles all struct method calls
type StructMethodHandler struct {
	*BasicTypeHandler
}

// NewStructMethodHandler creates a new struct method handler with all built-in methods
func NewStructMethodHandler() *StructMethodHandler {
	handler := &StructMethodHandler{
		BasicTypeHandler: NewBasicTypeHandler(),
	}

	// Register all built-in struct methods
	handler.AddMethod("get", handler.getMethod)

	return handler
}

// getMethod implements struct.get(fieldName)
func (h *StructMethodHandler) getMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeStruct {
		return nil, fmt.Errorf("expected struct")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("get method expects 1 argument")
	}

	if args[0].Type != ValueTypeString {
		return nil, fmt.Errorf("struct field name must be a string")
	}

	fieldName := args[0].Str
	if value, exists := target.Struct.Fields[fieldName]; exists {
		return value, nil
	}

	return nil, fmt.Errorf("struct %s has no field '%s'", target.Struct.Name, fieldName)
}

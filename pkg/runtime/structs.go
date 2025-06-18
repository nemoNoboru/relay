package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// evaluateStructExpr handles struct definitions
func (e *Evaluator) evaluateStructExpr(expr *parser.StructExpr, env *Environment) (*Value, error) {
	// Create struct definition
	fieldTypes := make(map[string]string)
	for _, field := range expr.Fields {
		// For now, we'll store the type name as a string
		// In a more complete implementation, we'd have a proper type system
		fieldTypes[field.Name] = e.getTypeName(field.Type)
	}

	structDef := &StructDefinition{
		Name:   expr.Name,
		Fields: fieldTypes,
	}

	// Store the struct definition
	e.structDefs[expr.Name] = structDef

	// Struct definitions don't return a value, they register a type
	return NewNil(), nil
}

// evaluateStructConstructor handles struct instantiation (User{name: "John"})
func (e *Evaluator) evaluateStructConstructor(expr *parser.StructConstructor, env *Environment) (*Value, error) {
	// Check if the struct type is defined
	structDef, exists := e.structDefs[expr.Name]
	if !exists {
		return nil, fmt.Errorf("undefined struct type: %s", expr.Name)
	}

	// Evaluate field values
	fields := make(map[string]*Value)
	for _, field := range expr.Fields {
		value, err := e.EvaluateWithEnv(field.Value, env)
		if err != nil {
			return nil, err
		}
		fields[field.Key] = value
	}

	// Validate that all required fields are provided
	for fieldName := range structDef.Fields {
		if _, provided := fields[fieldName]; !provided {
			return nil, fmt.Errorf("missing required field '%s' for struct %s", fieldName, expr.Name)
		}
	}

	// Create struct instance
	return NewStruct(expr.Name, fields), nil
}

// getTypeName extracts a type name from a TypeRef (simplified)
func (e *Evaluator) getTypeName(typeRef *parser.TypeRef) string {
	if typeRef == nil {
		return "unknown"
	}

	if typeRef.Name != "" {
		return typeRef.Name
	}

	if typeRef.Array != nil {
		return "[]" + e.getTypeName(typeRef.Array)
	}

	if typeRef.Function != nil {
		return "function"
	}

	if typeRef.Parameterized != nil {
		return typeRef.Parameterized.Name
	}

	return "unknown"
}

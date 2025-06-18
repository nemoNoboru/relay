package runtime

import (
	"fmt"
	"relay/pkg/parser"
	"strconv"
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
)

// Value represents a runtime value in the Relay language
type Value struct {
	Type     ValueType
	Number   float64
	Str      string
	Bool     bool
	Array    []*Value
	Object   map[string]*Value
	Function *Function // For functions and lambdas
	Struct   *Struct   // For struct instances
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
	default:
		return false
	}
}

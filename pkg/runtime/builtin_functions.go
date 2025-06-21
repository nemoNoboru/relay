package runtime

import (
	"fmt"
)

// Builtins is a map of all built-in functions
var Builtins = map[string]*Function{
	"len": {
		Name:      "len",
		IsBuiltin: true,
		Builtin: func(args []*Value) (*Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len() expected 1 argument, got %d", len(args))
			}
			switch args[0].Type {
			case ValueTypeString:
				// The String() method on Value returns the Go string, which is fine for len()
				return NewNumber(float64(len(args[0].Str))), nil
			case ValueTypeArray:
				return NewNumber(float64(len(args[0].Array))), nil
			case ValueTypeObject:
				return NewNumber(float64(len(args[0].Object))), nil
			default:
				return nil, fmt.Errorf("len() not supported for type %s", args[0].Type)
			}
		},
	},
	"string": {
		Name:      "string",
		IsBuiltin: true,
		Builtin: func(args []*Value) (*Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("string() expected 1 argument, got %d", len(args))
			}
			// If it's already a string, return it to avoid adding extra quotes.
			if args[0].Type == ValueTypeString {
				return args[0], nil
			}
			// Otherwise, use the canonical string representation.
			return NewString(args[0].String()), nil
		},
	},
}

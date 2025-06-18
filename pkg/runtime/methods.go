package runtime

import (
	"fmt"
	"relay/pkg/parser"
)

// evaluateMethodCall evaluates method calls
func (e *Evaluator) evaluateMethodCall(object *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	// Handle built-in methods based on object type
	switch object.Type {
	case ValueTypeArray:
		return e.evaluateArrayMethod(object, call, env)
	case ValueTypeObject:
		return e.evaluateObjectMethod(object, call, env)
	case ValueTypeServerState:
		return e.evaluateServerStateMethod(object, call, env)
	case ValueTypeString:
		return e.evaluateStringMethod(object, call, env)
	case ValueTypeStruct:
		return e.evaluateStructMethod(object, call, env)
	default:
		return nil, fmt.Errorf("method '%s' not supported for %v", call.Method, object.Type)
	}
}

// evaluateArrayMethod evaluates array methods
func (e *Evaluator) evaluateArrayMethod(array *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "length":
		return NewNumber(float64(len(array.Array))), nil

	case "get":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("get method expects 1 argument")
		}

		index, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if index.Type != ValueTypeNumber {
			return nil, fmt.Errorf("array index must be a number")
		}

		idx := int(index.Number)
		if idx < 0 || idx >= len(array.Array) {
			return nil, fmt.Errorf("array index %d out of bounds (length: %d)", idx, len(array.Array))
		}

		return array.Array[idx], nil

	case "set":
		if len(call.Args) != 2 {
			return nil, fmt.Errorf("set method expects 2 arguments")
		}

		index, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if index.Type != ValueTypeNumber {
			return nil, fmt.Errorf("array index must be a number")
		}

		value, err := e.EvaluateWithEnv(call.Args[1], env)
		if err != nil {
			return nil, err
		}

		idx := int(index.Number)
		if idx < 0 || idx >= len(array.Array) {
			return nil, fmt.Errorf("array index %d out of bounds (length: %d)", idx, len(array.Array))
		}

		array.Array[idx] = value
		return value, nil

	case "push":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("push method expects 1 argument")
		}

		value, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		array.Array = append(array.Array, value)
		return NewNumber(float64(len(array.Array))), nil

	case "pop":
		if len(array.Array) == 0 {
			return NewArray([]*Value{}), nil
		}

		lastIndex := len(array.Array) - 1
		value := array.Array[lastIndex]
		array.Array = array.Array[:lastIndex]
		return value, nil

	case "map":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("map method expects 1 argument")
		}

		fn, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if fn.Type != ValueTypeFunction {
			return nil, fmt.Errorf("map expects a function argument")
		}

		result := make([]*Value, len(array.Array))
		for i, item := range array.Array {
			// Call the function with the item and index
			args := []*Value{item, NewNumber(float64(i))}
			mapped, err := e.CallUserFunction(fn.Function, args, env)
			if err != nil {
				return nil, fmt.Errorf("error in map function: %v", err)
			}
			result[i] = mapped
		}

		return NewArray(result), nil

	case "filter":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("filter method expects 1 argument")
		}

		fn, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if fn.Type != ValueTypeFunction {
			return nil, fmt.Errorf("filter expects a function argument")
		}

		var result []*Value
		for i, item := range array.Array {
			// Call the function with the item and index
			args := []*Value{item, NewNumber(float64(i))}
			keep, err := e.CallUserFunction(fn.Function, args, env)
			if err != nil {
				return nil, fmt.Errorf("error in filter function: %v", err)
			}
			if keep.IsTruthy() {
				result = append(result, item)
			}
		}

		return NewArray(result), nil

	case "reduce":
		if len(call.Args) < 1 || len(call.Args) > 2 {
			return nil, fmt.Errorf("reduce method expects 1 or 2 arguments")
		}

		fn, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if fn.Type != ValueTypeFunction {
			return nil, fmt.Errorf("reduce expects a function argument")
		}

		if len(array.Array) == 0 {
			if len(call.Args) == 2 {
				return e.EvaluateWithEnv(call.Args[1], env)
			}
			return nil, fmt.Errorf("reduce of empty array without initial value")
		}

		var accumulator *Value
		startIndex := 0

		if len(call.Args) == 2 {
			accumulator, err = e.EvaluateWithEnv(call.Args[1], env)
			if err != nil {
				return nil, err
			}
		} else {
			accumulator = array.Array[0]
			startIndex = 1
		}

		for i := startIndex; i < len(array.Array); i++ {
			args := []*Value{accumulator, array.Array[i], NewNumber(float64(i))}
			accumulator, err = e.CallUserFunction(fn.Function, args, env)
			if err != nil {
				return nil, fmt.Errorf("error in reduce function: %v", err)
			}
		}

		return accumulator, nil

	case "find":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("find method expects 1 argument")
		}

		fn, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if fn.Type != ValueTypeFunction {
			return nil, fmt.Errorf("find expects a function argument")
		}

		for i, item := range array.Array {
			args := []*Value{item, NewNumber(float64(i))}
			found, err := e.CallUserFunction(fn.Function, args, env)
			if err != nil {
				return nil, fmt.Errorf("error in find function: %v", err)
			}
			if found.IsTruthy() {
				return item, nil
			}
		}

		return NewNil(), nil

	case "includes":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("includes method expects 1 argument")
		}

		searchValue, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		for _, item := range array.Array {
			if item.IsEqual(searchValue) {
				return NewBool(true), nil
			}
		}

		return NewBool(false), nil

	default:
		return nil, fmt.Errorf("unknown array method: %s", call.Method)
	}
}

// evaluateServerStateMethod evaluates server state methods (mutable object-like behavior)
func (e *Evaluator) evaluateServerStateMethod(object *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "get":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("get method expects 1 argument")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("state key must be a string")
		}

		object.ServerState.Mutex.RLock()
		defer object.ServerState.Mutex.RUnlock()

		if value, exists := (*object.ServerState.State)[key.Str]; exists {
			return value, nil
		}
		return NewNil(), nil

	case "set":
		if len(call.Args) != 2 {
			return nil, fmt.Errorf("set method expects 2 arguments")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("state key must be a string")
		}

		value, err := e.EvaluateWithEnv(call.Args[1], env)
		if err != nil {
			return nil, err
		}

		// Mutable update - modify the original state map
		object.ServerState.Mutex.Lock()
		(*object.ServerState.State)[key.Str] = value
		object.ServerState.Mutex.Unlock()

		// Return the value for chaining
		return value, nil

	default:
		return nil, fmt.Errorf("unknown server state method: %s", call.Method)
	}
}

// evaluateObjectMethod evaluates object methods
func (e *Evaluator) evaluateObjectMethod(object *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "get":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("get method expects 1 argument")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("object key must be a string")
		}

		if value, exists := object.Object[key.Str]; exists {
			return value, nil
		}
		return NewNil(), nil

	case "set":
		if len(call.Args) != 2 {
			return nil, fmt.Errorf("set method expects 2 arguments")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("object key must be a string")
		}

		value, err := e.EvaluateWithEnv(call.Args[1], env)
		if err != nil {
			return nil, err
		}

		// Create a new object with the updated field (immutable semantics)
		newObject := make(map[string]*Value)
		for k, v := range object.Object {
			newObject[k] = v
		}
		newObject[key.Str] = value

		return &Value{
			Type:   ValueTypeObject,
			Object: newObject,
		}, nil

	default:
		// Fallback: check if there's a field with this name that contains a function
		if fieldValue, exists := object.Object[call.Method]; exists {
			if fieldValue.Type == ValueTypeFunction {
				// Evaluate arguments
				args := make([]*Value, len(call.Args))
				for i, arg := range call.Args {
					val, err := e.EvaluateWithEnv(arg, env)
					if err != nil {
						return nil, err
					}
					args[i] = val
				}

				// Call the function
				if fieldValue.Function.IsBuiltin {
					return fieldValue.Function.Builtin(args)
				}
				return e.CallUserFunction(fieldValue.Function, args, env)
			}
		}
		return nil, fmt.Errorf("unknown object method: %s", call.Method)
	}
}

// evaluateStringMethod evaluates string methods
func (e *Evaluator) evaluateStringMethod(str *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "length":
		return NewNumber(float64(len(str.Str))), nil
	default:
		return nil, fmt.Errorf("unknown string method: %s", call.Method)
	}
}

// evaluateStructMethod evaluates struct methods
func (e *Evaluator) evaluateStructMethod(structVal *Value, call *parser.MethodCall, env *Environment) (*Value, error) {
	switch call.Method {
	case "get":
		if len(call.Args) != 1 {
			return nil, fmt.Errorf("get method expects 1 argument")
		}

		key, err := e.EvaluateWithEnv(call.Args[0], env)
		if err != nil {
			return nil, err
		}

		if key.Type != ValueTypeString {
			return nil, fmt.Errorf("struct field name must be a string")
		}

		if value, exists := structVal.Struct.Fields[key.Str]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("struct %s has no field '%s'", structVal.Struct.Name, key.Str)

	default:
		return nil, fmt.Errorf("unknown struct method: %s", call.Method)
	}
}

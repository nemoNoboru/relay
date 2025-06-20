package runtime

import (
	"fmt"
)

// FunctionExecutor interface for calling user functions in higher-order array methods
type FunctionExecutor interface {
	ExecuteFunction(fn *Value, args []*Value) (*Value, error)
}

// ArrayMethodHandler handles all array method calls
type ArrayMethodHandler struct {
	*BasicTypeHandler
	executor FunctionExecutor // Interface to execute functions
}

// NewArrayMethodHandler creates a new array method handler with all built-in methods
func NewArrayMethodHandler() *ArrayMethodHandler {
	handler := &ArrayMethodHandler{
		BasicTypeHandler: NewBasicTypeHandler(),
		executor:         nil, // Will be set by the evaluator
	}

	// Register all built-in array methods
	handler.AddMethod("length", handler.lengthMethod)
	handler.AddMethod("get", handler.getMethod)
	handler.AddMethod("set", handler.setMethod)
	handler.AddMethod("push", handler.pushMethod)
	handler.AddMethod("pop", handler.popMethod)
	handler.AddMethod("includes", handler.includesMethod)
	handler.AddMethod("map", handler.mapMethod)
	handler.AddMethod("filter", handler.filterMethod)
	handler.AddMethod("reduce", handler.reduceMethod)

	return handler
}

// SetFunctionExecutor sets the function executor for higher-order methods
func (h *ArrayMethodHandler) SetFunctionExecutor(executor FunctionExecutor) {
	h.executor = executor
}

// lengthMethod implements array.length()
func (h *ArrayMethodHandler) lengthMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	return NewNumber(float64(len(target.Array))), nil
}

// getMethod implements array.get(index)
func (h *ArrayMethodHandler) getMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("get method expects 1 argument")
	}

	if args[0].Type != ValueTypeNumber {
		return nil, fmt.Errorf("array index must be a number")
	}

	index := int(args[0].Number)
	if index < 0 || index >= len(target.Array) {
		return nil, fmt.Errorf("array index out of bounds")
	}

	return target.Array[index], nil
}

// setMethod implements array.set(index, value) - mutates original array
func (h *ArrayMethodHandler) setMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("set method expects 2 arguments")
	}

	if args[0].Type != ValueTypeNumber {
		return nil, fmt.Errorf("array index must be a number")
	}

	index := int(args[0].Number)
	if index < 0 || index >= len(target.Array) {
		return nil, fmt.Errorf("array index out of bounds")
	}

	value := args[1]

	// Mutate the original array (expected behavior based on tests)
	target.Array[index] = value

	return value, nil
}

// pushMethod implements array.push(value)
func (h *ArrayMethodHandler) pushMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("push method expects 1 argument")
	}

	value := args[0]

	// Create a new array with the pushed value
	newArray := make([]*Value, len(target.Array)+1)
	copy(newArray, target.Array)
	newArray[len(target.Array)] = value

	// Update the original array (mutable operation)
	target.Array = newArray

	return value, nil
}

// popMethod implements array.pop() - mutates original array
func (h *ArrayMethodHandler) popMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 0 {
		return nil, fmt.Errorf("pop method expects 0 arguments")
	}

	if len(target.Array) == 0 {
		// Return empty array for empty array (expected behavior based on tests)
		return NewArray([]*Value{}), nil
	}

	// Get the last element
	lastElement := target.Array[len(target.Array)-1]

	// Mutate the original array (mutable operation)
	target.Array = target.Array[:len(target.Array)-1]

	return lastElement, nil
}

// includesMethod implements array.includes(value)
func (h *ArrayMethodHandler) includesMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("includes method expects 1 argument")
	}

	searchValue := args[0]

	for _, item := range target.Array {
		if item.IsEqual(searchValue) {
			return NewBool(true), nil
		}
	}

	return NewBool(false), nil
}

// mapMethod implements array.map(function)
func (h *ArrayMethodHandler) mapMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("map method expects 1 argument")
	}

	if args[0].Type != ValueTypeFunction {
		return nil, fmt.Errorf("map expects a function argument")
	}

	if h.executor == nil {
		return nil, fmt.Errorf("function executor not available")
	}

	fn := args[0]
	result := make([]*Value, len(target.Array))

	for i, item := range target.Array {
		// Try calling with just the item first (single parameter)
		mapped, err := h.executor.ExecuteFunction(fn, []*Value{item})
		if err != nil {
			// If that fails, try with item and index (two parameters)
			mapped, err = h.executor.ExecuteFunction(fn, []*Value{item, NewNumber(float64(i))})
			if err != nil {
				return nil, fmt.Errorf("error in map function: %v", err)
			}
		}
		result[i] = mapped
	}

	return NewArray(result), nil
}

// filterMethod implements array.filter(function)
func (h *ArrayMethodHandler) filterMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("filter method expects 1 argument")
	}

	if args[0].Type != ValueTypeFunction {
		return nil, fmt.Errorf("filter expects a function argument")
	}

	if h.executor == nil {
		return nil, fmt.Errorf("function executor not available")
	}

	fn := args[0]
	var result []*Value

	for i, item := range target.Array {
		// Try calling with just the item first (single parameter)
		keep, err := h.executor.ExecuteFunction(fn, []*Value{item})
		if err != nil {
			// If that fails, try with item and index (two parameters)
			keep, err = h.executor.ExecuteFunction(fn, []*Value{item, NewNumber(float64(i))})
			if err != nil {
				return nil, fmt.Errorf("error in filter function: %v", err)
			}
		}

		if keep.IsTruthy() {
			result = append(result, item)
		}
	}

	return NewArray(result), nil
}

// reduceMethod implements array.reduce(function, initialValue?)
func (h *ArrayMethodHandler) reduceMethod(target *Value, args []*Value) (*Value, error) {
	if target.Type != ValueTypeArray {
		return nil, fmt.Errorf("expected array")
	}

	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("reduce method expects 1 or 2 arguments")
	}

	if args[0].Type != ValueTypeFunction {
		return nil, fmt.Errorf("reduce expects a function argument")
	}

	if h.executor == nil {
		return nil, fmt.Errorf("function executor not available")
	}

	fn := args[0]

	if len(target.Array) == 0 {
		if len(args) == 2 {
			return args[1], nil
		}
		return nil, fmt.Errorf("reduce of empty array without initial value")
	}

	var accumulator *Value
	startIndex := 0

	if len(args) == 2 {
		accumulator = args[1]
	} else {
		accumulator = target.Array[0]
		startIndex = 1
	}

	for i := startIndex; i < len(target.Array); i++ {
		item := target.Array[i]

		// For reduce, we always call with (accumulator, currentValue, index)
		newAcc, err := h.executor.ExecuteFunction(fn, []*Value{accumulator, item, NewNumber(float64(i))})
		if err != nil {
			// If 3 args fail, try with just accumulator and item
			newAcc, err = h.executor.ExecuteFunction(fn, []*Value{accumulator, item})
			if err != nil {
				return nil, fmt.Errorf("error in reduce function: %v", err)
			}
		}
		accumulator = newAcc
	}

	return accumulator, nil
}

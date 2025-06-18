package runtime

import "fmt"

// Note: applyBinaryOperation moved to core.go

// Arithmetic operations
func (e *Evaluator) add(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewNumber(left.Number + right.Number), nil
	}
	if left.Type == ValueTypeString && right.Type == ValueTypeString {
		return NewString(left.Str + right.Str), nil
	}
	return nil, fmt.Errorf("invalid operands for addition")
}

func (e *Evaluator) subtract(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewNumber(left.Number - right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for subtraction")
}

func (e *Evaluator) multiply(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewNumber(left.Number * right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for multiplication")
}

func (e *Evaluator) divide(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		if right.Number == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return NewNumber(left.Number / right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for division")
}

// Comparison operations
func (e *Evaluator) less(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number < right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

func (e *Evaluator) lessEqual(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number <= right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

func (e *Evaluator) greater(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number > right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

func (e *Evaluator) greaterEqual(left, right *Value) (*Value, error) {
	if left.Type == ValueTypeNumber && right.Type == ValueTypeNumber {
		return NewBool(left.Number >= right.Number), nil
	}
	return nil, fmt.Errorf("invalid operands for comparison")
}

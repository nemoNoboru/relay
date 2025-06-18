package runtime

import (
	"relay/pkg/parser"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluateNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"0", 0},
	}

	evaluator := NewEvaluator()

	for _, test := range tests {
		program, err := parser.Parse("test.relay", strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("Parse error for %s: %v", test.input, err)
		}

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
		}

		result, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Evaluation error for %s: %v", test.input, err)
		}

		if result.Type != ValueTypeNumber {
			t.Fatalf("Expected number, got %v", result.Type)
		}

		if result.Number != test.expected {
			t.Errorf("Expected %f, got %f", test.expected, result.Number)
		}
	}
}

func TestEvaluateStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"world"`, "world"},
		{`""`, ""},
	}

	evaluator := NewEvaluator()

	for _, test := range tests {
		program, err := parser.Parse("test.relay", strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("Parse error for %s: %v", test.input, err)
		}

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
		}

		result, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Evaluation error for %s: %v", test.input, err)
		}

		if result.Type != ValueTypeString {
			t.Fatalf("Expected string, got %v", result.Type)
		}

		if result.Str != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result.Str)
		}
	}
}

func TestEvaluateBooleans(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	evaluator := NewEvaluator()

	for _, test := range tests {
		program, err := parser.Parse("test.relay", strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("Parse error for %s: %v", test.input, err)
		}

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
		}

		result, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Evaluation error for %s: %v", test.input, err)
		}

		if result.Type != ValueTypeBool {
			t.Fatalf("Expected boolean, got %v", result.Type)
		}

		if result.Bool != test.expected {
			t.Errorf("Expected %t, got %t", test.expected, result.Bool)
		}
	}
}

func TestEvaluateArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"3 * 7", 21},
		{"15 / 3", 5},
		{"2 + 3 * 4", 20}, // Left to right evaluation: (2 + 3) * 4
		{"10 - 2 - 3", 5}, // Left to right: (10 - 2) - 3
	}

	evaluator := NewEvaluator()

	for _, test := range tests {
		program, err := parser.Parse("test.relay", strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("Parse error for %s: %v", test.input, err)
		}

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
		}

		result, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Evaluation error for %s: %v", test.input, err)
		}

		if result.Type != ValueTypeNumber {
			t.Fatalf("Expected number, got %v", result.Type)
		}

		if result.Number != test.expected {
			t.Errorf("For %s: expected %f, got %f", test.input, test.expected, result.Number)
		}
	}
}

func TestEvaluateComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"5 > 3", true},
		{"3 > 5", false},
		{"5 >= 5", true},
		{"3 >= 5", false},
		{"3 < 5", true},
		{"5 < 3", false},
		{"3 <= 3", true},
		{"5 <= 3", false},
		{"5 == 5", true},
		{"5 == 3", false},
		{"5 != 3", true},
		{"5 != 5", false},
	}

	evaluator := NewEvaluator()

	for _, test := range tests {
		program, err := parser.Parse("test.relay", strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("Parse error for %s: %v", test.input, err)
		}

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
		}

		result, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Evaluation error for %s: %v", test.input, err)
		}

		if result.Type != ValueTypeBool {
			t.Fatalf("Expected boolean, got %v", result.Type)
		}

		if result.Bool != test.expected {
			t.Errorf("For %s: expected %t, got %t", test.input, test.expected, result.Bool)
		}
	}
}

func TestEvaluateLogical(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
	}

	evaluator := NewEvaluator()

	for _, test := range tests {
		program, err := parser.Parse("test.relay", strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("Parse error for %s: %v", test.input, err)
		}

		if len(program.Expressions) != 1 {
			t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
		}

		result, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Evaluation error for %s: %v", test.input, err)
		}

		if result.Type != ValueTypeBool {
			t.Fatalf("Expected boolean, got %v", result.Type)
		}

		if result.Bool != test.expected {
			t.Errorf("For %s: expected %t, got %t", test.input, test.expected, result.Bool)
		}
	}
}

func TestEvaluateVariables(t *testing.T) {
	evaluator := NewEvaluator()

	// Test variable assignment
	program, err := parser.Parse("test.relay", strings.NewReader(`set x = 42`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}

	if result.Type != ValueTypeNumber || result.Number != 42 {
		t.Errorf("Expected 42, got %s", result.String())
	}

	// Test variable access
	program, err = parser.Parse("test.relay", strings.NewReader(`x`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}

	if result.Type != ValueTypeNumber || result.Number != 42 {
		t.Errorf("Expected 42 from variable, got %s", result.String())
	}
}

func TestEvaluateStringConcatenation(t *testing.T) {
	evaluator := NewEvaluator()

	program, err := parser.Parse("test.relay", strings.NewReader(`"hello" + " world"`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}

	if result.Type != ValueTypeString || result.Str != "hello world" {
		t.Errorf("Expected 'hello world', got %s", result.String())
	}
}

func TestEvaluateFunctionDefinition(t *testing.T) {
	evaluator := NewEvaluator()

	// Test named function definition
	program, err := parser.Parse("test.relay", strings.NewReader(`fn add(x, y) { x + y }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}

	if result.Type != ValueTypeFunction {
		t.Fatalf("Expected function, got %v", result.Type)
	}

	if result.Function.Name != "add" {
		t.Errorf("Expected function name 'add', got '%s'", result.Function.Name)
	}

	// Test that function is stored in environment
	funcValue, exists := evaluator.globalEnv.Get("add")
	if !exists {
		t.Error("Function 'add' should be stored in environment")
	}

	if funcValue.Type != ValueTypeFunction {
		t.Error("Stored value should be a function")
	}
}

func TestEvaluateFunctionCall(t *testing.T) {
	evaluator := NewEvaluator()

	// Define a function
	program, err := parser.Parse("test.relay", strings.NewReader(`fn add(x, y) { x + y }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function definition error: %v", err)
	}

	// Call the function
	program, err = parser.Parse("test.relay", strings.NewReader(`add(5, 3)`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function call error: %v", err)
	}

	if result.Type != ValueTypeNumber || result.Number != 8 {
		t.Errorf("Expected 8, got %s", result.String())
	}
}

func TestEvaluateFunctionWithReturn(t *testing.T) {
	evaluator := NewEvaluator()

	// Define a function with explicit return
	program, err := parser.Parse("test.relay", strings.NewReader(`fn multiply(x, y) { return x * y }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function definition error: %v", err)
	}

	// Call the function
	program, err = parser.Parse("test.relay", strings.NewReader(`multiply(4, 6)`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function call error: %v", err)
	}

	if result.Type != ValueTypeNumber || result.Number != 24 {
		t.Errorf("Expected 24, got %s", result.String())
	}
}

func TestEvaluateFunctionScope(t *testing.T) {
	evaluator := NewEvaluator()

	// Set a global variable
	program, err := parser.Parse("test.relay", strings.NewReader(`set global_var = 10`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Global variable error: %v", err)
	}

	// Define a function that uses a parameter with the same name
	program, err = parser.Parse("test.relay", strings.NewReader(`fn test_scope(global_var) { global_var + 5 }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function definition error: %v", err)
	}

	// Call the function - should use parameter, not global
	program, err = parser.Parse("test.relay", strings.NewReader(`test_scope(20)`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function call error: %v", err)
	}

	if result.Type != ValueTypeNumber || result.Number != 25 {
		t.Errorf("Expected 25 (parameter shadowing global), got %s", result.String())
	}

	// Check that global variable is unchanged
	globalVal, exists := evaluator.globalEnv.Get("global_var")
	if !exists || globalVal.Number != 10 {
		t.Error("Global variable should remain unchanged")
	}
}

func TestEvaluateFunctionParameterCount(t *testing.T) {
	evaluator := NewEvaluator()

	// Define a function
	program, err := parser.Parse("test.relay", strings.NewReader(`fn add(x, y) { x + y }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function definition error: %v", err)
	}

	// Call with wrong number of arguments
	program, err = parser.Parse("test.relay", strings.NewReader(`add(5)`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err == nil {
		t.Error("Expected error for wrong number of arguments")
	}

	if !strings.Contains(err.Error(), "expects 2 arguments, got 1") {
		t.Errorf("Expected parameter count error, got: %v", err)
	}
}

func TestEvaluateNestedFunctionCalls(t *testing.T) {
	evaluator := NewEvaluator()

	// Define helper functions
	program, err := parser.Parse("test.relay", strings.NewReader(`fn square(x) { x * x }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function definition error: %v", err)
	}

	program, err = parser.Parse("test.relay", strings.NewReader(`fn add(x, y) { x + y }`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	_, err = evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Function definition error: %v", err)
	}

	// Call nested functions: add(square(2), square(3)) = add(4, 9) = 13
	program, err = parser.Parse("test.relay", strings.NewReader(`add(square(2), square(3))`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	result, err := evaluator.Evaluate(program.Expressions[0])
	if err != nil {
		t.Fatalf("Nested function call error: %v", err)
	}

	if result.Type != ValueTypeNumber || result.Number != 13 {
		t.Errorf("Expected 13, got %s", result.String())
	}
}

// Test for spacing bug in function bodies
func TestEvaluateFunctionSpacing(t *testing.T) {
	evaluator := NewEvaluator()

	// Test function with no spaces around operators
	program, err := parser.Parse("test.relay", strings.NewReader("fn incrementNoSpace(c) {c+1}"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	// Test function with spaces around operators
	program, err = parser.Parse("test.relay", strings.NewReader("fn incrementWithSpace(c) { c + 1 }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	// Both should work the same way
	program, err = parser.Parse("test.relay", strings.NewReader("incrementNoSpace(5)"))
	require.NoError(t, err)
	result1, err := evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result1.Type)
	require.Equal(t, 6.0, result1.Number, "incrementNoSpace(5) should return 6, not %v", result1.Number)

	program, err = parser.Parse("test.relay", strings.NewReader("incrementWithSpace(5)"))
	require.NoError(t, err)
	result2, err := evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result2.Type)
	require.Equal(t, 6.0, result2.Number, "incrementWithSpace(5) should return 6, not %v", result2.Number)

	// Test with different inputs
	program, err = parser.Parse("test.relay", strings.NewReader("incrementNoSpace(10)"))
	require.NoError(t, err)
	result3, err := evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result3.Type)
	require.Equal(t, 11.0, result3.Number, "incrementNoSpace(10) should return 11, not %v", result3.Number)
}

// Test for closures - functions that capture their lexical environment
func TestEvaluateClosures(t *testing.T) {
	evaluator := NewEvaluator()

	// Test 1: Global variable capture
	program, err := parser.Parse("test.relay", strings.NewReader("set globalVar = 42"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("fn getGlobal() { globalVar }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("getGlobal()"))
	require.NoError(t, err)
	result, err := evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 42.0, result.Number, "Function should capture global variable")

	// Test 2: Parameter capture in returned function
	program, err = parser.Parse("test.relay", strings.NewReader("fn makeAdder(x) { fn(y) { x + y } }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("set add10 = makeAdder(10)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeFunction, result.Type)

	program, err = parser.Parse("test.relay", strings.NewReader("add10(5)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 15.0, result.Number, "Closure should capture parameter from outer function")

	// Test 3: Multiple closures with different captured values
	program, err = parser.Parse("test.relay", strings.NewReader("set add3 = makeAdder(3)"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("add3(7)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 10.0, result.Number, "Different closures should capture different values")

	// Test that original add10 still works
	program, err = parser.Parse("test.relay", strings.NewReader("add10(2)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 12.0, result.Number, "Original closure should still work independently")
}

// Test for first-class functions - functions that can be passed as arguments and returned as values
func TestEvaluateFirstClassFunctions(t *testing.T) {
	evaluator := NewEvaluator()

	// Test 1: Functions stored in variables
	program, err := parser.Parse("test.relay", strings.NewReader("fn double(x) { x * 2 }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("set myFunc = double"))
	require.NoError(t, err)
	result, err := evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeFunction, result.Type)

	program, err = parser.Parse("test.relay", strings.NewReader("myFunc(5)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 10.0, result.Number, "Function should work when called through variable")

	// Test 2: Functions passed as arguments
	program, err = parser.Parse("test.relay", strings.NewReader("fn applyTwice(f, x) { f(f(x)) }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("fn increment(x) { x + 1 }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("applyTwice(increment, 5)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 7.0, result.Number, "Function should work when passed as argument")

	// Test 3: Functions returned from functions (already tested in closures, but for completeness)
	program, err = parser.Parse("test.relay", strings.NewReader("fn multiplier(factor) { fn(x) { x * factor } }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("set times3 = multiplier(3)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeFunction, result.Type)

	program, err = parser.Parse("test.relay", strings.NewReader("times3(4)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 12.0, result.Number, "Returned function should work correctly")

	// Test 4: Higher-order functions (functions that take and return functions)
	program, err = parser.Parse("test.relay", strings.NewReader("fn compose(f, g) { fn(x) { f(g(x)) } }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("fn square(x) { x * x }"))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	program, err = parser.Parse("test.relay", strings.NewReader("set squareAndIncrement = compose(increment, square)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeFunction, result.Type)

	program, err = parser.Parse("test.relay", strings.NewReader("squareAndIncrement(3)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 10.0, result.Number, "Composed function should work: square(3)=9, then increment(9)=10")

	// Test 5: Anonymous functions as arguments
	program, err = parser.Parse("test.relay", strings.NewReader("applyTwice(fn(x) { x * 3 }, 2)"))
	require.NoError(t, err)
	result, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)
	require.Equal(t, ValueTypeNumber, result.Type)
	require.Equal(t, 18.0, result.Number, "Anonymous function as argument should work: 2*3=6, then 6*3=18")
}

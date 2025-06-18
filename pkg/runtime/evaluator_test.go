package runtime

import (
	"relay/pkg/parser"
	"strings"
	"testing"
	"time"

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

func TestEvaluateStructDefinitions(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple struct definition",
			input:    `struct User { name: string, age: number }`,
			expected: "nil", // Struct definitions return nil but register the type
		},
		{
			name:     "Struct with multiple field types",
			input:    `struct Post { id: number, title: string, published: bool }`,
			expected: "nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, test.expected, result.String())
		})
	}
}

func TestEvaluateStructConstructors(t *testing.T) {
	evaluator := NewEvaluator()

	// First define the struct
	structDef := `struct User { name: string, age: number }`
	program, err := parser.Parse("test", strings.NewReader(structDef))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple struct instantiation",
			input:    `User{ name: "John", age: 30 }`,
			expected: `User{name: "John", age: 30}`,
		},
		{
			name:     "Struct with different field order",
			input:    `User{ age: 25, name: "Alice" }`,
			expected: `User{age: 25, name: "Alice"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)

			// Check that it's a struct with the right name and fields
			require.Equal(t, ValueTypeStruct, result.Type)
			require.Equal(t, "User", result.Struct.Name)
			require.Contains(t, result.String(), "User{")
			require.Contains(t, result.String(), "name:")
			require.Contains(t, result.String(), "age:")
		})
	}
}

func TestEvaluateStructFieldAccess(t *testing.T) {
	evaluator := NewEvaluator()

	// Define struct and create instance
	setup := `
		struct User { name: string, age: number }
		set user = User{ name: "John", age: 30 }
	`
	program, err := parser.Parse("test", strings.NewReader(setup))
	require.NoError(t, err)

	for _, expr := range program.Expressions {
		_, err = evaluator.Evaluate(expr)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Access string field",
			input:    `user.get("name")`,
			expected: `"John"`,
		},
		{
			name:     "Access number field",
			input:    `user.get("age")`,
			expected: "30",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, test.expected, result.String())
		})
	}
}

func TestEvaluateStructErrors(t *testing.T) {

	tests := []struct {
		name        string
		setup       string
		input       string
		expectError string
	}{
		{
			name:        "Undefined struct type",
			setup:       "",
			input:       `UnknownStruct{ name: "test" }`,
			expectError: "undefined struct type: UnknownStruct",
		},
		{
			name:        "Missing required field",
			setup:       `struct User { name: string, age: number }`,
			input:       `User{ name: "John" }`,
			expectError: "missing required field 'age' for struct User",
		},
		{
			name:        "Access nonexistent field",
			setup:       `struct User { name: string } set user = User{ name: "John" }`,
			input:       `user.get("age")`,
			expectError: "struct User has no field 'age'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Fresh evaluator for each test
			eval := NewEvaluator()

			// Run setup if provided
			if test.setup != "" {
				setupProgram, err := parser.Parse("test", strings.NewReader(test.setup))
				require.NoError(t, err)
				for _, expr := range setupProgram.Expressions {
					_, err = eval.Evaluate(expr)
					require.NoError(t, err)
				}
			}

			// Run the test input
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			_, err = eval.Evaluate(program.Expressions[0])
			require.Error(t, err)
			require.Contains(t, err.Error(), test.expectError)
		})
	}
}

func TestEvaluateStructEquality(t *testing.T) {
	evaluator := NewEvaluator()

	// Define struct
	structDef := `struct User { name: string, age: number }`
	program, err := parser.Parse("test", strings.NewReader(structDef))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Equal structs",
			input:    `User{ name: "John", age: 30 } == User{ name: "John", age: 30 }`,
			expected: "true",
		},
		{
			name:     "Different field values",
			input:    `User{ name: "John", age: 30 } == User{ name: "Jane", age: 30 }`,
			expected: "false",
		},
		{
			name:     "Different field order but same values",
			input:    `User{ name: "John", age: 30 } == User{ age: 30, name: "John" }`,
			expected: "true",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, test.expected, result.String())
		})
	}
}

func TestEvaluateServerDefinitions(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers() // Clean up after tests

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple server definition",
			input: `server simple_server {
				state {
					count: number = 0
				}
				
				receive fn increment() -> number {
					state.set("count", state.get("count") + 1)
					state.get("count")
				}
			}`,
			expected: "<server simple_server: running>",
		},
		{
			name: "Server with multiple receive functions",
			input: `server math_server {
				state {
					total: number = 0
				}
				
				receive fn add(x: number) -> number {
					state.set("total", state.get("total") + x)
					state.get("total")
				}
				
				receive fn get_total() -> number {
					state.get("total")
				}
			}`,
			expected: "<server math_server: running>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, ValueTypeServer, result.Type)
			require.Contains(t, result.String(), "running")
		})
	}
}

func TestEvaluateServerMessagePassing(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Define a counter server
	serverDef := `server counter_server {
		state {
			count: number = 0
		}
		
		receive fn increment() -> number {
			state.set("count", state.get("count") + 1)
			state.get("count")
		}
		
		receive fn get_count() -> number {
			state.get("count")
		}
		
		receive fn add(x: number) -> number {
			state.set("count", state.get("count") + x)
			state.get("count")
		}
	}`

	program, err := parser.Parse("test", strings.NewReader(serverDef))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Call increment function",
			input:    `send "counter_server" increment {}`,
			expected: "1",
		},
		{
			name:     "Call increment again",
			input:    `send "counter_server" increment {}`,
			expected: "2",
		},
		{
			name:     "Get current count",
			input:    `send "counter_server" get_count {}`,
			expected: "2",
		},
		{
			name:     "Add to count",
			input:    `send "counter_server" add { x: 5 }`,
			expected: "7",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, test.expected, result.String())
		})
	}
}

func TestEvaluateServerState(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Define a server with complex state
	serverDef := `server user_server {
		state {
			users: [string] = [],
			count: number = 0
		}
		
		receive fn add_user(name: string) -> number {
			state.set("count", state.get("count") + 1)
			state.get("count")
		}
		
		receive fn get_user_count() -> number {
			state.get("count")
		}
	}`

	program, err := parser.Parse("test", strings.NewReader(serverDef))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	// Test state operations
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Initial count is 0",
			input:    `send "user_server" get_user_count {}`,
			expected: "0",
		},
		{
			name:     "Add first user",
			input:    `send "user_server" add_user { name: "Alice" }`,
			expected: "1",
		},
		{
			name:     "Add second user",
			input:    `send "user_server" add_user { name: "Bob" }`,
			expected: "2",
		},
		{
			name:     "Check final count",
			input:    `send "user_server" get_user_count {}`,
			expected: "2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, test.expected, result.String())
		})
	}
}

func TestEvaluateServerErrors(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Define a simple server
	serverDef := `server test_server {
		receive fn echo(msg: string) -> string {
			msg
		}
	}`

	program, err := parser.Parse("test", strings.NewReader(serverDef))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program.Expressions[0])
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name        string
		input       string
		expectError string
	}{
		{
			name:        "Send to nonexistent server",
			input:       `send "nonexistent_server" test {}`,
			expectError: "server 'nonexistent_server' not found",
		},
		{
			name:        "Call nonexistent method",
			input:       `send "test_server" nonexistent_method {}`,
			expectError: "", // Server returns nil for unknown methods
		},
		{
			name:        "Invalid message arguments",
			input:       `message(123, "method")`,
			expectError: "first argument to message must be server name (string)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			if test.expectError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectError)
			} else {
				require.NoError(t, err)
				// For unknown methods, server returns nil
				if test.name == "Call nonexistent method" {
					require.Equal(t, ValueTypeNil, result.Type)
				}
			}
		})
	}
}

func TestEvaluateMultipleServers(t *testing.T) {
	evaluator := NewEvaluator()
	defer evaluator.StopAllServers()

	// Define two servers that can communicate
	server1Def := `server ping_server {
		receive fn ping() -> string {
			"pong"
		}
		
		receive fn ping_other() -> string {
			send "pong_server" pong {}
		}
	}`

	server2Def := `server pong_server {
		receive fn pong() -> string {
			"ping"
		}
	}`

	// Create both servers
	program1, err := parser.Parse("test", strings.NewReader(server1Def))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program1.Expressions[0])
	require.NoError(t, err)

	program2, err := parser.Parse("test", strings.NewReader(server2Def))
	require.NoError(t, err)
	_, err = evaluator.Evaluate(program2.Expressions[0])
	require.NoError(t, err)

	// Give servers time to start
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Direct ping",
			input:    `send "ping_server" ping {}`,
			expected: `"pong"`,
		},
		{
			name:     "Direct pong",
			input:    `send "pong_server" pong {}`,
			expected: `"ping"`,
		},
		{
			name:     "Server-to-server communication",
			input:    `send "ping_server" ping_other {}`,
			expected: `"ping"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := parser.Parse("test", strings.NewReader(test.input))
			require.NoError(t, err)
			require.Len(t, program.Expressions, 1)

			result, err := evaluator.Evaluate(program.Expressions[0])
			require.NoError(t, err)
			require.Equal(t, test.expected, result.String())
		})
	}
}

package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArithmeticOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected float64
	}{
		// Basic arithmetic
		{"Addition", "2 + 3", 5.0},
		{"Subtraction", "10 - 4", 6.0},
		{"Multiplication", "3 * 7", 21.0},
		{"Division", "15 / 3", 5.0},

		// Operator precedence (left-to-right evaluation)
		{"Left-to-right addition and multiplication", "2 + 3 * 4", 20.0}, // (2 + 3) * 4
		{"Left-to-right subtraction", "10 - 2 - 3", 5.0},                 // (10 - 2) - 3

		// With variables
		{"Variables in arithmetic", "set a = 5\nset b = 3\na + b", 8.0},
		{"Complex expression", "set x = 10\nset y = 5\n(x - y) * 2", 10.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, test.expected, result.Number)
		})
	}
}

func TestStringOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"String concatenation", `"Hello" + " " + "World"`, "Hello World"},
		{"String with variables", `set first = "John"\nset last = "Doe"\nfirst + " " + last`, "John Doe"},
		{"Mixed concatenation", `"Count: " + "42"`, "Count: 42"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		})
	}
}

func TestComparisonOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		// Numeric comparisons
		{"Greater than (true)", "5 > 3", true},
		{"Greater than (false)", "3 > 5", false},
		{"Greater than or equal (greater)", "5 >= 3", true},
		{"Greater than or equal (equal)", "5 >= 5", true},
		{"Greater than or equal (false)", "3 >= 5", false},
		{"Less than (true)", "3 < 5", true},
		{"Less than (false)", "5 < 3", false},
		{"Less than or equal (less)", "3 <= 5", true},
		{"Less than or equal (equal)", "5 <= 5", true},
		{"Less than or equal (false)", "5 <= 3", false},

		// Equality
		{"Number equality (true)", "5 == 5", true},
		{"Number equality (false)", "5 == 3", false},
		{"String equality (true)", `"hello" == "hello"`, true},
		{"String equality (false)", `"hello" == "world"`, false},
		{"Boolean equality (true)", "true == true", true},
		{"Boolean equality (false)", "true == false", false},

		// Inequality
		{"Number inequality (true)", "5 != 3", true},
		{"Number inequality (false)", "5 != 5", false},
		{"String inequality (true)", `"hello" != "world"`, true},
		{"String inequality (false)", `"hello" != "hello"`, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeBool, result.Type)
			require.Equal(t, test.expected, result.Bool)
		})
	}
}

func TestLogicalOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		// AND operations
		{"true && true", "true && true", true},
		{"true && false", "true && false", false},
		{"false && true", "false && true", false},
		{"false && false", "false && false", false},

		// OR operations
		{"true || true", "true || true", true},
		{"true || false", "true || false", true},
		{"false || true", "false || true", true},
		{"false || false", "false || false", false},

		// With truthiness
		{"Truthy string && true", `"hello" && true`, true},
		{"Empty string && true", `"" && true`, false},
		{"Non-zero number || false", "42 || false", true},
		{"Zero number || false", "0 || false", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeBool, result.Type)
			require.Equal(t, test.expected, result.Bool)
		})
	}
}

func TestNullCoalescing(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Nil ?? String", "set x = nil\nx ?? \"default\"", "default"},
		{"String ?? String", "set x = \"value\"\nx ?? \"default\"", "value"},
		{"Empty string ?? String", "set x = \"\"\nx ?? \"default\"", ""}, // Empty string is not nil
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		})
	}
}

func TestUnaryOperations(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedType  ValueType
		expectedValue interface{}
	}{
		{"Negation of number", "-42", ValueTypeNumber, -42.0},
		{"Negation of variable", "set x = 5\n-x", ValueTypeNumber, -5.0},
		{"Logical NOT true", "!true", ValueTypeBool, false},
		{"Logical NOT false", "!false", ValueTypeBool, true},
		{"Logical NOT truthy", `!"hello"`, ValueTypeBool, false},
		{"Logical NOT falsy", `!""`, ValueTypeBool, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, test.expectedType, result.Type)

			switch test.expectedType {
			case ValueTypeNumber:
				require.Equal(t, test.expectedValue, result.Number)
			case ValueTypeBool:
				require.Equal(t, test.expectedValue, result.Bool)
			case ValueTypeString:
				require.Equal(t, test.expectedValue, result.Str)
			}
		})
	}
}

func TestComplexExpressions(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		expected  interface{}
		valueType ValueType
	}{
		{
			"Arithmetic with variables",
			`set a = 10\nset b = 5\n(a + b) * 2`,
			30.0,
			ValueTypeNumber,
		},
		{
			"String operations with variables",
			`set greeting = "Hello"\nset name = "World"\ngreeting + ", " + name + "!"`,
			"Hello, World!",
			ValueTypeString,
		},
		{
			"Mixed logical operations",
			`set a = true\nset b = false\na && !b`,
			true,
			ValueTypeBool,
		},
		{
			"Comparison with arithmetic",
			`set x = 10\nset y = 5\n(x + y) > (x - y)`,
			true,
			ValueTypeBool,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, test.valueType, result.Type)

			switch test.valueType {
			case ValueTypeNumber:
				require.Equal(t, test.expected, result.Number)
			case ValueTypeString:
				require.Equal(t, test.expected, result.Str)
			case ValueTypeBool:
				require.Equal(t, test.expected, result.Bool)
			}
		})
	}
}

func TestOperationErrors(t *testing.T) {
	errorTests := []struct {
		name string
		code string
	}{
		{"Division by zero", "10 / 0"},
		{"Invalid arithmetic", `"hello" + 42`},
		{"Invalid comparison", `"hello" < 42`},
		{"Invalid subtraction", `"hello" - 5`},
		{"Invalid multiplication", `"hello" * 2`},
		{"Invalid division", `"hello" / 2`},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			err := evalCodeError(t, test.code)
			require.Error(t, err, "Expected error for: %s", test.code)
		})
	}
}

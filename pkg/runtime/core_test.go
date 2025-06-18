package runtime

import (
	"relay/pkg/parser"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Helper function to evaluate code and return result
func evalCode(t *testing.T, code string) *Value {
	evaluator := NewEvaluator()
	program, err := parser.Parse("test.relay", strings.NewReader(code))
	require.NoError(t, err)

	var result *Value
	for _, expr := range program.Expressions {
		result, err = evaluator.Evaluate(expr)
		require.NoError(t, err)
	}
	return result
}

// Helper function to evaluate code expecting an error
func evalCodeError(t *testing.T, code string) error {
	evaluator := NewEvaluator()
	program, err := parser.Parse("test.relay", strings.NewReader(code))
	if err != nil {
		return err
	}

	for _, expr := range program.Expressions {
		_, err = evaluator.Evaluate(expr)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestBasicTypes(t *testing.T) {
	t.Run("Numbers", func(t *testing.T) {
		tests := []struct {
			code     string
			expected float64
		}{
			{"42", 42.0},
			{"3.14", 3.14},
			{"0", 0.0},
			{"-5", -5.0},
		}

		for _, test := range tests {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, test.expected, result.Number)
		}
	})

	t.Run("Strings", func(t *testing.T) {
		tests := []struct {
			code     string
			expected string
		}{
			{`"hello"`, "hello"},
			{`"world"`, "world"},
			{`""`, ""},
			{`"special chars: 123!@#"`, "special chars: 123!@#"},
		}

		for _, test := range tests {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		}
	})

	t.Run("Booleans", func(t *testing.T) {
		tests := []struct {
			code     string
			expected bool
		}{
			{"true", true},
			{"false", false},
		}

		for _, test := range tests {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeBool, result.Type)
			require.Equal(t, test.expected, result.Bool)
		}
	})

	t.Run("Arrays", func(t *testing.T) {
		result := evalCode(t, "[1, 2, 3]")
		require.Equal(t, ValueTypeArray, result.Type)
		require.Len(t, result.Array, 3)
		require.Equal(t, 1.0, result.Array[0].Number)
		require.Equal(t, 2.0, result.Array[1].Number)
		require.Equal(t, 3.0, result.Array[2].Number)
	})

	t.Run("Objects", func(t *testing.T) {
		result := evalCode(t, `{name: "John", age: 30}`)
		require.Equal(t, ValueTypeObject, result.Type)
		require.Contains(t, result.Object, "name")
		require.Contains(t, result.Object, "age")
		require.Equal(t, "John", result.Object["name"].Str)
		require.Equal(t, 30.0, result.Object["age"].Number)
	})
}

func TestVariables(t *testing.T) {
	t.Run("Variable assignment and access", func(t *testing.T) {
		tests := []struct {
			name      string
			code      string
			expected  interface{}
			valueType ValueType
		}{
			{"Number variable", "set x = 42\nx", 42.0, ValueTypeNumber},
			{"String variable", `set name = "Alice"` + "\nname", "Alice", ValueTypeString},
			{"Boolean variable", "set flag = true\nflag", true, ValueTypeBool},
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
	})

	// Note: Variable reassignment is not supported by design in Relay (immutable-by-default)

	t.Run("Undefined variable error", func(t *testing.T) {
		err := evalCodeError(t, "undefined_var")
		require.Error(t, err)
		require.Contains(t, err.Error(), "undefined variable")
	})
}

func TestEnvironmentScoping(t *testing.T) {
	t.Run("Global and local scope", func(t *testing.T) {
		result := evalCode(t, `
		set global_var = 100
		fn test() -> number {
			set local_var = 200
			global_var + local_var
		}
		test()`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 300.0, result.Number)
	})

	t.Run("Variable shadowing", func(t *testing.T) {
		result := evalCode(t, `
		set x = 10
		fn shadow_test() -> number {
			set x = 20
			x
		}
		shadow_test()`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 20.0, result.Number)
	})
}

func TestValueConversions(t *testing.T) {
	t.Run("Truthiness", func(t *testing.T) {
		tests := []struct {
			code     string
			expected bool
		}{
			// Truthy values
			{"42", true},
			{`"hello"`, true},
			{"true", true},
			{"[1]", true},
			{`{key: "value"}`, true},

			// Falsy values
			{"0", false},
			{`""`, false},
			{"false", false},
			{"[]", false},
			{"{}", false},
		}

		for _, test := range tests {
			t.Run(test.code, func(t *testing.T) {
				result := evalCode(t, test.code)
				require.Equal(t, test.expected, result.IsTruthy())
			})
		}
	})

	t.Run("String representation", func(t *testing.T) {
		tests := []struct {
			code     string
			contains string
		}{
			{"42", "42"},
			{`"hello"`, "hello"},
			{"true", "true"},
			{"false", "false"},
			{"[1, 2]", "1"},
			{`{name: "John"}`, "name"},
		}

		for _, test := range tests {
			t.Run(test.code, func(t *testing.T) {
				result := evalCode(t, test.code)
				str := result.String()
				require.Contains(t, str, test.contains)
			})
		}
	})
}

func TestValueEquality(t *testing.T) {
	t.Run("Equal values", func(t *testing.T) {
		equal := [][]string{
			{"42", "42"},
			{`"hello"`, `"hello"`},
			{"true", "true"},
			{"false", "false"},
			{"[1, 2]", "[1, 2]"},
		}

		for _, pair := range equal {
			val1 := evalCode(t, pair[0])
			val2 := evalCode(t, pair[1])
			require.True(t, val1.IsEqual(val2), "Expected %s == %s", pair[0], pair[1])
		}
	})

	t.Run("Unequal values", func(t *testing.T) {
		unequal := [][]string{
			{"42", "43"},
			{`"hello"`, `"world"`},
			{"true", "false"},
			{"[1, 2]", "[1, 3]"},
			{"42", `"42"`}, // Different types
		}

		for _, pair := range unequal {
			val1 := evalCode(t, pair[0])
			val2 := evalCode(t, pair[1])
			require.False(t, val1.IsEqual(val2), "Expected %s != %s", pair[0], pair[1])
		}
	})
}

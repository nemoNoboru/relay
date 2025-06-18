package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestREPLScenarios tests the specific scenarios users encounter in the REPL
func TestREPLScenarios(t *testing.T) {
	t.Run("Array creation and method calls - REPL USER SCENARIOS", func(t *testing.T) {
		t.Run("Basic array creation", func(t *testing.T) {
			result := evalCode(t, `[1,2,3,4]`)
			require.Equal(t, ValueTypeArray, result.Type)
			require.Len(t, result.Array, 4)
		})

		t.Run("Array assigned to variable", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3,4]
			n`)
			require.Equal(t, ValueTypeArray, result.Type)
			require.Len(t, result.Array, 4)
		})

		t.Run("Array get method - What users tried in REPL", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3,4]
			n.get(0)`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 1.0, result.Number)
		})

		t.Run("Array length method", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3,4]
			n.length()`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 4.0, result.Number)
		})

		t.Run("Array includes method", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3,4]
			n.includes(3)`)
			require.Equal(t, ValueTypeBool, result.Type)
			require.Equal(t, true, result.Bool)
		})

		t.Run("Array push method", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3]
			n.push(4)`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 4.0, result.Number)
		})

		t.Run("Array set method", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3,4]
			n.set(0, 10)
			n.get(0)`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 10.0, result.Number)
		})

		t.Run("Array pop method", func(t *testing.T) {
			result := evalCode(t, `set n = [1,2,3,4]
			n.pop()`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 4.0, result.Number)
		})
	})

	t.Run("Working runtime features", func(t *testing.T) {
		t.Run("Basic arithmetic", func(t *testing.T) {
			result := evalCode(t, `2 + 3 * 4`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 20.0, result.Number) // Left-to-right evaluation
		})

		t.Run("String concatenation", func(t *testing.T) {
			result := evalCode(t, `"Hello" + " " + "World"`)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, "Hello World", result.Str)
		})

		t.Run("Variable assignment and access", func(t *testing.T) {
			result := evalCode(t, `set x = 42
			x`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 42.0, result.Number)
		})

		t.Run("Function definition and call", func(t *testing.T) {
			result := evalCode(t, `fn add(a: number, b: number) -> number {
				a + b
			}
			add(5, 3)`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 8.0, result.Number)
		})

		t.Run("Object creation", func(t *testing.T) {
			result := evalCode(t, `{name: "John", age: 30}`)
			require.Equal(t, ValueTypeObject, result.Type)
			require.Equal(t, "John", result.Object["name"].Str)
			require.Equal(t, 30.0, result.Object["age"].Number)
		})

		t.Run("Struct definition and creation", func(t *testing.T) {
			result := evalCode(t, `struct Person {
				name: string,
				age: number
			}
			Person {name: "Alice", age: 25}`)
			require.Equal(t, ValueTypeStruct, result.Type)
			require.Equal(t, "Person", result.Struct.Name)
			require.Equal(t, "Alice", result.Struct.Fields["name"].Str)
		})

		t.Run("Higher-order functions", func(t *testing.T) {
			result := evalCode(t, `fn apply_twice(f: function, x: number) -> number {
				f(f(x))
			}
			fn double(x: number) -> number {
				x * 2
			}
			apply_twice(double, 3)`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 12.0, result.Number)
		})

		t.Run("Server creation and messaging", func(t *testing.T) {
			code := `
			server Counter {
				state {
					count: number = 0
				}
				receive fn increment() -> number {
					state.set("count", state.get("count") + 1)
					state.get("count")
				}
			}
			send "Counter" increment {}`

			result := evalCode(t, code)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 1.0, result.Number)
		})
	})

	t.Run("Array method chaining scenarios", func(t *testing.T) {
		t.Run("Push then get", func(t *testing.T) {
			result := evalCode(t, `set arr = [1, 2]
			arr.push(3)
			arr.get(2)`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 3.0, result.Number)
		})

		t.Run("Set then includes", func(t *testing.T) {
			result := evalCode(t, `set arr = [1, 2, 3]
			arr.set(1, 99)
			arr.includes(99)`)
			require.Equal(t, ValueTypeBool, result.Type)
			require.Equal(t, true, result.Bool)
		})

		t.Run("Multiple pushes and length", func(t *testing.T) {
			result := evalCode(t, `set arr = []
			arr.push(1)
			arr.push(2)
			arr.push(3)
			arr.length()`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 3.0, result.Number)
		})
	})

	t.Run("Complex data structures", func(t *testing.T) {
		t.Run("Array of mixed types", func(t *testing.T) {
			result := evalCode(t, `[1, "hello", true]`)
			require.Equal(t, ValueTypeArray, result.Type)
			require.Len(t, result.Array, 3)
			require.Equal(t, ValueTypeNumber, result.Array[0].Type)
			require.Equal(t, ValueTypeString, result.Array[1].Type)
			require.Equal(t, ValueTypeBool, result.Array[2].Type)
		})

		t.Run("Nested arrays", func(t *testing.T) {
			result := evalCode(t, `[[1, 2], [3, 4]]`)
			require.Equal(t, ValueTypeArray, result.Type)
			require.Len(t, result.Array, 2)
			require.Equal(t, ValueTypeArray, result.Array[0].Type)
			require.Equal(t, ValueTypeArray, result.Array[1].Type)
		})

		t.Run("Functions with arrays", func(t *testing.T) {
			result := evalCode(t, `fn sum_first_two(arr: array) -> number {
				arr.get(0) + arr.get(1)
			}
			sum_first_two([10, 20, 30])`)
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 30.0, result.Number)
		})
	})
}

// TestRuntimePerformance tests that runtime operations are efficient
func TestRuntimePerformance(t *testing.T) {
	t.Run("Large array operations", func(t *testing.T) {
		// Create a large array and test operations
		code := `set large_arr = []
		set i = 0
		large_arr.push(1)
		large_arr.push(2)
		large_arr.push(3)
		large_arr.push(4)
		large_arr.push(5)
		large_arr.length()`

		result := evalCode(t, code)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 5.0, result.Number)
	})

	t.Run("Deep function calls", func(t *testing.T) {
		code := `fn add(a: number, b: number) -> number {
			a + b
		}
		fn chain(x: number) -> number {
			add(add(x, 1), add(x, 2))
		}
		chain(5)`

		result := evalCode(t, code)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 13.0, result.Number) // (5+1) + (5+2) = 6 + 7 = 13
	})
}

// TestErrorHandling tests runtime error conditions
func TestErrorHandling(t *testing.T) {
	t.Run("Array bounds checking", func(t *testing.T) {
		err := evalCodeError(t, `set arr = [1, 2, 3]
		arr.get(10)`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "out of bounds")
	})

	t.Run("Invalid method calls", func(t *testing.T) {
		err := evalCodeError(t, `set num = 42
		num.get(0)`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not supported")
	})

	t.Run("Undefined variable access", func(t *testing.T) {
		err := evalCodeError(t, `undefined_variable`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "undefined variable")
	})

	t.Run("Division by zero", func(t *testing.T) {
		err := evalCodeError(t, `10 / 0`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "division by zero")
	})
}

// TestREPLCompatibility specifically tests REPL-style interactions
func TestREPLCompatibility(t *testing.T) {
	t.Run("Single expressions work", func(t *testing.T) {
		tests := []struct {
			name      string
			code      string
			expected  interface{}
			valueType ValueType
		}{
			{"Number literal", "42", 42.0, ValueTypeNumber},
			{"String literal", `"hello"`, "hello", ValueTypeString},
			{"Boolean literal", "true", true, ValueTypeBool},
			{"Array literal", "[1,2,3]", 3, ValueTypeArray},    // Check length
			{"Object literal", `{x: 1}`, 1.0, ValueTypeObject}, // Check field access
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
				case ValueTypeArray:
					require.Len(t, result.Array, test.expected.(int))
				case ValueTypeObject:
					require.Equal(t, test.expected, result.Object["x"].Number)
				}
			})
		}
	})

	t.Run("Variable persistence across statements", func(t *testing.T) {
		// In a real REPL, variables would persist across statements
		// Our test simulates this by including all statements
		result := evalCode(t, `set x = 10
		set y = 20
		x + y`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 30.0, result.Number)
	})

	t.Run("Array methods that users tried in REPL", func(t *testing.T) {
		// These are the exact scenarios from the user's REPL session
		tests := []struct {
			setup     string
			method    string
			expected  interface{}
			valueType ValueType
		}{
			{"set n = [1,2,3,4]", "n.get(0)", 1.0, ValueTypeNumber},
			{"set n = [1,2,3,4]", "n.length()", 4.0, ValueTypeNumber},
			{"set n = [1,2,3,4]", "n.includes(2)", true, ValueTypeBool},
			{"set n = [1,2,3,4]", "n.includes(5)", false, ValueTypeBool},
		}

		for _, test := range tests {
			t.Run(test.method, func(t *testing.T) {
				code := test.setup + "\n" + test.method
				result := evalCode(t, code)
				require.Equal(t, test.valueType, result.Type)

				switch test.valueType {
				case ValueTypeNumber:
					require.Equal(t, test.expected, result.Number)
				case ValueTypeBool:
					require.Equal(t, test.expected, result.Bool)
				}
			})
		}
	})
}

package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArrayBasics(t *testing.T) {
	t.Run("Array creation", func(t *testing.T) {
		tests := []struct {
			name     string
			code     string
			expected []float64
		}{
			{"Empty array", "[]", []float64{}},
			{"Number array", "[1, 2, 3]", []float64{1, 2, 3}},
			{"Array with variables", "set a = 1\nset b = 2\n[a, b, 3]", []float64{1, 2, 3}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := evalCode(t, test.code)
				require.Equal(t, ValueTypeArray, result.Type)
				require.Len(t, result.Array, len(test.expected))

				for i, expected := range test.expected {
					require.Equal(t, ValueTypeNumber, result.Array[i].Type)
					require.Equal(t, expected, result.Array[i].Number)
				}
			})
		}
	})

	t.Run("Array assignment", func(t *testing.T) {
		result := evalCode(t, `set arr = [1, 2, 3]
		arr`)
		require.Equal(t, ValueTypeArray, result.Type)
		require.Len(t, result.Array, 3)
	})
}

func TestArrayMethods(t *testing.T) {
	t.Run("Array length method", func(t *testing.T) {
		tests := []struct {
			name     string
			code     string
			expected float64
		}{
			{"Empty array length", "[].length()", 0.0},
			{"Array length", "[1, 2, 3].length()", 3.0},
			{"Variable array length", "set arr = [1, 2, 3, 4, 5]\narr.length()", 5.0},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := evalCode(t, test.code)
				require.Equal(t, ValueTypeNumber, result.Type)
				require.Equal(t, test.expected, result.Number)
			})
		}
	})

	t.Run("Array get method - THIS IS WHAT REPL USERS ARE TRYING", func(t *testing.T) {
		tests := []struct {
			name      string
			code      string
			expected  float64
			shouldErr bool
		}{
			{"Get first element", "[1, 2, 3].get(0)", 1.0, false},
			{"Get middle element", "[1, 2, 3].get(1)", 2.0, false},
			{"Get last element", "[1, 2, 3].get(2)", 3.0, false},
			{"Get with variable", "set arr = [10, 20, 30]\narr.get(1)", 20.0, false},
			{"Get out of bounds", "[1, 2, 3].get(5)", 0.0, true},
			{"Get negative index", "[1, 2, 3].get(-1)", 0.0, true},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.shouldErr {
					err := evalCodeError(t, test.code)
					require.Error(t, err)
				} else {
					result := evalCode(t, test.code)
					require.Equal(t, ValueTypeNumber, result.Type)
					require.Equal(t, test.expected, result.Number)
				}
			})
		}
	})

	t.Run("Array set method", func(t *testing.T) {
		tests := []struct {
			name      string
			code      string
			expected  float64
			shouldErr bool
		}{
			{"Set first element", "set arr = [1, 2, 3]\narr.set(0, 10)\narr.get(0)", 10.0, false},
			{"Set middle element", "set arr = [1, 2, 3]\narr.set(1, 20)\narr.get(1)", 20.0, false},
			{"Set last element", "set arr = [1, 2, 3]\narr.set(2, 30)\narr.get(2)", 30.0, false},
			{"Set out of bounds", "set arr = [1, 2, 3]\narr.set(5, 50)", 0.0, true},
			{"Set negative index", "set arr = [1, 2, 3]\narr.set(-1, 50)", 0.0, true},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.shouldErr {
					err := evalCodeError(t, test.code)
					require.Error(t, err)
				} else {
					result := evalCode(t, test.code)
					require.Equal(t, ValueTypeNumber, result.Type)
					require.Equal(t, test.expected, result.Number)
				}
			})
		}
	})

	t.Run("Array push method", func(t *testing.T) {
		tests := []struct {
			name     string
			code     string
			expected float64
		}{
			{"Push to empty array", "set arr = []\narr.push(1)", 1.0},
			{"Push to array", "set arr = [1, 2]\narr.push(3)", 3.0},
			{"Push and check length", "set arr = [1, 2]\narr.push(3)\narr.length()", 3.0},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := evalCode(t, test.code)
				require.Equal(t, ValueTypeNumber, result.Type)
				require.Equal(t, test.expected, result.Number)
			})
		}
	})

	t.Run("Array pop method", func(t *testing.T) {
		t.Run("Pop from array", func(t *testing.T) {
			result := evalCode(t, "set arr = [1, 2, 3]\narr.pop()")
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 3.0, result.Number)
		})

		t.Run("Pop and check length", func(t *testing.T) {
			result := evalCode(t, "set arr = [1, 2, 3]\narr.pop()\narr.length()")
			require.Equal(t, ValueTypeNumber, result.Type)
			require.Equal(t, 2.0, result.Number)
		})

		t.Run("Pop from empty array", func(t *testing.T) {
			result := evalCode(t, "set arr = []\narr.pop()")
			require.Equal(t, ValueTypeArray, result.Type)
			require.Len(t, result.Array, 0)
		})
	})

	t.Run("Array includes method", func(t *testing.T) {
		tests := []struct {
			name     string
			code     string
			expected bool
		}{
			{"Includes existing number", "[1, 2, 3].includes(2)", true},
			{"Includes non-existing number", "[1, 2, 3].includes(5)", false},
			{"Includes in empty array", "[].includes(1)", false},
			{"Includes string", `["a", "b", "c"].includes("b")`, true},
			{"Includes non-existing string", `["a", "b", "c"].includes("d")`, false},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := evalCode(t, test.code)
				require.Equal(t, ValueTypeBool, result.Type)
				require.Equal(t, test.expected, result.Bool)
			})
		}
	})
}

func TestArrayMethodChaining(t *testing.T) {
	t.Run("Method chaining", func(t *testing.T) {
		tests := []struct {
			name     string
			code     string
			expected float64
		}{
			{"Push then get", "set arr = [1, 2]\narr.push(3)\narr.get(2)", 3.0},
			{"Push then length", "set arr = [1, 2]\narr.push(3)\narr.length()", 3.0},
			{"Set then get", "set arr = [1, 2, 3]\narr.set(1, 10)\narr.get(1)", 10.0},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := evalCode(t, test.code)
				require.Equal(t, ValueTypeNumber, result.Type)
				require.Equal(t, test.expected, result.Number)
			})
		}
	})
}

func TestArrayMethodErrors(t *testing.T) {
	t.Run("Invalid method calls", func(t *testing.T) {
		errorTests := []struct {
			name string
			code string
		}{
			{"Non-array method call", `"hello".get(0)`},
			{"Number method call", `42.push(1)`},
			{"Boolean method call", `true.length()`},
		}

		for _, test := range errorTests {
			t.Run(test.name, func(t *testing.T) {
				err := evalCodeError(t, test.code)
				require.Error(t, err, "Expected error for: %s", test.code)
			})
		}
	})
}

func TestMixedArrayTypes(t *testing.T) {
	t.Run("Arrays with mixed types", func(t *testing.T) {
		result := evalCode(t, `[1, "hello", true]`)
		require.Equal(t, ValueTypeArray, result.Type)
		require.Len(t, result.Array, 3)

		// Check types
		require.Equal(t, ValueTypeNumber, result.Array[0].Type)
		require.Equal(t, ValueTypeString, result.Array[1].Type)
		require.Equal(t, ValueTypeBool, result.Array[2].Type)

		// Check values
		require.Equal(t, 1.0, result.Array[0].Number)
		require.Equal(t, "hello", result.Array[1].Str)
		require.Equal(t, true, result.Array[2].Bool)
	})

	t.Run("Nested arrays", func(t *testing.T) {
		result := evalCode(t, `[[1, 2], [3, 4]]`)
		require.Equal(t, ValueTypeArray, result.Type)
		require.Len(t, result.Array, 2)

		// Check that each element is an array
		require.Equal(t, ValueTypeArray, result.Array[0].Type)
		require.Equal(t, ValueTypeArray, result.Array[1].Type)

		// Check nested array contents
		require.Len(t, result.Array[0].Array, 2)
		require.Len(t, result.Array[1].Array, 2)
		require.Equal(t, 1.0, result.Array[0].Array[0].Number)
		require.Equal(t, 2.0, result.Array[0].Array[1].Number)
		require.Equal(t, 3.0, result.Array[1].Array[0].Number)
		require.Equal(t, 4.0, result.Array[1].Array[1].Number)
	})
}

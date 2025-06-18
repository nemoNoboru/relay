package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFunctionDefinition(t *testing.T) {
	t.Run("Simple function definition", func(t *testing.T) {
		result := evalCode(t, `fn add(a: number, b: number) -> number {
			a + b
		}
		add(5, 3)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 8.0, result.Number)
	})

	t.Run("Function without parameters", func(t *testing.T) {
		result := evalCode(t, `fn get_answer() -> number {
			42
		}
		get_answer()`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 42.0, result.Number)
	})

	t.Run("Function with string return", func(t *testing.T) {
		result := evalCode(t, `fn greet(name: string) -> string {
			"Hello, " + name
		}
		greet("World")`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Hello, World", result.Str)
	})

	t.Run("Function with boolean return", func(t *testing.T) {
		result := evalCode(t, `fn is_positive(x: number) -> bool {
			x > 0
		}
		is_positive(5)`)
		require.Equal(t, ValueTypeBool, result.Type)
		require.Equal(t, true, result.Bool)
	})
}

func TestFunctionScoping(t *testing.T) {
	t.Run("Local variable scope", func(t *testing.T) {
		result := evalCode(t, `
		set global_var = 100
		fn test() -> number {
			set local_var = 50
			global_var + local_var
		}
		test()`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 150.0, result.Number)
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

	t.Run("Parameter shadowing global", func(t *testing.T) {
		result := evalCode(t, `
		set value = 100
		fn test(value: number) -> number {
			value * 2
		}
		test(5)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 10.0, result.Number)
	})

	t.Run("Function accessing global after local scope", func(t *testing.T) {
		// Relay is immutable-by-default, so this test is invalid
		// Variables cannot be reassigned, only new variables can be created
		result := evalCode(t, `
		set global_var = 100
		fn get_global() -> number {
			set local_var = 50
			global_var + local_var
		}
		get_global()`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 150.0, result.Number)
	})
}

func TestRecursiveFunctions(t *testing.T) {
	t.Run("Factorial function", func(t *testing.T) {
		result := evalCode(t, `
		fn factorial(n: number) -> number {
			if (n <= 1) {
				1
			} else {
				n * factorial(n - 1)
			}
		}
		factorial(5)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 120.0, result.Number)
	})

	t.Run("Fibonacci function", func(t *testing.T) {
		result := evalCode(t, `
		fn fib(n: number) -> number {
			if (n <= 1) {
				n
			} else {
				fib(n - 1) + fib(n - 2)
			}
		}
		fib(6)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 8.0, result.Number)
	})
}

func TestFunctionWithArrays(t *testing.T) {
	t.Run("Function returning array", func(t *testing.T) {
		result := evalCode(t, `
		fn create_array() -> array {
			[1, 2, 3]
		}
		create_array()`)
		require.Equal(t, ValueTypeArray, result.Type)
		require.Len(t, result.Array, 3)
		require.Equal(t, 1.0, result.Array[0].Number)
	})

	t.Run("Function taking array parameter", func(t *testing.T) {
		result := evalCode(t, `
		fn sum_array(arr: array) -> number {
			arr.get(0) + arr.get(1) + arr.get(2)
		}
		sum_array([10, 20, 30])`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 60.0, result.Number)
	})

	t.Run("Function modifying array", func(t *testing.T) {
		result := evalCode(t, `
		fn double_first(arr: array) -> number {
			set first = arr.get(0)
			arr.set(0, first * 2)
			arr.get(0)
		}
		set my_array = [5, 10, 15]
		double_first(my_array)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 10.0, result.Number)
	})
}

func TestBuiltinFunctions(t *testing.T) {
	t.Run("Print function", func(t *testing.T) {
		// Print function should work without error
		result := evalCode(t, `print("Hello, World!")`)
		require.Equal(t, ValueTypeNil, result.Type)
	})

	t.Run("Print with variable", func(t *testing.T) {
		result := evalCode(t, `set message = "Test"
		print(message)`)
		require.Equal(t, ValueTypeNil, result.Type)
	})

	t.Run("Print with number", func(t *testing.T) {
		result := evalCode(t, `print(42)`)
		require.Equal(t, ValueTypeNil, result.Type)
	})

	t.Run("Print with boolean", func(t *testing.T) {
		result := evalCode(t, `print(true)`)
		require.Equal(t, ValueTypeNil, result.Type)
	})

	t.Run("Print with array", func(t *testing.T) {
		result := evalCode(t, `print([1, 2, 3])`)
		require.Equal(t, ValueTypeNil, result.Type)
	})
}

func TestHigherOrderFunctions(t *testing.T) {
	t.Run("Function as parameter (simple)", func(t *testing.T) {
		result := evalCode(t, `
		fn apply_twice(f: function, x: number) -> number {
			f(f(x))
		}
		fn double(x: number) -> number {
			x * 2
		}
		apply_twice(double, 3)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 12.0, result.Number) // double(double(3)) = double(6) = 12
	})

	t.Run("Function returning function", func(t *testing.T) {
		result := evalCode(t, `
		fn make_adder(x: number) -> function {
			fn(y: number) -> number {
				x + y
			}
		}
		set add5 = make_adder(5)
		add5(3)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 8.0, result.Number)
	})
}

func TestFunctionErrors(t *testing.T) {
	t.Run("Undefined function call", func(t *testing.T) {
		err := evalCodeError(t, "undefined_function()")
		require.Error(t, err)
		require.Contains(t, err.Error(), "undefined function")
	})

	t.Run("Wrong number of parameters", func(t *testing.T) {
		err := evalCodeError(t, `
		fn add(a: number, b: number) -> number {
			a + b
		}
		add(5)`) // Missing one parameter
		require.Error(t, err)
	})

	t.Run("Too many parameters", func(t *testing.T) {
		err := evalCodeError(t, `
		fn add(a: number, b: number) -> number {
			a + b
		}
		add(5, 3, 7)`) // Extra parameter
		require.Error(t, err)
	})

	t.Run("Function redefinition", func(t *testing.T) {
		// This should either work (overwrite) or error consistently
		code := `
		fn test() -> number { 1 }
		fn test() -> number { 2 }
		test()`

		// Just check it doesn't crash
		result := evalCode(t, code)
		require.Equal(t, ValueTypeNumber, result.Type)
	})
}

func TestComplexFunctionScenarios(t *testing.T) {
	t.Run("Nested function calls", func(t *testing.T) {
		result := evalCode(t, `
		fn add(a: number, b: number) -> number {
			a + b
		}
		fn multiply(a: number, b: number) -> number {
			a * b
		}
		fn complex_calc(x: number) -> number {
			multiply(add(x, 5), add(x, 3))
		}
		complex_calc(2)`) // (2+5) * (2+3) = 7 * 5 = 35
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 35.0, result.Number)
	})

	t.Run("Function with conditional logic", func(t *testing.T) {
		result := evalCode(t, `
		fn max(a: number, b: number) -> number {
			if (a > b) {
				a
			} else {
				b
			}
		}
		max(10, 7)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 10.0, result.Number)
	})

	t.Run("Function creating and returning objects", func(t *testing.T) {
		result := evalCode(t, `
		fn create_person(name: string, age: number) -> object {
			{name: name, age: age}
		}
		set person = create_person("Alice", 30)
		person.name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Alice", result.Str)
	})
}

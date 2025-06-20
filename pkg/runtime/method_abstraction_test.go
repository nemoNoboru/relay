package runtime

import (
	"fmt"
	"testing"
)

// Test-Driven Development for Method Abstraction
// These tests define the desired behavior of our abstraction layer

func TestMethodDispatcherInterface(t *testing.T) {
	t.Run("Method dispatcher should handle array methods", func(t *testing.T) {
		// Create a method dispatcher (now this exists!)
		dispatcher := NewMethodDispatcher()

		// Register the dedicated array handler with all built-in methods
		arrayHandler := NewArrayMethodHandler()
		dispatcher.RegisterHandler(ValueTypeArray, arrayHandler)

		// Array with some data
		array := NewArray([]*Value{NewNumber(1), NewNumber(2), NewNumber(3)})

		// Test length method
		result, err := dispatcher.CallMethod(array, "length", []*Value{})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 3 {
			t.Errorf("Expected length 3, got: %v", result)
		}

		// Test get method
		result, err = dispatcher.CallMethod(array, "get", []*Value{NewNumber(1)})
		if err != nil {
			t.Fatalf("Expected no error for get(1), got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 2 {
			t.Errorf("Expected get(1) = 2, got: %v", result)
		}

		// Test push method
		result, err = dispatcher.CallMethod(array, "push", []*Value{NewNumber(4)})
		if err != nil {
			t.Fatalf("Expected no error for push(4), got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 4 {
			t.Errorf("Expected push to return length 4, got: %v", result)
		}

		// Verify the array was modified
		if len(array.Array) != 4 || array.Array[3].Number != 4 {
			t.Errorf("Expected array to be modified by push")
		}
	})
}

func TestTypeMethodHandler(t *testing.T) {
	t.Run("Should be able to register custom array methods", func(t *testing.T) {
		// Now we can actually implement this!
		handler := NewArrayMethodHandler()

		// Add a custom 'sum' method to arrays
		handler.AddMethod("sum", func(target *Value, args []*Value) (*Value, error) {
			if target.Type != ValueTypeArray {
				return nil, fmt.Errorf("expected array")
			}
			sum := 0.0
			for _, val := range target.Array {
				if val.Type == ValueTypeNumber {
					sum += val.Number
				}
			}
			return NewNumber(sum), nil
		})

		dispatcher := NewMethodDispatcher()
		dispatcher.RegisterHandler(ValueTypeArray, handler)

		// Test the custom method
		array := NewArray([]*Value{NewNumber(1), NewNumber(2), NewNumber(3)})
		result, err := dispatcher.CallMethod(array, "sum", []*Value{})
		if err != nil {
			t.Fatalf("Expected no error for sum method, got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 6 {
			t.Errorf("Expected sum = 6, got: %v", result)
		}
	})
}

func TestMethodCallWithoutParser(t *testing.T) {
	t.Run("Method calls should work without parser MethodCall struct", func(t *testing.T) {
		// Goal: be able to call methods with just:
		// - method name (string)
		// - arguments ([]*Value)
		// - no parser.MethodCall dependency

		obj := NewObject(map[string]*Value{
			"name": NewString("test"),
			"age":  NewNumber(25),
		})

		// We want: result, err := dispatcher.CallMethod(obj, "get", []*Value{NewString("name")})
		// Instead of current: evaluator.evaluateMethodCall(obj, parser.MethodCall{...}, env)

		// For now, verify object creation
		if obj.Object["name"].Str != "test" {
			t.Errorf("Object creation failed")
		}
	})
}

func TestMethodHandlerRegistration(t *testing.T) {
	t.Run("Should be able to extend methods for existing types", func(t *testing.T) {
		// Goal: Add new methods to arrays without modifying core code
		// dispatcher := NewMethodDispatcher()
		//
		// // Add a custom 'sum' method to arrays
		// arrayHandler := dispatcher.GetHandler(ValueTypeArray)
		// arrayHandler.AddMethod("sum", func(array *Value, args []*Value) (*Value, error) {
		//     sum := 0.0
		//     for _, val := range array.Array {
		//         if val.Type == ValueTypeNumber {
		//             sum += val.Number
		//         }
		//     }
		//     return NewNumber(sum), nil
		// })

		// This would allow extending functionality without parser changes
		array := NewArray([]*Value{NewNumber(1), NewNumber(2), NewNumber(3)})
		if len(array.Array) != 3 {
			t.Errorf("Array setup failed")
		}
	})
}

func TestArgumentEvaluationSeparation(t *testing.T) {
	t.Run("Argument evaluation should be separate from method execution", func(t *testing.T) {
		// Current problem: method implementations do their own argument evaluation
		// Better: arguments come pre-evaluated as []*Value

		// Goal architecture:
		// 1. Evaluator evaluates arguments to []*Value
		// 2. Method dispatcher receives pre-evaluated arguments
		// 3. Method handlers work with concrete values only

		args := []*Value{NewNumber(42), NewString("hello")}

		// These are already evaluated - no parser.Expression needed
		if args[0].Type != ValueTypeNumber || args[0].Number != 42 {
			t.Errorf("Pre-evaluated arguments test failed")
		}
	})
}

func TestImmutableObjectMethods(t *testing.T) {
	t.Run("Object methods should return new objects (immutable)", func(t *testing.T) {
		original := NewObject(map[string]*Value{
			"x": NewNumber(1),
		})

		// Goal: obj.set("y", 2) returns new object, doesn't mutate original
		// result, err := dispatcher.CallMethod(original, "set", []*Value{NewString("y"), NewNumber(2)})

		// Original should be unchanged
		if len(original.Object) != 1 {
			t.Errorf("Original object should have 1 field")
		}

		// Result should have 2 fields (when we implement it)
	})
}

func TestMutableServerStateMethods(t *testing.T) {
	t.Run("Server state methods should mutate in place", func(t *testing.T) {
		state := make(map[string]*Value)
		state["count"] = NewNumber(0)

		// ServerState should allow mutation
		// This is different from regular objects
		if state["count"].Number != 0 {
			t.Errorf("Server state setup failed")
		}
	})
}

func TestHigherOrderArrayMethods(t *testing.T) {
	t.Run("Array methods like map/filter should work with function arguments", func(t *testing.T) {
		array := NewArray([]*Value{NewNumber(1), NewNumber(2), NewNumber(3)})

		// Goal: array.map(fn(x) { x * 2 })
		// Method handler should receive the function as an evaluated argument
		// and be able to call it without knowing about evaluator internals

		if len(array.Array) != 3 {
			t.Errorf("Array setup failed")
		}
	})
}

// Test what the interface should look like
func TestDesiredMethodDispatcherAPI(t *testing.T) {
	t.Run("Define the API we want", func(t *testing.T) {
		// This test documents the desired API - and now it works!

		// 1. Create a dispatcher
		dispatcher := NewMethodDispatcher()

		// 2. Register built-in method handlers
		dispatcher.RegisterHandler(ValueTypeArray, NewArrayMethodHandler())
		// dispatcher.RegisterHandler(ValueTypeObject, NewObjectMethodHandler()) // TODO
		// dispatcher.RegisterHandler(ValueTypeStruct, NewStructMethodHandler()) // TODO

		// 3. Call methods without parser dependency
		array := NewArray([]*Value{NewNumber(1), NewNumber(2)})
		result, err := dispatcher.CallMethod(array, "length", []*Value{})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 2 {
			t.Errorf("Expected length 2, got: %v", result)
		}

		// 4. Add custom methods
		arrayHandler := dispatcher.GetHandler(ValueTypeArray)
		arrayHandler.AddMethod("sum", func(target *Value, args []*Value) (*Value, error) {
			sum := 0.0
			for _, val := range target.Array {
				if val.Type == ValueTypeNumber {
					sum += val.Number
				}
			}
			return NewNumber(sum), nil
		})

		// Test the custom method
		result, err = dispatcher.CallMethod(array, "sum", []*Value{})
		if err != nil {
			t.Fatalf("Expected no error for sum, got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 3 {
			t.Errorf("Expected sum = 3, got: %v", result)
		}
	})
}

func TestEvaluatorMethodIntegration(t *testing.T) {
	t.Run("Evaluator should have method dispatcher integrated", func(t *testing.T) {
		// Create an evaluator with integrated method dispatcher
		evaluator := NewEvaluator()

		// Test that we can call methods directly through the evaluator's dispatcher
		array := NewArray([]*Value{NewNumber(1), NewNumber(2), NewNumber(3)})

		// Call method through integrated dispatcher
		result, err := evaluator.methodDispatcher.CallMethod(array, "length", []*Value{})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 3 {
			t.Errorf("Expected length 3, got: %v", result)
		}

		// Test array methods work
		result, err = evaluator.methodDispatcher.CallMethod(array, "push", []*Value{NewNumber(4)})
		if err != nil {
			t.Fatalf("Expected no error for push, got: %v", err)
		}

		if result.Type != ValueTypeNumber || result.Number != 4 {
			t.Errorf("Expected push to return 4, got: %v", result)
		}

		// Test that the array was actually modified
		if len(array.Array) != 4 {
			t.Errorf("Expected array length 4 after push, got: %d", len(array.Array))
		}
	})
}

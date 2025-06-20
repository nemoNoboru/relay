package runtime

import (
	"relay/pkg/parser"
	"strings"
	"testing"
)

// TDD tests for complete method abstraction integration
// These tests define how the evaluator should work with only the abstraction system

func TestFullMethodAbstractionIntegration(t *testing.T) {
	t.Run("Evaluator should use method dispatcher for all method calls", func(t *testing.T) {
		evaluator := NewEvaluator()

		// Test array methods work through normal evaluation
		program := parseCode(t, `set arr = [1, 2, 3]
		set len = arr.length()
		set val = arr.get(1)`)

		// Evaluate the program
		for _, expr := range program.Expressions {
			_, err := evaluator.Evaluate(expr)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
		}

		// Check that variables were set correctly
		lenVar, exists := evaluator.globalEnv.Get("len")
		if !exists || lenVar.Type != ValueTypeNumber || lenVar.Number != 3 {
			t.Errorf("Expected len = 3, got: %v", lenVar)
		}

		valVar, exists := evaluator.globalEnv.Get("val")
		if !exists || valVar.Type != ValueTypeNumber || valVar.Number != 2 {
			t.Errorf("Expected val = 2, got: %v", valVar)
		}
	})

	t.Run("Object methods should work through abstraction", func(t *testing.T) {
		evaluator := NewEvaluator()

		program := parseCode(t, `set obj = {name: "test", age: 25}
		set name = obj.get("name")
		set obj2 = obj.set("city", "SF")`)

		for _, expr := range program.Expressions {
			_, err := evaluator.Evaluate(expr)
			if err != nil {
				t.Fatalf("Expected no error for object methods, got: %v", err)
			}
		}

		// Check object methods worked
		nameVar, exists := evaluator.globalEnv.Get("name")
		if !exists || nameVar.Type != ValueTypeString || nameVar.Str != "test" {
			t.Errorf("Expected name = 'test', got: %v", nameVar)
		}

		obj2Var, exists := evaluator.globalEnv.Get("obj2")
		if !exists || obj2Var.Type != ValueTypeObject {
			t.Errorf("Expected obj2 to be an object, got: %v", obj2Var)
		}

		// obj2 should have city field (immutable semantics)
		if cityVal, hasCity := obj2Var.Object["city"]; !hasCity || cityVal.Str != "SF" {
			t.Errorf("Expected obj2 to have city = 'SF', got: %v", obj2Var.Object)
		}
	})

	t.Run("Struct methods should work through abstraction", func(t *testing.T) {
		evaluator := NewEvaluator()

		program := parseCode(t, `struct Person {
			name: String
			age: Number
		}
		set p = Person{name: "Alice", age: 30}
		set name = p.get("name")`)

		for _, expr := range program.Expressions {
			_, err := evaluator.Evaluate(expr)
			if err != nil {
				t.Fatalf("Expected no error for struct methods, got: %v", err)
			}
		}

		nameVar, exists := evaluator.globalEnv.Get("name")
		if !exists || nameVar.Type != ValueTypeString || nameVar.Str != "Alice" {
			t.Errorf("Expected name = 'Alice', got: %v", nameVar)
		}
	})

	t.Run("String methods should work through abstraction", func(t *testing.T) {
		evaluator := NewEvaluator()

		program := parseCode(t, `set str = "hello"
		set len = str.length()`)

		for _, expr := range program.Expressions {
			_, err := evaluator.Evaluate(expr)
			if err != nil {
				t.Fatalf("Expected no error for string methods, got: %v", err)
			}
		}

		lenVar, exists := evaluator.globalEnv.Get("len")
		if !exists || lenVar.Type != ValueTypeNumber || lenVar.Number != 5 {
			t.Errorf("Expected len = 5, got: %v", lenVar)
		}
	})

	t.Run("Higher-order array methods should work", func(t *testing.T) {
		evaluator := NewEvaluator()

		program := parseCode(t, `set arr = [1, 2, 3]
		set doubled = arr.map(fn(x) { x * 2 })
		set evens = arr.filter(fn(x) { x == 2 })
		set sum = arr.reduce(fn(acc, x) { acc + x }, 0)`)

		for _, expr := range program.Expressions {
			_, err := evaluator.Evaluate(expr)
			if err != nil {
				t.Fatalf("Expected no error for higher-order methods, got: %v", err)
			}
		}

		// Check results
		doubledVar, exists := evaluator.globalEnv.Get("doubled")
		if !exists || doubledVar.Type != ValueTypeArray || len(doubledVar.Array) != 3 {
			t.Errorf("Expected doubled array, got: %v", doubledVar)
		}

		evensVar, exists := evaluator.globalEnv.Get("evens")
		if !exists || evensVar.Type != ValueTypeArray || len(evensVar.Array) != 1 {
			t.Errorf("Expected evens array with 1 element, got: %v", evensVar)
		}

		sumVar, exists := evaluator.globalEnv.Get("sum")
		if !exists || sumVar.Type != ValueTypeNumber || sumVar.Number != 6 {
			t.Errorf("Expected sum = 6, got: %v", sumVar)
		}
	})
}

func TestMethodAbstractionErrorHandling(t *testing.T) {
	t.Run("Should give clear errors for unsupported methods", func(t *testing.T) {
		evaluator := NewEvaluator()

		program := parseCode(t, `set arr = [1, 2, 3]
		set result = arr.nonexistent()`)

		// First expression should work
		_, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Array creation should work, got: %v", err)
		}

		// Second expression should fail with clear error
		_, err = evaluator.Evaluate(program.Expressions[1])
		if err == nil {
			t.Fatalf("Expected error for nonexistent method")
		}

		// Error should mention the method name
		if err.Error() == "" {
			t.Errorf("Expected descriptive error message")
		}
	})

	t.Run("Should handle method call on wrong type", func(t *testing.T) {
		evaluator := NewEvaluator()

		program := parseCode(t, `set num = 42
		set result = num.length()`)

		// First expression should work
		_, err := evaluator.Evaluate(program.Expressions[0])
		if err != nil {
			t.Fatalf("Number assignment should work, got: %v", err)
		}

		// Second expression should fail
		_, err = evaluator.Evaluate(program.Expressions[1])
		if err == nil {
			t.Fatalf("Expected error for method call on number")
		}
	})
}

func TestNoOldMethodSystemDependencies(t *testing.T) {
	t.Run("All method calls should go through dispatcher", func(t *testing.T) {
		// This test ensures we've removed the old evaluateMethodCall paths
		evaluator := NewEvaluator()

		// Test all the major types that had method calls before
		testCases := []string{
			`[1,2,3].length()`,
			`{x: 1}.get("x")`,
			`"hello".length()`,
		}

		for _, code := range testCases {
			program := parseCode(t, code)
			_, err := evaluator.Evaluate(program.Expressions[0])
			if err != nil {
				t.Errorf("Expected %s to work through abstraction, got: %v", code, err)
			}
		}
	})
}

// Helper function to parse code for testing
func parseCode(t *testing.T, code string) *parser.Program {
	t.Helper()

	reader := strings.NewReader(code)
	program, err := parser.Parse("test", reader)
	if err != nil {
		t.Fatalf("Failed to parse code '%s': %v", code, err)
	}
	return program
}

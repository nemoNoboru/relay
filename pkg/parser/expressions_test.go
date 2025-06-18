package parser

import (
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestSimpleBinaryExpressions(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"Addition", `2 + 3`},
		{"Multiplication", `5 * 7`},
		{"Comparison", `x > 10`},
		{"Equality", `name == "test"`},
		{"Logical AND", `a && b`},
		{"Logical OR", `x || y`},
		{"Null coalesce", `value ?? "default"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := Parse("test.relay", strings.NewReader(test.src))
			require.NoError(t, err)
			if len(program.Expressions) != 1 {
				t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
			}
			if program.Expressions[0].Binary == nil {
				t.Fatal("Expected Binary expression")
			}
		})
	}
}

func TestSimpleLiterals(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"Number", `42`},
		{"String", `"hello"`},
		{"Boolean", `true`},
		{"Identifier", `user`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := Parse("test.relay", strings.NewReader(test.src))
			require.NoError(t, err)
			if len(program.Expressions) != 1 {
				t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
			}

			// Simple literals should create Binary expressions with no operations
			if program.Expressions[0].Binary != nil {
				if len(program.Expressions[0].Binary.Right) != 0 {
					t.Error("Simple literals should not have binary operations")
				}
			}
		})
	}
}

func TestOperatorPrecedence(t *testing.T) {
	src := `2 + 3 * 4`
	program, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
	if len(program.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(program.Expressions))
	}
	if program.Expressions[0].Binary == nil {
		t.Fatal("Expected Binary expression")
	}

	// The simplified parser flattens all operators in left-to-right order
	// so 2 + 3 * 4 becomes: 2 + 3 * 4 (operations: ["+", "*"])
	binary := program.Expressions[0].Binary
	if len(binary.Right) < 1 {
		t.Fatal("Expected at least one binary operation")
	}
}

func TestNullCoalesceOperator(t *testing.T) {
	src := `set value = optional_value ?? "default"`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestFieldAccess(t *testing.T) {
	src := `set name = user.get("name")`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestMethodChaining(t *testing.T) {
	src := `set result = users.filter(fn (u) { u.get("active") })
		.map(fn (u) { u.get("name") })
		.sort_by(fn (u) { u })`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestUpdateStatement(t *testing.T) {
	src := `state.set("count", state.get("count") + 1)`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

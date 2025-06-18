package runtime

import (
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestIfElseBasics(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		expected  interface{}
		valueType ValueType
	}{
		{
			"Simple if true",
			`if true { "success" }`,
			"success",
			ValueTypeString,
		},
		{
			"Simple if false",
			`if false { "success" }`,
			nil,
			ValueTypeNil,
		},
		{
			"If else true condition",
			`if true { "then" } else { "else" }`,
			"then",
			ValueTypeString,
		},
		{
			"If else false condition",
			`if false { "then" } else { "else" }`,
			"else",
			ValueTypeString,
		},
		{
			"If with number comparison",
			"set x = 10\nif x > 5 { \"greater\" } else { \"lesser\" }",
			"greater",
			ValueTypeString,
		},
		{
			"If with string comparison",
			"set name = \"Alice\"\nif name == \"Alice\" { \"found\" } else { \"not found\" }",
			"found",
			ValueTypeString,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, test.valueType, result.Type)

			switch test.valueType {
			case ValueTypeString:
				require.Equal(t, test.expected.(string), result.Str)
			case ValueTypeNumber:
				require.Equal(t, test.expected.(float64), result.Number)
			case ValueTypeBool:
				require.Equal(t, test.expected.(bool), result.Bool)
			case ValueTypeNil:
				// nil is expected
			}
		})
	}
}

func TestIfElseTruthiness(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		expected  string
	}{
		{"Non-empty string", `"hello"`, "truthy"},
		{"Empty string", `""`, "falsy"},
		{"Positive number", `42`, "truthy"},
		{"Zero", `0`, "falsy"},
		{"Negative number", `-5`, "truthy"},
		{"True boolean", `true`, "truthy"},
		{"False boolean", `false`, "falsy"},
		{"Nil value", `nil`, "falsy"},
		{"Non-empty array", `[1, 2, 3]`, "truthy"},
		{"Empty array", `[]`, "falsy"},
		{"Non-empty object", `{name: "test"}`, "truthy"},
		{"Empty object", `{}`, "falsy"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code := `if ` + test.condition + ` { "truthy" } else { "falsy" }`
			result := evalCode(t, code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		})
	}
}

func TestNestedIfElse(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Nested if-else positive",
			`set x = 10
			if x > 0 {
				if x > 5 {
					"large positive"
				} else {
					"small positive"
				}
			} else {
				"not positive"
			}`,
			"large positive",
		},
		{
			"Nested if-else small positive",
			`set x = 3
			if x > 0 {
				if x > 5 {
					"large positive"
				} else {
					"small positive"
				}
			} else {
				"not positive"
			}`,
			"small positive",
		},
		{
			"Nested if-else negative",
			`set x = -2
			if x > 0 {
				if x > 5 {
					"large positive"
				} else {
					"small positive"
				}
			} else {
				"not positive"
			}`,
			"not positive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		})
	}
}

func TestIfElseWithFunctions(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"Function with if-else",
			`fn getSign(x) {
				if x > 0 {
					"positive"
				} else {
					if x < 0 {
						"negative"
					} else {
						"zero"
					}
				}
			}
			getSign(5)`,
			"positive",
		},
		{
			"Function with if-else negative",
			`fn getSign(x) {
				if x > 0 {
					"positive"
				} else {
					if x < 0 {
						"negative"
					} else {
						"zero"
					}
				}
			}
			getSign(-3)`,
			"negative",
		},
		{
			"Function with if-else zero",
			`fn getSign(x) {
				if x > 0 {
					"positive"
				} else {
					if x < 0 {
						"negative"
					} else {
						"zero"
					}
				}
			}
			getSign(0)`,
			"zero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		})
	}
}

func TestIfElseWithNullCoalescing(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"If returns value, no coalescing",
			`fn getValue(hasValue) {
				if hasValue {
					"actual"
				} else {
					nil
				}
			}
			getValue(true) ?? "default"`,
			"actual",
		},
		{
			"If returns nil, coalescing kicks in",
			`fn getValue(hasValue) {
				if hasValue {
					"actual"
				} else {
					nil
				}
			}
			getValue(false) ?? "default"`,
			"default",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, ValueTypeString, result.Type)
			require.Equal(t, test.expected, result.Str)
		})
	}
}

func TestIfElseExpressionBlocks(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		expected  interface{}
		valueType ValueType
	}{
		{
			"If block with multiple statements",
			`set x = 5
			if x > 0 {
				set y = x * 2
				set z = y + 1
				z
			}`,
			11.0,
			ValueTypeNumber,
		},
		{
			"Else block with multiple statements",
			`set x = -5
			if x > 0 {
				"positive"
			} else {
				set abs_x = x * -1
				abs_x
			}`,
			5.0,
			ValueTypeNumber,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := evalCode(t, test.code)
			require.Equal(t, test.valueType, result.Type)

			switch test.valueType {
			case ValueTypeNumber:
				require.Equal(t, test.expected.(float64), result.Number)
			case ValueTypeString:
				require.Equal(t, test.expected.(string), result.Str)
			}
		})
	}
}

func TestIfElseErrors(t *testing.T) {
	errorTests := []struct {
		name string
		code string
	}{
		{
			"Error in condition evaluation",
			`if undefined_variable { "test" }`,
		},
		{
			"Error in then block",
			`if true { undefined_variable }`,
		},
		{
			"Error in else block",
			`if false { "ok" } else { undefined_variable }`,
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			err := evalCodeError(t, test.code)
			require.Error(t, err, "Expected error for: %s", test.code)
		})
	}
}

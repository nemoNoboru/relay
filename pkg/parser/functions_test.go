package parser

import (
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestFunctionDefinition(t *testing.T) {
	src := `fn calculate_total(items: [object]) -> number {
		items.map(fn (item) { item.get("price") }).reduce(fn (acc, price) { acc + price }, 0)
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestLambdaExpression(t *testing.T) {
	src := `set doubled = numbers.map(fn (x) { x * 2 })`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestFunctionTypes(t *testing.T) {
	src := `fn apply_operation(op: fn(number) -> number, value: number) -> number {
		op(value)
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

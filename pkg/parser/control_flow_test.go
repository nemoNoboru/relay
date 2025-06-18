package parser

import (
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestBlockExpression(t *testing.T) {
	src := `set result = {
		set x = 10
		set y = 20
		x + y
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestIfExpression(t *testing.T) {
	src := `set message = if count > 0 { "items found" } else { "no items" }`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestControlFlow(t *testing.T) {
	src := `for user in users {
		set processed = user.get("name")
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestErrorHandling(t *testing.T) {
	src := `set result = risky_operation()
	if result.get("error") != nil {
		set result = "failed"
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

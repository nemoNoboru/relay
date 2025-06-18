package parser

import (
	"strings"
	"testing"

	require "github.com/alecthomas/assert/v2"
)

func TestSendExpression(t *testing.T) {
	src := `set posts = send "blog_service" get_posts {}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestTemplateDeclaration(t *testing.T) {
	src := `template "index.html" from get_posts()`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestTemplateWithParams(t *testing.T) {
	src := `template "post.html" from get_post(id: string)`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestConfigDefinition(t *testing.T) {
	src := `config {
		app_name: "myblog",
		port: 8080,
		federation: {
			auto_discover: true,
			cache_duration: "1h"
		}
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

func TestSymbolLiterals(t *testing.T) {
	// TODO: Implement dispatch expressions in parser
	// This test is skipped until dispatch syntax is implemented
	t.Skip("Dispatch expressions not yet implemented in simplified parser")

	src := `dispatch action {
		:create: fn (data) { "created" },
		:update: fn (data) { "updated" },
		:delete: fn (data) { "deleted" }
	}`
	_, err := Parse("test.relay", strings.NewReader(src))
	require.NoError(t, err)
}

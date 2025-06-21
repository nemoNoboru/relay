package runtime

import (
	"bytes"
	"relay/pkg/parser"
	"testing"
)

// MustParse is a test helper that parses a string into a program or fails the test.
// It simplifies the setup of tests that require a valid AST.
func MustParse(t *testing.T, code string) *parser.Program {
	t.Helper()
	prog, err := parser.Parse("test.rl", bytes.NewBufferString(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	return prog
}

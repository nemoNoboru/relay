package lexer

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestRelayFileTokenization(t *testing.T) {
	// Read the example relay file
	examplePath := filepath.Join("..", "..", "examples", "simple_blog.relay")
	content, err := ioutil.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example file: %v", err)
	}

	l := New(string(content))

	var tokens []Token
	keywordCounts := make(map[TokenType]int)

	// Tokenize the entire file
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)

		// Count important keywords
		switch tok.Type {
		case STRUCT, PROTOCOL, SERVER, STATE, RECEIVE, TEMPLATE, CONFIG:
			keywordCounts[tok.Type]++
		}

		if tok.Type == EOF {
			break
		}

		// Safety check to prevent infinite loops
		if len(tokens) > 1000 {
			t.Fatal("Too many tokens, possible infinite loop")
		}
	}

	// Verify we got a reasonable number of tokens
	if len(tokens) < 50 {
		t.Fatalf("Expected at least 50 tokens, got %d", len(tokens))
	}

	// Verify we found the expected keywords from simple_blog.relay
	// Make the test more flexible since there might be variations
	requiredKeywords := map[TokenType]int{
		STRUCT:   1, // struct Post
		PROTOCOL: 1, // protocol BlogService
		SERVER:   1, // server blog_service
		STATE:    1, // state block
		RECEIVE:  2, // receive get_posts, receive create_post
		TEMPLATE: 2, // template declarations
		CONFIG:   1, // config block
	}

	for keyword, minCount := range requiredKeywords {
		if keywordCounts[keyword] < minCount {
			t.Errorf("Expected at least %d instances of keyword %v, got %d",
				minCount, keyword, keywordCounts[keyword])
		}
	}

	// Print some stats for debugging
	t.Logf("Successfully tokenized %d tokens from simple_blog.relay", len(tokens))
	t.Logf("Keyword counts: %+v", keywordCounts)
}

func TestTokenizeAndPrint(t *testing.T) {
	// Simple test input
	input := `struct User {
  name: string.min(5)
}

server {
  receive greet {name: string} -> string {
    return "Hello " + name
  }
}`

	l := New(input)

	t.Log("Tokenizing Relay code:")
	t.Log(input)
	t.Log("\nTokens:")

	for i := 0; i < 30; i++ { // Limit to first 30 tokens
		tok := l.NextToken()
		if tok.Type == EOF {
			t.Logf("%d: %s", i, tok.String())
			break
		}
		t.Logf("%d: %s", i, tok.String())
	}
}

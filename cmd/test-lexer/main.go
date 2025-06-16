package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudpunks/relay/pkg/lexer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-lexer <relay-file> or test-lexer -")
		fmt.Println("  Use '-' to read from stdin")
		os.Exit(1)
	}

	var input string

	if os.Args[1] == "-" {
		// Read from stdin
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		input = string(data)
	} else {
		// Read from file
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", os.Args[1], err)
			os.Exit(1)
		}
		input = string(data)
	}

	fmt.Printf("Tokenizing:\n%s\n", input)
	fmt.Println("=" + string(make([]byte, 50, 50)) + "=")

	l := lexer.New(input)

	tokenCount := 0
	for {
		tok := l.NextToken()

		fmt.Printf("%3d: %s\n", tokenCount, tok.String())

		tokenCount++
		if tok.Type == lexer.EOF {
			break
		}

		// Safety check
		if tokenCount > 200 {
			fmt.Println("... (stopping after 200 tokens)")
			break
		}
	}

	fmt.Printf("\nTotal tokens: %d\n", tokenCount)
}

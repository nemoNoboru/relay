// Package main provides the Relay language command-line interface.
//
// The Relay CLI supports multiple modes:
// - Interactive REPL for development and testing
// - File execution for running .relay/.rl programs
// - Build mode for compilation (future feature)
// - File loading with REPL for debugging
//
// The runtime features a unified evaluation architecture with full support for:
// - First-class functions with closures
// - Struct definitions and instantiation
// - Server state management with concurrency
// - Immutable-by-default semantics
// - Comprehensive expression evaluation (739 tests passing)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"relay/pkg/parser"
	"relay/pkg/repl"
	"relay/pkg/runtime"
)

const version = "0.3.0-dev"

func main() {
	var (
		versionFlag = flag.Bool("version", false, "Show version")
		helpFlag    = flag.Bool("help", false, "Show help")
		runFlag     = flag.String("run", "", "Run a .relay/.rl file")
		buildFlag   = flag.String("build", "", "Build a .relay/.rl file to binary")
		replFlag    = flag.Bool("repl", false, "Start interactive REPL")
		portFlag    = flag.Int("port", 8080, "Port to run server on")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Relay v%s - Cloudpunks Edition\n", version)
		return
	}

	if *helpFlag || (flag.NArg() == 0 && *runFlag == "" && *buildFlag == "" && !*replFlag) {
		showHelp()
		return
	}

	switch {
	case *replFlag:
		if err := startREPL(""); err != nil {
			log.Fatalf("Error starting REPL: %v", err)
		}
	case *runFlag != "":
		if err := runRelayFile(*runFlag, *portFlag, false); err != nil {
			log.Fatalf("Error running %s: %v", *runFlag, err)
		}
	case *buildFlag != "":
		if err := buildRelayFile(*buildFlag); err != nil {
			log.Fatalf("Error building %s: %v", *buildFlag, err)
		}
	default:
		// Handle positional arguments
		if flag.NArg() > 0 {
			filename := flag.Arg(0)

			// Check if user wants to load file and start REPL
			if flag.NArg() > 1 && flag.Arg(1) == "-repl" {
				if err := runRelayFile(filename, *portFlag, true); err != nil {
					log.Fatalf("Error loading %s: %v", filename, err)
				}
			} else {
				if err := runRelayFile(filename, *portFlag, false); err != nil {
					log.Fatalf("Error running %s: %v", filename, err)
				}
			}
		}
	}
}

func showHelp() {
	fmt.Println("Relay - Federated Web Programming Language")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  relay [options] <file.relay|file.rl>")
	fmt.Println("  relay <file.relay|file.rl> -repl     # Load file and start REPL")
	fmt.Println("  relay -run <file.relay|file.rl>")
	fmt.Println("  relay -build <file.relay|file.rl>")
	fmt.Println("  relay -repl")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version     Show version")
	fmt.Println("  -help        Show this help")
	fmt.Println("  -run         Run a .relay/.rl file")
	fmt.Println("  -build       Build a .relay/.rl file to binary")
	fmt.Println("  -repl        Start interactive REPL")
	fmt.Println("  -port        Port to run server on (default: 8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  relay blog.rl")
	fmt.Println("  relay debug.rl -repl                 # Load debug.rl then start REPL")
	fmt.Println("  relay -run blog.relay -port 3000")
	fmt.Println("  relay -build blog.relay")
	fmt.Println("  relay -repl")
}

func startREPL(preloadFile string) error {
	r := repl.New(os.Stdin, os.Stdout)

	// If a file was specified, load it first
	if preloadFile != "" {
		fmt.Printf("Loading %s...\n", preloadFile)
		if err := r.LoadFile(preloadFile); err != nil {
			return fmt.Errorf("failed to load %s: %v", preloadFile, err)
		}
		fmt.Printf("File loaded successfully!\n")
	}

	r.Start()
	return nil
}

func runRelayFile(filename string, port int, startREPL bool) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}

	// Check file extension
	ext := filepath.Ext(filename)
	if ext != ".relay" && ext != ".rl" {
		return fmt.Errorf("unsupported file extension %s (expected .relay or .rl)", ext)
	}

	// Read the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Parse the file
	program, err := parser.Parse(filename, file)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	// Create evaluator
	evaluator := runtime.NewEvaluator()

	// Execute the program
	fmt.Printf("Executing %s...\n", filename)
	var lastResult *runtime.Value
	for i, expr := range program.Expressions {
		result, err := evaluator.Evaluate(expr)
		if err != nil {
			return fmt.Errorf("runtime error at expression %d: %v", i+1, err)
		}
		lastResult = result
	}

	// Show the result of the last expression (if any)
	if lastResult != nil && len(program.Expressions) > 0 {
		// Only show result if it's not nil and not a function/struct definition
		if lastResult.Type != runtime.ValueTypeNil {
			fmt.Printf("Result: %s\n", lastResult.String())
		}
	}

	// If requested, start REPL with the loaded context
	if startREPL {
		fmt.Println("\nStarting REPL with loaded context...")
		r := repl.NewWithEvaluator(os.Stdin, os.Stdout, evaluator)
		r.Start()
	} else {
		fmt.Printf("Execution completed successfully.\n")
	}

	return nil
}

func buildRelayFile(filename string) error {
	fmt.Printf("Building %s...\n", filename)
	// TODO: Implement relay file compilation
	return fmt.Errorf("not implemented yet")
}

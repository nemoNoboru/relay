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
	"time"

	"relay/pkg/actor"
	"relay/pkg/repl"
)

var (
	version = "0.3.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		versionFlag = flag.Bool("version", false, "Show version")
		helpFlag    = flag.Bool("help", false, "Show help")
		runFlag     = flag.String("run", "", "Run a .relay/.rl file")
		buildFlag   = flag.String("build", "", "Build a .relay/.rl file to binary")
		replFlag    = flag.Bool("repl", false, "Start interactive REPL")
		serverFlag  = flag.Bool("server", false, "Start HTTP server with JSON-RPC 2.0 endpoints")
		portFlag    = flag.Int("port", 8080, "Port to run server on")
		hostFlag    = flag.String("host", "0.0.0.0", "Host to bind server to")
		// New P2P flags
		nodeIDFlag     = flag.String("node-id", "", "Node ID for peer-to-peer networking (auto-generated if not provided)")
		addPeerFlag    = flag.String("add-peer", "", "Add a peer node (format: http://host:port)")
		disableRegFlag = flag.Bool("disable-registry", false, "Disable server registry for peer-to-peer functionality")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Relay v%s - Cloudpunks Edition\n", version)
		return
	}

	if *helpFlag || (flag.NArg() == 0 && *runFlag == "" && *buildFlag == "" && !*replFlag && !*serverFlag) {
		showHelp()
		return
	}

	// Determine the main action based on flags
	action := determineAction(runFlag, buildFlag, replFlag, serverFlag)

	// Execute the action
	switch action {
	case "run_with_server":
		if err := runRelayFileWithServer(*runFlag, *hostFlag, *portFlag, *nodeIDFlag, *addPeerFlag, *disableRegFlag); err != nil {
			log.Fatalf("Error starting server with %s: %v", *runFlag, err)
		}
	case "server":
		if err := startHTTPServer(*hostFlag, *portFlag, *nodeIDFlag, *addPeerFlag, *disableRegFlag); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	case "repl":
		if err := startREPL(""); err != nil {
			log.Fatalf("Error starting REPL: %v", err)
		}
	case "run":
		if err := runRelayFile(*runFlag, *portFlag, false); err != nil {
			log.Fatalf("Error running %s: %v", *runFlag, err)
		}
	case "build":
		if err := buildRelayFile(*buildFlag); err != nil {
			log.Fatalf("Error building %s: %v", *buildFlag, err)
		}
	default:
		// Default to REPL if no specific action is determined
		if err := startREPL(""); err != nil {
			log.Fatalf("Error starting REPL: %v", err)
		}
	}
}

func determineAction(runFlag, buildFlag *string, replFlag, serverFlag *bool) string {
	if *serverFlag && *runFlag != "" {
		return "run_with_server"
	}
	if *serverFlag {
		return "server"
	}
	if *replFlag {
		return "repl"
	}
	if *runFlag != "" {
		return "run"
	}
	if *buildFlag != "" {
		return "build"
	}
	return "repl" // Default action
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
	fmt.Println("  relay -server")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version     Show version")
	fmt.Println("  -help        Show this help")
	fmt.Println("  -run         Run a .relay/.rl file")
	fmt.Println("  -build       Build a .relay/.rl file to binary")
	fmt.Println("  -repl        Start interactive REPL")
	fmt.Println("  -server      Start HTTP server with JSON-RPC 2.0 endpoints")
	fmt.Println("  -port        Port to run server on (default: 8080)")
	fmt.Println("  -host        Host to bind server to (default: 0.0.0.0)")
	fmt.Println("  -node-id     Node ID for peer-to-peer networking (auto-generated if not provided)")
	fmt.Println("  -add-peer    Add a peer node (format: http://host:port)")
	fmt.Println("  -disable-registry Disable server registry for peer-to-peer functionality")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  relay blog.rl")
	fmt.Println("  relay debug.rl -repl                 # Load debug.rl then start REPL")
	fmt.Println("  relay -run blog.relay -port 3000")
	fmt.Println("  relay -build blog.relay")
	fmt.Println("  relay -repl")
	fmt.Println("  relay -server -port 9090 -host 127.0.0.1")
}

func startREPL(preloadFile string) error {
	// Setup the main actor system components
	router := actor.NewRouter()

	supervisor := actor.NewSupervisorActor("supervisor", router)
	router.Register(supervisor.Actor.Name, supervisor.Actor)
	go supervisor.Start()

	// Create a primary RelayServerActor for the REPL session
	// by sending a message to the supervisor and waiting for a reply.
	replyChan := make(chan actor.ActorMsg, 1)
	router.Send(actor.ActorMsg{
		To:        "supervisor",
		From:      "main",
		Type:      "create_child",
		Data:      "RelayServerActor",
		ReplyChan: replyChan,
	})

	// Wait for the supervisor's reply
	var replActorName string
	select {
	case reply := <-replyChan:
		if name, ok := reply.Data.(string); ok {
			replActorName = name
		} else {
			return fmt.Errorf("failed to get REPL actor name from supervisor")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for REPL actor to be created")
	}

	fmt.Printf("Successfully created REPL actor: %s\n", replActorName)
	fmt.Println("Type ':help' for a list of commands.")

	// Start the new actor-based REPL
	r := repl.New(os.Stdin, os.Stdout, router, replActorName)

	// If a file was specified, load it first
	if preloadFile != "" {
		fmt.Printf("Loading %s...\n", preloadFile)
		if err := r.LoadFile(preloadFile); err != nil {
			return fmt.Errorf("failed to load %s: %v", preloadFile, err)
		}
		fmt.Printf("File loaded successfully!\n")
	}

	r.Start()
	// Add a small delay to ensure all actor messages are processed before exit.
	time.Sleep(100 * time.Millisecond)
	return nil
}

func startHTTPServer(host string, port int, nodeID, addPeer string, disableRegistry bool) error {
	fmt.Println("Server mode is not yet implemented with the new actor model.")
	return nil
}

func runRelayFile(filename string, port int, shouldStartREPL bool) error {
	// Setup the main actor system components
	router := actor.NewRouter()

	supervisor := actor.NewSupervisorActor("supervisor", router)
	router.Register(supervisor.Actor.Name, supervisor.Actor)
	go supervisor.Start()

	// Create a primary RelayServerActor for the script and wait for the reply.
	replyChan := make(chan actor.ActorMsg, 1)
	router.Send(actor.ActorMsg{
		To:        "supervisor",
		From:      "main",
		Type:      "create_child",
		Data:      "RelayServerActor",
		ReplyChan: replyChan,
	})

	var scriptActorName string
	select {
	case reply := <-replyChan:
		if name, ok := reply.Data.(string); ok {
			scriptActorName = name
		} else {
			return fmt.Errorf("failed to get script actor name from supervisor")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for script actor to be created")
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Send the script content to the newly created actor for evaluation.
	fmt.Printf("Executing %s via actor %s...\n", filename, scriptActorName)
	router.Send(actor.ActorMsg{
		To:   scriptActorName,
		From: "main",
		Type: "eval",
		Data: string(content),
	})

	// NOTE: This is now asynchronous. We need a way to wait for the result
	// if we want to print it or start a REPL with the resulting state.
	time.Sleep(500 * time.Millisecond) // Simple wait for evaluation to complete.

	// Check file extension
	ext := filepath.Ext(filename)
	if ext != ".relay" && ext != ".rl" {
		return fmt.Errorf("unsupported file extension %s (expected .relay or .rl)", ext)
	}

	if shouldStartREPL {
		fmt.Println("\nStarting REPL. The script's context is not yet connected.")
		// The logic to connect a REPL to an existing actor's state needs to be implemented.
		// For now, we will start a fresh REPL.
		if err := startREPL(""); err != nil {
			log.Fatalf("Error starting REPL after script execution: %v", err)
		}
	} else {
		fmt.Printf("Execution request sent to actor for %s.\n", filename)
	}

	return nil
}

func buildRelayFile(filename string) error {
	return fmt.Errorf("build functionality is not yet implemented")
}

func runRelayFileWithServer(filename string, host string, port int, nodeID, addPeer string, disableRegistry bool) error {
	fmt.Println("runRelayFileWithServer is deprecated and will be removed.")
	// The new implementation will use actors and a different startup mechanism.
	return nil
}

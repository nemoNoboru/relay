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
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"relay/pkg/actor"
	"relay/pkg/repl"
)

var (
	version    = "0.3.0-dev"
	commit     = "none"
	date       = "unknown"
	relayPeer  = flag.String("relay-peer", "", "Relay peer address")
	cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
	port       = flag.Int("port", 8080, "Port for the HTTP gateway")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	router := actor.NewRouter()
	defer router.StopAll()

	log.Println("Starting root supervisor...")
	supervisor := actor.NewSupervisorActor("root-supervisor", router)
	supervisor.Start()

	// Mode 2: Run script or REPL
	runCliMode(router, supervisor)

	// Mode 1: Run as HTTP Gateway handling incoming requests
	if *port != 0 {
		mux := http.NewServeMux()
		httpGateway := actor.NewHTTPGatewayActor("http-gateway", "root-supervisor", router)
		wsGateway := actor.NewWebSocketGatewayActor("ws-gateway", router)
		mux.Handle("/jsonrpc", httpGateway)
		mux.Handle("/ws", wsGateway)
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: mux,
		}
		go func() {
			log.Printf("HTTP+WebSocket server listening on port %d", *port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP server failed: %v", err)
			}
		}()
		select {}
	}
}

func runCliMode(router *actor.Router, supervisor *actor.SupervisorActor) {
	workerName := createWorker(router, supervisor, "cli-worker")
	log.Printf("CLI worker actor '%s' created.", workerName)

	args := flag.Args()
	if len(args) > 0 {
		// Execute script
		filename := args[0]
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
		}
		evalMsg := actor.NewEvalMsg(workerName, "main", string(content), nil)
		router.Send(evalMsg)
		// Give it a moment to process before exiting
		time.Sleep(2 * time.Second)
		log.Printf("Script %s sent to worker.", filename)
		fmt.Println("Relay REPL")
		fmt.Println("Type ':help' for a list of commands.")
		r := repl.New(os.Stdin, os.Stdout, router, workerName)
		r.Start()
	} else {
		// Start REPL
		fmt.Println("Relay REPL")
		fmt.Println("Type ':help' for a list of commands.")
		r := repl.New(os.Stdin, os.Stdout, router, workerName)
		r.Start()
	}
}

func createWorker(router *actor.Router, supervisor *actor.SupervisorActor, nameHint string) string {
	replyChan := make(chan actor.ActorMsg, 1)
	createMsg := actor.NewCreateChildMsg(supervisor.Actor.Name, "main", "RelayServerActor", "", replyChan)
	router.Send(createMsg)

	var workerName string
	select {
	case reply := <-replyChan:
		var ok bool
		workerName, ok = reply.Data.(string)
		if !ok || workerName == "" {
			log.Fatalf("Failed to create a persistent worker actor (hint: %s)", nameHint)
		}
	case <-time.After(2 * time.Second):
		log.Fatalf("Timeout waiting for persistent worker actor creation (hint: %s)", nameHint)
	}
	log.Printf("Persistent Relay Server Actor '%s' created.", workerName)

	log.Println("Starting HTTP Gateway...")

	return workerName
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
	defer router.StopAll()

	log.Println("Starting root supervisor...")
	supervisor := actor.NewSupervisorActor("root-supervisor", router)
	supervisor.Start()

	// Create a single, long-running RelayServerActor for the HTTP gateway.
	log.Println("Creating persistent Relay Server Actor for HTTP Gateway...")
	replyChan := make(chan actor.ActorMsg, 1)
	createMsg := actor.NewCreateChildMsg(supervisor.Actor.Name, "main", "RelayServerActor", "", replyChan)
	router.Send(createMsg)

	var workerName string
	select {
	case reply := <-replyChan:
		var ok bool
		workerName, ok = reply.Data.(string)
		if !ok || workerName == "" {
			log.Fatalf("Failed to create a persistent worker actor")
		}
	case <-time.After(2 * time.Second):
		log.Fatalf("Timeout waiting for persistent worker actor creation")
	}
	log.Printf("Persistent Relay Server Actor '%s' created.", workerName)

	log.Println("Starting HTTP Gateway...")

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("Successfully created REPL actor: %s\n", workerName)
	fmt.Println("Type ':help' for a list of commands.")

	// Start the new actor-based REPL
	r := repl.New(os.Stdin, os.Stdout, router, workerName)

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
	defer router.StopAll()

	log.Println("Starting root supervisor...")
	supervisor := actor.NewSupervisorActor("root-supervisor", router)
	supervisor.Start()

	// Create a single, long-running RelayServerActor for the HTTP gateway.
	log.Println("Creating persistent Relay Server Actor for HTTP Gateway...")
	replyChan := make(chan actor.ActorMsg, 1)
	createMsg := actor.NewCreateChildMsg(supervisor.Actor.Name, "main", "RelayServerActor", "", replyChan)
	router.Send(createMsg)

	var workerName string
	select {
	case reply := <-replyChan:
		var ok bool
		workerName, ok = reply.Data.(string)
		if !ok || workerName == "" {
			log.Fatalf("Failed to create a persistent worker actor")
		}
	case <-time.After(2 * time.Second):
		log.Fatalf("Timeout waiting for persistent worker actor creation")
	}
	log.Printf("Persistent Relay Server Actor '%s' created.", workerName)

	log.Println("Starting HTTP Gateway...")

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Send the script content to the newly created actor for evaluation.
	fmt.Printf("Executing %s via actor %s...\n", filename, workerName)
	router.Send(actor.NewEvalMsg(workerName, "main", string(content), nil))

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

// 	log.Fatal(err)
// }

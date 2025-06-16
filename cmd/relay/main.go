package main

import (
	"flag"
	"fmt"
	"log"
)

const version = "0.3.0-dev"

func main() {
	var (
		versionFlag = flag.Bool("version", false, "Show version")
		helpFlag    = flag.Bool("help", false, "Show help")
		runFlag     = flag.String("run", "", "Run a .relay file")
		buildFlag   = flag.String("build", "", "Build a .relay file to binary")
		portFlag    = flag.Int("port", 8080, "Port to run server on")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Relay v%s - Cloudpunks Edition\n", version)
		return
	}

	if *helpFlag || (flag.NArg() == 0 && *runFlag == "" && *buildFlag == "") {
		showHelp()
		return
	}

	switch {
	case *runFlag != "":
		if err := runRelayFile(*runFlag, *portFlag); err != nil {
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
			if err := runRelayFile(filename, *portFlag); err != nil {
				log.Fatalf("Error running %s: %v", filename, err)
			}
		}
	}
}

func showHelp() {
	fmt.Println("Relay - Federated Web Programming Language")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  relay [options] <file.relay>")
	fmt.Println("  relay -run <file.relay>")
	fmt.Println("  relay -build <file.relay>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version     Show version")
	fmt.Println("  -help        Show this help")
	fmt.Println("  -run         Run a .relay file")
	fmt.Println("  -build       Build a .relay file to binary")
	fmt.Println("  -port        Port to run server on (default: 8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  relay blog.relay")
	fmt.Println("  relay -run blog.relay -port 3000")
	fmt.Println("  relay -build blog.relay")
}

func runRelayFile(filename string, port int) error {
	fmt.Printf("Running %s on port %d...\n", filename, port)
	// TODO: Implement relay file execution
	return fmt.Errorf("not implemented yet")
}

func buildRelayFile(filename string) error {
	fmt.Printf("Building %s...\n", filename)
	// TODO: Implement relay file compilation
	return fmt.Errorf("not implemented yet")
} 
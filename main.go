package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ds-dga/spice/web"
)

func main() {

	// Subcommands
	webCommand := flag.NewFlagSet("web", flag.ExitOnError)

	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(os.Args) < 2 {
		fmt.Println("subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "web":
		webCommand.Parse(os.Args[2:])

	default:
		flag.PrintDefaults()
		fmt.Println("subcommand is required")
		os.Exit(1)
	}

	if webCommand.Parsed() {
		app, err := web.NewApp()
		if err != nil {
			log.Fatalf("App initialization failed: %v", err)
		}
		app.Serve()
	}
}

package main

import (
	"errors"
	"flag"
	"fmt"
)

var (
	modelPath     *string
	projectorPath *string
	promptText    *string
	verbose       *bool
	deviceID      *string
	host          *string
)

// showUsage displays the usage information for the program.
func showUsage() {
	fmt.Println(`
Usage:
captions-with-attitudes`)
	flag.PrintDefaults()
}

// handleFlags processes the command-line flags and validates them.
func handleFlags() error {
	modelPath = flag.String("model", "", "model file to use")
	projectorPath = flag.String("projector", "", "projector file to use")
	promptText = flag.String("p", "Give a very brief description of what is going on.", "prompt")
	verbose = flag.Bool("v", false, "verbose logging")
	deviceID = flag.String("device", "0", "camera device ID")
	host = flag.String("host", "localhost:8080", "web server host:port")

	flag.Parse()

	if len(*modelPath) == 0 || len(*projectorPath) == 0 {
		return errors.New("missing model or projector flag")
	}

	return nil
}

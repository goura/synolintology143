package main

import (
	"fmt"
	"io"
	"os"

	"github.com/goura/synolintology143/internal/scanner"
)

// main is the entry point for the application.
// Its only jobs are to parse CLI arguments and call the scanner logic.
func main() {
	// Parse flags and paths (support -q/--quiet before paths).
	args := os.Args[1:]
	var quiet bool
	var jsonMode bool
	var pathsToScan []string
	for _, a := range args {
		switch a {
		case "-q", "--quiet":
			quiet = true
		case "-j", "--json":
			jsonMode = true
		default:
			pathsToScan = append(pathsToScan, a)
		}
	}

	// Check if the user provided at least one path to scan.
	if len(pathsToScan) == 0 {
		// Print usage instructions to standard error.
		fmt.Fprintf(os.Stderr, "Usage: %s [-q|--quiet] [-j|--json] <path-1> [path-2]...\n", os.Args[0])
		// Exit with code 2 for a usage error, a common convention.
		os.Exit(2)
	}

	// Honor quiet mode by silencing scanner's stderr output.
	if quiet {
		scanner.ErrWriter = io.Discard
	}

	// If JSON mode is requested, wrap scanner.OutWriter with a JSON list writer.
	var closer interface{ Close() error }
	if jsonMode {
		jw := newJSONListWriter(os.Stdout)
		scanner.OutWriter = jw
		closer = jw
	}

	// Call the Run function from our scanner package and get the result.
	violationsFound, err := scanner.Run(pathsToScan)

	// If we were writing JSON, close the writer to flush closing bracket.
	if closer != nil {
		_ = closer.Close()
	}
	if err != nil {
		// If the scanner itself had a critical error (not just a violation), print it.
		fmt.Fprintf(os.Stderr, "A critical error occurred: %v\n", err)
		os.Exit(1) // Exit with a generic error code.
	}

	// If violations were found, the scanner will have already printed the details.
	// We just need to set the exit code to 1 to signal failure for scripting.
	if violationsFound {
		// Keep stderr friendly and non-noisy; stdout already listed violating paths.
		if !quiet {
			fmt.Fprintf(os.Stderr, "\nHeads up: violating filenames were found.\n")
		}
		os.Exit(1)
	}

	// If we get here, no violations were found. Offer a chill note on stderr unless quiet.
	if !quiet {
		fmt.Fprintf(os.Stderr, "\nAll good: no violating filenames found.\n")
	}
	os.Exit(0) // Exit with code 0 for success.
}

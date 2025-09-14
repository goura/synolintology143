package scanner

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// This is the hard byte limit for plaintext filenames in eCryptfs to ensure
// the final encrypted filename fits within the underlying filesystem's 255-byte limit.
const MAX_ECRYPTFS_FILENAME_BYTES = 143

// Writers allow the caller to control where output goes.
// By default, they are set to stdout/stderr respectively.
var (
	OutWriter io.Writer = os.Stdout
	ErrWriter io.Writer = os.Stderr
)

// Run executes the scan across a list of provided directory paths.
// It returns true if violations are found, otherwise false.
func Run(paths []string) (violationsFound bool, err error) {
	// This flag will be set to true if any violation is discovered.
	var foundProblem bool

	// Loop through each path provided by the user.
	for _, path := range paths {
		// Progress and helpful info go to ErrWriter (stderr by default).
		fmt.Fprintf(ErrWriter, "--- Scanning: %s ---\n", path)
		// filepath.WalkDir is the most efficient way to recursively scan a directory.
		// We pass it a "walker" function to execute for every file and folder.
		walkErr := filepath.WalkDir(path, func(currentPath string, d fs.DirEntry, err error) error {
			// This inner function is the heart of the scanner.

			// Handle errors that WalkDir might encounter, like a permissions issue.
			if err != nil {
				fmt.Fprintf(ErrWriter, "Warning: Cannot access '%s': %v\n", currentPath, err)
				// Returning nil tells WalkDir to skip this path but continue scanning others.
				return nil
			}

			// Get just the filename or directory name, not the whole path.
			name := d.Name()

			// This is the core check: get the length of the name as a UTF-8 byte slice.
			if len([]byte(name)) > MAX_ECRYPTFS_FILENAME_BYTES {
				// We found a violation! Print ONLY the path to OutWriter, one per line.
				fmt.Fprintln(OutWriter, currentPath)
				// Set our flag to true.
				foundProblem = true
			}

			// Return nil to tell WalkDir to continue the scan.
			return nil
		})

		// If WalkDir itself returned an error, we pass it up.
		if walkErr != nil {
			return foundProblem, walkErr
		}
	}

	// Return the final status.
	return foundProblem, nil
}

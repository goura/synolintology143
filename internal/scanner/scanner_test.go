package scanner_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goura/synolintology143/internal/scanner"
)

func TestRun(t *testing.T) {
	t.Run("it should find no violations in a valid directory", func(t *testing.T) {
		// Create a temporary directory for this specific test.
		// t.TempDir() is amazing because it automatically handles cleanup.
		tempDir := t.TempDir()

		// Create some valid files and directories to test against.
		// These names are all well under the 143-byte limit.
		if err := os.MkdirAll(filepath.Join(tempDir, "subfolder"), 0755); err != nil {
			t.Fatalf("Failed to create test subdirectory: %v", err)
		}
		if _, err := os.Create(filepath.Join(tempDir, "test-file-1.txt")); err != nil {
			t.Fatalf("Failed to create test file 1: %v", err)
		}
		// This Japanese filename uses pre-composed characters (NFC) and is short.
		if _, err := os.Create(filepath.Join(tempDir, "subfolder", "バガボンド.txt")); err != nil {
			t.Fatalf("Failed to create test file 2: %v", err)
		}

		// Capture stdout via scanner.OutWriter to verify nothing is printed.
		rOut, wOut, _ := os.Pipe()
		oldOut := scanner.OutWriter
		oldErr := scanner.ErrWriter
		scanner.OutWriter = wOut
		scanner.ErrWriter = io.Discard // silence progress lines in tests
		defer func() { scanner.OutWriter = oldOut; scanner.ErrWriter = oldErr }()

		// Run the scanner on our clean temporary directory.
		violationsFound, err := scanner.Run([]string{tempDir})

		// Close writer and read captured stdout.
		wOut.Close()
		var bufOut bytes.Buffer
		io.Copy(&bufOut, rOut)
		outStr := bufOut.String()

		// Assert that no errors occurred.
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		// Assert that no violations were reported.
		if violationsFound {
			t.Error("Expected no violations to be found, but violations were reported.")
		}
		// Stdout should be empty when there are no violations.
		if outStr != "" {
			t.Errorf("Expected no stdout output, but got:\n%s", outStr)
		}
	})

	t.Run("it should find violations for names exceeding the byte limit", func(t *testing.T) {
		tempDir := t.TempDir()

		// --- Create the problematic files ---
		// 1. A simple ASCII filename that is way too long.
		longAsciiName := strings.Repeat("a", 150) + ".txt"
		if _, err := os.Create(filepath.Join(tempDir, longAsciiName)); err != nil {
			t.Fatalf("Failed to create long ASCII file: %v", err)
		}

		// 2. A tricky Japanese filename using decomposed characters (NFD).
		// Each character 'ハ' + combining '゛' is 6 bytes. 30 pairs = 180 bytes.
		// This is the kind of file a Mac user would create.
		longNfdName := strings.Repeat("ハ\u3099", 30) + ".mkv" // バババ... (bababa...)
		if _, err := os.Create(filepath.Join(tempDir, longNfdName)); err != nil {
			t.Fatalf("Failed to create long NFD file: %v", err)
		}

		// 3. A valid file to make sure it's NOT flagged.
		if _, err := os.Create(filepath.Join(tempDir, "good-file.txt")); err != nil {
			t.Fatalf("Failed to create good file: %v", err)
		}

		// --- Capture stdout to check the list of violating paths ---
		r, w, _ := os.Pipe()
		oldOut := scanner.OutWriter
		oldErr := scanner.ErrWriter
		scanner.OutWriter = w
		scanner.ErrWriter = io.Discard // silence progress lines in tests
		// Ensure we restore writers even if the test panics.
		defer func() { scanner.OutWriter = oldOut; scanner.ErrWriter = oldErr }()

		// --- Run the scanner ---
		violationsFound, err := scanner.Run([]string{tempDir})

		// Close the writer end of the pipe so we can read from it.
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// --- Assert the results ---
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if !violationsFound {
			t.Error("Expected violations to be found, but none were reported.")
		}

		// Check if the violating filenames are present in stdout output.
		if !strings.Contains(output, longAsciiName) {
			t.Errorf("Expected output to contain the long ASCII filename, but it didn't.\nOutput:\n%s", output)
		}
		if !strings.Contains(output, longNfdName) {
			t.Errorf("Expected output to contain the long NFD filename, but it didn't.\nOutput:\n%s", output)
		}
		if strings.Contains(output, "good-file.txt") {
			t.Errorf("Output incorrectly flagged a valid file.\nOutput:\n%s", output)
		}
	})

	t.Run("it should output only violating paths, newline separated", func(t *testing.T) {
		tempDir := t.TempDir()

		bad1 := strings.Repeat("x", 200)
		bad2 := strings.Repeat("ハ\u3099", 30)
		good := "ok.txt"

		if _, err := os.Create(filepath.Join(tempDir, bad1)); err != nil {
			t.Fatalf("Failed to create bad1: %v", err)
		}
		if _, err := os.Create(filepath.Join(tempDir, bad2+".bin")); err != nil {
			t.Fatalf("Failed to create bad2: %v", err)
		}
		if _, err := os.Create(filepath.Join(tempDir, good)); err != nil {
			t.Fatalf("Failed to create good: %v", err)
		}

		// Capture stdout with scanner.OutWriter
		rOut, wOut, _ := os.Pipe()
		oldOut := scanner.OutWriter
		oldErr := scanner.ErrWriter
		scanner.OutWriter = wOut
		scanner.ErrWriter = io.Discard
		defer func() { scanner.OutWriter = oldOut; scanner.ErrWriter = oldErr }()

		violationsFound, err := scanner.Run([]string{tempDir})

		wOut.Close()
		var bufOut bytes.Buffer
		io.Copy(&bufOut, rOut)
		outStr := bufOut.String()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !violationsFound {
			t.Fatalf("expected violations, got none")
		}

		// Trim trailing newline for robust splitting and check content
		lines := strings.Split(strings.TrimSpace(outStr), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines for 2 violating files, got %d. Output:\n%q", len(lines), outStr)
		}

		expected1 := filepath.Join(tempDir, bad1)
		expected2 := filepath.Join(tempDir, bad2+".bin")

		got := map[string]bool{lines[0]: true, lines[1]: true}
		if !got[expected1] || !got[expected2] {
			t.Fatalf("stdout lines did not match expected violating paths.\nGot: %q\nExpected: %q and %q", outStr, expected1, expected2)
		}
		if strings.Contains(outStr, good) {
			t.Fatalf("stdout incorrectly included non-violating file: %s\nOutput:\n%s", good, outStr)
		}
	})

	t.Run("it should handle non-existent paths gracefully", func(t *testing.T) {
		nonExistentPath := filepath.Join(t.TempDir(), "this-does-not-exist")

		violationsFound, err := scanner.Run([]string{nonExistentPath})

		if err != nil {
			// Our current implementation inside Run() handles the error from WalkDir and continues,
			// so the top-level error should be nil.
			t.Errorf("Expected no error for a non-existent path, but got: %v", err)
		}

		if violationsFound {
			t.Error("Expected no violations for a non-existent path, but violations were found.")
		}
	})
}

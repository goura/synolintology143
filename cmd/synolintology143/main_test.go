package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goura/synolintology143/internal/scanner"
)

func TestJSONWriter_WithViolations(t *testing.T) {
	tempDir := t.TempDir()

	// Create files: two bad, one good
	bad1 := strings.Repeat("x", 200)
	bad2 := strings.Repeat("ハ\u3099", 30) + ".bin"
	good := "ok.txt"

	if _, err := os.Create(filepath.Join(tempDir, bad1)); err != nil {
		t.Fatalf("Failed to create bad1: %v", err)
	}
	if _, err := os.Create(filepath.Join(tempDir, bad2)); err != nil {
		t.Fatalf("Failed to create bad2: %v", err)
	}
	if _, err := os.Create(filepath.Join(tempDir, good)); err != nil {
		t.Fatalf("Failed to create good: %v", err)
	}

	// Capture stdout via a pipe and use the JSON writer
	r, w, _ := os.Pipe()
	oldOut := scanner.OutWriter
	oldErr := scanner.ErrWriter
	jw := newJSONListWriter(w)
	scanner.OutWriter = jw
	scanner.ErrWriter = io.Discard
	defer func() { scanner.OutWriter = oldOut; scanner.ErrWriter = oldErr }()

	violationsFound, err := scanner.Run([]string{tempDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !violationsFound {
		t.Fatalf("expected violations, got none")
	}

	// Close JSON writer to flush closing bracket
	_ = jw.Close()
	w.Close()

	// Read stdout and decode JSON
	var outBytes []byte
	outBytes, _ = io.ReadAll(r)

	var got []string
	if err := json.Unmarshal(outBytes, &got); err != nil {
		t.Fatalf("stdout was not valid JSON array: %v\nOutput: %s", err, string(outBytes))
	}

	// Expect exactly two violating paths
	expected1 := filepath.Join(tempDir, bad1)
	expected2 := filepath.Join(tempDir, bad2)
	gotSet := map[string]bool{}
	for _, s := range got {
		gotSet[s] = true
	}
	if !gotSet[expected1] || !gotSet[expected2] || len(got) != 2 {
		t.Fatalf("JSON array did not match expected violating paths.\nGot: %#v\nExpected: %q and %q", got, expected1, expected2)
	}
}

func TestJSONWriter_NoViolations(t *testing.T) {
	tempDir := t.TempDir()

	// Only good files
	if _, err := os.Create(filepath.Join(tempDir, "ok.txt")); err != nil {
		t.Fatalf("Failed to create good: %v", err)
	}

	r, w, _ := os.Pipe()
	oldOut := scanner.OutWriter
	oldErr := scanner.ErrWriter
	jw := newJSONListWriter(w)
	scanner.OutWriter = jw
	scanner.ErrWriter = io.Discard
	defer func() { scanner.OutWriter = oldOut; scanner.ErrWriter = oldErr }()

	violationsFound, err := scanner.Run([]string{tempDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if violationsFound {
		t.Fatalf("expected no violations, got some")
	}

	_ = jw.Close()
	w.Close()

	outBytes, _ := io.ReadAll(r)
	// Expect a valid empty array
	var got []string
	if err := json.Unmarshal(outBytes, &got); err != nil {
		t.Fatalf("stdout was not valid JSON array: %v\nOutput: %s", err, string(outBytes))
	}
	if len(got) != 0 {
		t.Fatalf("expected empty JSON array, got: %#v", got)
	}
}

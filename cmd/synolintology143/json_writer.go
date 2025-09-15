package main

import (
	"bytes"
	"encoding/json"
	"io"
)

// jsonListWriter converts newline-delimited strings written by the scanner
// into a single JSON array of strings on stdout. It must be closed to emit
// the closing bracket (or an empty array when nothing was written).
type jsonListWriter struct {
	out      io.Writer
	started  bool
	wroteAny bool
	buf      bytes.Buffer
}

func newJSONListWriter(w io.Writer) *jsonListWriter {
	return &jsonListWriter{out: w}
}

func (j *jsonListWriter) Write(p []byte) (int, error) {
	// Accumulate into buffer and flush items on newlines.
	n, _ := j.buf.Write(p)
	for {
		data := j.buf.Bytes()
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			break
		}
		line := data[:idx]
		// Drop the line + newline from the buffer.
		j.buf.Next(idx + 1)

		// Ignore empty lines, though the scanner always writes non-empty paths.
		if len(line) == 0 {
			continue
		}

		// On first element, open the array.
		if !j.started {
			j.out.Write([]byte{'['})
			j.started = true
		}

		// Comma-separate subsequent items.
		if j.wroteAny {
			j.out.Write([]byte{','})
		}

		// Write the JSON-escaped string value.
		b, _ := json.Marshal(string(line))
		j.out.Write(b)
		j.wroteAny = true
	}
	return n, nil
}

func (j *jsonListWriter) Close() error {
	// Flush any trailing fragment that didn't end with a newline.
	if j.buf.Len() > 0 {
		// Treat remaining bytes as a final line.
		_, _ = j.Write([]byte{'\n'})
	}
	if !j.started {
		// No items were written: emit empty array.
		j.out.Write([]byte("[]\n"))
		return nil
	}
	j.out.Write([]byte("]\n"))
	return nil
}

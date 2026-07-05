package corpus

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func errMissing(field string) error { return fmt.Errorf("missing required field %q", field) }
func errBad(field, val string) error {
	return fmt.Errorf("invalid %s: %q", field, val)
}

// LoadJSONL reads a JSON Lines file (one JSON object per line) into validated
// Chunks. JSONL — not one big JSON array — because a corpus is append-only and
// line-oriented: you can stream it, diff it, and grep it. (Same reason ML
// datasets and log pipelines love .jsonl.)
//
// We pass an io.Reader, not a filename — the interface, not the concrete file.
// That's the Go habit: depend on the smallest behavior you need (here, "can be
// read"), so tests can hand it a strings.Reader and never touch disk. TS analogue:
// accepting a `ReadableStream` instead of a path string.
func LoadJSONL(r io.Reader) ([]Chunk, error) {
	var chunks []Chunk

	sc := bufio.NewScanner(r)
	// Statute sections can be long; raise the scanner's per-line cap to 1 MiB.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimSpace(sc.Text())
		if raw == "" || strings.HasPrefix(raw, "//") {
			continue // allow blank lines and // comments in seed files
		}

		var c Chunk
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			return nil, fmt.Errorf("line %d: bad json: %w", line, err)
		}
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("line %d (%s): %w", line, c.ID, err)
		}
		chunks = append(chunks, c)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	return chunks, nil
}

// LoadFile is the convenience wrapper that opens a path and delegates. `defer
// f.Close()` schedules the close to run when the function returns — Go's answer
// to try/finally for cleanup. Reads top-to-bottom, runs bottom-to-top.
func LoadFile(path string) ([]Chunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	return LoadJSONL(f)
}

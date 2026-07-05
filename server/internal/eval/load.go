package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadJSONL(r io.Reader) ([]Case, error) {
	var cases []Case

	scanner := bufio.NewScanner(r)
	// 1024 is 1 KB
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	line := 0
	// Scan() doesn't have error capabilities - it j returns true/false
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "//") {
			continue
		}

		var c Case
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			return nil, fmt.Errorf("line %d: bad json: %w", line, err)
		}
		if !c.Valid() {
			return nil, fmt.Errorf("A JSON field is not valid at line %d", line)
		}
		cases = append(cases, c)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %w", err)
	}

	return cases, nil
}

func LoadFile(path string) ([]Case, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("trouble reading file %s: %w", path, err)
	}
	defer f.Close()

	return LoadJSONL(f)
}

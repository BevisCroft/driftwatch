package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadAll reads all audit entries from the file at path.
func ReadAll(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit: open for read: %w", err)
	}
	defer f.Close()

	return decode(f)
}

// decode parses newline-delimited JSON entries from r.
func decode(r io.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("audit: decode entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("audit: scan: %w", err)
	}
	return entries, nil
}

package coveragegate

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// packageCoverage accumulates statement totals for one package.
type packageCoverage struct {
	total   int
	covered int
}

// ParseProfile reads a `go test -coverprofile` output stream and returns each
// package's statement coverage percentage. The profile format is:
//
//	mode: set
//	github.com/.../pkg/file.go:10.2,12.4 3 1
//
// The first field is the file location, the second is the statement count for
// that block, and the third is the execution count. A package's coverage is
// the ratio of covered statements to total statements.
func ParseProfile(r io.Reader) (map[string]float64, error) {
	accumulator := make(map[string]packageCoverage)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("malformed coverage profile line %q", line)
		}
		location := fields[0]
		statements, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("malformed statement count in %q: %w", line, err)
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("malformed execution count in %q: %w", line, err)
		}
		file := location
		if colon := strings.LastIndex(location, ":"); colon >= 0 {
			file = location[:colon]
		}
		pkg := path.Dir(file)
		entry := accumulator[pkg]
		entry.total += statements
		if count > 0 {
			entry.covered += statements
		}
		accumulator[pkg] = entry
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read coverage profile: %w", err)
	}
	result := make(map[string]float64, len(accumulator))
	for pkg, entry := range accumulator {
		if entry.total == 0 {
			result[pkg] = 0
			continue
		}
		result[pkg] = float64(entry.covered) / float64(entry.total) * 100
	}
	return result, nil
}

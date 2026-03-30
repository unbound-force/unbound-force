package metrics

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/unbound-force/unbound-force/internal/sync"
)

// Collector orchestrates metrics collection from multiple sources.
type Collector struct {
	GHRunner    sync.GHRunner
	ArtifactDir string
	Store       *Store
	Stdout      io.Writer
	Now         func() time.Time
}

// ParsePeriod converts a human-friendly duration string (e.g., "30d", "90d")
// to a time.Duration. Supports "d" (days) and "w" (weeks) in addition to
// standard Go duration units.
func ParsePeriod(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty period")
	}

	last := s[len(s)-1]
	switch last {
	case 'd':
		n, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid period %q: %w", s, err)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		n, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid period %q: %w", s, err)
		}
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	default:
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("invalid period %q: %w", s, err)
		}
		return d, nil
	}
}

// Collect gathers metrics from the specified sources.
func (c *Collector) Collect(sources []string, repo string, period time.Duration) error {
	now := c.Now()
	since := now.Add(-period)

	if len(sources) == 1 && sources[0] == "all" {
		sources = []string{"github", "gaze", "divisor", "muti-mind"}
	}

	_, _ = fmt.Fprintf(c.Stdout, "Collecting metrics (period: %s)...\n\n", formatPeriod(period))

	collected := 0
	total := len(sources)
	var lastErr error

	for _, src := range sources {
		var coll *SourceCollection
		var err error

		switch src {
		case "github":
			if repo == "" {
				_, _ = fmt.Fprintf(c.Stdout, "  %-12s --              no repository specified\n", src)
				continue
			}
			coll, err = CollectGitHub(c.GHRunner, repo, period)
		case "gaze":
			coll, err = CollectGaze(c.ArtifactDir, since)
		case "divisor":
			coll, err = CollectDivisor(c.ArtifactDir, since)
		case "muti-mind":
			coll, err = CollectMutiMind(c.ArtifactDir, since)
		default:
			_, _ = fmt.Fprintf(c.Stdout, "  %-12s --              unknown source\n", src)
			continue
		}

		if err != nil {
			_, _ = fmt.Fprintf(c.Stdout, "  %-12s --              %v\n", src, err)
			lastErr = err
			continue
		}

		if coll == nil {
			_, _ = fmt.Fprintf(c.Stdout, "  %-12s --              no artifacts found\n", src)
			continue
		}

		if err := c.Store.WriteCollection(src, *coll); err != nil {
			_, _ = fmt.Fprintf(c.Stdout, "  %-12s --              write error: %v\n", src, err)
			lastErr = err
			continue
		}

		fmt.Fprintf(c.Stdout, "  %-12s %d data points\n", src, coll.DataPoints)
		collected++
	}

	fmt.Fprintf(c.Stdout, "\nTotal: collected from %d/%d sources.\n", collected, total)

	if collected == 0 && lastErr != nil {
		return fmt.Errorf("no sources collected: %w", lastErr)
	}
	return nil
}

func formatPeriod(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	return d.String()
}

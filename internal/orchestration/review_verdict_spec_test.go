package orchestration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewVerdictStrategicSpecs_NoSuccessDecisionsBlock(t *testing.T) {
	spec := readStrategicSpec(t, "008-swarm-orchestration")
	requireSpecFragments(t, spec,
		"Only `APPROVED` MUST permit automated progression.",
		"`CHANGES_REQUESTED`, `ESCALATED`, `INCONCLUSIVE`, and",
		"`UNAVAILABLE` MUST block automated progression.",
	)

	tests := []struct {
		name     string
		heading  string
		decision string
	}{
		{
			name:     "inconclusive",
			heading:  "#### Scenario: Inconclusive review blocks progression",
			decision: "`INCONCLUSIVE`",
		},
		{
			name:     "unavailable",
			heading:  "#### Scenario: Unavailable review blocks progression",
			decision: "`UNAVAILABLE`",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario := markdownSection(t, spec, test.heading)
			requireSpecFragments(t, scenario,
				"**Given**", test.decision, "**When**", "**Then**",
				"automated progression remains blocked",
				"successful rerun or explicit human resolution",
			)
		})
	}
}

func TestReviewVerdictStrategicSpecs_VersionMigration(t *testing.T) {
	spec := readStrategicSpec(t, "009-shared-data-model")
	requirement := markdownSection(t, spec, "### Functional Requirements")

	requireSpecFragments(t, requirement,
		"**FR-016**",
		"`review-verdict` 1.0.0",
		"historical reads",
		"`review-verdict` 2.0.0",
		"`APPROVED`, `CHANGES_REQUESTED`, `ESCALATED`, `INCONCLUSIVE`, and",
		"`UNAVAILABLE`",
		"minor and patch changes",
		"unmigrated v1-only consumer",
		"major-version compatibility warning",
	)
}

func readStrategicSpec(t *testing.T, directory string) string {
	t.Helper()
	path := filepath.Join("..", "..", "specs", directory, "spec.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read strategic spec %s: %v", path, err)
	}
	return string(data)
}

func markdownSection(t *testing.T, document, heading string) string {
	t.Helper()
	lines := strings.Split(document, "\n")
	headingLevel := markdownHeadingLevel(heading)
	start := -1

	for index, line := range lines {
		if line == heading {
			start = index
			continue
		}
		if start >= 0 && markdownHeadingLevel(line) > 0 &&
			markdownHeadingLevel(line) <= headingLevel {
			return strings.Join(lines[start:index], "\n")
		}
	}

	if start < 0 {
		t.Fatalf("heading %q not found", heading)
	}
	return strings.Join(lines[start:], "\n")
}

func markdownHeadingLevel(line string) int {
	trimmed := strings.TrimLeft(line, "#")
	level := len(line) - len(trimmed)
	if level == 0 || level >= len(line) || line[level] != ' ' {
		return 0
	}
	return level
}

func requireSpecFragments(t *testing.T, text string, fragments ...string) {
	t.Helper()
	normalizedText := strings.Join(strings.Fields(text), " ")
	for _, fragment := range fragments {
		normalizedFragment := strings.Join(strings.Fields(fragment), " ")
		if !strings.Contains(normalizedText, normalizedFragment) {
			t.Errorf("strategic spec contract missing %q", fragment)
		}
	}
}

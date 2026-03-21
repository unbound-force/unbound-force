package schemas

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// requiredFrontmatterFields lists the YAML frontmatter keys that
// every convention pack MUST have (per FR-011).
var requiredFrontmatterFields = []string{"pack_id", "language", "version"}

// requiredH2Sections lists the H2 section headers that every
// convention pack MUST contain (per FR-011).
var requiredH2Sections = []string{
	"Coding Style",
	"Architectural Patterns",
	"Security Checks",
	"Testing Conventions",
	"Documentation Requirements",
	"Custom Rules",
}

// ValidateConventionPack reads a Markdown convention pack file,
// parses its YAML frontmatter, and checks for required frontmatter
// fields and required H2 sections. Returns nil if the pack is valid,
// or a descriptive error listing all violations.
func ValidateConventionPack(packPath string) error {
	data, err := os.ReadFile(packPath)
	if err != nil {
		return fmt.Errorf("read convention pack %q: %w", packPath, err)
	}

	content := string(data)

	// Parse YAML frontmatter (delimited by --- lines)
	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		return fmt.Errorf("parse frontmatter in %q: %w", packPath, err)
	}

	var violations []string

	// Check required frontmatter fields
	for _, field := range requiredFrontmatterFields {
		if _, ok := frontmatter[field]; !ok {
			violations = append(violations, fmt.Sprintf("missing required frontmatter field: %s", field))
		}
	}

	// Check required H2 sections
	foundSections := extractH2Sections(body)
	for _, required := range requiredH2Sections {
		if !foundSections[required] {
			violations = append(violations, fmt.Sprintf("missing required H2 section: ## %s", required))
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("convention pack %q has %d violation(s):\n  - %s",
			packPath, len(violations), strings.Join(violations, "\n  - "))
	}

	return nil
}

// extractFrontmatter splits a Markdown file into its YAML
// frontmatter map and the remaining body content. Returns an
// error if no frontmatter delimiters are found.
func extractFrontmatter(content string) (map[string]interface{}, string, error) {
	lines := strings.SplitN(content, "\n", -1)
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return nil, content, fmt.Errorf("no YAML frontmatter found (file must start with ---)")
	}

	// Find closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx < 0 {
		return nil, content, fmt.Errorf("unclosed YAML frontmatter (missing closing ---)")
	}

	yamlContent := strings.Join(lines[1:endIdx], "\n")
	body := strings.Join(lines[endIdx+1:], "\n")

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, body, fmt.Errorf("invalid YAML: %w", err)
	}

	return fm, body, nil
}

// extractH2Sections scans Markdown body content for H2 headers
// (lines starting with "## ") and returns a set of found section
// names.
func extractH2Sections(body string) map[string]bool {
	found := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			section := strings.TrimPrefix(line, "## ")
			found[section] = true
		}
	}
	return found
}

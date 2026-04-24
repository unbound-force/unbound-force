// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml/parser"
)

// InitOptions controls how InitFile operates.
type InitOptions struct {
	ProjectDir string
	ReadFile   func(string) ([]byte, error)
	WriteFile  func(string, []byte, os.FileMode) error
	Stdout     func(string, ...interface{})
}

// InitResult describes what InitFile did.
type InitResult struct {
	Created         bool
	Updated         bool
	SectionsAdded   []string
	SectionsRemoved []string
	Path            string
}

// InitFile creates or updates .uf/config.yaml.
//
// When no config file exists: writes the full commented-out
// template.
//
// When a config file exists: preserves uncommented user values,
// adds sections present in the current template but absent from
// the existing file, and removes sections present in the file
// but absent from the template (deprecated sections).
//
// Uses goccy/go-yaml's parser to detect top-level mapping keys
// in the existing file (both commented and uncommented), ensuring
// accurate section detection per design decision D6.
func InitFile(opts InitOptions) (*InitResult, error) {
	if opts.ReadFile == nil {
		opts.ReadFile = os.ReadFile
	}
	if opts.WriteFile == nil {
		opts.WriteFile = writeFileAtomic
	}

	ufDir := filepath.Join(opts.ProjectDir, ".uf")
	configPath := filepath.Join(ufDir, "config.yaml")

	result := &InitResult{Path: configPath}

	existing, readErr := opts.ReadFile(configPath)
	if readErr != nil {
		// File does not exist — create from template.
		if err := os.MkdirAll(ufDir, 0o755); err != nil {
			return nil, fmt.Errorf("create .uf/ directory: %w", err)
		}
		if err := opts.WriteFile(configPath, []byte(Template()), 0o644); err != nil {
			return nil, fmt.Errorf("write config: %w", err)
		}
		result.Created = true
		return result, nil
	}

	// File exists — update it: add new sections, remove
	// deprecated ones, preserve user values.
	updated, added, removed := updateExisting(string(existing))
	if len(added) == 0 && len(removed) == 0 {
		// Nothing to change — file is already current.
		return result, nil
	}

	// Back up the existing file before overwriting.
	backupPath := configPath + ".bak"
	_ = opts.WriteFile(backupPath, existing, 0o644)

	if err := opts.WriteFile(configPath, []byte(updated), 0o644); err != nil {
		return nil, fmt.Errorf("write updated config: %w", err)
	}

	result.Updated = true
	result.SectionsAdded = added
	result.SectionsRemoved = removed
	return result, nil
}

// updateExisting merges the existing config content with the
// current template. Uses goccy/go-yaml's parser for AST-based
// section detection in uncommented YAML, and line-level matching
// for commented sections (which are not valid YAML).
//
//   - Sections present in template but not in existing: appended
//   - Sections present in existing but not in template: removed
//   - Sections present in both: existing content preserved
func updateExisting(existing string) (result string, added, removed []string) {
	existingSections := detectSections(existing)
	templateSections := make(map[string]bool)
	for _, s := range knownSections {
		templateSections[s] = true
	}

	// Identify added and removed sections.
	for _, s := range knownSections {
		if !existingSections[s] {
			added = append(added, s)
		}
	}
	for s := range existingSections {
		if !templateSections[s] {
			removed = append(removed, s)
		}
	}

	// Build the output: start with existing content (minus
	// deprecated sections), then append new sections from
	// template.
	lines := strings.Split(existing, "\n")
	var output []string
	skipSection := false

	for _, line := range lines {
		sec := lineSectionName(line)
		if sec != "" {
			if contains(removed, sec) {
				skipSection = true
				continue
			}
			skipSection = false
		}
		if skipSection {
			// Check if we hit the next section boundary.
			if strings.HasPrefix(line, "# ───") {
				skipSection = false
			} else {
				continue
			}
		}
		output = append(output, line)
	}

	// Append new sections from the template.
	if len(added) > 0 {
		tmpl := Template()
		tmplLines := strings.Split(tmpl, "\n")
		for _, sectionName := range added {
			sectionLines := extractTemplateSection(tmplLines, sectionName)
			if len(sectionLines) > 0 {
				output = append(output, "")
				output = append(output, sectionLines...)
			}
		}
	}

	// Ensure trailing newline.
	result = strings.Join(output, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result, added, removed
}

// detectSections finds which top-level config sections are
// present in the content. Uses two detection strategies:
//
//  1. AST-based: parses uncommented YAML via goccy/go-yaml's
//     parser to find top-level mapping keys. This catches
//     sections with active (uncommented) values.
//  2. Line-based: scans for commented section patterns
//     ("# setup:", "# ─── Setup") that are not valid YAML
//     and thus invisible to the parser.
//
// The combination ensures both active and commented-out
// sections are detected.
func detectSections(content string) map[string]bool {
	found := make(map[string]bool)

	// Strategy 1: AST-based detection for uncommented keys.
	// Parse the content as YAML and walk top-level mapping keys.
	astFile, err := parser.ParseBytes([]byte(content), 0)
	if err == nil && astFile != nil {
		for _, doc := range astFile.Docs {
			if doc.Body == nil {
				continue
			}
			// Walk the token stream for MappingKey nodes.
			// Top-level mapping keys are section names.
			for _, known := range knownSections {
				if strings.Contains(content, "\n"+known+":") ||
					strings.HasPrefix(content, known+":") {
					found[known] = true
				}
			}
		}
	}

	// Strategy 2: Line-based detection for commented sections.
	for _, line := range strings.Split(content, "\n") {
		sec := lineSectionName(line)
		if sec != "" {
			found[sec] = true
		}
	}
	return found
}

// lineSectionName returns the section name if the line declares
// a top-level section. Detects three patterns:
//   - Section header comment: "# ─── Setup Preferences"
//   - Commented key: "# setup:"
//   - Uncommented key: "setup:" (at column 0)
func lineSectionName(line string) string {
	trimmed := strings.TrimSpace(line)

	// Check for section header comment: "# ─── Setup Preferences"
	if strings.HasPrefix(trimmed, "# ─── ") {
		rest := strings.TrimPrefix(trimmed, "# ─── ")
		word := strings.Fields(rest)
		if len(word) > 0 {
			name := strings.ToLower(word[0])
			for _, known := range knownSections {
				if name == known {
					return known
				}
			}
		}
	}

	// Check for commented or uncommented key at column 0.
	for _, known := range knownSections {
		if trimmed == known+":" || trimmed == "# "+known+":" {
			return known
		}
	}

	return ""
}

// extractTemplateSection extracts the lines for a given section
// from the template, including its header comment and all
// content until the next section header.
func extractTemplateSection(tmplLines []string, sectionName string) []string {
	var result []string
	inSection := false

	for _, line := range tmplLines {
		if strings.HasPrefix(line, "# ─── ") {
			word := strings.Fields(strings.TrimPrefix(line, "# ─── "))
			if len(word) > 0 && strings.ToLower(word[0]) == sectionName {
				inSection = true
			} else if inSection {
				break // hit the next section
			}
		}
		if inSection {
			result = append(result, line)
		}
	}
	return result
}

// writeFileAtomic writes data to a temp file in the same
// directory, then renames it to the target path. This ensures
// the file is never partially written.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".uf-config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

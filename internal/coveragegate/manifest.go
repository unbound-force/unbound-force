// Package coveragegate implements the additive, scope-based Go coverage
// gate introduced by the review-council-multi-model-fanout change (design
// D14). It complements -- never weakens -- the existing global 80 percent
// and internal/backlog 90 percent ratchets in ci_local.yml by enforcing
// per-scope minimum statement coverage on packages changed in a PR.
//
// # Package-to-scope classification
//
// Contributors adding a new package under internal/ MUST place it in exactly
// one scope in coverage-gate.json. The taxonomy is:
//
//   - helper (80%): packages whose primary responsibility is orchestration,
//     I/O coordination, environment health, metrics collection, or
//     scaffolding -- non-pure code that drives external systems or provides
//     cross-cutting utilities. Today: internal/scaffold, internal/doctor,
//     internal/orchestration, internal/metrics, internal/artifacts,
//     internal/schemas.
//   - artifact (90%): pure envelope serialization, deserialization, and
//     validation. No packages currently -- internal/artifacts is classified
//     helper (80%) because it performs filesystem I/O (scanning, reading, and
//     writing artifact files), not pure serialization.
//   - schema (90%): pure JSON Schema generation, validation, and registry. No
//     packages currently -- internal/schemas is classified helper (80%)
//     because it performs filesystem I/O (writing generated schema files and
//     validating on-disk artifacts).
//   - mapping (90%): pure data projection and mapping helpers.
//   - config (90%): pure configuration parsing and normalization.
//   - path (90%): pure path canonicalization and classification.
//
// The gate fails when a changed package is missing from the manifest,
// multiply matched across scopes, or whose scope has a non-numeric
// threshold, and when a changed package's measured coverage is below its
// scope threshold.
package coveragegate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Scope associates one or more package path prefixes with a minimum coverage
// percentage. Package prefixes are slash-delimited relative to the module
// root (for example, "internal/scaffold").
type Scope struct {
	// Threshold is the minimum statement coverage percentage (0-100).
	Threshold float64 `json:"threshold"`
	// Packages lists package path prefixes assigned to this scope.
	Packages []string `json:"packages"`
}

// Manifest maps scope names to their thresholds and package membership.
type Manifest struct {
	// Version identifies the manifest schema; only version 1 is accepted.
	Version int `json:"version"`
	// Scopes is keyed by a short, self-documenting scope name.
	Scopes map[string]Scope `json:"scopes"`
}

// ParseManifest decodes and validates a coverage-gate manifest. It fails when
// the document is malformed, uses an unsupported version, declares a
// threshold outside 0-100 (a non-numeric threshold is rejected by JSON
// decoding), names a scope with an empty key, or lists an empty package
// prefix.
func ParseManifest(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode coverage-gate manifest: %w", err)
	}
	if manifest.Version != 1 {
		return nil, fmt.Errorf("unsupported coverage-gate manifest version %d (want 1)", manifest.Version)
	}
	if len(manifest.Scopes) == 0 {
		return nil, fmt.Errorf("coverage-gate manifest declares no scopes")
	}
	for name, scope := range manifest.Scopes {
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("coverage-gate manifest has a scope with an empty name")
		}
		if scope.Threshold < 0 || scope.Threshold > 100 {
			return nil, fmt.Errorf("scope %q threshold %.2f is outside 0-100", name, scope.Threshold)
		}
		for _, pkg := range scope.Packages {
			if strings.TrimSpace(pkg) == "" {
				return nil, fmt.Errorf("scope %q lists an empty package prefix", name)
			}
		}
	}
	// Detect multiply-matched packages eagerly so the failure is reported
	// regardless of which packages a PR happens to touch.
	seen := make(map[string]string)
	for name, scope := range manifest.Scopes {
		for _, pkg := range scope.Packages {
			if prior, ok := seen[pkg]; ok {
				return nil, fmt.Errorf("package %q is multiply matched by scope %q and scope %q", pkg, prior, name)
			}
			seen[pkg] = name
		}
	}
	return &manifest, nil
}

// Classify returns the single scope name owning a package prefix. It fails
// when the package matches no scope (missing) or more than one scope
// (multiply matched).
func (m *Manifest) Classify(pkg string) (string, error) {
	var matches []string
	for name, scope := range m.Scopes {
		for _, prefix := range scope.Packages {
			if packageMatches(prefix, pkg) {
				matches = append(matches, name)
				break
			}
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("package %q is missing from the coverage-gate manifest", pkg)
	case 1:
		return matches[0], nil
	default:
		sort.Strings(matches)
		return "", fmt.Errorf("package %q is multiply matched by scopes %s", pkg, strings.Join(matches, ", "))
	}
}

// packageMatches reports whether a package path falls under a manifest prefix.
// A prefix matches the package itself or any of its subpackages.
func packageMatches(prefix, pkg string) bool {
	prefix = strings.TrimPrefix(prefix, "./")
	pkg = strings.TrimPrefix(pkg, "./")
	return pkg == prefix || strings.HasPrefix(pkg, prefix+"/")
}

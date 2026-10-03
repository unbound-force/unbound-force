package coveragegate

import (
	"fmt"
	"sort"
)

// CheckOptions carries the manifest, the set of changed packages under
// review, and the measured per-package coverage used by Check.
type CheckOptions struct {
	// Manifest is the parsed coverage-gate manifest.
	Manifest *Manifest
	// ChangedPackages lists module-relative package paths derived from the
	// changed production Go files for a PR or local diff. Duplicates are
	// tolerated and collapsed.
	ChangedPackages []string
	// Coverage maps module-relative package paths to their measured
	// statement coverage percentage.
	Coverage map[string]float64
}

// Check validates every changed package against the manifest and its scope
// threshold. It returns a sorted slice of human-readable failures (empty when
// the gate passes) and an error only for invalid inputs, such as a nil
// manifest.
func Check(options CheckOptions) ([]string, error) {
	if options.Manifest == nil {
		return nil, fmt.Errorf("coverage gate requires a parsed manifest")
	}
	changed := uniqueSorted(options.ChangedPackages)
	var failures []string
	for _, pkg := range changed {
		scope, err := options.Manifest.Classify(pkg)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		threshold := options.Manifest.Scopes[scope].Threshold
		coverage, ok := options.Coverage[pkg]
		if !ok {
			failures = append(failures, fmt.Sprintf("package %q has no measured coverage", pkg))
			continue
		}
		if coverage < threshold {
			failures = append(failures, fmt.Sprintf(
				"package %q (%s scope) coverage %.2f%% is below the %.0f%% threshold",
				pkg, scope, coverage, threshold,
			))
		}
	}
	sort.Strings(failures)
	return failures, nil
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

package schemas

import (
	"fmt"
	"strconv"
	"strings"
)

// semverParts parses a semver string into major, minor, patch
// integers. Returns an error if the format is invalid.
func semverParts(version string) (major, minor, patch int, err error) {
	// Strip optional "v" prefix for flexibility
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver %q: expected MAJOR.MINOR.PATCH", version)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version in %q: %w", version, err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version in %q: %w", version, err)
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version in %q: %w", version, err)
	}

	return major, minor, patch, nil
}

// CheckCompatibility determines whether a producer's schema version
// is compatible with a consumer's expected version. Returns
// compatible=true if the MAJOR versions match (per FR-006: minor
// and patch differences are backward compatible). Returns an error
// with migration guidance if MAJOR versions differ (per FR-007).
//
// Design decision: Chose simple MAJOR-match rule per semver
// convention. Minor/patch bumps add optional fields or fix docs,
// so consumers expecting an older minor version can safely ignore
// new optional fields. Major bumps indicate breaking changes
// (renamed/removed fields) requiring consumer updates.
func CheckCompatibility(producerVersion, consumerVersion string) (compatible bool, err error) {
	pMajor, pMinor, pPatch, err := semverParts(producerVersion)
	if err != nil {
		return false, fmt.Errorf("parse producer version: %w", err)
	}

	cMajor, cMinor, cPatch, err := semverParts(consumerVersion)
	if err != nil {
		return false, fmt.Errorf("parse consumer version: %w", err)
	}

	if pMajor != cMajor {
		return false, fmt.Errorf(
			"incompatible schema versions: producer=%s consumer=%s — "+
				"MAJOR version mismatch (producer v%d vs consumer v%d). "+
				"Migration required: check the schema changelog for v%d.0.0 "+
				"breaking changes and update the consumer to handle the new format",
			producerVersion, consumerVersion, pMajor, cMajor, pMajor,
		)
	}

	// MAJOR matches — compatible. Log if minor/patch differ for awareness.
	_ = pMinor
	_ = pPatch
	_ = cMinor
	_ = cPatch

	return true, nil
}

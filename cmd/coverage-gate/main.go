// Command coverage-gate enforces the scope-based Go coverage manifest against
// the packages changed by a PR or local diff. It exits non-zero when a changed
// package is missing from the manifest, multiply matched across scopes, has a
// non-numeric scope threshold, or falls below its scope threshold.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/unbound-force/unbound-force/internal/coveragegate"
)

const defaultManifest = "coverage-gate.json"
const defaultProfile = "coverage.out"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses flags, loads the manifest and coverage profile, derives changed
// packages from git, and delegates the threshold check to runGate. stdout and
// stderr are injected for testability.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("coverage-gate", flag.ExitOnError)
	manifestPath := flags.String("manifest", defaultManifest, "path to the coverage-gate manifest JSON")
	profilePath := flags.String("profile", defaultProfile, "path to the go test -coverprofile output")
	base := flags.String("base", "", "base ref or SHA for git diff (defaults to $GITHUB_BASE_SHA, then main)")
	_ = flags.Parse(args)

	manifest, err := loadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
		return 1
	}

	changed, err := changedPackages(resolveBase(*base))
	if err != nil {
		fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
		return 1
	}

	coverage, err := loadCoverage(*profilePath)
	if err != nil {
		fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
		return 1
	}

	return runGate(manifest, changed, coverage, stdout, stderr)
}

// runGate checks the derived changed packages against the manifest and writes
// results. It is separated from run so the threshold logic is testable without
// filesystem or git access.
func runGate(manifest *coveragegate.Manifest, changed []string, coverage map[string]float64, stdout, stderr io.Writer) int {
	failures, err := coveragegate.Check(coveragegate.CheckOptions{
		Manifest:        manifest,
		ChangedPackages: changed,
		Coverage:        coverage,
	})
	if err != nil {
		fmt.Fprintf(stderr, "coverage-gate: %v\n", err)
		return 1
	}
	if len(failures) > 0 {
		fmt.Fprintf(stderr, "coverage-gate: %d failure(s):\n", len(failures))
		for _, failure := range failures {
			fmt.Fprintf(stderr, "  - %s\n", failure)
		}
		return 1
	}
	fmt.Fprintf(stdout, "coverage-gate: %d changed package(s) within scope thresholds\n", len(changed))
	return 0
}

func loadManifest(path string) (*coveragegate.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	return coveragegate.ParseManifest(data)
}

func resolveBase(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv("GITHUB_BASE_SHA"); env != "" {
		return env
	}
	return "main"
}

func loadCoverage(path string) (map[string]float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open coverage profile %q: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	coverage, err := coveragegate.ParseProfile(file)
	if err != nil {
		return nil, err
	}
	module, err := modulePath()
	if err != nil {
		return nil, err
	}
	prefix := module + "/"
	normalized := make(map[string]float64, len(coverage))
	for pkg, pct := range coverage {
		normalized[strings.TrimPrefix(pkg, prefix)] = pct
	}
	return normalized, nil
}

func modulePath() (string, error) {
	output, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return "", fmt.Errorf("resolve module path: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// changedPackages derives module-relative package paths from the production Go
// files changed since the merge base of baseRef and HEAD.
func changedPackages(baseRef string) ([]string, error) {
	mergeBase, err := exec.Command("git", "merge-base", baseRef, "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("resolve merge base with %q: %w", baseRef, err)
	}
	output, err := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", strings.TrimSpace(string(mergeBase)), "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("list changed files: %w", err)
	}
	return coveragegate.PackagesFromFiles(strings.Split(strings.TrimSpace(string(output)), "\n")), nil
}

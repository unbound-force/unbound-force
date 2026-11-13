package refs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// objectNamePattern matches a full lowercase object name: 40 hex
// digits in SHA-1 repositories, 64 in SHA-256 ones.
var objectNamePattern = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// Lookup resolves ref to the object name of the commit it points at
// in the repository at dir, peeling annotated tags.
//
// The name reaches git as a single argument after --end-of-options
// (git 2.24 or later), so it can never be read as an option or reach
// a shell. Ambient GIT_* variables are removed so that dir alone
// selects the repository.
func Lookup(dir, ref string) (string, error) {
	name := Normalize(ref)
	if name == "" {
		return "", fmt.Errorf("resolve ref %q: empty name", ref)
	}
	cmd := exec.Command("git", "rev-parse", "--verify", "--end-of-options", name+"^{commit}")
	cmd.Dir = dir
	cmd.Env = gitEnv(os.Environ())
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("resolve ref %q: %w: %s", ref, err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("resolve ref %q: %w", ref, err)
	}
	sha, err := parseObjectName(out)
	if err != nil {
		return "", fmt.Errorf("resolve ref %q: %w", ref, err)
	}
	return sha, nil
}

// parseObjectName validates rev-parse output as one full object name.
func parseObjectName(out []byte) (string, error) {
	sha := strings.TrimSpace(string(out))
	if !objectNamePattern.MatchString(sha) {
		return "", fmt.Errorf("unexpected rev-parse output %q", sha)
	}
	return sha, nil
}

// gitEnv returns environ without GIT_* variables and with LC_ALL=C,
// so the repository is chosen by the working directory alone and
// git's diagnostics are stable.
func gitEnv(environ []string) []string {
	env := make([]string, 0, len(environ)+1)
	for _, kv := range environ {
		if strings.HasPrefix(kv, "GIT_") || strings.HasPrefix(kv, "LC_ALL=") {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "LC_ALL=C")
}

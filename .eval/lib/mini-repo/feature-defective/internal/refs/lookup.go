package refs

import (
	"fmt"
	"os/exec"
	"strings"
)

// Lookup resolves ref to a commit SHA in the repository at dir.
// The ref usually comes from a CLI flag or a PR's head branch name.
func Lookup(dir, ref string) (string, error) {
	cmd := exec.Command("sh", "-c", "git rev-parse "+Normalize(ref))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

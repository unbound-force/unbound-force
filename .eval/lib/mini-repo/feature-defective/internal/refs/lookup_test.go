package refs

import "testing"

func TestLookup_HEAD(t *testing.T) {
	sha, err := Lookup("../..", "HEAD")
	if err != nil {
		t.Skipf("not in a git checkout: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("Lookup(HEAD) = %q, want a 40-character SHA", sha)
	}
}

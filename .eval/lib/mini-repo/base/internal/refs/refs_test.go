package refs

import "testing"

func TestNormalize_StripsPrefixAndSpace(t *testing.T) {
	cases := map[string]string{
		"main":               "main",
		" refs/heads/main\n": "main",
		"refs/heads/feat/x":  "feat/x",
		"refs/tags/v1.0.0":   "refs/tags/v1.0.0",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

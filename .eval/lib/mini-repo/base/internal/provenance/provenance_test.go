package provenance

import "testing"

func TestStamp_NormalizesRef(t *testing.T) {
	for _, ref := range []string{"main", "refs/heads/main", "  main\n"} {
		if got := Stamp(ref); got.Ref != "main" {
			t.Errorf("Stamp(%q).Ref = %q, want %q", ref, got.Ref, "main")
		}
	}
}

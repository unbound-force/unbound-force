package provenance

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// newRepo creates a repository with one commit on main and returns
// its path and the commit's object name as git reports it.
func newRepo(t *testing.T) (dir, head string) {
	t.Helper()
	if testing.Short() {
		t.Skip("runs git subprocesses")
	}
	dir = t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return dir, strings.TrimSpace(string(out))
}

func TestStamp_PinsResolvedCommit(t *testing.T) {
	dir, head := newRepo(t)
	got, err := Stamp(dir, " refs/heads/main\n")
	if err != nil {
		t.Fatalf("Stamp: %v", err)
	}
	if want := (Record{Ref: "main", Commit: head}); got != want {
		t.Errorf("Stamp = %+v, want %+v", got, want)
	}
}

func TestStamp_FailsWithoutCommit(t *testing.T) {
	dir, _ := newRepo(t)
	got, err := Stamp(dir, "missing")
	if err == nil {
		t.Fatalf("Stamp(missing) = %+v, want an error", got)
	}
	if got != (Record{}) {
		t.Errorf("Stamp(missing) returned %+v with its error, want a zero Record", got)
	}
	for _, want := range []string{"stamp provenance", `"missing"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestRecord_JSONShape(t *testing.T) {
	out, err := json.Marshal(Record{Ref: "main", Commit: "abc"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if want := `{"ref":"main","commit":"abc"}`; string(out) != want {
		t.Errorf("JSON = %s, want %s", out, want)
	}
}

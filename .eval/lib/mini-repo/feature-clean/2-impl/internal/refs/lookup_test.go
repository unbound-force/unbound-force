package refs

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// newRepo creates a repository with one commit on main and an
// annotated tag v1 on it, and returns its path and the commit's
// object name as git itself reports it. initArgs are passed to
// git init.
func newRepo(t *testing.T, initArgs ...string) (dir, head string) {
	t.Helper()
	if testing.Short() {
		t.Skip("runs git subprocesses")
	}
	dir = t.TempDir()
	id := []string{"-c", "user.name=t", "-c", "user.email=t@example.invalid"}
	runGit(t, dir, append([]string{"init", "-q", "-b", "main"}, initArgs...)...)
	// The directory name makes each repository's commit unique.
	runGit(t, dir, append(id, "commit", "-q", "--allow-empty", "-m", "init "+filepath.Base(dir))...)
	runGit(t, dir, append(id, "tag", "-a", "-m", "v1", "v1")...)
	return dir, strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv(os.Environ())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func TestLookup_ResolvesBranchSpellingsToHeadCommit(t *testing.T) {
	dir, head := newRepo(t)
	for _, ref := range []string{"main", "refs/heads/main", " main\n", "HEAD"} {
		got, err := Lookup(dir, ref)
		if err != nil {
			t.Errorf("Lookup(%q): %v", ref, err)
		} else if got != head {
			t.Errorf("Lookup(%q) = %q, want %q", ref, got, head)
		}
	}
}

func TestLookup_PeelsAnnotatedTag(t *testing.T) {
	dir, head := newRepo(t)
	tagObject := strings.TrimSpace(runGit(t, dir, "rev-parse", "v1"))
	if tagObject == head {
		t.Fatal("test setup: v1 is not an annotated tag")
	}
	got, err := Lookup(dir, "v1")
	if err != nil {
		t.Fatalf("Lookup(v1): %v", err)
	}
	if got != head {
		t.Errorf("Lookup(v1) = %q, want commit %q (tag object is %q)", got, head, tagObject)
	}
}

func TestLookup_SHA256Repository(t *testing.T) {
	dir, head := newRepo(t, "--object-format=sha256")
	got, err := Lookup(dir, "main")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got != head || len(got) != 64 {
		t.Errorf("Lookup = %q, want 64-digit %q", got, head)
	}
}

func TestLookup_RejectsHostileNamesWithoutSideEffects(t *testing.T) {
	dir, _ := newRepo(t)
	marker := filepath.Join(t.TempDir(), "pwned")
	for _, ref := range []string{
		"",
		"  ",
		"--output=" + marker,
		"main; touch " + marker,
		"$(touch " + marker + ")",
		"main`touch " + marker + "`",
	} {
		sha, err := Lookup(dir, ref)
		if err == nil {
			t.Errorf("Lookup(%q) = %q, want an error", ref, sha)
			continue
		}
		if sha != "" {
			t.Errorf("Lookup(%q) returned %q alongside its error, want \"\"", ref, sha)
		}
		if !strings.HasPrefix(err.Error(), "resolve ref ") {
			t.Errorf("Lookup(%q) error does not name the ref: %v", ref, err)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("a hostile ref created %s (stat err %v)", marker, err)
	}
}

func TestLookup_IgnoresAmbientGitDir(t *testing.T) {
	dir, head := newRepo(t)
	other, otherHead := newRepo(t)
	if head == otherHead {
		t.Fatal("test setup: repositories share a commit")
	}
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	got, err := Lookup(dir, "main")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got != head {
		t.Errorf("Lookup = %q, want %q from dir, not %q from GIT_DIR", got, head, otherHead)
	}
}

func TestLookup_UnknownRefCarriesGitDiagnostic(t *testing.T) {
	dir, _ := newRepo(t)
	sha, err := Lookup(dir, "does-not-exist")
	if err == nil {
		t.Fatalf("Lookup(does-not-exist) = %q, want an error", sha)
	}
	if sha != "" {
		t.Errorf("Lookup returned %q alongside its error, want \"\"", sha)
	}
	for _, want := range []string{`"does-not-exist"`, "Needed a single revision"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestParseObjectName_AcceptsOnlyFullObjectNames(t *testing.T) {
	sha1 := strings.Repeat("a1", 20)
	sha256 := strings.Repeat("b2", 32)
	for _, tc := range []struct {
		out  string
		want string
		ok   bool
	}{
		{sha1 + "\n", sha1, true},
		{sha256, sha256, true},
		{"", "", false},
		{"main", "", false},
		{strings.ToUpper(sha1), "", false},
		{sha1 + "0", "", false},
		{sha1 + "\n" + sha1, "", false},
	} {
		got, err := parseObjectName([]byte(tc.out))
		if (err == nil) != tc.ok || got != tc.want {
			t.Errorf("parseObjectName(%q) = %q, %v; want %q, ok=%v", tc.out, got, err, tc.want, tc.ok)
		}
	}
}

func TestGitEnv_DropsGitVariablesAndFixesLocale(t *testing.T) {
	got := gitEnv([]string{"PATH=/bin", "GIT_DIR=/x", "GIT_WORK_TREE=/y", "LC_ALL=fr_FR.UTF-8", "HOME=/h"})
	want := []string{"PATH=/bin", "HOME=/h", "LC_ALL=C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("gitEnv = %q, want %q", got, want)
	}
}

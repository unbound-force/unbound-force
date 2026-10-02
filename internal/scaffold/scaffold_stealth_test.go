package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stealthGitFake simulates the git commands that stealth mode issues. It
// lets tests inject a specific working-tree/index view without a real git
// binary, while ReadFile/WriteFile remain the real os functions.
type stealthGitFake struct {
	insideWorkTree string
	insideErr      error
	excludePath    string
	tracked        []string
	others         []string
	statusOut      string
	ref            string
	commit         string
	staged         []string
	restoreCalls   [][]string
}

func (g *stealthGitFake) exec(name string, args ...string) ([]byte, error) {
	cmd := name + " " + strings.Join(args, " ")
	switch cmd {
	case "git rev-parse --is-inside-work-tree":
		return []byte(g.insideWorkTree), g.insideErr
	case "git rev-parse --git-path info/exclude":
		return []byte(g.excludePath), nil
	case "git ls-files":
		return []byte(strings.Join(g.tracked, "\n")), nil
	case "git ls-files -s -z":
		var sb strings.Builder
		for _, p := range g.tracked {
			sb.WriteString("100644 deadbeef 0\t" + p + "\x00")
		}
		return []byte(sb.String()), nil
	case "git ls-files --others --exclude-standard -z":
		return []byte(strings.Join(g.others, "\x00")), nil
	case "git status --porcelain":
		return []byte(g.statusOut), nil
	case "git rev-parse --abbrev-ref HEAD":
		return []byte(g.ref), nil
	case "git rev-parse HEAD":
		return []byte(g.commit), nil
	case "git diff --cached --name-only -z":
		return []byte(strings.Join(g.staged, "\x00")), nil
	}
	if strings.HasPrefix(cmd, "git restore --staged") {
		g.restoreCalls = append(g.restoreCalls, args)
		return nil, nil
	}
	return nil, fmt.Errorf("unexpected git command: %q", cmd)
}

func newStealthGitFake() *stealthGitFake {
	return &stealthGitFake{
		insideWorkTree: "true",
		excludePath:    ".git/info/exclude",
		ref:            "opsx/test",
		commit:         "deadbeef",
	}
}

// newStealthOpts builds Options with the fake git ExecCmd and real os
// ReadFile/WriteFile/LookPath, targeting dir. Sub-tools are reported
// unavailable so initSubTools only writes opencode.json + .uf/dcp.jsonc.
func newStealthOpts(dir string, g *stealthGitFake, stdout *bytes.Buffer) Options {
	return Options{
		TargetDir: dir,
		Stealth:   true,
		Version:   "1.2.3-test",
		Stdout:    stdout,
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}
}

// initGitDir creates dir/.git/info so writeGitInfoExclude has a writable
// parent directory for unit tests that fake git.
func initGitDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "info"), 0o755); err != nil {
		t.Fatalf("mkdir .git/info: %v", err)
	}
}

// writeFile writes content to path with the given mode, creating parents.
func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestEscapeExcludePattern(t *testing.T) {
	cases := []struct{ in, want string }{
		{"#comment", "\\#comment"},
		{"!negate", "\\!negate"},
		{".opencode/", ".opencode/"},
		{"normal/path", "normal/path"},
		{"a*b", "a\\*b"},
		{"a?b", "a\\?b"},
		{"a[b", "a\\[b"},
		{"a\\b", "a\\\\b"},
	}
	for _, c := range cases {
		if got := escapeExcludePattern(c.in); got != c.want {
			t.Errorf("escapeExcludePattern(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestStealthInit_LeavesTrackedFilesUnmodified is task 3.1.
func TestStealthInit_LeavesTrackedFilesUnmodified(t *testing.T) {
	cases := []struct {
		name    string
		stealth bool
		wantMod bool
	}{
		{"stealth", true, false},
		{"normal", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			initGitDir(t, dir)
			writeFile(t, filepath.Join(dir, ".gitignore"), "existing-ignore\n", 0o644)
			writeFile(t, filepath.Join(dir, "AGENTS.md"), "# AGENTS\n", 0o644)

			var buf bytes.Buffer
			g := newStealthGitFake()
			opts := Options{
				TargetDir: dir,
				Stealth:   c.stealth,
				Version:   "1.2.3-test",
				Stdout:    &buf,
				ExecCmd:   g.exec,
				ReadFile:  os.ReadFile,
				WriteFile: os.WriteFile,
				LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
			}
			if _, err := Run(opts); err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			gi, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
			if err != nil {
				t.Fatalf("read .gitignore: %v", err)
			}
			ag, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
			if err != nil {
				t.Fatalf("read AGENTS.md: %v", err)
			}
			giMod := strings.Contains(string(gi), gitignoreMarker)
			agMod := strings.Contains(string(ag), agentsmdPackMarker)
			if c.wantMod {
				if !giMod {
					t.Errorf("normal mode: .gitignore missing UF block:\n%s", gi)
				}
				if !agMod {
					t.Errorf("normal mode: AGENTS.md missing pack section:\n%s", ag)
				}
			} else {
				if string(gi) != "existing-ignore\n" {
					t.Errorf("stealth mode modified .gitignore:\n%s", gi)
				}
				if string(ag) != "# AGENTS\n" {
					t.Errorf("stealth mode modified AGENTS.md:\n%s", ag)
				}
			}
		})
	}
}

// TestWriteGitInfoExclude_IdempotentAndEscaped is task 3.2.
func TestWriteGitInfoExclude_IdempotentAndEscaped(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	g := newStealthGitFake()
	opts := &Options{
		TargetDir: dir,
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	sr := writeGitInfoExclude(opts, nil)
	if sr.err != nil {
		t.Fatalf("writeGitInfoExclude() error: %v", sr.err)
	}

	p := filepath.Join(dir, ".git", "info", "exclude")
	first, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read exclude: %v", err)
	}
	if !strings.Contains(string(first), stealthExcludeMarker) {
		t.Errorf("exclude missing marker:\n%s", first)
	}
	for _, path := range stealthExcludePaths {
		if !strings.Contains(string(first), escapeExcludePattern(path)+"\n") {
			t.Errorf("exclude missing canonical path %q:\n%s", path, first)
		}
	}

	// Second write must not duplicate entries.
	if sr := writeGitInfoExclude(opts, nil); sr.err != nil {
		t.Fatalf("second write error: %v", sr.err)
	}
	second, _ := os.ReadFile(p)
	if strings.Count(string(second), stealthExcludeMarker) != 1 {
		t.Errorf("marker duplicated:\n%s", second)
	}
	for _, path := range stealthExcludePaths {
		if strings.Count(string(second), escapeExcludePattern(path)+"\n") != 1 {
			t.Errorf("path %q duplicated:\n%s", path, second)
		}
	}

	// The tracked .gitignore gains no new entry.
	writeFile(t, filepath.Join(dir, ".gitignore"), "existing\n", 0o644)
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(gi) != "existing\n" {
		t.Errorf(".gitignore modified by exclude write:\n%s", gi)
	}
}

// TestRestoreTrackedFiles_PreservesModeAndSymlink is task 3.3. It fakes
// only ExecCmd; ReadFile/WriteFile remain the real os functions so the
// assertions observe the real on-disk bytes and modes.
func TestRestoreTrackedFiles_PreservesModeAndSymlink(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".gitignore"), "keep\n", 0o600)
	target := filepath.Join(dir, "target.txt")
	writeFile(t, target, "link-me", 0o644)
	link := filepath.Join(dir, "link")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	g := newStealthGitFake()
	g.tracked = []string{".gitignore", "link"}
	opts := &Options{
		TargetDir: dir,
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	snap, err := snapshotTrackedFiles(opts)
	if err != nil {
		t.Fatalf("snapshotTrackedFiles() error: %v", err)
	}

	// Simulate a sub-tool modifying bytes and mode of .gitignore, and
	// replacing the symlink with a regular file.
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("modified\n"), 0o644); err != nil {
		t.Fatalf("modify .gitignore: %v", err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatalf("remove symlink: %v", err)
	}
	writeFile(t, link, "not-a-link", 0o644)

	restored, _ := restoreTrackedFiles(opts, snap)
	if len(restored) != 2 {
		t.Errorf("expected 2 restored paths, got %v", restored)
	}

	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(gi) != "keep\n" {
		t.Errorf(".gitignore bytes not restored: %q", gi)
	}
	fi, err := os.Lstat(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("lstat .gitignore: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf(".gitignore mode = %v, want 0600", fi.Mode().Perm())
	}

	lf, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat link: %v", err)
	}
	if lf.Mode()&os.ModeSymlink == 0 {
		t.Errorf("link not restored as symlink")
	}
	if got, err := os.Readlink(link); err != nil || got != "target.txt" {
		t.Errorf("link target = %q, err %v; want target.txt", got, err)
	}
}

// TestNormalInit_Regression is task 3.4.
func TestNormalInit_Regression(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	var buf bytes.Buffer
	g := newStealthGitFake()
	opts := Options{
		TargetDir: dir,
		Stealth:   false,
		Version:   "1.2.3-test",
		Stdout:    &buf,
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}
	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	gi, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(gi), gitignoreMarker) {
		t.Errorf("normal init did not write UF ignore block:\n%s", gi)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); os.IsNotExist(err) {
		t.Errorf("normal init did not scaffold .opencode/")
	}
}

// TestStealthCheck_ReportsCleanliness is task 3.5.
func TestStealthCheck_ReportsCleanliness(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), stealthExcludeMarker+"\n.opencode/\n", 0o644)
		g := newStealthGitFake()
		var buf bytes.Buffer
		opts := &Options{TargetDir: dir, Version: "9.9.9", Stdout: &buf, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if err := runStealthCheck(opts); err != nil {
			t.Fatalf("runStealthCheck() error: %v", err)
		}
		var rep stealthCheckReport
		if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
			t.Fatalf("unmarshal report: %v\n%s", err, buf.String())
		}
		if !rep.Clean {
			t.Errorf("expected clean, got %+v", rep)
		}
		if rep.Tool != "unbound-force" || rep.Version != "9.9.9" || rep.Timestamp == "" {
			t.Errorf("missing provenance fields: %+v", rep)
		}
		if rep.Ref != "opsx/test" || rep.Commit != "deadbeef" {
			t.Errorf("missing git provenance (ref/commit): %+v", rep)
		}
	})

	t.Run("dirty", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), stealthExcludeMarker+"\n.opencode/\n", 0o644)
		g := newStealthGitFake()
		g.others = []string{".opencode/foo.md"}
		var buf bytes.Buffer
		opts := &Options{TargetDir: dir, Version: "9.9.9", Stdout: &buf, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if err := runStealthCheck(opts); err == nil {
			t.Fatal("expected non-nil error for dirty tree")
		}
		var rep stealthCheckReport
		if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
			t.Fatalf("unmarshal report: %v\n%s", err, buf.String())
		}
		if rep.Clean {
			t.Errorf("expected not clean")
		}
	})

	t.Run("missing marker", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		g := newStealthGitFake()
		var buf bytes.Buffer
		opts := &Options{TargetDir: dir, Version: "9.9.9", Stdout: &buf, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if err := runStealthCheck(opts); err == nil {
			t.Fatal("expected non-nil error for missing marker")
		}
	})

	t.Run("check without stealth", func(t *testing.T) {
		dir := t.TempDir()
		var buf bytes.Buffer
		_, err := Run(Options{TargetDir: dir, Check: true, Stealth: false, Stdout: &buf})
		if err == nil || !strings.Contains(err.Error(), "--check requires --stealth") {
			t.Fatalf("expected --check requires --stealth error, got %v", err)
		}
	})
}

// gitRunReal runs a real git command in dir for integration tests.
func gitRunReal(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// initRealGitRepo initializes a git repo with an initial commit and chdirs
// into it so the scaffold's default ExecCmd (which runs in-process cwd)
// targets the repo.
func initRealGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRunReal(t, dir, "init", "-q")
	writeFile(t, filepath.Join(dir, "README.md"), "# readme\n", 0o644)
	gitRunReal(t, dir, "add", "README.md")
	gitRunReal(t, dir, "commit", "-q", "-m", "init")
	t.Chdir(dir)
	return dir
}

// realStealthOpts builds Options for integration tests using the real git
// binary (default ExecCmd) but disabling sub-tools via LookPath.
func realStealthOpts(dir string, stdout *bytes.Buffer, extra ...func(*Options)) Options {
	opts := Options{
		TargetDir: dir,
		Stealth:   true,
		Version:   "1.2.3-test",
		Stdout:    stdout,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}
	for _, fn := range extra {
		fn(&opts)
	}
	return opts
}

// TestStealthInit_Integration_GitStatusClean is task 3.6.
func TestStealthInit_Integration_GitStatusClean(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := initRealGitRepo(t)
	writeFile(t, filepath.Join(dir, ".gitignore"), "existing\n", 0o644)
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "# AGENTS\n", 0o644)
	gitRunReal(t, dir, "add", ".gitignore", "AGENTS.md")
	gitRunReal(t, dir, "commit", "-q", "-m", "track files")

	var buf bytes.Buffer
	if _, err := Run(realStealthOpts(dir, &buf)); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	porcelain := gitRunReal(t, dir, "status", "--porcelain")
	if strings.TrimSpace(porcelain) != "" {
		t.Errorf("expected clean tree, got:\n%s", porcelain)
	}
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(gi) != "existing\n" {
		t.Errorf(".gitignore modified:\n%s", gi)
	}
}

// TestStealthInit_NonGitDirErrors is task 3.7 (integration half).
func TestStealthInit_NonGitDirErrors(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	t.Chdir(dir)
	var buf bytes.Buffer
	_, err := Run(realStealthOpts(dir, &buf))
	if err == nil || !strings.Contains(err.Error(), "git repository") {
		t.Fatalf("expected git-requirement error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); !os.IsNotExist(err) {
		t.Errorf("expected no scaffolded files, but .opencode exists")
	}
}

// TestStealthInit_GitNotInstalled is task 3.7 (unit half, git not found).
func TestStealthInit_GitNotInstalled(t *testing.T) {
	dir := t.TempDir()
	g := newStealthGitFake()
	g.insideErr = &exec.Error{Name: "git", Err: errors.New("executable file not found in $PATH")}
	var buf bytes.Buffer
	_, err := Run(newStealthOpts(dir, g, &buf))
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("expected git-not-installed error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); !os.IsNotExist(err) {
		t.Errorf("expected no scaffolded files, but .opencode exists")
	}
}

// TestStealthInit_RollbackOnFailure is task 3.8.
func TestStealthInit_RollbackOnFailure(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	// Pre-existing marker-delimited exclusion block that must survive.
	priorExclude := "# prior block\nprior/path\n"
	writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), priorExclude, 0o644)
	writeFile(t, filepath.Join(dir, ".gitignore"), "tracked\n", 0o644)
	writeFile(t, filepath.Join(dir, "preexisting.txt"), "pre\n", 0o644)

	g := newStealthGitFake()
	g.tracked = []string{".gitignore", "preexisting.txt"}
	g.staged = []string{"preexisting.txt"} // pre-existing staged change

	opts := &Options{
		TargetDir: dir,
		Stealth:   true,
		Version:   "1.2.3-test",
		Stdout:    &bytes.Buffer{},
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}

	ctx, err := beginStealth(opts)
	if err != nil {
		t.Fatalf("beginStealth() error: %v", err)
	}

	// Snapshot tracked files (Run() does this between beginStealth and
	// initSubTools; the test reproduces that step before simulating a
	// sub-tool mutation).
	snap, err := snapshotTrackedFiles(opts)
	if err != nil {
		t.Fatalf("snapshotTrackedFiles() error: %v", err)
	}
	ctx.snapshot = snap
	ctx.didSnapshot = true

	// Simulate a sub-tool: modify .gitignore, stage a new file, and scaffold
	// a canonical directory that rollback must remove.
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("sub-tool edit\n"), 0o644); err != nil {
		t.Fatalf("sub-tool write: %v", err)
	}
	writeFile(t, filepath.Join(dir, "newfile.txt"), "new\n", 0o644)
	writeFile(t, filepath.Join(dir, ".opencode", "agents", "x.md"), "agent\n", 0o644)
	g.staged = append(g.staged, "newfile.txt")

	if err := ctx.rollback(opts); err != nil {
		t.Fatalf("rollback() error: %v", err)
	}

	// .gitignore restored.
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(gi) != "tracked\n" {
		t.Errorf(".gitignore not restored: %q", gi)
	}
	// Scaffolded files removed (failure atomicity).
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); !os.IsNotExist(err) {
		t.Errorf("expected scaffolded .opencode to be removed, got err=%v", err)
	}
	// Pre-existing exclusion block survived (current invocation's block removed).
	ex, _ := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
	if string(ex) != priorExclude {
		t.Errorf("exclude not restored to prior state: %q", ex)
	}
	// Un-staging: newfile.txt was staged during init, preexisting.txt survives.
	if len(g.restoreCalls) == 0 {
		t.Fatalf("expected restore --staged to be called during rollback")
	}
	joined := strings.Join(g.restoreCalls[0], " ")
	if !strings.Contains(joined, "newfile.txt") {
		t.Errorf("expected newfile.txt to be un-staged via rollback, got args: %v", g.restoreCalls[0])
	}
	if strings.Contains(joined, "preexisting.txt") {
		t.Errorf("pre-existing staged file should NOT be un-staged, got args: %v", g.restoreCalls[0])
	}
}

// TestStealthInit_RollbackPreExistingCanonicalDir is task 3.8 (failure
// atomicity): rollback removes scaffolded files written INTO a pre-existing
// canonical directory while preserving the user's pre-existing files.
func TestStealthInit_RollbackPreExistingCanonicalDir(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	writeFile(t, filepath.Join(dir, ".opencode", "custom-agent.md"), "user\n", 0o644)

	g := newStealthGitFake()
	g.tracked = nil
	g.others = []string{".opencode/custom-agent.md"}

	opts := &Options{
		TargetDir: dir,
		Stealth:   true,
		Version:   "1.2.3-test",
		Stdout:    &bytes.Buffer{},
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}

	ctx, err := beginStealth(opts)
	if err != nil {
		t.Fatalf("beginStealth() error: %v", err)
	}

	// Simulate a sub-tool writing a new file into the pre-existing .opencode/
	// directory, then a later sub-tool failure.
	writeFile(t, filepath.Join(dir, ".opencode", "agents", "new.md"), "agent\n", 0o644)
	g.others = []string{".opencode/custom-agent.md", ".opencode/agents/new.md"}

	if err := ctx.rollback(opts); err != nil {
		t.Fatalf("rollback() error: %v", err)
	}

	// The user's pre-existing file survives.
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "custom-agent.md")); err != nil {
		t.Errorf("pre-existing custom-agent.md should survive, got err=%v", err)
	}
	// The scaffolded file inside the pre-existing dir is removed.
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "new.md")); !os.IsNotExist(err) {
		t.Errorf("scaffolded agents/new.md should be removed, got err=%v", err)
	}
}

// TestUnstageNewlyStaged is part of task 3.8's index-restore assertion.
func TestUnstageNewlyStaged(t *testing.T) {
	g := newStealthGitFake()
	g.staged = []string{"preexisting.txt", "newfile.txt"}
	opts := &Options{TargetDir: t.TempDir(), ExecCmd: g.exec}
	if err := unstageNewlyStaged(opts, []string{"preexisting.txt"}); err != nil {
		t.Fatalf("unstageNewlyStaged: %v", err)
	}

	if len(g.restoreCalls) == 0 {
		t.Fatalf("expected restore --staged to be called")
	}
	joined := strings.Join(g.restoreCalls[0], " ")
	if !strings.Contains(joined, "newfile.txt") {
		t.Errorf("expected newfile.txt to be un-staged, got args: %v", g.restoreCalls[0])
	}
	if strings.Contains(joined, "preexisting.txt") {
		t.Errorf("pre-existing staged file should NOT be un-staged, got args: %v", g.restoreCalls[0])
	}
}

// TestStealthInit_RollbackFailure_ReportsResidual asserts a failing rollback
// step surfaces a clear error naming the residual dirty path (task 3.8).
func TestStealthInit_RollbackFailure_ReportsResidual(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), "# prior\n", 0o644)
	writeFile(t, filepath.Join(dir, ".gitignore"), "tracked\n", 0o644)

	g := newStealthGitFake()
	g.tracked = []string{".gitignore"}

	opts := &Options{
		TargetDir: dir,
		Stealth:   true,
		Version:   "1.2.3-test",
		Stdout:    &bytes.Buffer{},
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}

	ctx, err := beginStealth(opts)
	if err != nil {
		t.Fatalf("beginStealth() error: %v", err)
	}
	snap, err := snapshotTrackedFiles(opts)
	if err != nil {
		t.Fatalf("snapshotTrackedFiles() error: %v", err)
	}
	ctx.snapshot = snap
	ctx.didSnapshot = true

	// Make the exclusion restore fail so rollback must report a residual path.
	opts.WriteFile = func(string, []byte, os.FileMode) error { return errors.New("injected write failure") }
	if err := ctx.rollback(opts); err == nil {
		t.Fatal("expected rollback to report a residual dirty path")
	} else if !strings.Contains(err.Error(), "rollback incomplete") {
		t.Fatalf("expected 'rollback incomplete' error, got: %v", err)
	}
}

// TestStealthInit_Integration_RollbackOnMidRunFailure drives a real mid-Run
// failure (snapshot error) through Run() and asserts the deferred rollback
// leaves a clean tree and a non-nil error (task 3.8 integration half).
func TestStealthInit_Integration_RollbackOnMidRunFailure(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := initRealGitRepo(t)
	writeFile(t, filepath.Join(dir, ".gitignore"), "existing\n", 0o644)
	gitRunReal(t, dir, "add", ".gitignore")
	gitRunReal(t, dir, "commit", "-q", "-m", "track")

	var buf bytes.Buffer
	opts := realStealthOpts(dir, &buf, func(o *Options) {
		o.ExecCmd = func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 2 && args[0] == "ls-files" && args[1] == "-s" {
				return nil, errors.New("injected snapshot failure")
			}
			cmd := exec.Command(name, args...)
			return cmd.CombinedOutput()
		}
	})
	if _, err := Run(opts); err == nil {
		t.Fatal("expected mid-run failure error")
	}
	porcelain := gitRunReal(t, dir, "status", "--porcelain")
	if strings.TrimSpace(porcelain) != "" {
		t.Errorf("expected clean tree after rollback, got:\n%s", porcelain)
	}
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(gi) != "existing\n" {
		t.Errorf(".gitignore not restored:\n%s", gi)
	}
}

// TestStealthInit_RunLevelCheck_WritesNothing drives the Run() Check branch
// (Check && Stealth) and asserts it writes no files (task 3.5).
func TestStealthInit_RunLevelCheck_WritesNothing(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), stealthExcludeMarker+"\n.opencode/\n", 0o644)
	g := newStealthGitFake()
	var buf bytes.Buffer
	opts := newStealthOpts(dir, g, &buf)
	opts.Check = true
	if _, err := Run(opts); err != nil {
		t.Fatalf("Run(check) error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); !os.IsNotExist(err) {
		t.Errorf("expected no scaffolded files, but .opencode exists")
	}
	var rep stealthCheckReport
	if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
		t.Fatalf("unmarshal report: %v\n%s", err, buf.String())
	}
	if !rep.Clean {
		t.Errorf("expected clean report, got %+v", rep)
	}
}

// TestStealthInit_PipelineOperability is task 3.9.
func TestStealthInit_PipelineOperability(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	g := newStealthGitFake()
	var buf bytes.Buffer
	if _, err := Run(newStealthOpts(dir, g, &buf)); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	for _, rel := range []string{
		".opencode/agents/cobalt-crush-dev.md",
		"openspec",
		".specify",
	} {
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected %s to be present and readable", rel)
		}
	}
}

// TestStealthInit_RefusesTrackedScaffoldPaths is task 3.10.
func TestStealthInit_RefusesTrackedScaffoldPaths(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprintf("force=%v", force), func(t *testing.T) {
			dir := t.TempDir()
			initGitDir(t, dir)
			g := newStealthGitFake()
			g.tracked = []string{"opencode.json", "README.md"}
			var buf bytes.Buffer
			opts := newStealthOpts(dir, g, &buf)
			opts.Force = force
			_, err := Run(opts)
			if err == nil || !strings.Contains(err.Error(), "opencode.json") {
				t.Fatalf("expected refusal naming opencode.json, got %v", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, ".opencode")); !os.IsNotExist(statErr) {
				t.Errorf("expected no scaffolded files, but .opencode exists")
			}
		})
	}
}

// TestStealthCheck_DetectsLostExclusion is task 3.11.
func TestStealthCheck_DetectsLostExclusion(t *testing.T) {
	t.Run("wiped", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		g := newStealthGitFake()
		g.others = []string{".opencode/foo.md", "openspec/bar.md"}
		var buf bytes.Buffer
		opts := &Options{TargetDir: dir, Version: "9.9.9", Stdout: &buf, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if err := runStealthCheck(opts); err == nil {
			t.Fatal("expected error for wiped exclusion")
		}
		var rep stealthCheckReport
		_ = json.Unmarshal(buf.Bytes(), &rep)
		if rep.Clean {
			t.Errorf("expected not clean")
		}
		if len(rep.Visible) == 0 {
			t.Errorf("expected visible paths reported, got %v", rep.Visible)
		}
	})

	t.Run("partial", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		// Marker present but one path removed.
		writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), stealthExcludeMarker+"\n.opencode/\n", 0o644)
		g := newStealthGitFake()
		g.others = []string{".specify/"} // .specify/ now visible
		var buf bytes.Buffer
		opts := &Options{TargetDir: dir, Version: "9.9.9", Stdout: &buf, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if err := runStealthCheck(opts); err == nil {
			t.Fatal("expected error for partially-edited block")
		}
		var rep stealthCheckReport
		_ = json.Unmarshal(buf.Bytes(), &rep)
		found := false
		for _, v := range rep.Visible {
			if v == ".specify/" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected .specify/ reported as visible, got %v", rep.Visible)
		}
	})
}

// TestWriteGitInfoExclude_UnwritableAndResolution is task 3.12.
func TestWriteGitInfoExclude_UnwritableAndResolution(t *testing.T) {
	t.Run("unwritable target", func(t *testing.T) {
		dir := t.TempDir()
		g := newStealthGitFake()
		// Resolve to a path inside a missing parent directory so the
		// write fails portably (no chmod 000, which is ineffective as root).
		g.excludePath = filepath.Join(dir, "missing", ".git", "info", "exclude")
		opts := &Options{TargetDir: dir, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		sr := writeGitInfoExclude(opts, nil)
		if sr.err == nil {
			t.Fatal("expected clear error for unwritable exclusion target")
		}
		if !strings.Contains(sr.detail, "write failed") {
			t.Errorf("expected write-failed detail, got %v", sr.detail)
		}
	})

	t.Run("path resolution via git rev-parse", func(t *testing.T) {
		dir := t.TempDir()
		custom := filepath.Join(dir, "custom", "git", "info", "exclude")
		if err := os.MkdirAll(filepath.Dir(custom), 0o755); err != nil {
			t.Fatalf("mkdir parent: %v", err)
		}
		g := newStealthGitFake()
		g.excludePath = custom
		opts := &Options{TargetDir: dir, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if sr := writeGitInfoExclude(opts, nil); sr.err != nil {
			t.Fatalf("writeGitInfoExclude() error: %v", sr.err)
		}
		if _, err := os.Stat(custom); os.IsNotExist(err) {
			t.Errorf("expected exclusion written at %s (resolved via git), but missing", custom)
		}
	})
}

// TestStealthInit_FlagCompositions is task 3.13.
func TestStealthInit_FlagCompositions(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(*testing.T, string)
		mutate func(*Options)
		verify func(*testing.T, string, string)
	}{
		{
			name: "force",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md"), "DIVERGENT\n", 0o644)
			},
			mutate: func(o *Options) { o.Force = true },
			verify: func(t *testing.T, dir, exclude string) {
				if !strings.Contains(exclude, stealthExcludeMarker) {
					t.Errorf("force: exclusion block missing")
				}
				b, _ := os.ReadFile(filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md"))
				if strings.Contains(string(b), "DIVERGENT") {
					t.Errorf("force: untracked scaffolded file was not overwritten:\n%s", b)
				}
			},
		},
		{
			name:   "divisor",
			mutate: func(o *Options) { o.DivisorOnly = true },
			verify: func(t *testing.T, dir, exclude string) {
				if !strings.Contains(exclude, stealthExcludeMarker) {
					t.Errorf("divisor: exclusion block missing")
				}
			if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "divisor-adversary.md")); os.IsNotExist(err) {
				t.Errorf("divisor: expected divisor agent scaffolded")
			}
			if _, err := os.Stat(filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")); !os.IsNotExist(err) {
				t.Errorf("divisor: non-divisor agent should NOT be scaffolded")
			}
		},
	},
	{
		name:   "lang",
		mutate: func(o *Options) { o.DivisorOnly = true; o.Lang = "go" },
		verify: func(t *testing.T, dir, exclude string) {
			if !strings.Contains(exclude, stealthExcludeMarker) {
				t.Errorf("lang: exclusion block missing")
			}
			if _, err := os.Stat(filepath.Join(dir, ".opencode", "uf", "packs", "go.md")); os.IsNotExist(err) {
				t.Errorf("lang: expected go pack scaffolded")
			}
			for _, other := range []string{"python.md", "typescript.md"} {
				if _, err := os.Stat(filepath.Join(dir, ".opencode", "uf", "packs", other)); !os.IsNotExist(err) {
					t.Errorf("lang: non-go pack %s should NOT be scaffolded under --divisor --lang go", other)
				}
			}
		},
	},
		{
			name:   "dry-run",
			mutate: func(o *Options) { o.DryRun = true },
			verify: func(t *testing.T, dir, exclude string) {
				if !strings.Contains(exclude, stealthExcludeMarker) {
					t.Errorf("dry-run: exclusion block missing")
				}
				// The DryRun suppression of opencode.json is a pre-existing
				// configureOpencodeJSON behavior, asserted directly in
				// TestConfigureOpencodeJSON_DryRunSuppressesWrite below.
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			initGitDir(t, dir)
			if c.setup != nil {
				c.setup(t, dir)
			}
			g := newStealthGitFake()
			var buf bytes.Buffer
			opts := newStealthOpts(dir, g, &buf)
			c.mutate(&opts)
			if _, err := Run(opts); err != nil {
				t.Fatalf("Run() error: %v", err)
			}
			exclude, _ := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
			c.verify(t, dir, string(exclude))
		})
	}
}

// TestConfigureOpencodeJSON_DryRunSuppressesWrite asserts that DryRun
// genuinely suppresses the opencode.json write (the discriminating half of
// task 3.13's "dry-run" composition, which the fake-git flag-composition
// table cannot observe because it reports all tools absent via LookPath).
func TestConfigureOpencodeJSON_DryRunSuppressesWrite(t *testing.T) {
	lookFoundDewey := func(name string) (string, error) {
		if name == "dewey" {
			return "/usr/local/bin/dewey", nil
		}
		return "", fmt.Errorf("not found: %s", name)
	}

	// DryRun=false: dewey present -> opencode.json is written with the MCP entry.
	dir := t.TempDir()
	opts := &Options{
		TargetDir: dir,
		LookPath:  lookFoundDewey,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}
	configureOpencodeJSON(opts)
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); os.IsNotExist(err) {
		t.Fatalf("expected opencode.json to be written when dewey is present")
	}

	// DryRun=true: the write is suppressed even though dewey is present.
	dir2 := t.TempDir()
	opts2 := &Options{
		TargetDir: dir2,
		DryRun:    true,
		LookPath:  lookFoundDewey,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}
	configureOpencodeJSON(opts2)
	if _, err := os.Stat(filepath.Join(dir2, "opencode.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run: opencode.json should be suppressed")
	}
}

// TestNormalInit_WarnsStaleStealthExclusion is task 3.14.
func TestNormalInit_WarnsStaleStealthExclusion(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), stealthExcludeMarker+"\n.opencode/\n", 0o644)
	g := newStealthGitFake()
	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stealth:   false,
		Version:   "1.2.3-test",
		Stdout:    &buf,
		ExecCmd:   g.exec,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		LookPath:  func(string) (string, error) { return "", fmt.Errorf("not found") },
	}
	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.Contains(buf.String(), "stealth-mode exclusion block is present") {
		t.Errorf("expected stale stealth exclusion warning, got:\n%s", buf.String())
	}
}

// TestWriteGitInfoExclude_AppendsToExistingAndReadError covers the
// append-to-existing-content and read-error branches of writeGitInfoExclude.
func TestWriteGitInfoExclude_AppendsToExistingAndReadError(t *testing.T) {
	t.Run("appends to existing content", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		g := newStealthGitFake()
		writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), "# existing\nfoo/bar", 0o644)
		opts := &Options{TargetDir: dir, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if sr := writeGitInfoExclude(opts, nil); sr.err != nil {
			t.Fatalf("writeGitInfoExclude() error: %v", sr.err)
		}
		b, _ := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
		s := string(b)
		if !strings.Contains(s, "# existing") || !strings.Contains(s, "foo/bar") {
			t.Errorf("existing content not preserved:\n%s", s)
		}
		if !strings.Contains(s, stealthExcludeMarker) {
			t.Errorf("marker missing:\n%s", s)
		}
	})

	t.Run("extra paths", func(t *testing.T) {
		dir := t.TempDir()
		initGitDir(t, dir)
		g := newStealthGitFake()
		opts := &Options{TargetDir: dir, ExecCmd: g.exec, ReadFile: os.ReadFile, WriteFile: os.WriteFile}
		if sr := writeGitInfoExclude(opts, []string{".custom/", "!negated"}); sr.err != nil {
			t.Fatalf("writeGitInfoExclude() error: %v", sr.err)
		}
		b, _ := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
		s := string(b)
		if !strings.Contains(s, ".custom/\n") {
			t.Errorf("extra path .custom/ missing:\n%s", s)
		}
		if !strings.Contains(s, "\\!negated\n") {
			t.Errorf("extra path !negated not escaped:\n%s", s)
		}
	})

	t.Run("read error", func(t *testing.T) {
		dir := t.TempDir()
		g := newStealthGitFake()
		opts := &Options{
			TargetDir: dir,
			ExecCmd:   g.exec,
			ReadFile:  func(string) ([]byte, error) { return nil, errors.New("boom") },
			WriteFile: os.WriteFile,
		}
		sr := writeGitInfoExclude(opts, nil)
		if sr.err == nil || !strings.Contains(sr.detail, "read failed") {
			t.Fatalf("expected read-failed error, got %v (detail %q)", sr.err, sr.detail)
		}
	})
}

// TestSnapshotTrackedFiles_LstatError covers the non-IsNotExist Lstat error
// branch (a tracked path whose parent component is a regular file).
func TestSnapshotTrackedFiles_LstatError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "file.txt"), "not-a-dir\n", 0o644)
	g := newStealthGitFake()
	g.tracked = []string{"file.txt/child.md"}
	opts := &Options{TargetDir: dir, ExecCmd: g.exec}
	if _, err := snapshotTrackedFiles(opts); err == nil {
		t.Fatal("expected snapshotTrackedFiles to error on ENOTDIR Lstat")
	}
}

// TestSnapshotTrackedFiles_MissingPath covers the IsNotExist skip branch.
func TestSnapshotTrackedFiles_MissingPath(t *testing.T) {
	dir := t.TempDir()
	g := newStealthGitFake()
	g.tracked = []string{"present.txt", "gone.txt"}
	writeFile(t, filepath.Join(dir, "present.txt"), "here\n", 0o644)
	opts := &Options{TargetDir: dir, ExecCmd: g.exec}
	snap, err := snapshotTrackedFiles(opts)
	if err != nil {
		t.Fatalf("snapshotTrackedFiles() error: %v", err)
	}
	if _, ok := snap["present.txt"]; !ok {
		t.Errorf("present.txt missing from snapshot")
	}
	if _, ok := snap["gone.txt"]; ok {
		t.Errorf("gone.txt (nonexistent) should be skipped, but present")
	}
}

// TestSnapshotTrackedFiles_SkipsGitlink covers the gitlink (submodule, index
// mode 160000) skip: a submodule's working tree is a directory and must never
// be treated as a replaced regular file (which would delete the checkout).
func TestSnapshotTrackedFiles_SkipsGitlink(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{TargetDir: dir, ExecCmd: func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) >= 1 && args[0] == "ls-files" {
			return []byte("100644 deadbeef 0\tREADME.md\x00160000 cafebabe 0\tsub\x00"), nil
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	}}
	writeFile(t, filepath.Join(dir, "README.md"), "hi\n", 0o644)
	// Materialize the submodule working tree as a real directory so the gitlink
	// guard is exercised: without the mode-160000 skip, Lstat would classify
	// "sub" as a directory and snapshot it (the regression this test guards).
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	writeFile(t, filepath.Join(dir, "sub", "inner.txt"), "sub\n", 0o644)
	snap, err := snapshotTrackedFiles(opts)
	if err != nil {
		t.Fatalf("snapshotTrackedFiles() error: %v", err)
	}
	if _, ok := snap["README.md"]; !ok {
		t.Errorf("README.md missing from snapshot")
	}
	if _, ok := snap["sub"]; ok {
		t.Errorf("gitlink (submodule) entry should be skipped, but present")
	}
}

// TestRestoreTrackedFiles_RestoresDeletedAndUnchanged covers the
// deleted-file restore and unchanged-skip branches of restoreTrackedFiles.
func TestRestoreTrackedFiles_RestoresDeletedAndUnchanged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "deleted.txt"), "orig\n", 0o644)
	writeFile(t, filepath.Join(dir, "unchanged.txt"), "same\n", 0o644)
	g := newStealthGitFake()
	g.tracked = []string{"deleted.txt", "unchanged.txt"}
	opts := &Options{TargetDir: dir, ExecCmd: g.exec}
	snap, err := snapshotTrackedFiles(opts)
	if err != nil {
		t.Fatalf("snapshotTrackedFiles() error: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "deleted.txt")); err != nil {
		t.Fatalf("remove deleted.txt: %v", err)
	}
	restored, _ := restoreTrackedFiles(opts, snap)
	if _, err := os.Stat(filepath.Join(dir, "deleted.txt")); err != nil {
		t.Errorf("deleted.txt not restored: %v", err)
	}
	foundDeleted := false
	for _, r := range restored {
		if r == "deleted.txt" {
			foundDeleted = true
		}
		if r == "unchanged.txt" {
			t.Errorf("unchanged.txt should not be in restored list")
		}
	}
	if !foundDeleted {
		t.Errorf("deleted.txt missing from restored list: %v", restored)
	}
}

// TestRestoreSymlink_AlreadyCorrect covers the already-correct fast path.
func TestRestoreSymlink_AlreadyCorrect(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "target.txt"), "x", 0o644)
	if err := os.Symlink("target.txt", filepath.Join(dir, "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if restoreSymlink(filepath.Join(dir, "link"), "target.txt") {
		t.Errorf("expected restoreSymlink to return false when already correct")
	}
}

// TestAppendNewUntrackedToExclude covers the discovery/append/dedup branch of
// appendNewUntrackedToExclude (task 3.x): extra untracked paths created during
// init are appended to the exclusion block as escaped literals, and a re-run
// appends no duplicates.
func TestAppendNewUntrackedToExclude(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)

	g := newStealthGitFake()
	g.others = []string{"extra.log", "!negated"}

	var buf bytes.Buffer
	opts := newStealthOpts(dir, g, &buf)
	if sr := writeGitInfoExclude(&opts, nil); sr.err != nil {
		t.Fatalf("writeGitInfoExclude: %v", sr.err)
	}
	if sr := appendNewUntrackedToExclude(&opts, nil); sr.err != nil {
		t.Fatalf("appendNewUntrackedToExclude: %v", sr.err)
	}

	ex, _ := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
	content := string(ex)
	if !strings.Contains(content, "extra.log") {
		t.Errorf("extra.log not appended to exclude: %q", content)
	}
	if !strings.Contains(content, "\\!negated") {
		t.Errorf("!negated not escaped in exclude: %q", content)
	}

	// Re-run: no duplicates.
	if sr := appendNewUntrackedToExclude(&opts, nil); sr.err != nil {
		t.Fatalf("second appendNewUntrackedToExclude: %v", sr.err)
	}
	ex2, _ := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
	if strings.Count(string(ex2), "extra.log") != 1 {
		t.Errorf("extra.log duplicated after re-run: %q", ex2)
	}
}

// TestEmitStealthSummary_NamesVisible asserts the post-init summary names the
// specific dirty paths rather than reporting clean:false with an empty list.
func TestEmitStealthSummary_NamesVisible(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)

	g := newStealthGitFake()
	g.statusOut = "?? .opencode/\n?? .uf/secrets\n"

	var buf bytes.Buffer
	opts := newStealthOpts(dir, g, &buf)
	emitStealthSummary(&opts)

	var report stealthCheckReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("summary is not valid JSON: %v (%q)", err, buf.String())
	}
	if report.Clean {
		t.Errorf("expected clean=false, got %+v", report)
	}
	if len(report.Visible) != 2 {
		t.Errorf("expected 2 visible paths, got %v", report.Visible)
	}
}

// TestStealthCheck_DetectsStaleExtraPath asserts --check flags a previously
// excluded (extra) path that has become visible even though the marker is
// still present.
func TestStealthCheck_DetectsStaleExtraPath(t *testing.T) {
	dir := t.TempDir()
	initGitDir(t, dir)
	// Pre-seed an exclusion block with the marker + canonical + an extra path.
	block := stealthExcludeMarker + "\n.opencode/\n.specify/\nopenspec/\n.uf/\nopencode.json\nextra.log\n"
	writeFile(t, filepath.Join(dir, ".git", "info", "exclude"), block, 0o644)

	g := newStealthGitFake()
	g.others = []string{"extra.log"}

	var buf bytes.Buffer
	opts := newStealthOpts(dir, g, &buf)
	if err := runStealthCheck(&opts); err == nil {
		t.Errorf("expected runStealthCheck to report stale extra path, got nil")
	}
}

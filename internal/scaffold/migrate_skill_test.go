package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newSkillMigrationOpts returns Options rooted at dir with real file
// I/O and a captured Stdout for migrateSkillDir tests.
func newSkillMigrationOpts(dir string) (*Options, *bytes.Buffer) {
	var buf bytes.Buffer
	return &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}, &buf
}

// assertFileContent fails the test unless relPath under dir holds
// exactly want.
func assertFileContent(t *testing.T, dir, relPath, want string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(dir, relPath))
	if err != nil {
		t.Fatalf("read %s: %v", relPath, err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", relPath, got, want)
	}
}

// assertAbsent fails the test if relPath under dir exists.
func assertAbsent(t *testing.T, dir, relPath string) {
	t.Helper()
	if _, err := os.Lstat(filepath.Join(dir, relPath)); !os.IsNotExist(err) {
		t.Errorf("%s should not exist (err=%v)", relPath, err)
	}
}

func TestMigrateSkillDir_NoOldDir(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/a/SKILL.md", "new")
	opts, _ := newSkillMigrationOpts(dir)

	if result := migrateSkillDir(opts); result != nil {
		t.Errorf("expected nil (no-op) when only skills/ exists, got %+v", result)
	}
}

func TestMigrateSkillDir_RenameOnly(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skill/alpha/SKILL.md", "alpha")
	createFile(t, dir, ".opencode/skill/alpha/references/notes.md", "notes")
	createFile(t, dir, ".opencode/skill/bravo/SKILL.md", "bravo")
	opts, _ := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || result.action != "migrated" {
		t.Fatalf("expected migrated result, got %+v", result)
	}
	if result.detail != "2 skills renamed" {
		t.Errorf("detail = %q, want %q", result.detail, "2 skills renamed")
	}
	assertFileContent(t, dir, ".opencode/skills/alpha/SKILL.md", "alpha")
	assertFileContent(t, dir, ".opencode/skills/alpha/references/notes.md", "notes")
	assertFileContent(t, dir, ".opencode/skills/bravo/SKILL.md", "bravo")
	assertAbsent(t, dir, ".opencode/skill")
}

func TestMigrateSkillDir_RenameFailureKeepsLegacyDir(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "a")
	// A read-only .opencode parent makes the atomic rename of the
	// whole legacy directory fail, exercising the rename-only
	// branch's failure path (merge-branch failures are covered by
	// the Move/Remove failure tests below).
	opencodeDir := filepath.Join(dir, ".opencode")
	if err := os.Chmod(opencodeDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(opencodeDir, 0o755) })
	opts, _ := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || result.action != "failed" {
		t.Fatalf("expected failed result, got %+v", result)
	}
	if !strings.HasPrefix(result.detail, "rename: ") {
		t.Errorf("detail = %q, want prefix %q", result.detail, "rename: ")
	}
	assertFileContent(t, dir, ".opencode/skill/a/SKILL.md", "a")
	assertAbsent(t, dir, ".opencode/skills")
}

func TestMigrateSkillDir_MergeMovesDropsAndKeepsConflicts(t *testing.T) {
	dir := t.TempDir()
	scaffoldSkill, err := assetContent("opencode/skills/speckit-workflow/SKILL.md")
	if err != nil {
		t.Fatalf("read embedded speckit-workflow skill: %v", err)
	}
	// Current scaffold copy under skills/, stale legacy copy under
	// skill/ (the #933 duplicate-skill warning).
	createFile(t, dir, ".opencode/skills/speckit-workflow/SKILL.md", string(scaffoldSkill))
	createFile(t, dir, ".opencode/skill/speckit-workflow/SKILL.md", "old scaffold version")
	// Byte-identical duplicate.
	createFile(t, dir, ".opencode/skills/same/SKILL.md", "same")
	createFile(t, dir, ".opencode/skill/same/SKILL.md", "same")
	// Legacy-only skill and legacy-only file inside a shared skill.
	createFile(t, dir, ".opencode/skill/legacy-only/SKILL.md", "legacy")
	createFile(t, dir, ".opencode/skills/shared/SKILL.md", "shared")
	createFile(t, dir, ".opencode/skill/shared/extra.md", "extra")
	// User-owned skill modified in both places: never clobbered.
	createFile(t, dir, ".opencode/skills/custom/SKILL.md", "custom new")
	createFile(t, dir, ".opencode/skill/custom/SKILL.md", "custom edited")
	opts, buf := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || result.action != "migrated" {
		t.Fatalf("expected migrated result, got %+v", result)
	}
	wantDetail := "moved 2, removed 1 duplicates and 1 stale scaffold copies, 1 conflicts kept, 0 warnings"
	if result.detail != wantDetail {
		t.Errorf("detail = %q, want %q", result.detail, wantDetail)
	}
	assertFileContent(t, dir, ".opencode/skills/speckit-workflow/SKILL.md", string(scaffoldSkill))
	assertAbsent(t, dir, ".opencode/skill/speckit-workflow")
	assertFileContent(t, dir, ".opencode/skills/same/SKILL.md", "same")
	assertAbsent(t, dir, ".opencode/skill/same")
	assertFileContent(t, dir, ".opencode/skills/legacy-only/SKILL.md", "legacy")
	assertFileContent(t, dir, ".opencode/skills/shared/extra.md", "extra")
	assertAbsent(t, dir, ".opencode/skill/shared")
	// Conflict: both copies survive untouched, with a warning.
	assertFileContent(t, dir, ".opencode/skills/custom/SKILL.md", "custom new")
	assertFileContent(t, dir, ".opencode/skill/custom/SKILL.md", "custom edited")
	if !strings.Contains(buf.String(), "skill/custom/SKILL.md: conflict — kept both copies") {
		t.Errorf("expected conflict warning for custom skill, got:\n%s", buf.String())
	}
}

func TestMigrateSkillDir_MergeRemovesEmptiedLegacyDir(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/a/SKILL.md", "a")
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "a")
	createFile(t, dir, ".opencode/skill/b/nested/deep.md", "deep")
	opts, _ := newSkillMigrationOpts(dir)

	if result := migrateSkillDir(opts); result == nil {
		t.Fatal("expected non-nil result")
	}
	assertFileContent(t, dir, ".opencode/skills/b/nested/deep.md", "deep")
	assertAbsent(t, dir, ".opencode/skill")

	// Idempotent: a second run is a silent no-op.
	if again := migrateSkillDir(opts); again != nil {
		t.Errorf("second run should be a no-op, got %+v", again)
	}
}

func TestMigrateSkillDir_ReadFailureKeepsLegacyFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/a/SKILL.md", "new")
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "old")
	opts, buf := newSkillMigrationOpts(dir)
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.Contains(path, filepath.Join(".opencode", "skill", "a")) {
			return nil, os.ErrPermission
		}
		return os.ReadFile(path)
	}

	result := migrateSkillDir(opts)

	if result == nil || !strings.HasSuffix(result.detail, "1 warnings") {
		t.Fatalf("expected one warning, got %+v", result)
	}
	assertFileContent(t, dir, ".opencode/skill/a/SKILL.md", "old")
	if !strings.Contains(buf.String(), "skill/a/SKILL.md: read failed") {
		t.Errorf("expected read-failure warning, got:\n%s", buf.String())
	}
}

func TestMigrateSkillDir_MkdirFailureKeepsLegacyFile(t *testing.T) {
	dir := t.TempDir()
	// skills/blocker is a file where the merge needs a directory, so
	// MkdirAll for the nested destination fails.
	createFile(t, dir, ".opencode/skills/keeper/SKILL.md", "keeper")
	createFile(t, dir, ".opencode/skills/blocker", "in the way")
	createFile(t, dir, ".opencode/skill/blocker/SKILL.md", "legacy")
	opts, buf := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || !strings.HasSuffix(result.detail, "1 warnings") {
		t.Fatalf("expected one warning, got %+v", result)
	}
	if !strings.Contains(buf.String(), "skill/blocker/SKILL.md: move failed") {
		t.Errorf("expected move-failure warning, got:\n%s", buf.String())
	}
	assertFileContent(t, dir, ".opencode/skill/blocker/SKILL.md", "legacy")
}

func TestMigrateSkillDir_MoveFailureKeepsLegacyFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "legacy")
	newDir := filepath.Join(dir, ".opencode", "skills", "a")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// A read-only destination directory makes both the rename and the
	// write fallback in moveFile fail, exercising the move-failure path.
	if err := os.Chmod(newDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(newDir, 0o755) })
	opts, buf := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || !strings.HasSuffix(result.detail, "1 warnings") {
		t.Fatalf("expected one warning, got %+v", result)
	}
	if !strings.Contains(buf.String(), "skill/a/SKILL.md: move failed") {
		t.Errorf("expected move-failure warning, got:\n%s", buf.String())
	}
	assertFileContent(t, dir, ".opencode/skill/a/SKILL.md", "legacy")
	assertAbsent(t, dir, ".opencode/skills/a/SKILL.md")
}

func TestMigrateSkillDir_NewPathReadFailureKeepsLegacyFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/a/SKILL.md", "new")
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "old")
	opts, buf := newSkillMigrationOpts(dir)
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.Contains(path, filepath.Join(".opencode", "skills", "a")) {
			return nil, os.ErrPermission
		}
		return os.ReadFile(path)
	}

	result := migrateSkillDir(opts)

	if result == nil || !strings.HasSuffix(result.detail, "1 warnings") {
		t.Fatalf("expected one warning, got %+v", result)
	}
	assertFileContent(t, dir, ".opencode/skill/a/SKILL.md", "old")
	if !strings.Contains(buf.String(), "skill/a/SKILL.md: read failed") {
		t.Errorf("expected read-failure warning, got:\n%s", buf.String())
	}
}

func TestMigrateSkillDir_RemoveFailureKeepsLegacyFileAndWarns(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/a/SKILL.md", "same")
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "same")
	legacyDir := filepath.Join(dir, ".opencode", "skill", "a")
	// A read-only legacy directory makes removing the duplicate fail,
	// so the file must be neither counted as removed nor dropped
	// silently.
	if err := os.Chmod(legacyDir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(legacyDir, 0o755) })
	opts, buf := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || !strings.HasSuffix(result.detail, "1 warnings") {
		t.Fatalf("expected one warning, got %+v", result)
	}
	if !strings.Contains(result.detail, "removed 0 duplicates and 0 stale") {
		t.Errorf("failed remove must not count as removed: %q", result.detail)
	}
	if !strings.Contains(buf.String(), "skill/a/SKILL.md: remove failed") {
		t.Errorf("expected remove-failure warning, got:\n%s", buf.String())
	}
	assertFileContent(t, dir, ".opencode/skill/a/SKILL.md", "same")
}

func TestMigrateSkillDir_SymlinkedDirSkipped(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "elsewhere/a/SKILL.md", "a")
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(filepath.Join(dir, "elsewhere"), filepath.Join(dir, ".opencode", "skill")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	opts, _ := newSkillMigrationOpts(dir)

	result := migrateSkillDir(opts)

	if result == nil || result.action != "skipped" {
		t.Fatalf("expected skipped result for symlinked dir, got %+v", result)
	}
	assertFileContent(t, dir, "elsewhere/a/SKILL.md", "a")
}

func TestMigrateSkillDir_SymlinkedFileLeftInPlace(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/a/SKILL.md", "a")
	createFile(t, dir, "target.md", "t")
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "skill", "b"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	link := filepath.Join(dir, ".opencode", "skill", "b", "SKILL.md")
	if err := os.Symlink(filepath.Join(dir, "target.md"), link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	opts, buf := newSkillMigrationOpts(dir)

	if result := migrateSkillDir(opts); result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, err := os.Lstat(link); err != nil {
		t.Errorf("symlinked legacy file should stay in place: %v", err)
	}
	if !strings.Contains(buf.String(), "symlink — manual migration required") {
		t.Errorf("expected symlink warning, got:\n%s", buf.String())
	}
}

func TestMigrateSkillDir_DivisorOnlySkips(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skill/a/SKILL.md", "a")
	opts, _ := newSkillMigrationOpts(dir)
	opts.DivisorOnly = true

	if result := migrateSkillDir(opts); result != nil {
		t.Errorf("expected nil in DivisorOnly mode, got %+v", result)
	}
	assertFileContent(t, dir, ".opencode/skill/a/SKILL.md", "a")
}

func TestMigrateSkillDir_RerunWithKeptConflictReportsSkipped(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skills/custom/SKILL.md", "new")
	createFile(t, dir, ".opencode/skill/custom/SKILL.md", "edited")
	opts, _ := newSkillMigrationOpts(dir)

	first := migrateSkillDir(opts)
	second := migrateSkillDir(opts)

	for i, result := range []*subToolResult{first, second} {
		if result == nil || result.action != "skipped" {
			t.Errorf("run %d: expected skipped (nothing changed), got %+v", i+1, result)
		}
	}
	assertFileContent(t, dir, ".opencode/skill/custom/SKILL.md", "edited")
}

// TestRun_MigratesLegacyDirsExceptStealth runs the full init and
// checks the legacy command and skill directories are migrated in
// normal mode and left untouched in stealth mode, which runs after
// the stealth restore and must not reshape the tree.
func TestRun_MigratesLegacyDirsExceptStealth(t *testing.T) {
	for _, stealth := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "stealth"}[stealth], func(t *testing.T) {
			dir := t.TempDir()
			initGitDir(t, dir)
			createFile(t, dir, ".opencode/command/legacy.md", "legacy command")
			createFile(t, dir, ".opencode/skill/legacy-only/SKILL.md", "legacy")
			createFile(t, dir, ".opencode/skill/speckit-workflow/SKILL.md", "old scaffold version")

			var buf bytes.Buffer
			// Stealth refuses tracked .opencode/ paths up front, so the
			// legacy files are untracked here; stealth still must not
			// move them.
			g := newStealthGitFake()
			opts := newStealthOpts(dir, g, &buf)
			opts.Stealth = stealth
			if _, err := Run(opts); err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			if stealth {
				assertFileContent(t, dir, ".opencode/command/legacy.md", "legacy command")
				assertFileContent(t, dir, ".opencode/skill/legacy-only/SKILL.md", "legacy")
				assertFileContent(t, dir, ".opencode/skill/speckit-workflow/SKILL.md", "old scaffold version")
				return
			}
			// The scaffold writes .opencode/commands/ before the
			// migration runs, so the legacy command file merges into
			// the scaffolded directory.
			assertFileContent(t, dir, ".opencode/commands/legacy.md", "legacy command")
			assertAbsent(t, dir, ".opencode/command")
			assertFileContent(t, dir, ".opencode/skills/legacy-only/SKILL.md", "legacy")
			assertAbsent(t, dir, ".opencode/skill")
			for _, line := range []string{
				".opencode/command/ → commands/",
				".opencode/skill/ → skills/",
			} {
				if !strings.Contains(buf.String(), line) {
					t.Errorf("expected migration summary line %q, got:\n%s", line, buf.String())
				}
			}
		})
	}
}

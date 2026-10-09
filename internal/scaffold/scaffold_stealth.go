package scaffold

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// stealthExcludeMarker is the sentinel comment that delimits the block of
// exclusion entries appended to .git/info/exclude by stealth mode. It is
// used for idempotency detection and scoped rollback (only the block
// appended by the current invocation is removed on failure).
const stealthExcludeMarker = "# Unbound Force — stealth exclusion (managed by uf init)"

// gitlinkMode is the index mode git reports for submodule (gitlink) entries.
// Snapshotting must skip these: their working tree is a directory, and
// treating one as a replaced regular file would delete the submodule checkout.
const gitlinkMode = "160000"

// stealthExcludePaths is the canonical set of top-level paths scaffolded by
// uf init that stealth mode must hide from git. Sub-tools may create
// additional untracked paths; those are discovered and appended after init.
var stealthExcludePaths = []string{
	".opencode/",
	".specify/",
	"openspec/",
	".uf/",
	"opencode.json",
}

// trackedFileState captures the pre-init state of a tracked file so it can
// be restored if a sub-tool mutates it during stealth init.
type trackedFileState struct {
	mode       os.FileMode
	bytes      []byte
	isSymlink  bool
	linkTarget string
}

// stealthContext carries the rollback state for a single stealth init. Its
// rollback method is invoked on failure paths and disarmed on success.
type stealthContext struct {
	excludePath     string
	preExclude      []byte
	preExcludeExist bool
	snapshot        map[string]trackedFileState
	didSnapshot     bool
	preStaged       []string
	preExisting     map[string]bool
	preUntracked    map[string]bool
}

// stealthCheckReport is the machine-parseable JSON payload emitted by
// --check and by the post-init stealth summary. It carries Constitution III
// provenance fields (tool, version, timestamp, ref, commit).
type stealthCheckReport struct {
	Tool      string   `json:"tool"`
	Version   string   `json:"version"`
	Timestamp string   `json:"timestamp"`
	Ref       string   `json:"ref,omitempty"`
	Commit    string   `json:"commit,omitempty"`
	Clean     bool     `json:"clean"`
	Visible   []string `json:"visible"`
}

// escapeExcludePattern emits a path as a literal gitignore pattern by
// backslash-escaping the gitignore metacharacters '\', '*', '?', '[' and a
// leading '#' (comment) or '!' (negation) so the path is matched literally.
func escapeExcludePattern(p string) string {
	var b strings.Builder
	for i, r := range p {
		switch r {
		case '#', '!':
			if i == 0 {
				b.WriteByte('\\')
			}
			b.WriteRune(r)
		case '\\', '*', '?', '[':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// resolveGitInfoExcludePath returns the absolute path to the repository's
// .git/info/exclude file, resolved through git so linked worktrees and
// submodules (where .git is a file) are handled correctly.
func resolveGitInfoExcludePath(opts *Options) (string, error) {
	out, err := opts.ExecCmd("git", "rev-parse", "--git-path", "info/exclude")
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", errors.New("git returned an empty path for info/exclude")
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(opts.TargetDir, p)
	}
	return p, nil
}

// verifyInsideGitRepo confirms the target directory is inside a git work
// tree. It distinguishes "git not installed / not on PATH" (a more
// actionable error) from "not inside a work tree".
func verifyInsideGitRepo(opts *Options) error {
	out, err := opts.ExecCmd("git", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 127 {
			return errors.New("git is not installed or not on PATH; stealth mode requires git")
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return errors.New("git is not installed or not on PATH; stealth mode requires git")
		}
		return errors.New("stealth mode requires a git repository (run `git init` first)")
	}
	if strings.TrimSpace(string(out)) != "true" {
		return errors.New("stealth mode requires a git repository (run `git init` first)")
	}
	return nil
}

// isCanonicalScaffoldPath reports whether rel is one of the canonical
// top-level scaffolded paths (or a path beneath one).
func isCanonicalScaffoldPath(rel string) bool {
	return matchesScaffoldPath(rel, stealthExcludePaths)
}

// trackedScaffoldPaths returns the subset of the canonical scaffolded paths
// that are already tracked by git. Stealth mode refuses to run when any are
// tracked, because .git/info/exclude cannot hide tracked files.
func trackedScaffoldPaths(opts *Options) ([]string, error) {
	out, err := opts.ExecCmd("git", "ls-files")
	if err != nil {
		return nil, err
	}
	var found []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isCanonicalScaffoldPath(line) {
			found = append(found, line)
		}
	}
	return found, nil
}

// writeGitInfoExclude appends the stealth exclusion block (marker + escaped
// canonical paths + any extra paths) to .git/info/exclude. Idempotent: if
// the marker is already present, it is a no-op.
func writeGitInfoExclude(opts *Options, extra []string) subToolResult {
	p, err := resolveGitInfoExcludePath(opts)
	if err != nil {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: err.Error(), err: err}
	}
	existing, readErr := opts.ReadFile(p)
	if readErr != nil && !os.IsNotExist(readErr) {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: fmt.Sprintf("read failed: %v", readErr), err: readErr}
	}
	if readErr == nil && strings.Contains(string(existing), stealthExcludeMarker) {
		return subToolResult{name: ".git/info/exclude", action: "already configured"}
	}

	var b strings.Builder
	if readErr == nil {
		content := string(existing)
		b.WriteString(content)
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			b.WriteString("\n")
		}
		if len(strings.TrimSpace(content)) > 0 {
			b.WriteString("\n")
		}
	}
	b.WriteString(stealthExcludeMarker)
	b.WriteString("\n")
	for _, pp := range stealthExcludePaths {
		b.WriteString(escapeExcludePattern(pp))
		b.WriteString("\n")
	}
	for _, pp := range extra {
		b.WriteString(escapeExcludePattern(pp))
		b.WriteString("\n")
	}
	if writeErr := opts.WriteFile(p, []byte(b.String()), 0o644); writeErr != nil {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: fmt.Sprintf("write failed: %v", writeErr), err: writeErr}
	}
	return subToolResult{name: ".git/info/exclude", action: "configured"}
}

// beginStealth runs the stealth pre-flight: git-repo verification, the
// already-tracked refusal, and the exclusion write (before any scaffolding
// so a visibility window is never opened). It captures the pre-init
// exclusion-file state for scoped rollback.
func beginStealth(opts *Options) (*stealthContext, error) {
	if err := verifyInsideGitRepo(opts); err != nil {
		return nil, err
	}
	tracked, err := trackedScaffoldPaths(opts)
	if err != nil {
		return nil, fmt.Errorf("list tracked scaffolded paths: %w", err)
	}
	if len(tracked) > 0 {
		return nil, fmt.Errorf("refusing stealth init: these scaffolded paths are already tracked by git: %s", strings.Join(tracked, ", "))
	}

	ctx := &stealthContext{}
	ctx.preExisting = make(map[string]bool)
	for _, p := range stealthExcludePaths {
		if _, err := os.Lstat(filepath.Join(opts.TargetDir, strings.TrimSuffix(p, "/"))); err == nil {
			ctx.preExisting[p] = true
		}
	}
	stagedOut, err := opts.ExecCmd("git", "diff", "--cached", "--name-only", "-z")
	if err != nil {
		return nil, fmt.Errorf("capture staged files: %w", err)
	}
	for _, p := range strings.Split(string(stagedOut), "\x00") {
		if p != "" {
			ctx.preStaged = append(ctx.preStaged, p)
		}
	}

	untrackedOut, err := opts.ExecCmd("git", "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, fmt.Errorf("capture untracked files: %w", err)
	}
	ctx.preUntracked = make(map[string]bool)
	for _, p := range strings.Split(string(untrackedOut), "\x00") {
		if p != "" {
			ctx.preUntracked[p] = true
		}
	}
	if p, err := resolveGitInfoExcludePath(opts); err == nil {
		ctx.excludePath = p
		if b, rerr := opts.ReadFile(p); rerr == nil {
			ctx.preExclude = b
			ctx.preExcludeExist = true
		}
	}
	if sr := writeGitInfoExclude(opts, nil); sr.err != nil {
		return nil, fmt.Errorf("write .git/info/exclude: %w", sr.err)
	}
	return ctx, nil
}

// rollback removes scaffolded files created by this init, restores the
// exclusion file to its pre-init state, and reverts any tracked-file
// mutations. It returns an error naming residual dirty paths when any
// rollback step fails.
func (c *stealthContext) rollback(opts *Options) error {
	var residual []string
	for _, p := range stealthExcludePaths {
		if c.preExisting[p] {
			continue
		}
		if err := os.RemoveAll(filepath.Join(opts.TargetDir, strings.TrimSuffix(p, "/"))); err != nil {
			residual = append(residual, p)
		}
	}
	// Restore the exclusion file to its pre-init state BEFORE discovering
	// untracked files, so scaffolded files written into a pre-existing
	// canonical directory (e.g. a prior uf init, or the user's own .opencode/)
	// become visible again and are removed by the sweep below.
	if c.excludePath != "" {
		if c.preExcludeExist {
			if err := opts.WriteFile(c.excludePath, c.preExclude, 0o644); err != nil {
				residual = append(residual, c.excludePath)
			}
		} else if err := os.Remove(c.excludePath); err != nil && !os.IsNotExist(err) {
			residual = append(residual, c.excludePath)
		}
	}
	// Remove any untracked files created during init that were not present
	// before (the exclusion block is now restored, so scaffolded files under
	// pre-existing canonical directories surface here too).
	if out, err := opts.ExecCmd("git", "ls-files", "--others", "--exclude-standard", "-z"); err == nil {
		for _, p := range strings.Split(string(out), "\x00") {
			if p == "" || c.preUntracked[p] {
				continue
			}
			// Only remove files this init scaffolded (canonical paths). Any
			// other untracked file created during the init window belongs to
			// the user or a concurrent process, so leave it in place and
			// report it as residual rather than deleting unrelated data.
			if !matchesScaffoldPath(p, stealthExcludePaths) {
				residual = append(residual, p)
				continue
			}
			if err := os.RemoveAll(filepath.Join(opts.TargetDir, p)); err != nil {
				residual = append(residual, p)
			}
		}
	}
	if c.didSnapshot {
		_, failed := restoreTrackedFiles(opts, c.snapshot)
		residual = append(residual, failed...)
	}
	if err := unstageNewlyStaged(opts, c.preStaged); err != nil {
		residual = append(residual, err.Error())
	}
	if len(residual) > 0 {
		return fmt.Errorf("rollback incomplete; residual dirty paths: %s", strings.Join(residual, ", "))
	}
	return nil
}

// unstageNewlyStaged un-stages any path staged during this init while
// preserving pre-existing staged changes. It never performs a blanket
// git reset, which could clobber unrelated user staging.
func unstageNewlyStaged(opts *Options, preStaged []string) error {
	out, err := opts.ExecCmd("git", "diff", "--cached", "--name-only", "-z")
	if err != nil {
		return fmt.Errorf("list staged files: %w", err)
	}
	staged := map[string]bool{}
	for _, p := range strings.Split(string(out), "\x00") {
		if p != "" {
			staged[p] = true
		}
	}
	pre := map[string]bool{}
	for _, p := range preStaged {
		pre[p] = true
	}
	var toUnstage []string
	for p := range staged {
		if !pre[p] {
			toUnstage = append(toUnstage, p)
		}
	}
	if len(toUnstage) > 0 {
		args := append([]string{"restore", "--staged", "--"}, toUnstage...)
		if _, err := opts.ExecCmd("git", args...); err != nil {
			return fmt.Errorf("unstage %s: %w", strings.Join(toUnstage, ", "), err)
		}
	}
	return nil
}

// snapshotTrackedFiles records the bytes and mode of every tracked file
// (using real os funcs, not opts.ReadFile, so tests can fake only ExecCmd).
// Symlinks are captured as symlinks with their link target. Gitlink entries
// (submodules, index mode 160000) are skipped: their working tree is a
// directory, and treating it as a replaced regular file would destroy it.
func snapshotTrackedFiles(opts *Options) (map[string]trackedFileState, error) {
	out, err := opts.ExecCmd("git", "ls-files", "-s", "-z")
	if err != nil {
		return nil, err
	}
	snap := make(map[string]trackedFileState)
	for _, entry := range strings.Split(string(out), "\x00") {
		if entry == "" {
			continue
		}
		// `git ls-files -s -z` emits "<mode> <object> <stage>\t<path>" per entry.
		meta, rel, ok := strings.Cut(entry, "\t")
		if !ok || rel == "" {
			continue
		}
		if mode := strings.Fields(meta); len(mode) > 0 && mode[0] == gitlinkMode {
			// Submodule gitlink entry; skip to avoid deleting its checkout.
			continue
		}
		full := filepath.Join(opts.TargetDir, rel)
		fi, err := os.Lstat(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		st := trackedFileState{mode: fi.Mode()}
		if fi.Mode()&os.ModeSymlink != 0 {
			st.isSymlink = true
			if target, err := os.Readlink(full); err == nil {
				st.linkTarget = target
			}
		} else if fi.Mode().IsRegular() {
			b, err := os.ReadFile(full)
			if err != nil {
				return nil, err
			}
			st.bytes = b
		}
		snap[rel] = st
	}
	return snap, nil
}

// restoreTrackedFiles reverts any tracked file that was mutated during init
// to its snapshot state (bytes and mode; symlinks as symlinks). It returns
// the list of restored paths and the list of paths it failed to restore.
func restoreTrackedFiles(opts *Options, snap map[string]trackedFileState) (restored []string, failed []string) {
	for rel, st := range snap {
		full := filepath.Join(opts.TargetDir, rel)
		if st.isSymlink {
			if restoreSymlink(full, st.linkTarget) {
				restored = append(restored, rel)
				continue
			}
			fi, lerr := os.Lstat(full)
			if lerr != nil || fi.Mode()&os.ModeSymlink == 0 {
				failed = append(failed, rel)
			} else if cur, rerr := os.Readlink(full); rerr != nil || cur != st.linkTarget {
				failed = append(failed, rel)
			}
			continue
		}
		// Lstat (not Stat) so a symlink swapped in by a sub-tool is detected,
		// and we never read/write through it to an unintended target.
		curFi, err := os.Lstat(full)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted during init; restore it from the snapshot.
				if werr := restoreRegularFile(full, st); werr == nil {
					restored = append(restored, rel)
				} else {
					failed = append(failed, rel)
				}
			} else {
				failed = append(failed, rel)
			}
			continue
		}
		if !curFi.Mode().IsRegular() {
			// A sub-tool replaced this tracked file with a symlink or
			// directory. Remove it and recreate the original regular file.
			_ = os.RemoveAll(full)
			if werr := restoreRegularFile(full, st); werr == nil {
				restored = append(restored, rel)
			} else {
				failed = append(failed, rel)
			}
			continue
		}
		cur, err := os.ReadFile(full)
		if err != nil {
			failed = append(failed, rel)
			continue
		}
		if string(cur) == string(st.bytes) && curFi.Mode() == st.mode {
			continue
		}
		if err := restoreRegularFile(full, st); err != nil {
			failed = append(failed, rel)
			continue
		}
		restored = append(restored, rel)
	}
	return restored, failed
}

// restoreRegularFile writes the snapshot bytes and re-applies the snapshot
// mode (including setuid/setgid/sticky bits) to a regular file. It returns
// an error if either the write or the chmod fails so the caller can report
// the path as residual rather than silently leaving permission bits wrong.
func restoreRegularFile(full string, st trackedFileState) error {
	if err := os.WriteFile(full, st.bytes, st.mode); err != nil {
		return err
	}
	return os.Chmod(full, st.mode)
}

// restoreSymlink ensures full is a symlink pointing at target, recreating it
// if it was replaced by a regular file or its target changed.
func restoreSymlink(full, target string) bool {
	fi, err := os.Lstat(full)
	if err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			if cur, err := os.Readlink(full); err == nil && cur == target {
				return false
			}
		}
		_ = os.Remove(full)
	}
	return os.Symlink(target, full) == nil
}

// appendNewUntrackedToExclude discovers untracked paths created during init
// (beyond the canonical set) and appends them to the exclusion block so the
// tree stays git-clean. Returns an empty result when there is nothing to add.
func appendNewUntrackedToExclude(opts *Options, preUntracked map[string]bool) subToolResult {
	out, err := opts.ExecCmd("git", "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: fmt.Sprintf("discover untracked: %v", err), err: err}
	}
	var extra []string
	for _, p := range strings.Split(string(out), "\x00") {
		if p != "" && !preUntracked[p] {
			extra = append(extra, p)
		}
	}
	if len(extra) == 0 {
		return subToolResult{}
	}

	p, err := resolveGitInfoExcludePath(opts)
	if err != nil {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: err.Error(), err: err}
	}
	existing, readErr := opts.ReadFile(p)
	if readErr != nil {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: fmt.Sprintf("read failed: %v", readErr), err: readErr}
	}
	content := string(existing)
	var b strings.Builder
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteString("\n")
	}
	added := 0
	for _, pp := range extra {
		escaped := escapeExcludePattern(pp)
		if strings.Contains(content, "\n"+escaped+"\n") || strings.HasSuffix(content, "\n"+escaped) {
			continue
		}
		b.WriteString(escaped)
		b.WriteString("\n")
		added++
	}
	if added == 0 {
		return subToolResult{}
	}
	if writeErr := opts.WriteFile(p, []byte(b.String()), 0o644); writeErr != nil {
		return subToolResult{name: ".git/info/exclude", action: "failed", detail: fmt.Sprintf("write failed: %v", writeErr), err: writeErr}
	}
	return subToolResult{name: ".git/info/exclude", action: "configured"}
}

// stealthExcludedPaths returns the union of the canonical scaffolded paths
// and any extra paths recorded in the exclusion block. It is used to detect
// a stale (partially-edited) exclusion block where a previously-excluded
// path has become visible.
func stealthExcludedPaths(opts *Options) ([]string, error) {
	set := map[string]bool{}
	for _, p := range stealthExcludePaths {
		set[p] = true
	}
	p, err := resolveGitInfoExcludePath(opts)
	if err != nil {
		return nil, err
	}
	existing, err := opts.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return stealthExcludePaths, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(existing), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		set[strings.TrimPrefix(line, "\\")] = true
	}
	paths := make([]string, 0, len(set))
	for p := range set {
		paths = append(paths, p)
	}
	return paths, nil
}

// matchesScaffoldPath reports whether p is equal to, or beneath, one of the
// excluded scaffolded paths.
func matchesScaffoldPath(p string, excluded []string) bool {
	for _, e := range excluded {
		prefix := strings.TrimSuffix(e, "/")
		if p == prefix || strings.HasPrefix(p, e) {
			return true
		}
	}
	return false
}

// visibleScaffoldPaths returns the scaffolded paths (canonical and any extra
// recorded in the exclusion block) that are currently visible to git
// (untracked and not excluded). An empty result with an intact exclusion
// marker means the tree is clean.
func visibleScaffoldPaths(opts *Options) ([]string, error) {
	excluded, err := stealthExcludedPaths(opts)
	if err != nil {
		return nil, err
	}
	out, err := opts.ExecCmd("git", "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	var visible []string
	for _, p := range strings.Split(string(out), "\x00") {
		if p == "" {
			continue
		}
		if matchesScaffoldPath(p, excluded) {
			visible = append(visible, p)
		}
	}
	return visible, nil
}

// gitRefCommit returns the current branch name and HEAD commit (best-effort).
func gitRefCommit(opts *Options) (ref, commit string) {
	if out, err := opts.ExecCmd("git", "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		ref = strings.TrimSpace(string(out))
	}
	if out, err := opts.ExecCmd("git", "rev-parse", "HEAD"); err == nil {
		commit = strings.TrimSpace(string(out))
	}
	return ref, commit
}

// newStealthCheckReport builds the JSON payload shared by --check and the
// post-init summary.
func newStealthCheckReport(opts *Options, clean bool, visible []string) stealthCheckReport {
	ref, commit := gitRefCommit(opts)
	return stealthCheckReport{
		Tool:      "unbound-force",
		Version:   opts.Version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Ref:       ref,
		Commit:    commit,
		Clean:     clean,
		Visible:   visible,
	}
}

// runStealthCheck verifies the stealth invariant without writing anything.
// It emits a JSON report and returns a non-nil error (non-zero exit) when
// the tree is not clean or the exclusion marker is missing/stale.
func runStealthCheck(opts *Options) error {
	if err := verifyInsideGitRepo(opts); err != nil {
		return err
	}
	p, err := resolveGitInfoExcludePath(opts)
	if err != nil {
		return err
	}
	existing, err := opts.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	markerMissing := !strings.Contains(string(existing), stealthExcludeMarker)
	visible, err := visibleScaffoldPaths(opts)
	if err != nil {
		return err
	}
	clean := !markerMissing && len(visible) == 0

	if err := writeStealthReport(opts, newStealthCheckReport(opts, clean, visible)); err != nil {
		return err
	}

	if !clean {
		return fmt.Errorf("stealth exclusion is not intact (missing marker: %v, visible scaffolded paths: %d)", markerMissing, len(visible))
	}
	return nil
}

// writeStealthReport marshals a stealth report to JSON and writes it to
// opts.Stdout, shared by --check and the post-init summary.
func writeStealthReport(opts *Options, r stealthCheckReport) error {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	_, err = fmt.Fprintln(opts.Stdout, string(b))
	return err
}

// emitStealthSummary writes the machine-parseable cleanliness report after a
// successful stealth init. The invariant holds when git status --porcelain
// reports no output.
func emitStealthSummary(opts *Options) {
	excluded, exErr := stealthExcludedPaths(opts)
	out, err := opts.ExecCmd("git", "status", "--porcelain")
	var visible []string
	rawEmpty := true
	if err == nil {
		rawEmpty = len(strings.TrimSpace(string(out))) == 0
		if exErr == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || len(line) < 4 {
					continue
				}
				path := strings.TrimSpace(line[3:])
				if idx := strings.Index(path, " -> "); idx >= 0 {
					path = strings.TrimSpace(path[idx+4:])
				}
				if path != "" && matchesScaffoldPath(path, excluded) {
					visible = append(visible, path)
				}
			}
		}
	}
	clean := err == nil && rawEmpty
	_ = writeStealthReport(opts, newStealthCheckReport(opts, clean, visible))
}

// warnStaleStealthExclusion warns on a normal (non-stealth) init when a
// stealth-mode exclusion block is present in .git/info/exclude, since that
// block would silently keep scaffolded files out of git.
func warnStaleStealthExclusion(opts *Options) {
	if opts.ExecCmd == nil {
		return
	}
	if _, err := opts.LookPath("git"); err != nil {
		return
	}
	p, err := resolveGitInfoExcludePath(opts)
	if err != nil {
		return
	}
	existing, err := opts.ReadFile(p)
	if err != nil {
		return
	}
	if strings.Contains(string(existing), stealthExcludeMarker) {
		_, _ = fmt.Fprintln(opts.Stdout, "  warning: a stealth-mode exclusion block is present in .git/info/exclude; scaffolded files may be hidden from git. Remove the marker-delimited block to commit them.")
	}
}

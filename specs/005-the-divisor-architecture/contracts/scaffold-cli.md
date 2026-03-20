# Contract: Scaffold CLI (unbound init)

**Spec**: 005-the-divisor-architecture
**Date**: 2026-03-19
**Type**: CLI command schema

## Command: `unbound init`

Scaffolds the Unbound Force specification framework and
Divisor review agents into the current directory.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Overwrite all existing files regardless of ownership |
| `--divisor` | bool | false | Deploy only Divisor agents, command, and convention packs |
| `--lang` | string | "" | Project language for convention pack (auto-detected if omitted) |

### Valid `--lang` Values

| Value | Convention Pack | Detection Marker |
|-------|----------------|-----------------|
| `go` | `go.md` | `go.mod` |
| `typescript` | `typescript.md` | `tsconfig.json` or `package.json` |
| `python` | `python.md` (future) | `pyproject.toml` |
| `rust` | `rust.md` (future) | `Cargo.toml` |
| (empty) | auto-detected or `default.md` | N/A |

### Behavior Matrix

| Flags | Behavior |
|-------|----------|
| `unbound init` | Deploy all 45 scaffold files. Language auto-detected for convention pack selection. |
| `unbound init --divisor` | Deploy only Divisor subset (~10-12 files). Language auto-detected. Skip speckit, openspec, non-divisor agents. |
| `unbound init --divisor --lang go` | Deploy Divisor subset with Go convention pack (skip typescript, etc.). |
| `unbound init --divisor --lang typescript` | Deploy Divisor subset with TypeScript convention pack. |
| `unbound init --force` | Deploy all files, overwriting existing regardless of ownership. |
| `unbound init --divisor --force` | Deploy Divisor subset, overwriting existing Divisor files. |

### File Deployment (Divisor Subset)

When `--divisor` is set, only these files are deployed:

**Agents** (user-owned):
- `.opencode/agents/divisor-guard.md`
- `.opencode/agents/divisor-architect.md`
- `.opencode/agents/divisor-adversary.md`
- `.opencode/agents/divisor-sre.md`
- `.opencode/agents/divisor-testing.md`

**Command** (tool-owned):
- `.opencode/command/review-council.md`

**Convention Packs** (tool-owned canonical + user-owned custom):
- `.opencode/divisor/packs/{lang}.md`
- `.opencode/divisor/packs/{lang}-custom.md`
- `.opencode/divisor/packs/default.md`
- `.opencode/divisor/packs/default-custom.md`

### Output Format

```text
unbound v1.2.3 initialized (divisor)

  + .opencode/agents/divisor-guard.md
  + .opencode/agents/divisor-architect.md
  + .opencode/agents/divisor-adversary.md
  + .opencode/agents/divisor-sre.md
  + .opencode/agents/divisor-testing.md
  + .opencode/command/review-council.md
  + .opencode/divisor/packs/go.md
  + .opencode/divisor/packs/go-custom.md
  + .opencode/divisor/packs/default.md
  + .opencode/divisor/packs/default-custom.md

Created: 10  Skipped: 0  Updated: 0

Run /review-council to start a code review.
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (filesystem failure, invalid flag combination) |

### Error Cases

| Scenario | Behavior |
|----------|----------|
| Cannot detect language, no `--lang` | Fall back to `default.md` pack. Print informational note. |
| `--lang` value has no matching pack | Error: "No convention pack available for language '{lang}'" |
| Target directory not writable | Error with `os.MkdirAll` context |
| Partial write failure | Error with file-level context. Files written before failure remain. |

## Go API: `scaffold.Options`

```go
type Options struct {
    TargetDir   string    // Root dir (default: cwd)
    Force       bool      // Overwrite all files
    DivisorOnly bool      // Divisor subset only
    Lang        string    // Language override (auto-detect if empty)
    Version     string    // Version for marker comment
    Stdout      io.Writer // Summary output writer
}
```

## Go API: `scaffold.Result`

```go
type Result struct {
    Created     []string // New files
    Skipped     []string // Existing, not overwritten
    Overwritten []string // Existing, replaced by --force
    Updated     []string // Tool-owned, content changed
}
```

## New Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `isDivisorAsset` | `func(relPath string) bool` | Predicate: is this asset in the Divisor subset? |
| `isConventionPack` | `func(relPath string) bool` | Predicate: is this a convention pack file? |
| `shouldDeployPack` | `func(relPath, lang string) bool` | Filter: should this pack be deployed for the resolved language? |
| `detectLang` | `func(targetDir string) string` | Auto-detect project language from marker files |

## File Ownership

| Pattern | Ownership | Auto-update? |
|---------|-----------|:---:|
| `opencode/agents/divisor-*.md` | User-owned | No |
| `opencode/command/review-council.md` | Tool-owned | Yes |
| `opencode/divisor/packs/{lang}.md` | Tool-owned | Yes |
| `opencode/divisor/packs/{lang}-custom.md` | User-owned | No |
<!-- scaffolded by unbound vdev -->

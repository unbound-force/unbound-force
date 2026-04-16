# Quickstart: Unleash OpenSpec Support

**Branch**: `031-unleash-openspec` | **Date**: 2026-04-16

## Verification Guide

### Prerequisites

- An existing OpenSpec change created via `/opsx-propose`
- The `opsx/<name>` branch checked out
- `openspec/changes/<name>/tasks.md` exists

### Verify US1: Single Command for Both Workflows

```bash
# 1. Create an OpenSpec change (if none exists)
/opsx-propose test-unleash-openspec

# 2. Verify you're on the opsx branch
git rev-parse --abbrev-ref HEAD
# Expected: opsx/test-unleash-openspec

# 3. Run /unleash
/unleash

# Expected output:
# - "Detected OpenSpec change: test-unleash-openspec"
# - "OpenSpec mode — artifacts from /opsx-propose,
#    skipping clarify/plan/tasks"
# - Spec review runs
# - Implementation runs
# - Code review runs
# - Retrospective runs
# - Demo output displayed
```

### Verify US2: Resumability

```bash
# 1. Run /unleash on an opsx/* branch
/unleash

# 2. After spec review passes, interrupt (Ctrl+C)

# 3. Re-run /unleash
/unleash

# Expected output:
# - "Detected: clarify ✓ plan ✓ tasks ✓ spec-review ✓"
# - "Resuming at step 5/8: Implementing..."
# - Skips spec review (marker present)
```

### Verify US3: Skip Clarify/Plan/Tasks

```bash
# 1. Run /unleash on an opsx/* branch
/unleash

# Expected: NO clarify, plan, or tasks steps execute
# Expected: Output includes "skipping clarify/plan/tasks"
# Expected: Pipeline starts at spec review
```

### Verify US4: Spec Review with OpenSpec Artifacts

```bash
# 1. Run /unleash on an opsx/* branch
/unleash

# Expected: Review council receives
#   openspec/changes/<name>/ as review scope
# Expected: Review council announces
#   "Detected Spec Review Mode (OpenSpec)"
```

### Verify Backward Compatibility

```bash
# 1. Switch to a Speckit branch
git checkout 031-unleash-openspec

# 2. Run /unleash
/unleash

# Expected: Full Speckit pipeline runs unchanged
# Expected: No OpenSpec-related output
```

### Verify Error Handling

```bash
# 1. Create an opsx branch without /opsx-propose
git checkout -b opsx/no-artifacts

# 2. Run /unleash
/unleash

# Expected: STOP with "No tasks.md found for change
#   `no-artifacts`. Run `/opsx-propose` first."

# 3. Clean up
git checkout main && git branch -d opsx/no-artifacts
```

### Verify Scaffold Asset Sync

```bash
# After implementation, run the test suite
go test -race -count=1 ./internal/scaffold/...

# Expected: TestEmbeddedAssets_MatchSource passes
# (live file matches scaffold asset copy)
```

## Implementation Checklist

- [ ] Step 1 modified: `opsx/*` STOP replaced with
      OpenSpec detection + change name extraction
- [ ] Step 1 modified: OpenSpec prerequisite check
      (tasks.md in change directory)
- [ ] Step 2 modified: OpenSpec resumability (clarify/
      plan/tasks always "done")
- [ ] Steps 1-3 modified: Skip for OpenSpec with
      announcement
- [ ] Step 4 modified: Pass FEATURE_DIR to review
      council
- [ ] Steps 5-8 modified: Use FEATURE_DIR for both modes
- [ ] Step 8 modified: Use proposal.md (not spec.md)
      for OpenSpec demo
- [ ] Guardrails updated: Reflect both branch patterns
- [ ] Scaffold asset synced:
      `internal/scaffold/assets/opencode/command/unleash.md`
- [ ] Tests pass: `go test -race -count=1 ./...`

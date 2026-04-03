# Quickstart: Hivemind-to-Dewey Memory Migration

**Branch**: `022-hivemind-dewey-migration`
**Date**: 2026-04-03

## Prerequisites

- Go 1.24+ installed
- Repository cloned and on the
  `022-hivemind-dewey-migration` branch

## Phase A: Verify Migration (automated)

Run the full test suite to verify the migration:

```bash
make check
```

This runs `go test -race -count=1 ./...` which includes:
- `TestEmbeddedAssets_MatchSource` — verifies scaffold
  assets match canonical sources (drift detection)
- `TestScaffoldOutput_NoHivemindReferences` — verifies
  no `hivemind_store` or `hivemind_find` references in
  scaffolded output
- `TestScaffoldOutput_NoGraphthulhuReferences` — existing
  regression guard (should continue to pass)

All tests must pass.

## Phase B: Verify Migration (manual)

### B1: Verify `/unleash` retrospective step

```bash
# Should return 0 (no hivemind_store references)
grep -c "hivemind_store" .opencode/command/unleash.md
# Expected: 0

# Should return >= 1 (dewey_store_learning present)
grep -c "dewey_store_learning" .opencode/command/unleash.md
# Expected: >= 1

# Verify scaffold copy matches
diff .opencode/command/unleash.md \
  internal/scaffold/assets/opencode/command/unleash.md
# Expected: no output (files identical)
```

### B2: Verify Divisor agent Prior Learnings

```bash
# Should return 0 (no hivemind_find references)
grep -c "hivemind_find" .opencode/agents/divisor-*.md
# Expected: 0 for each file

# Should return >= 1 per file (dewey_semantic_search)
grep -c "dewey_semantic_search" \
  .opencode/agents/divisor-*.md
# Expected: >= 1 for each of 5 files

# Verify scaffold copies match
for f in divisor-adversary divisor-architect \
         divisor-guard divisor-sre divisor-testing; do
  diff ".opencode/agents/${f}.md" \
    "internal/scaffold/assets/opencode/agents/${f}.md"
done
# Expected: no output (all files identical)
```

### B3: Verify documentation

```bash
# AGENTS.md should not reference Hivemind as active
grep -n "Hivemind" AGENTS.md
# Expected: only in historical "Recent Changes" entries
# and the new Spec 022 entry. No references in the
# "Embedding Model Alignment" section.

# setup.go should not reference Hivemind
grep -n "Hivemind" internal/setup/setup.go
# Expected: 0 matches
```

### B4: Verify scaffold output

```bash
# Scaffold to a temp directory
tmpdir=$(mktemp -d)
go run ./cmd/unbound-force/ init "$tmpdir"

# Check scaffolded files
grep -r "hivemind_store\|hivemind_find" "$tmpdir"
# Expected: no output (zero matches)

# Clean up
rm -rf "$tmpdir"
```

## Phase C: End-to-End Verification (optional)

These steps require Dewey to be installed and running.

### C1: Verify learning retrieval works

1. Start a Dewey-enabled OpenCode session
2. Trigger a Divisor review on any file
3. Verify the Prior Learnings step queries
   `dewey_semantic_search` (visible in agent output)
4. If Dewey is unavailable, verify the agent proceeds
   with a warning rather than an error

### C2: Verify learning storage works

1. Run `/unleash` on a small feature
2. At the retrospective step, verify learnings are
   stored via `dewey_store_learning` (or warned if
   dewey#25 has not landed)
3. In a subsequent session, verify the stored learnings
   are retrievable via `dewey_semantic_search`

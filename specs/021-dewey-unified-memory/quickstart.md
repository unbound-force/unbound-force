# Quickstart: Dewey Unified Memory

**Branch**: `021-dewey-unified-memory`
**Date**: 2026-04-03

## Verification Steps

These steps verify the implementation is correct after
all phases are complete. They map to the spec's
acceptance scenarios and success criteria.

### Phase A: Dewey Repo Changes (Ollama + Learnings)

**V1: Ollama auto-start (SC-001)**

```bash
# Stop Ollama if running
pkill ollama || true
sleep 2

# Start Dewey — should auto-start Ollama
dewey serve &
DEWEY_PID=$!
sleep 5

# Verify Ollama is now running
curl -s http://localhost:11434/ | grep -q "Ollama"
echo "PASS: Ollama auto-started by Dewey"

# Verify semantic search works (not keyword-only)
# (requires MCP client — use OpenCode or test harness)

kill $DEWEY_PID
```

**V2: Ollama already running (US1-S2)**

```bash
# Start Ollama manually
ollama serve &
sleep 3

# Start Dewey — should detect existing Ollama
dewey serve &
DEWEY_PID=$!
sleep 3

# Verify Dewey logs "using existing Ollama instance"
# (check Dewey stderr/logs)

kill $DEWEY_PID
# Ollama should still be running
curl -s http://localhost:11434/ | grep -q "Ollama"
echo "PASS: Existing Ollama preserved"
```

**V3: Ollama not installed (US1-S3)**

```bash
# Temporarily hide ollama from PATH
PATH_BACKUP=$PATH
export PATH=$(echo $PATH | tr ':' '\n' | grep -v ollama | tr '\n' ':')

dewey serve &
DEWEY_PID=$!
sleep 3

# Verify Dewey logs "keyword-only mode"
# Verify dewey_search works (keyword)
# Verify dewey_semantic_search returns degraded results

kill $DEWEY_PID
export PATH=$PATH_BACKUP
```

**V4: Learning storage (SC-002, SC-003)**

```bash
# With Dewey running and Ollama serving:

# Store a learning via MCP tool
# (use OpenCode or MCP test client)
# dewey_store_learning({
#   text: "scaffold.go requires initSubTools nil guard",
#   tags: ["021-dewey-unified-memory", "gotcha"]
# })

# Search for the learning
# dewey_semantic_search({ query: "scaffold nil guard" })
# Verify the learning appears in results
# Verify source_type is "learning"

# Search with unified query
# dewey_semantic_search({ query: "scaffold patterns" })
# Verify learning appears alongside spec and code results
```

**V5: Learning filtering (US2-S4)**

```bash
# Store a learning with specific tags
# dewey_store_learning({
#   text: "doctor checks should verify embedding capability",
#   tags: ["021-dewey-unified-memory", "2026-04-03", "pattern"]
# })

# Filter by tag
# dewey_semantic_search_filtered({
#   query: "doctor checks",
#   has_tag: "pattern"
# })
# Verify only tagged learnings returned
```

### Phase B: This Repo Changes (Agent Migration)

**V6: Unleash retrospective uses Dewey (SC-004)**

```bash
# Verify unleash.md references dewey_store_learning
grep -c "dewey_store_learning" .opencode/command/unleash.md
# Expected: >= 1

# Verify no hivemind_store references
grep -c "hivemind_store" .opencode/command/unleash.md
# Expected: 0
```

**V7: Divisor agents use Dewey (US3-S2)**

```bash
# Verify all 5 Divisor agents reference Dewey
for f in .opencode/agents/divisor-*.md; do
  echo "=== $f ==="
  grep -c "dewey_semantic_search" "$f"
done
# Expected: >= 1 for each file

# Verify no hivemind_find references
grep -c "hivemind_find" .opencode/agents/divisor-*.md
# Expected: 0
```

**V8: Graceful degradation preserved (FR-014)**

```bash
# Verify degradation instructions exist
for f in .opencode/agents/divisor-*.md; do
  grep -l "Dewey is unavailable\|graceful\|skip" "$f"
done
# Expected: all 5 files match

grep -l "Dewey is unavailable\|graceful\|skip" \
  .opencode/command/unleash.md
# Expected: match
```

**V9: AGENTS.md updated (FR-015)**

```bash
# Verify unified memory documentation
grep -c "unified memory\|replaces Hivemind" AGENTS.md
# Expected: >= 1

# Verify no "complements Hivemind" framing
grep -c "complements Hivemind" AGENTS.md
# Expected: 0
```

**V10: Scaffold asset sync**

```bash
# Verify drift detection passes
go test -race -count=1 ./internal/scaffold/...
# Expected: PASS
```

### Phase C: Swarm Fork

**V11: Fork functional (SC-005)**

```bash
# Install forked plugin
npm install -g @unbound-force/opencode-swarm-plugin@latest

# Verify all Swarm tools work
swarm doctor
# Expected: all checks pass
```

**V12: Setup installs fork (US4-S2)**

```bash
# Verify setup references forked package
grep -c "unbound-force" internal/setup/setup.go
# Expected: references to forked package name
```

### Phase D: Doctor/Setup Updates

**V13: Doctor embedding check (SC-006)**

```bash
uf doctor
# Verify "Dewey Knowledge Layer" group shows:
# - dewey binary: found
# - embedding model: granite-embedding:30m installed
# - embedding capability: available (or keyword-only)
# - workspace: initialized
```

**V14: Full test suite**

```bash
make check
# Expected: all tests pass, zero regressions
```

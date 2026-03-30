# Quickstart: Divisor Council Refinement

**Spec**: 019 | **Date**: 2026-03-30

## Verification Steps

After implementation, use these steps to verify each user
story is satisfied.

### US1: Zero-Waste Cleanup

1. **Scaffold assets removed**:
   ```bash
   # Verify no reviewer-*.md in scaffold assets
   ls internal/scaffold/assets/opencode/agents/reviewer-*.md
   # Expected: "No such file or directory"
   ```

2. **Tests pass with updated counts**:
   ```bash
   go test -race -count=1 ./internal/scaffold/...
   # Expected: PASS
   # Verify expectedAssetPaths count decreased by 4,
   # increased by 1 (severity.md) = net -3
   ```

3. **Legacy file warning**:
   ```bash
   # In a temp dir with legacy files:
   mkdir -p /tmp/test-legacy/.opencode/agents
   touch /tmp/test-legacy/.opencode/agents/reviewer-{adversary,architect,guard,sre,testing}.md
   uf init /tmp/test-legacy
   # Expected: warning listing legacy files + removal command
   ```

4. **Review council discovery**:
   ```bash
   # Verify /review-council only discovers divisor-* agents
   ls .opencode/agents/divisor-*.md
   # Expected: 5 files
   # /review-council discovery pattern is "divisor-*.md"
   ```

### US2: De-duplicated Review Findings

1. **Ownership boundaries in agent files**:
   ```bash
   # Verify each agent has an "Out of Scope" section
   grep -l "Out of Scope" .opencode/agents/divisor-*.md
   # Expected: 5 files
   ```

2. **No duplicate dimensions**:
   ```bash
   # Verify "hardcoded secrets" only in adversary
   grep -l "hardcoded secrets\|Hardcoded.*secret" \
     .opencode/agents/divisor-*.md
   # Expected: only divisor-adversary.md

   # Verify "test isolation" only in tester
   grep -l "Test.*[Ii]solation\|test isolation" \
     .opencode/agents/divisor-*.md
   # Expected: only divisor-testing.md (as a primary section)
   ```

### US3: Consistent Severity Classification

1. **Severity pack exists**:
   ```bash
   cat .opencode/unbound/packs/severity.md
   # Expected: CRITICAL/HIGH/MEDIUM/LOW definitions
   ```

2. **All agents reference severity pack**:
   ```bash
   grep -l "severity.md\|severity definitions" \
     .opencode/agents/divisor-*.md
   # Expected: 5 files
   ```

3. **Scaffold deploys severity pack**:
   ```bash
   ls internal/scaffold/assets/opencode/unbound/packs/severity.md
   # Expected: file exists
   ```

### US4: Qualified FR References

1. **No bare FR references**:
   ```bash
   # Search for bare "FR-" without "Spec NNN" qualifier
   grep -n "FR-[0-9]" .opencode/agents/divisor-*.md \
     | grep -v "per Spec [0-9]"
   # Expected: no output (all FR refs are qualified)
   ```

### US5: Learning-Informed Reviews

1. **Prior Learnings step in agents**:
   ```bash
   grep -l "Prior Learnings\|hivemind_find" \
     .opencode/agents/divisor-*.md
   # Expected: 5 files
   ```

2. **Graceful degradation documented**:
   ```bash
   grep -l "Hivemind.*not available\|skip.*informational" \
     .opencode/agents/divisor-*.md
   # Expected: 5 files
   ```

### US6: Static Analysis in CI

1. **CI workflow updated**:
   ```bash
   grep -c "golangci-lint\|govulncheck" \
     .github/workflows/test.yml
   # Expected: >= 2 (both tools present)
   ```

2. **Tools run locally**:
   ```bash
   golangci-lint run
   # Expected: exit 0 (no findings)

   govulncheck ./...
   # Expected: exit 0 (no vulnerabilities)
   ```

3. **Setup installs tools**:
   ```bash
   grep -l "golangci-lint\|govulncheck" \
     internal/setup/setup.go
   # Expected: file contains install logic
   ```

## Full Test Suite

```bash
# Run all tests (must pass before merge)
make check
# or: go build ./... && go vet ./... && go test -race -count=1 ./...

# Run scaffold tests specifically
go test -race -count=1 -v ./internal/scaffold/...

# Run setup tests specifically
go test -race -count=1 -v ./internal/setup/...
```

## Expected File Count Change

| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| Scaffold assets (expectedAssetPaths) | 52 | 49 | -3 |
| Legacy reviewer assets removed | 4 | 0 | -4 |
| New severity.md pack added | 0 | 1 | +1 |
| knownNonEmbeddedFiles entries | reviewer-testing.md | removed | -1 |

Note: `reviewer-testing.md` was already in
`knownNonEmbeddedFiles` (not embedded). It should be
removed from that list since the canonical source file
will remain on disk but is no longer expected to be
embedded (it was never embedded).

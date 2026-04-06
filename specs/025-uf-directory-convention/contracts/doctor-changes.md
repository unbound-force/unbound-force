# Contract: Doctor Changes

**Spec**: 025-uf-directory-convention
**Type**: Internal contract

## Purpose

Defines the exact changes required in the doctor package
(`internal/doctor/checks.go`) and its test file.

## Function Changes

### checkDewey(opts *Options) CheckGroup

Path changes:
1. Workspace directory check:
   **Before**: `filepath.Join(opts.TargetDir, ".dewey")`
   **After**: `filepath.Join(opts.TargetDir, ".uf", "dewey")`

No changes to: dewey binary check, embedding model check,
embedding capability check. These are tool-level checks
unrelated to the workspace path.

### checkReplicator(opts *Options) CheckGroup

Path changes:
1. `.hive/` existence check:
   **Before**: `filepath.Join(opts.TargetDir, ".hive")`
   **After**: `filepath.Join(opts.TargetDir, ".uf", "replicator")`

2. Display name:
   **Before**: `.hive/`
   **After**: `.uf/replicator/`

No changes to: replicator binary check, replicator doctor
delegation, MCP config check.

### checkScaffoldedFiles(opts *Options) CheckGroup

Path changes:
1. Convention packs directory:
   **Before**: `filepath.Join(opts.TargetDir, ".opencode", "unbound", "packs")`
   **After**: `filepath.Join(opts.TargetDir, ".opencode", "uf", "packs")`

2. Display name:
   **Before**: `.opencode/unbound/packs/`
   **After**: `.opencode/uf/packs/`

No changes to: `.opencode/agents/`, `.opencode/command/`,
`.specify/`, `AGENTS.md` checks.

## Test Changes

### doctor_test.go

All test setup that creates `.dewey/` directories must
create `.uf/dewey/` instead.

All test setup that creates `.hive/` directories must
create `.uf/replicator/` instead.

All test assertions checking for `.dewey/` or `.hive/`
in output must check for `.uf/dewey/` or `.uf/replicator/`.

All test setup that creates `.opencode/unbound/packs/`
must create `.opencode/uf/packs/` instead.

## Behavior Matrix

| Condition | Old Behavior | New Behavior |
|-----------|-------------|-------------|
| `.uf/dewey/` exists | N/A | PASS: "initialized" |
| `.uf/dewey/` missing, dewey installed | N/A | WARN: "not initialized", hint: "dewey init" |
| `.uf/replicator/` exists | N/A | PASS: "initialized" |
| `.uf/replicator/` missing, replicator installed | N/A | WARN: "not initialized", hint: "Run: uf init" |
| `.opencode/uf/packs/` exists with .md files | N/A | PASS: "N convention packs" |
| `.opencode/uf/packs/` missing | N/A | FAIL: "not found", hint: "Run: uf init" |

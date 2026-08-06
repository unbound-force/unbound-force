<!-- spec-review: passed -->
<!--
  All tasks in Group 1 modify the same file (uf.init.md),
  so no [P] markers — parallel execution would cause merge
  conflicts. Tasks run sequentially in order.

  Group 2 tasks modify a different file (*_test.go) and
  can run in parallel with each other but depend on Group 1.

  Group 3 modifies the 3 speckit command files and depends
  on Group 1 being complete to verify correctness.
-->

## 1. Update Step 6 in uf.init.md

All subtasks modify
`internal/scaffold/assets/opencode/commands/uf.init.md`.
After Group 1 is complete, the deployed copy at
`.opencode/commands/uf.init.md` MUST be synced to match
(byte-identical, enforced by
`TestEmbeddedAssets_MatchSource`).

- [x] 1.1 Replace the "Execution/utility commands" category
  (lines 561–565) with three individual command entries,
  each referencing its own guardrail block below. Update the
  Step 6 preamble (lines 548–550) to say "Use five variants
  depending on the command type" instead of "two variants".

- [x] 1.2 Replace the "Execution/utility guardrails block"
  (lines 607–625) with three command-specific guardrail
  blocks:

  **Implement guardrails** (`speckit.implement.md`):
  - This command writes source code — implementation is
    its purpose
  - Scope: implement only what is defined in the active
    feature's `tasks.md`
  - MUST NOT contain "NEVER modify source code"
  - Correctness marker: "writes source code"

  **Constitution guardrails** (`speckit.constitution.md`):
  - Allowed write targets: `.specify/memory/constitution.md`,
    `.specify/templates/`
  - MUST NOT restrict writes to `FEATURE_SPEC`/`FEATURE_DIR`
  - MUST NOT contain "NEVER modify source code"
  - Correctness marker: ".specify/memory/"

  **Taskstoissues guardrails** (`speckit.taskstoissues.md`):
  - This command creates GitHub issues via the MCP API
  - Issues MUST only be created in the repository matching
    the current Git remote
  - This command does NOT write local files
  - MUST NOT contain "files this command may write"
  - Correctness marker: "GitHub issues via"

- [x] 1.3 Remove the workaround Note (lines 627–632) that
  documents the implement override conflict. The fix
  eliminates the conflict it worked around.

- [x] 1.4 Update the idempotency logic (lines 567–583) to
  handle replace-if-incorrect for execution/utility
  commands: when `## Guardrails` exists but the
  command-specific correctness marker is absent, replace
  the entire guardrails section. Add report line:
  `✅ <filename>: guardrails corrected`
  Note: The existing secondary check for spec-phase
  commands (review-rationale sentence detection) MUST
  remain unchanged. The `## Guardrails` heading check
  MUST match headings outside of fenced code blocks.

- [x] 1.5 Copy the updated
  `internal/scaffold/assets/opencode/commands/uf.init.md`
  to `.opencode/commands/uf.init.md` to maintain byte
  parity. Verify with diff.

## 2. Add Guardrail Content Regression Tests

- [x] 2.1 Add a test function in the scaffold test file that
  reads the `uf.init.md` content and verifies:
  - The spec-phase guardrail block contains
    "NEVER modify source code"
  - The implement guardrail block contains
    "writes source code" and does NOT contain
    "NEVER modify source code"
  - The constitution guardrail block contains
    ".specify/memory/" and does NOT restrict writes
    to "FEATURE_DIR" only
  - The taskstoissues guardrail block contains
    "GitHub issues" and does NOT contain
    "files this command may write"

- [x] 2.2 Run `make test` to verify all tests pass,
  including the new regression tests and the existing
  `TestEmbeddedAssets_MatchSource` drift detection.

## 3. Verify Corrected Guardrails in Command Files

- [x] 3.1 [P] Verify `speckit.implement.md` guardrails are
  correct: contains "writes source code", does NOT contain
  "NEVER modify source code"

- [x] 3.2 [P] Verify `speckit.constitution.md` guardrails
  are correct: contains ".specify/memory/", does NOT
  restrict to "FEATURE_DIR" only

- [x] 3.3 [P] Verify `speckit.taskstoissues.md` guardrails
  are correct: contains "GitHub issues via", does NOT
  contain "files this command may write"

## 4. Constitution Alignment Verification

- [x] 4.1 Verify Phase Discipline: each command's guardrails
  accurately describe the command's operational scope
  (no false restrictions, no missing constraints)

- [x] 4.2 Verify Gatekeeping Integrity: guardrails serve as
  governance gates with correct content — no
  known-incorrect gates that require override workarounds

- [x] 4.3 Verify Zero-Waste: the workaround Note is
  removed, no aspirational or contradictory content
  remains
<!-- code-review: passed -->

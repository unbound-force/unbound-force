<!--
  All tasks modify the same file (.opencode/commands/uf-init.md),
  so no [P] markers — parallel execution would cause merge
  conflicts. Tasks run sequentially in order.
-->

## 1. Tighten Branch Enforcement Idempotency (Step 2)

- [x] 1.1 In Step 2 "Branch Enforcement: Propose", split the
  idempotency check into two independent checks:
  (a) basic branch check — look for `opsx/<name>` or
  `opsx/<change-name>` as branch references;
  (b) dirty tree check — look for `git status --short`
  in a pre-branch-creation context. If (a) is present but
  (b) is not, insert only the dirty tree check portion.
  Update the report to show separate status for each.

- [x] 1.2 In Step 2 "Branch Enforcement: Archive", add an
  idempotency check for the commit-before-archive flow:
  look for `git add` and `git commit` appearing before
  the archive move step. If the basic branch/return-to-main
  content is present but commit-before-archive is not,
  insert only the commit step. Add the following insertion
  content: a step that runs `git status --short`, stages
  and commits all changes with a descriptive message, and
  pushes. Mark it CRITICAL: "Do NOT move to archive with
  uncommitted changes."

- [x] 1.3 In Step 2 "Branch Enforcement: Explore" (currently
  noted as excluded), add a new sub-section for explore's
  guardrail-only branch enforcement: check if the explore
  SKILL.md guardrails section contains "uncommitted changes"
  or "switch branches". If not, insert a guardrail bullet:
  "Don't switch branches without confirmation — If
  exploration leads to creating a proposal (which requires
  a new `opsx/` branch), check for uncommitted changes
  first and ask the user before switching."

## 2. Add STOP HERE Block Injection (New Step 10)

- [x] 2.1 Add a new "### Step 10: STOP HERE Blocks" section
  after Step 9 (Report Results). Define target files as
  the 6 spec-phase speckit commands: `speckit.specify.md`,
  `speckit.plan.md`, `speckit.tasks.md`,
  `speckit.analyze.md`, `speckit.checklist.md`,
  `speckit.clarify.md`. Explicitly exclude
  `speckit.implement.md`, `speckit.constitution.md`,
  `speckit.taskstoissues.md`.

- [x] 2.2 Define the idempotency check: search for "STOP HERE"
  (case-sensitive). If found, report skip.

- [x] 2.3 Define the insertion content: a bold directive
  "**STOP HERE. Do NOT proceed to implementation.**"
  followed by instructions to report results and prompt
  the user to invoke `/speckit.implement`, `/unleash`,
  or `/cobalt-crush` for implementation.

- [x] 2.4 Define the insertion point: after the main workflow
  instructions, before the `## Guardrails` section. If
  no Guardrails section exists, insert at the end of the
  file.

## 3. Add Review-Rationale to Guardrails (Modify Step 6)

- [x] 3.1 Modify Step 6 "Speckit Command Guardrails" to define
  two Guardrails variants. Spec-phase variant (for specify,
  plan, tasks, analyze, checklist, clarify): includes the
  standard block PLUS the review-rationale sentence.
  Execution/utility variant (implement, constitution,
  taskstoissues): standard block only, no review-rationale.

- [x] 3.2 Add a secondary idempotency check: if `## Guardrails`
  already exists on a spec-phase command, check whether the
  review-rationale sentence is present (search for "review
  defeats the purpose"). If the heading exists but the
  sentence is missing, append the sentence to the existing
  Guardrails section. Report:
  `✅ <filename>: review-rationale added`.

## 4. Add Scaffold Comment Deduplication (New Step 11)

- [x] 4.1 Add a new "### Step 11: Scaffold Comment
  Deduplication" section. Define the pattern to match:
  `<!-- scaffolded by uf ... -->` (any version string
  after "uf").

- [x] 4.2 Define the deduplication logic: if multiple
  matches are found, keep only the last occurrence, remove
  all earlier occurrences. Report count of removed
  duplicates.

- [x] 4.3 Define the target scope: all files that `/uf-init`
  processes (the 7 OpenSpec files and 9 speckit commands).
  This step runs after all other insertions to catch any
  duplicates introduced by earlier steps.

## 5. Add Legacy Directory Cleanup (New Step 12)

- [x] 5.1 Add a new "### Step 12: Legacy Directory Cleanup"
  section with two sub-tasks:

- [x] 5.2 Sub-task A: `unbound/packs/` removal. Check if
  `.opencode/unbound/packs/` exists. If yes, verify
  `.opencode/uf/packs/default.md` and
  `.opencode/uf/packs/severity.md` exist. If both
  pre-conditions met, remove `.opencode/unbound/`
  recursively. Report result.

- [x] 5.3 Sub-task B: `command/` (singular) migration
  hardening. Add a note to Step 0 that after migration,
  verify the target files are now findable in `commands/`
  (plural). If Step 0 ran but speckit commands are still
  in `command/` (check for `speckit.specify.md` in both
  directories), report a warning.

## 6. Update Report Template (Step 9)

- [x] 6.1 Add new sections to the Step 9 report template
  for the three new steps:
  ```
  ### STOP HERE Blocks
    [status] [filename]: [action]
    ...

  ### Scaffold Comment Deduplication
    [status] [filename]: [action]
    ...

  ### Legacy Directory Cleanup
    [status] [item]: [action]
    ...
  ```

- [x] 6.2 Update the Summary line to include the new
  categories in the count.

## 7. Verification

- [x] 7.1 After all edits to `uf-init.md` are complete,
  re-read the file and verify: all 12 steps are present
  (0-8 original + 10-12 new), the report template in
  Step 9 includes sections for all customization
  categories, and the file is valid Markdown.

- [x] 7.2 Constitution alignment verification: confirm
  the change maintains Autonomous Collaboration (all
  customizations are file-based insertions, no runtime
  coupling), Composability First (missing-file handling
  is preserved in all new steps), and Observable Quality
  (all new steps report status using the existing
  indicator pattern).

- [x] 7.3 Run `git diff` to review all changes to
  `uf-init.md` and verify no existing content was
  accidentally removed or reordered.

- [x] 7.4 Sync the scaffold asset copy: copy the
  modified `.opencode/commands/uf-init.md` to
  `internal/scaffold/assets/opencode/commands/uf-init.md`
  and run `go test -race -count=1 -run
  TestEmbeddedAssets_MatchSource ./internal/scaffold/`
  to verify byte-identity.

<!-- spec-review: passed -->
<!-- code-review: passed -->

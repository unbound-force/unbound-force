## ADDED Requirements

### Requirement: STOP HERE Block Injection

The `/uf-init` command MUST inject a STOP HERE block into
each spec-phase speckit command file. The block MUST
appear after the main workflow instructions and before
the Guardrails section.

Spec-phase commands: `speckit.specify.md`,
`speckit.plan.md`, `speckit.tasks.md`,
`speckit.analyze.md`, `speckit.checklist.md`,
`speckit.clarify.md`.

The STOP HERE block MUST contain:
1. A bold directive: "STOP HERE. Do NOT proceed to
   implementation."
2. Instructions to report results and prompt the user
3. Direction to invoke a separate command for
   implementation

The idempotency check MUST search for the phrase
"STOP HERE" (case-sensitive). If found, skip.

#### Scenario: First run on a file without STOP HERE

- **GIVEN** `speckit.specify.md` exists and does not
  contain "STOP HERE"
- **WHEN** `/uf-init` runs
- **THEN** the STOP HERE block is inserted after the
  main workflow, before Guardrails
- **AND** the report shows
  `✅ speckit.specify.md: STOP HERE inserted`

#### Scenario: Subsequent run on a file with STOP HERE

- **GIVEN** `speckit.specify.md` already contains
  "STOP HERE"
- **WHEN** `/uf-init` runs
- **THEN** no modification is made
- **AND** the report shows
  `⊘ speckit.specify.md: STOP HERE already present (skipped)`

#### Scenario: Execution command excluded

- **GIVEN** `speckit.implement.md` exists
- **WHEN** `/uf-init` runs
- **THEN** no STOP HERE block is inserted into
  `speckit.implement.md`

#### Scenario: File without Guardrails section

- **GIVEN** `speckit.specify.md` does not contain
  "STOP HERE"
- **AND** does not contain a `## Guardrails` section
- **WHEN** `/uf-init` runs
- **THEN** the STOP HERE block is inserted at the end
  of the file

---

### Requirement: Review-Rationale Guardrail Sentence

The Guardrails section injected by Step 6 MUST include a
review-rationale sentence for spec-phase commands only.

The sentence: "The user needs to review the plan before
implementation begins. Implementing without review defeats
the purpose of the spec-first workflow."

This sentence MUST NOT be added to execution or utility
commands (`speckit.implement.md`,
`speckit.constitution.md`, `speckit.taskstoissues.md`).

#### Scenario: Spec-phase command gets review-rationale

- **GIVEN** `speckit.plan.md` does not have a Guardrails
  section
- **WHEN** `/uf-init` runs Step 6
- **THEN** the Guardrails section includes the
  review-rationale sentence

#### Scenario: Execution command omits review-rationale

- **GIVEN** `speckit.implement.md` does not have a
  Guardrails section
- **WHEN** `/uf-init` runs Step 6
- **THEN** the Guardrails section does NOT include the
  review-rationale sentence

#### Scenario: Review-rationale already present

- **GIVEN** `speckit.plan.md` has a `## Guardrails`
  section
- **AND** contains "review defeats the purpose"
- **WHEN** `/uf-init` runs Step 6
- **THEN** no modification is made
- **AND** the report shows
  `⊘ speckit.plan.md: guardrails already present
  (skipped)`

---

### Requirement: Scaffold Comment Deduplication

The `/uf-init` command MUST deduplicate scaffold comments
matching the pattern `<!-- scaffolded by uf ... -->` in
all files it processes.

When multiple scaffold comments are found:
- Keep only the LAST occurrence
- Remove all earlier occurrences
- Report: `✅ <filename>: deduplicated scaffold comments
  (N removed)`

When zero or one scaffold comment is found:
- No action needed
- Report: `⊘ <filename>: scaffold comments clean`

#### Scenario: File with accumulated scaffold comments

- **GIVEN** a file contains 5 lines matching
  `<!-- scaffolded by uf ... -->`
- **WHEN** `/uf-init` runs the deduplication step
- **THEN** only the last scaffold comment remains
- **AND** the report shows 4 removed

#### Scenario: File with single scaffold comment

- **GIVEN** a file contains exactly 1 scaffold comment
- **WHEN** `/uf-init` runs the deduplication step
- **THEN** no modification is made

---

### Requirement: Legacy Directory Cleanup

The `/uf-init` command MUST remove the legacy
`unbound/packs/` directory when the current `uf/packs/`
directory exists and contains core files.

Pre-conditions for removal:
- `.opencode/uf/packs/` MUST exist
- `.opencode/uf/packs/default.md` MUST exist
- `.opencode/uf/packs/severity.md` MUST exist

If pre-conditions are met and `.opencode/unbound/packs/`
exists, remove `.opencode/unbound/packs/` recursively
(NOT the parent `unbound/` directory). Then remove
`.opencode/unbound/` only if it is empty after `packs/`
removal. If `.opencode/unbound/` contains other content,
leave it and report a warning.

If pre-conditions are NOT met, report an error and skip.

#### Scenario: Legacy directory present, uf/packs valid

- **GIVEN** `.opencode/unbound/packs/` exists
- **AND** `.opencode/uf/packs/default.md` exists
- **AND** `.opencode/uf/packs/severity.md` exists
- **WHEN** `/uf-init` runs the cleanup step
- **THEN** `.opencode/unbound/` is removed
- **AND** the report shows
  `✅ unbound/packs/: removed (migrated to uf/packs/)`

#### Scenario: Legacy directory absent

- **GIVEN** `.opencode/unbound/packs/` does not exist
- **WHEN** `/uf-init` runs the cleanup step
- **THEN** no action is taken
- **AND** the report shows
  `⊘ unbound/packs/: not present`

#### Scenario: uf/packs missing core files

- **GIVEN** `.opencode/unbound/packs/` exists
- **AND** `.opencode/uf/packs/default.md` does NOT exist
- **WHEN** `/uf-init` runs the cleanup step
- **THEN** `.opencode/unbound/` is NOT removed
- **AND** the report shows
  `❌ unbound/packs/: uf/packs/ missing core files, skipped`

---

## MODIFIED Requirements

### Requirement: Branch Enforcement Idempotency (Step 2)

Previously: The idempotency check searched semantically
for "any branch management content" (does the file
describe creating, validating, or cleaning up an
`opsx/<name>` branch?).

The idempotency check MUST distinguish between three
branch enforcement variants:

1. **Basic branch check**: Look for `opsx/<name>` or
   `opsx/<change-name>` as branch references
2. **Dirty tree check** (propose only): Look for
   `git status --short` in the context of pre-branch
   creation. If the basic branch check is present but
   the dirty tree check is not, insert the dirty tree
   check only.
3. **Commit-before-archive** (archive-change only): Look
   for `git add` and `git commit` appearing before the
   archive move step. If absent, insert the commit step.
4. **Branch-switch confirmation** (explore only): Look
   for `uncommitted changes` in the guardrails section.
   If absent, insert the guardrail bullet.

Each variant MUST be checked independently. Presence of
one variant MUST NOT cause other variants to be skipped.

#### Scenario: Basic branch check present, dirty tree missing

- **GIVEN** `openspec-propose/SKILL.md` contains
  `opsx/<name>` branch references
- **AND** does NOT contain `git status --short` in a
  pre-branch context
- **WHEN** `/uf-init` runs Step 2 for propose
- **THEN** the dirty tree check is inserted
- **AND** the basic branch check is NOT re-inserted
- **AND** the report shows
  `✅ SKILL.md: dirty tree check inserted`
  `⊘ SKILL.md: branch check already present`

#### Scenario: Return-to-main present, commit-before-archive missing

- **GIVEN** `openspec-archive-change/SKILL.md` contains
  `git checkout main`
- **AND** does NOT contain `git add` and `git commit`
  before the archive move step
- **WHEN** `/uf-init` runs Step 2 for archive
- **THEN** the commit-before-archive step is inserted
- **AND** the return-to-main content is NOT re-inserted

#### Scenario: Branch-switch confirmation missing in explore

- **GIVEN** `openspec-explore/SKILL.md` has a guardrails
  section
- **AND** does NOT contain `uncommitted changes` or
  `switch branches` in the guardrails
- **WHEN** `/uf-init` runs Step 2 for explore
- **THEN** the branch-switch confirmation bullet is
  appended to the guardrails

#### Scenario: All variants already present

- **GIVEN** all branch enforcement variants are present
  in their respective files
- **WHEN** `/uf-init` runs Step 2
- **THEN** no modifications are made
- **AND** all variants report skip status

---

### Requirement: Guardrails Injection Variants (Step 6)

Previously: A single Guardrails block was injected into
all 9 speckit command files identically.

The Guardrails injection MUST use two variants:

**Spec-phase variant** (specify, plan, tasks, analyze,
checklist, clarify): Includes the standard Guardrails
block PLUS the review-rationale sentence.

**Execution/utility variant** (implement, constitution,
taskstoissues): Includes only the standard Guardrails
block, without the review-rationale sentence.

The idempotency check remains: search for `## Guardrails`
heading. The review-rationale sentence SHOULD be checked
separately: if Guardrails exists but the sentence is
missing on a spec-phase command, append the sentence.

#### Scenario: Guardrails present but review-rationale missing

- **GIVEN** `speckit.tasks.md` has a `## Guardrails`
  section
- **AND** does NOT contain "review defeats the purpose"
- **WHEN** `/uf-init` runs Step 6
- **THEN** the review-rationale sentence is appended to
  the existing Guardrails section
- **AND** the report shows
  `✅ speckit.tasks.md: review-rationale added`

---

## REMOVED Requirements

None.

## Why

Divisor workflows currently run each reviewer persona once on the
host model. One model can miss defects another model finds. Commands
also cannot declare ordered model runs or preserve model-level cost and
provenance. Issue #635 requires configurable fan-out and deterministic
verdict consolidation.

A working implementation exists in the local `fullsend-ai/fullsend`
checkout. This change treats that implementation as validated input,
not code to reconstruct. It ports the Option B responsibility split,
then corrects gaps found during target-repository review.

The validated reference is commit
`b37dbc597baf0180008d3e4a5a764521450b0996`. File digests and all target
adaptations are recorded in `design.md`.

## What Changes

- Add a version 2 review matrix with named model profiles and ordered,
  per-mode, per-agent run arrays.
- Make explicit runs authoritative. Agents without explicit runs use
  advisor generation only when configured; otherwise they run once on
  the host model.
- Add a shared `dispatch-advisor` skill with deterministic persona
  selection, total tier rules, bounded opt-in fan-out, and a versioned
  machine-readable plan.
- Add a policy-free `invoke_agent` plugin with strict validation while
  preserving existing OpenCode and persona permissions.
- Add a deterministic review-dispatch policy plugin for plan
  validation, semantic checks, and atomic artifact persistence.
- Route review council, issue triage, feedback escalation, and Speckit
  test review through the shared plan and invocation contracts.
- Preserve each workflow's verdict policy and return a cause-based
  `INCONCLUSIVE` or `UNAVAILABLE` when no assessment succeeds.
- Emit a Hero Interface Contract envelope containing the complete plan,
  runs, consolidated findings, advisories, verdict, and provenance.
- Version the canonical `review-verdict` artifact to 2.0.0 so
  `INCONCLUSIVE` and `UNAVAILABLE` are native decisions. Migrate its
  consumers while keeping `review-dispatch` as additive provenance.
- Add bounded sibling-repository evidence and schema-valid,
  source-grounded Dewey lesson proposals. After deterministic
  validation,
  parent commands store ready proposals through the existing
  `dewey_store_learning` interface as normal learnings.
- Add `--full`, immutable PR base/head input context, and model
  self-reporting from the validated reference.
- Extend `uf init` with canonical assets, exact npm dependencies, a
  tracked lockfile, safe staged plugin activation, and doctor checks.
- Apply the approved narrow commit-scope amendment so scaffold parity
  required by this change can ship atomically with the feature.
- Add measurable TypeScript, Go, schema, command-contract, and smoke
  test coverage without external services.

The user explicitly approved full reference parity and the OpenSpec
workflow before proposal creation. The expanded features remain one
deployable slice because the commands, policy, execution plugin,
artifact, context boundary, and scaffold must agree before generated
repositories can use multi-model review safely. Splitting those runtime
contracts would leave partially configured or unobservable behavior.

## Capabilities

### New Capabilities

- `review-dispatch`: Resolve explicit or advised model runs, execute
  them within limits, and emit a versioned review artifact.
- `divisor-workflows`: Apply the dispatch and workflow-specific
  consolidation contracts to triage, feedback, and test review.
- `review-context`: Acquire confined sibling evidence and validate
  source-grounded lesson proposals.
- `scaffold`: Provision and safely activate the complete runtime in
  repositories managed by `uf init`.
- `review-verdict`: Version the canonical Divisor decision contract and
  migrate consumers for fail-closed no-success outcomes.

### Modified Capabilities

- `review-council`: Replace single-host delegation with explicit-first
  multi-model dispatch while preserving review gates and iteration
  behavior.

### Removed Capabilities

- None.

## Impact

The change affects `.opencode/commands/`, `.opencode/skills/`, two new
plugins registered in `opencode.json`, `.uf/` review configuration, schemas,
artifact writing, doctor checks, and the Go scaffold and tests. A
closed reviewer-capabilities manifest classifies all nine Divisor
personas without adding unknown OpenCode agent frontmatter.

The implementation also updates the strategic Divisor and documentation
curation contracts. They will record six review personas, three content
personas, plan-scoped reruns, Curator pruning rules, and the canonical
`review-verdict` 2.0.0 consumer migration.

The migration also updates Specs 008 and 009. Orchestration MUST
block both no-success decisions. The shared-data contract MUST
register the version 2 enum and compatibility behavior.

The repository will track `.opencode/package.json` and
`.opencode/package-lock.json`. It will continue to ignore installed
modules and the inconsistent Bun lock. Exact direct versions include
`@opencode-ai/plugin` 1.4.10, Zod 4.1.8, Vitest 5.0.3, and
`@vitest/coverage-v8` 5.0.3.

These four dependencies are necessary and bounded. The official
plugin package is the supported OpenCode ABI; hand-written session
calls would duplicate an unstable protocol. It supplies the typed
client used by the thin plugin. Zod is the plugin SDK's schema type.
The Go test stack cannot execute TypeScript ESM or enforce TypeScript
branch
coverage. Vitest supplies
isolated ESM mocking and thresholds, while its coverage-v8 adapter
supplies branch and statement instrumentation.

Node's built-in test runner and a custom coverage parser were
considered. They would add bespoke mocking and threshold code without
reducing supply chain risk. All four packages are MIT licensed and
maintained by their respective OpenCode, Zod, and Vitest projects.
Exact versions and npm lockfile integrity hashes constrain resolution.
Implementation MUST re-verify SPDX licenses and integrity entries.
Rollback removes both registered plugins, scripts, and new direct
dependencies, restores host-model Task dispatch, and restores the
tracked lock whose baseline SHA256 is
`b628c764236019785a620deadc06373949236ce84fb38bc68a39f619411874e3`,
removes the newly tracked manifest, and restores its prior ignore rule.

This is a user-facing workflow change. Implementation MUST assess
`CHANGELOG.md`, `README.md`, and `AGENTS.md`; file the required website
documentation issue; and assess whether a blog post or tutorial issue
would help users adopt matrix configuration.

The user approved a narrow governance amendment for this change.
Scaffold assets and tests required by and traced to an active
approved change MAY ship with that feature. Unrelated scaffold
refreshes still require a separate branch, and explicit staging
remains mandatory.

Out of scope are new personas, provider credentials, provider-specific
variant guarantees, weakening or replacing existing governance gates,
and automatic execution of sibling content. Additive TypeScript and
scoped Go coverage gates are explicitly in scope.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Independent child sessions exchange a versioned plan and a Hero
Interface Contract artifact. Partial failures do not stop successful
runs. No-success dispatches are explicit and never imply approval.

### II. Composability First

**Assessment**: PASS

The matrix declares runs, the advisor creates a deterministic plan, the
plugin executes one run, and commands consolidate results. Explicit,
advisor, and host sources remain distinguishable and replaceable.

### III. Observable Quality

**Assessment**: PASS

Every invocation records immutable base and head refs and SHAs, branch,
commit, correlation id, plan, all run outcomes, usage when available,
consolidated findings, advisories, and final verdict in a versioned
envelope. Writes are atomic and collision safe. Artifact failure is
visible and prevents a fully successful observability result.

### IV. Testability

**Assessment**: PASS

TypeScript statements MUST remain at or above 90 percent. Branches
at or above 85 percent. New Go helpers MUST reach 80 percent; pure
mapping and configuration helpers MUST reach 90 percent. Existing
global 80 percent, backlog 90 percent, race, drift, and CRAP ratchets
remain unchanged. Integration and smoke tests use fake clients and
scratch repositories without provider, GitHub, Dewey, or network use.
The TypeScript thresholds MUST run through `make check` and Local CI
so a regression blocks the build.
Scoped Go thresholds MUST be checked from a committed manifest by the
same local and CI entry points. Missing or non-numeric coverage MUST
fail rather than silently pass.

### V. Security by Default

**Assessment**: PASS

The invocation plugin validates and bounds inputs, preserves existing
persona permissions, propagates cancellation, and redacts sensitive
errors. Sibling evidence is identity-checked, commit-pinned,
size-bounded, and hash-recorded. Lesson proposals pass deterministic
schema, grounding, secret, tool, and dedupe validation before a parent
may call the existing `dewey_store_learning` interface. Dewey
unavailability skips storage without changing the review verdict.

# Design: stop-shipping-plugin-lock

## Context

`review-council-multi-model-fanout` (SC-FR-003, design "Track exact
npm dependencies and `package-lock.json`") made the lock a tracked,
scaffolded, tool-owned input and installed it with
`npm ci --ignore-scripts --omit=dev`. That gave hero repositories a
pinned, integrity-checked tree. It also assumed uf is the only writer
of `.opencode/` package files. opencode also installs into
`.opencode/`, and users add their own plugin dependencies there, so
that assumption doesn't hold.

## Cross-Change Reconciliation

The contracts touched by this change carry
replaced text in five unarchived sibling changes. This change
cannot ship a MODIFIED `ci` delta of its own: OpenSpec requires
a MODIFIED requirement to match canonical text that will not
exist until its predecessors archive. The reconciliation lives
in the sibling deltas:

- `review-council-multi-model-fanout` authors the requirements
  this change modifies (SC-FR-001, SC-FR-003, SC-FR-004,
  SC-FR-005) and creates the scaffold capability; its archived
  text is the base this delta's `Previously` paragraphs quote.
- `clarification-marker-superset` amended SC-FR-003 after that:
  its still-unarchived delta carries the current stack text
  (lock tracking, integrity hashes, baseline-lock rollback SHA
  `5ebee4af`, and the MIT license clause) that this change's
  MODIFIED requirement replaces.
- `ts-plugin-coverage-gate` (`specs/ci/spec.md:12-14`,
  `proposal.md:169`): its locked-`npm ci` MUST for the
  `plugin-test` target and its proposal's "install discipline
  is unchanged" note are updated to the lockless
  `PLUGIN_TEST_NPM` install in that change's delta.
- `init-override-stubs-gitignore-sync` (SC-FR-015): its `npm
  ci` references (`spec.md:324,356`, `proposal.md:46`,
  `design.md:30,222`, `tasks.md:211,230`) and its
  exactly-four-lines `.opencode/.gitignore` MUST
  (`spec.md:318-322`) are amended for the fifth ignored line
  (`package-lock.json`) and the lockless install.
- `fix-triage-issue-dispatch-docs` (FR-007, Dependency Pin Lockstep
  and Engine Floor): its lockstep enumeration names the deleted
  `.opencode/package-lock.json` and its scaffold mirror, the deleted
  `plugin-package-rollback.json` testdata, and the pre-rename
  `TestPluginPackageInputs_ExactVersionsAndLockIntegrity`; the
  enumeration is amended to the lockless lockstep set.
- The security bump to Zod 4.6.6 and `toml` 5.0.1 (D6) amends this
  delta's pin sentences, override enumeration, and drift scenario in
  place. Predecessor deltas that still name Zod 4.1.8
  (`review-council-multi-model-fanout`,
  `clarification-marker-superset`) carry superseded staging text
  that this change's MODIFIED SC-FR-003 replaces at archive; they
  are not edited.

This change archives after `review-council-multi-model-fanout`
and `clarification-marker-superset`, so its MODIFIED deltas
match the text those changes leave canonical, and after, or
together with, the three reconciled siblings so the canonical
text lines up.

## Decisions

### D1. Stop shipping and tracking the lock

The scaffold drops the embedded `opencode/package-lock.json` asset,
and the path leaves both tool-owned path lists. This repository
removes its tracked `.opencode/package-lock.json` too. The lock is
generated state, and both npm and opencode write it.

### D2. Install without reading or writing a lock

Activation runs `npm install --ignore-scripts --omit=dev
--no-package-lock --no-audit --no-fund`.

- `--no-package-lock` makes npm ignore any lock already present (for
  example a stale tracked copy in an existing hero repository) and
  skip writing one, so activation leaves no new file behind.
- `--ignore-scripts` keeps lifecycle scripts disabled, as before.
- `--no-audit` and `--no-fund` keep the install quiet and
  network-light.

The remediation text in both `uf init` and `uf doctor` names the
same command. The scripts-disabled posture applies to every
lockless install, including `make plugin-test` in this
repository, which runs `npm install --ignore-scripts
--no-package-lock`.

### D2a. Ship no devDependencies

A lockless `npm install --omit=dev` still builds the full ideal tree,
including `devDependencies`, and only skips them when writing
`node_modules`. On node 22's bundled npm 10.9.9 the vitest 5 peer
graph crashes that build ("Cannot read properties of null (reading
'edgesOut')"), so activation would fail on the supported runtime.

- The scaffolded `.opencode/package.json` is this repository's
  manifest without its `devDependencies` block. Tests derive and
  compare it byte for byte.
- `uf init` re-inserts a target's existing `devDependencies` (sorted,
  before `overrides`) when it refreshes the tool-owned manifest, so
  this repository's dev manifest stays byte-identical and skipped.
- `uf doctor` expects only the runtime packages under `node_modules`.
- `make plugin-test` installs the full graph through `npx -y
  npm@11.6.2` (`PLUGIN_TEST_NPM`), which builds it without crashing.

### D3. Ignore the lock in `.opencode/.gitignore`

`package-lock.json` joins the canonical `opencodeGitignore` lines.
The existing merge-only writer (SC-FR-015) appends it to existing
repositories without touching user lines. A file that is already
tracked stays tracked until a user removes it. Deleting it
automatically would destroy a user-owned lock, so uf doesn't.

### D4. Doctor compares pins with the install

Without a lock, the reproducibility contract is the exact pins in
`package.json`.

- `manifest-install` reads each declared package's installed
  `node_modules/<name>/package.json` version.
- It fails on a mismatch, an unreadable install, or a missing
  declaration, and warns on a missing or malformed manifest.
- Packages that aren't installed remain the `dependencies` check's
  concern.

### D5. Record the historical rollback values in prose

The pre-change rollback record
(`internal/scaffold/testdata/plugin-package-rollback.json` and
`TestPluginPackageInputs_RollbackPrerequisites`) verified a
baseline lock that this change deletes, so the pair could only
assert its own testdata against identical constants. Both are
removed; git history preserves them. The recorded values:
baseline lock SHA256
`1ac9cdf171a743c69b18498e15711d628280a4748defd6334cd4276e194d32dc`;
security-override lock SHA256
`fef6f7f145d9dc8398233566934ff7a4e36d9505dbcb662660c773c03c1cdcba`;
prior `.opencode/.gitignore` lines `node_modules`,
`package.json`, `bun.lock`, `.gitignore`; and the override pins
`toml` 4.1.1 and `uuid` 13.0.0.

### D6. Bump zod to 4.6.6 and toml to 5.0.1 for CVE remediation

The throwaway-lock OSV scan added under Risks found five HIGH
advisories in the pinned runtime tree: GHSA-mrhx-c7xc-m3w5 in zod
4.1.8 (fixed in 4.6.6) and GHSA-4qx8-w7c3-pj3p, GHSA-9j65-hcg3-gc64,
GHSA-9qmh-m7c6-f92p, and GHSA-h895-cgqg-jf52 in toml 4.2.0 (all
fixed in 5.0.1). The bump executes the gatekeeping security-fix
path: documented advisories, upstream verification recorded in the
provenance record (git head and registry integrity per bump),
review and CI approval, and rollback by reverting the commit.

`@opencode-ai/plugin` 1.19.0 declares zod 4.1.8 as its own exact
dependency, so bumping only the root pin would leave a nested
vulnerable copy under the plugin. The manifest therefore gains a
zod 4.6.6 override at the same version as the root pin (npm
rejects a conflicting override), which forces the nested copy to
the patched version.

The `toml` override moves across a major version because all four
advisories are fixed only in 5.0.1. Its sole requiring package,
`effect` 4.0.0-beta.83, uses toml only in an unimported unstable
CLI fileParse path (`src/unstable/cli/Primitive.ts`), so forcing
5.0.1 parallels the accepted `uuid` 13.0.1-versus-^14.0.0
divergence. The provenance record carries the new versions, git
heads, and integrity hashes, and `verified_at` moves to 2026-11-06.

## Risks / Trade-offs

- **Transitive drift.** Transitive versions other than `toml` and
  `uuid` can change between installs. This is mitigated by the exact
  direct pins, the security overrides, `--ignore-scripts`, and the
  provider-free import and tool-definition probes, which still gate
  activation.
- **Lock-based scanners.** OSV-Scanner and Dependabot can no
  longer scan the npm transitive tree from a tracked lock. CI
  compensates with a throwaway-lock step: it generates an
  ephemeral lock via the pinned npm (`npx -y npm@11.6.2 install
  --package-lock-only --ignore-scripts`, coordinated with
  `PLUGIN_TEST_NPM`), runs osv-scanner against it, and discards
  it — the lock is never tracked or shipped, so the delta
  holds. Direct `package.json` pins remain scannable;
  `--no-audit` on the install commands silences npm's own
  advisory report, which the throwaway scan replaces.
- **No byte-for-byte reproducible install** in hero repositories.
  Moving to bundled plugin sources would restore it, but is out of
  scope here.

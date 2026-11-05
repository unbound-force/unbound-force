# Proposal: stop-shipping-plugin-lock

## Why

`uf init` treats `.opencode/package-lock.json` as a tool-owned asset.
It overwrites the file on every run and installs the review plugins
with `npm ci`, which fails unless the lock matches `package.json`
exactly. But `.opencode/` is also opencode's own dependency
directory, and opencode installs into it on its own:

- In the website repo, the committed `.opencode/package.json` is
  opencode's own (`{"dependencies": {"@opencode-ai/plugin":
  "1.4.10"}}`). gaze's `package-lock.json` no longer matches the
  scaffold copy.
- In this repo, running the CLI tests with opencode on `PATH`
  recreates `.opencode/package-lock.json` at the repository root.
- When a user adds a plugin or dependency for their own `.opencode/`
  plugins, the lock changes. The next `uf init` then either
  overwrites that change or fails `npm ci` on the mismatch.

Keeping the lock under source means uf and opencode fight over the
same file. The canonical lock also drifts from its scaffold mirror,
which turned main's Local CI red after b9e096a9 (fixed in #953).

## What Changes

- **The scaffold no longer ships a lock.** The embedded asset
  `opencode/package-lock.json` is removed, and the path is no longer
  tool-owned.
- **Activation installs without a lock.** `uf init` runs
  `npm install --ignore-scripts --omit=dev --no-package-lock
  --no-audit --no-fund` in `.opencode/`. It neither reads a stale
  lock nor writes a new one. Lifecycle scripts stay disabled.
- **`.opencode/.gitignore` ignores `package-lock.json`.** The
  existing merge-only writer appends the line to existing
  repositories.
- **This repository stops tracking `.opencode/package-lock.json`.**
  `make plugin-test` installs with `npm install --no-package-lock`.
- **`uf doctor` checks pins against the install.** The manifest-lock
  check becomes `manifest-install`, which compares each declared
  `package.json` pin with the version installed under
  `.opencode/node_modules/`.

`package.json` stays tool-owned, with exact direct-dependency pins
and the `toml` and `uuid` security overrides.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `scaffold`: stops shipping, tracking, and installing from
  `.opencode/package-lock.json`. `uf init` provisions the plugin
  package inputs without a lock, activation installs with
  `npm install --ignore-scripts --omit=dev --no-package-lock`,
  `.opencode/.gitignore` ignores the lock, and `uf doctor`
  compares `package.json` pins with installed versions
  (`manifest-install`) instead of validating a lock.

### Removed Capabilities
- None.

## Impact

- **Reproducibility is weaker.** Direct dependencies and the two
  overridden transitive packages stay pinned exactly, but other
  transitive versions now resolve at install time, and there are no
  lock integrity hashes.
- **Rollout:** existing hero repositories keep their committed
  `.opencode/package-lock.json` until someone removes it with `git rm
  --cached`. `uf init` stops overwriting and reading it, and the
  `.gitignore` merge starts ignoring it.
- **Supersedes** parts of SC-FR-001, SC-FR-003, SC-FR-004 and
  SC-FR-005 from `review-council-multi-model-fanout` (see the
  delta).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change removes a file that two autonomous tools (uf and
opencode) both write, replacing the conflict with an ownership
boundary expressed through `.gitignore` and install flags. The
trade-off and its acceptance are recorded here and in the
design's risk register.

### II. Composability First

**Assessment**: PASS

Each hero repository remains independently usable: the
scaffold ships fewer assets, the install path adds no new
dependencies, and `package.json` retains its exact pins so the
plugin suite composes with the surrounding tooling unchanged.

### III. Observable Quality

**Assessment**: PASS

`uf doctor` reports an observable, machine-parseable
`manifest-install` comparison (pin vs installed version) in
place of the removed lock check, and `make plugin-test` fails
closed if a lock file appears after the run.

### IV. Testability

**Assessment**: PASS

Every flipped behavior carries a test: lock absence at
scaffold, ignored-lock status, foreign-lock survival through
`uf init`, `manifest-install` outcomes including the malformed
manifest, and the lockless install command pins.

### V. Security by Default

**Assessment**: PASS (trade-off documented)

The lock's integrity hashes and its transitive-tree freeze are
gone. The accepted risk is recorded under Impact; direct pins
and the `toml`/`uuid` security overrides remain, lifecycle
scripts stay disabled on every install path, and transitive
vulnerability scanning moves to a throwaway-lock CI step
(design Risks).

### VI. Deterministic Execution

**Assessment**: PASS (trade-off documented)

Installs are no longer byte-for-byte reproducible without a
lock. Determinism is bounded by the exact direct pins plus the
security overrides, and the residual transitive drift is
recorded as a risk with its mitigations.

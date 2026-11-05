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
same command.

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

### D5. Keep the historical rollback record

`internal/scaffold/testdata/plugin-package-rollback.json` and
`TestPluginPackageInputs_RollbackPrerequisites` document past
baseline lock SHAs. They stay as a historical record; nothing
deploys or verifies those locks any more.

## Risks / Trade-offs

- **Transitive drift.** Transitive versions other than `toml` and
  `uuid` can change between installs. This is mitigated by the exact
  direct pins, the security overrides, `--ignore-scripts`, and the
  provider-free import and tool-definition probes, which still gate
  activation.
- **Lock-based scanners.** OSV-Scanner and Dependabot can no longer
  scan the npm transitive tree from a lock in this repository.
  `package.json` remains scannable.
- **No byte-for-byte reproducible install** in hero repositories.
  Moving to bundled plugin sources would restore it, but is out of
  scope here.

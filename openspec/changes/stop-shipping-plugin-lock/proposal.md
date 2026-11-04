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

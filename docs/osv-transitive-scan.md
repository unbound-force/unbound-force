# OSV Transitive Scan for Plugin Suites

This repository (and any hero repository scaffolded by `uf init`) ships
no committed `.opencode/package-lock.json`. Without a lock,
OSV-Scanner and Dependabot lose lock-based transitive dependency
scanning: Dependabot sees no lockfile to watch, and a manifest-only
OSV scan resolves only direct pins.

The `stop-shipping-plugin-lock` change compensates with a
**throwaway-lock scan** in `ci_security.yml` (the
`osv_plugin_transitive` job): a pinned npm builds an ephemeral lock,
osv-scanner reads it, and the lock is discarded. The scan has already
caught vulnerable pins in practice (zod 4.1.8, toml 4.2.0 — both now
pinned to fixed versions with an override forcing the nested
`@opencode-ai/plugin` copy onto the fixed zod).

`uf init` deploys manifests, plugin sources, commands, and agents —
never workflow jobs. Hero repos following the rollout must copy this
job themselves. This page is that copyable recipe. The canonical job
lives in `unbound-force/unbound-force`
`.github/workflows/ci_security.yml`; keep your copy aligned with it
when it changes.

## The job

```yaml
  osv_plugin_transitive:
    name: OSV-Scanner (plugin transitive)
    permissions:
      contents: read
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1  # v7.0.1

      - name: Set up Go
        uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e  # v7.0.0
        with:
          go-version-file: 'go.mod'

      - name: Set up Node 22
        uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1  # v7.1.0
        with:
          node-version: 22

      - name: Generate throwaway lock
        # npm pin: node 22's bundled npm 10.9.9 crashes in Arborist
        # building the lockless vitest 5 peer graph. If your repo also
        # runs plugin tests through a pinned npm (like Makefile
        # PLUGIN_TEST_NPM in the meta repo), keep the two in sync.
        run: cd .opencode && npx -y npm@11.6.2 install --package-lock-only --ignore-scripts --no-audit --no-fund

      - name: Scan plugin transitive dependencies
        # setup-go v7 pins GOTOOLCHAIN=local (from go.mod), but
        # osv-scanner v2.7.0 requires go >= 1.27.0. The explicit floor
        # pin encodes the pinned scanner's declared toolchain
        # requirement under repo control; bump it in lockstep with the
        # osv-scanner pin below (paired-pin convention, like the npm
        # pin above).
        run: |
          export GOTOOLCHAIN=go1.27.0
          go install github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.0
          "$(go env GOPATH)/bin/osv-scanner" scan source --lockfile=.opencode/package-lock.json
```

## Gotchas

- **Why `--package-lock-only`**: the lock is generated, never
  installed from, and never committed. `.opencode/.gitignore` (kept
  in sync by `uf init`) ignores `package-lock.json`, so the scan
  leaves no residue. Do not "fix" the job by committing the lock.
- **Why the npm pin (`npx -y npm@11.6.2`)**: node 22's bundled npm
  10.9.9 crashes in Arborist resolving the lockless vitest 5 peer
  graph. The pin is a twin of this repository's `Makefile`
  `PLUGIN_TEST_NPM`; if your repo pins npm anywhere else, keep every
  occurrence on the same version.
- **Why `GOTOOLCHAIN=go1.27.0` and not `auto`**: setup-go v7 sets
  `GOTOOLCHAIN=local` from the repo's go.mod, which breaks `go
  install` for any tool whose floor exceeds it. `auto` would silently
  download whatever Go is newest. The explicit floor pin encodes the
  pinned scanner's declared requirement (osv-scanner v2.7.0 declares
  go >= 1.27.0) under repo control. Bump it when you bump the
  osv-scanner pin — they are a pair — and on Go toolchain patch
  security releases. Downloads are verified through
  `sum.golang.org` either way; the pin is about determinism, not
  integrity.
- **`go-version-file` presumes a Go repo.** If your hero repo has no
  go.mod (the scanner is Go-built but your project need not be Go),
  replace it with an explicit floor, e.g. `go-version: '1.27.x'`, and
  drop `GOTOOLCHAIN` or keep it — with a matching setup-go version it
  becomes a no-op belt-and-suspenders.
- **Action pins are SHAs on purpose** (supply-chain posture: tags are
  mutable). Resolve your own trusted SHAs for `actions/checkout`,
  `actions/setup-go`, and `actions/setup-node` through your org's
  normal review rather than trusting this page's copies; the pattern
  to copy is "pin by commit SHA, comment the version".
- **Least privilege**: the job needs only `contents: read`. Do not
  grant `security-events: write` unless you add SARIF upload, and do
  not add SARIF upload without a security-tab retention decision.
- **Exact tool pins**: `osv-scanner@v2.7.0` is a version pin, not a
  mutable branch. Bump it deliberately, with the GOTOOLCHAIN floor,
  and re-review the scan output.

## Running it locally

The same scan, without the CI wrapper:

```bash
(cd .opencode && npx -y npm@11.6.2 install --package-lock-only --ignore-scripts --no-audit --no-fund)
GOTOOLCHAIN=auto go install github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.0
"$(go env GOPATH)/bin/osv-scanner" scan source --lockfile=.opencode/package-lock.json
rm .opencode/package-lock.json
```

(`auto` is fine interactively: your local `go` will resolve the
tool's floor once; CI pins the floor for reproducibility.)

## Scope note

This scan covers the `.opencode/` plugin suite only. Repo-level
dependency scanning (Go modules, GitHub Actions, containers) stays
with whatever your repo already runs — in the meta repo that is the
reusable `vuln_scan` and Scorecards jobs in the same workflow file.
The meta repo also repeats this transitive job on the daily schedule
(`ci_scheduled.yml`), so newly-disclosed advisories surface without
a push; hero repos copying the recipe should wire their copy into
their own scheduled workflow the same way.

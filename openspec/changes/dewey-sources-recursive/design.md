## Context

`generateDeweySources()` in `scaffold.go` creates a
multi-repo Dewey sources config with per-repo entries
(`disk-<repo>`) and an org-level entry (`disk-org`).
The per-repo entries correctly index each sibling repo
individually. The `disk-org` entry points to `../` and
recursively indexes everything, duplicating the
per-repo work and hitting embedding context limits on
large files.

## Goals / Non-Goals

### Goals
- Add `recursive: false` to the `disk-org` entry
- Make `uf init --force` regenerate `sources.yaml`
- Overwrite customized `sources.yaml` on force

### Non-Goals
- Changing per-repo `disk-<repo>` entries (they should
  remain recursive to index all repo contents)
- Adding chunking/truncation to Dewey (separate
  concern -- Dewey repo issue)
- Fixing the UUID collision issue (Dewey issue #17)

## Decisions

### D1: `recursive: false` on disk-org only

Add a single line after the `path` config for `disk-org`:
```yaml
recursive: false
```

Per-repo entries (`disk-<repo>`) stay recursive -- the
user wants full indexing of each sibling repo. Only
the org-level entry should be non-recursive because
its purpose is to pick up top-level design documents,
not to re-index repos.

### D2: Force regeneration overrides customization

Currently `generateDeweySources()` checks
`isDefaultSourcesConfig()` and skips if the file has
been customized. On `--force`, bypass this check and
regenerate the file entirely.

Implementation: add a `force bool` parameter to
`generateDeweySources()`. When `force=true`, skip the
`isDefaultSourcesConfig` check. The caller in the
force block passes `true`; the caller in the first-run
block passes `false` (preserving existing behavior).

### D3: Call order in force block

In the `opts.Force` block of `initSubTools()`, call
`generateDeweySources` BEFORE `dewey index` so the
updated sources config is used for the re-index:

```
} else if opts.Force {
    generateDeweySources(opts, true)  // regenerate
    dewey index                       // re-index with new config
}
```

## Risks / Trade-offs

### Risk: Force overwrites user-customized sources

If a user has manually added custom Dewey sources and
runs `uf init --force`, their custom sources will be
overwritten with the auto-detected config.

**Mitigation**: `--force` is an explicit opt-in that
already overwrites `opencode.json` and re-indexes.
The user accepts that force means "reset to defaults."
They can re-add custom sources after force.

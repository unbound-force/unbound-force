<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add JSONC and port parsing helpers

- [x] 1.1 In `internal/sandbox/config.go`, add
  `stripJSONComments()` to remove `//` and `/* */`
  comments from JSONC input, preserving string contents.
- [x] 1.2 In `internal/sandbox/config.go`, add
  `stripTrailingCommas()` to remove trailing commas
  before `]` or `}`, preserving string contents.
- [x] 1.3 In `internal/sandbox/config.go`, add
  `portMapping` struct with `host` and `container` fields.
- [x] 1.4 In `internal/sandbox/config.go`, add
  `parsePortEntry()` using `(portMapping, bool)` ok-idiom.
  Handle JSON numbers and `"host:container"` strings.
- [x] 1.5 In `internal/sandbox/config.go`, add
  `parseDevcontainerPorts()` that reads devcontainer.json
  via `opts.ReadFile`, strips JSONC, unmarshals
  `forwardPorts`, validates ranges, and deduplicates
  against `excludePorts` and a `seen` map.

## 2. Integrate into sandbox builders

- [x] 2.1 In `internal/sandbox/config.go`, call
  `parseDevcontainerPorts()` from `buildRunArgs()` after
  `DefaultServerPort`, excluding `DefaultServerPort`.
- [x] 2.2 In `internal/sandbox/podman.go`, call
  `parseDevcontainerPorts()` from
  `buildPersistentRunArgs()` after demo ports, excluding
  `DefaultServerPort` and all demo ports.

## 3. Add tests [P]

- [x] 3.1 `TestParseDevcontainerPorts_HappyPath`
- [x] 3.2 `TestParseDevcontainerPorts_NoFile`
- [x] 3.3 `TestParseDevcontainerPorts_NoForwardPorts`
- [x] 3.4 `TestParseDevcontainerPorts_StringPorts`
- [x] 3.5 `TestParseDevcontainerPorts_JSONC`
- [x] 3.6 `TestParseDevcontainerPorts_InvalidRange`
- [x] 3.7 `TestParseDevcontainerPorts_AllExcluded`
- [x] 3.8 `TestParseDevcontainerPorts_TrailingComma`
- [x] 3.9 `TestParseDevcontainerPorts_DuplicateHostPorts`
- [x] 3.10 `TestBuildPersistentRunArgs_DevcontainerPorts`
- [x] 3.11 `TestBuildPersistentRunArgs_DevcontainerPortsAbsentFile`
- [x] 3.12 `TestBuildRunArgs_DevcontainerPorts`
- [x] 3.13 `TestBuildRunArgs_DevcontainerHostContainerMapping`
- [x] 3.14 `TestBuildPersistentRunArgs_DevcontainerHostContainerMapping`
- [x] 3.15 `TestBuildPersistentRunArgs_DevcontainerDemoDedup`

## 4. Update documentation [P]

- [x] 4.1 Add `CHANGELOG.md` entry under Unreleased/Fixed
- [x] 4.2 Add port forwarding note to `docs/cli-reference.md`
  under `uf sandbox init`, `uf sandbox create`, and
  `uf sandbox start` sections
- [x] 4.3 Add note to `docs/configuration.md` sandbox
  configuration section

## 5. Verification

- [x] 5.1 Run `go test -race -count=1 ./internal/sandbox/...`
- [x] 5.2 Run `make check` (lint + test + build)
- [x] 5.3 Verify constitution alignment

<!-- spec-review: retroactive -->
<!-- code-review: passed -->

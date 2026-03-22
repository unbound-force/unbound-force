# Contract: CLI Binary Names and Help Output

**Date**: 2026-03-22
**Branch**: `013-binary-rename`

## Binary Names

| Name | Type | Purpose |
|------|------|---------|
| `unbound-force` | Primary binary | Canonical name, used in Homebrew, `go install`, formal docs |
| `uf` | Symlink alias | Daily-use shorthand, identical behavior |

## Help Output Format

Both `unbound-force --help` and `uf --help` produce
identical output:

```
Unbound Force specification framework toolkit (alias: uf)

Usage:
  unbound-force [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  doctor      Diagnose the Unbound Force development environment
  help        Help about any command
  init        Scaffold specification framework into current directory
  setup       Install and configure the Unbound Force development tool chain
  version     Print the unbound-force version

Flags:
  -h, --help      help for unbound-force
  -v, --version   version for unbound-force

Use "unbound-force [command] --help" for more information about a command.
```

### Changes from previous output

- Root command `Use` field: `unbound` → `unbound-force`
- Description: Added `(alias: uf)` suffix
- All self-references in flags/usage: `unbound` →
  `unbound-force`

## Subcommand Help

Each subcommand's help references the parent as
`unbound-force`:

```
unbound-force init --help
```

produces:

```
Initialize the Unbound Force specification framework...

Usage:
  unbound-force init [flags]
...
```

## Makefile Targets

```makefile
install:
	go build -o $(GOPATH)/bin/unbound-force ./cmd/unbound-force/
	ln -sf $(GOPATH)/bin/unbound-force $(GOPATH)/bin/uf
```

## GoReleaser Config

```yaml
builds:
  - id: unbound-force
    binary: unbound-force
    main: ./cmd/unbound-force/
    ...

archives:
  - name_template: >-
      unbound-force_{{ .Version }}_{{ .Os }}_{{ .Arch }}

homebrew_casks:
  - name: unbound-force
    ...
    hooks:
      post:
        install: |
          if OS.mac?
            system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/unbound-force"]
          end
```

Note: The `uf` symlink for Homebrew casks requires
either a `binary` stanza with `target:` parameter or
a post-install hook. Research the correct cask syntax
during T017.

## Version Output Format

`unbound-force version` and `unbound-force --version`
produce:

```
unbound-force vVERSION (commit COMMIT, built DATE)
```

`unbound-force --version` (flag) produces:

```
unbound-force version VERSION
```

## Doctor Hint Strings

All doctor output hints reference `uf`:

| Old Hint | New Hint |
|----------|----------|
| `Run: unbound init` | `Run: uf init` |
| `Run: unbound setup` | `Run: uf setup` |
| `Fix: unbound setup` | `Fix: uf setup` |
| `Then run: unbound setup` | `Then run: uf setup` |

## Setup Progress Messages

All setup progress messages reference `uf`:

| Old Message | New Message |
|------------|-------------|
| `Running unbound init...` | `Running uf init...` |

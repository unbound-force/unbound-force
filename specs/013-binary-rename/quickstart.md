# Quickstart: Binary Rename

**Date**: 2026-03-22
**Branch**: `013-binary-rename`

## Overview

After this change, the Unbound Force CLI is invoked as
`unbound-force` (canonical) or `uf` (alias) instead of
the bare `unbound` name that collides with the NLnet
Labs DNS resolver.

## For Developers Using the CLI

### Installing

```bash
# Via Homebrew (installs both unbound-force and uf)
brew install unbound-force/tap/unbound-force

# Via go install (binary only, no uf symlink)
go install github.com/unbound-force/unbound-force/cmd/unbound-force@latest
# Then manually create the alias:
ln -sf $(go env GOPATH)/bin/unbound-force $(go env GOPATH)/bin/uf
```

### Daily Usage

```bash
# Initialize a project
uf init

# Check environment health
uf doctor

# Install missing tools
uf setup

# Check version
uf version
```

Both `uf` and `unbound-force` work identically:

```bash
uf init              # same as:
unbound-force init   # this
```

### If You Had the Old Binary

If you previously installed the `unbound` binary via
`go install`, remove the stale binary:

```bash
rm $(go env GOPATH)/bin/unbound
```

Then install the new one:

```bash
go install ./cmd/unbound-force/
```

## For Contributors

### Building and Installing Locally

```bash
# Build and install with alias
make install

# This produces:
#   $GOPATH/bin/unbound-force
#   $GOPATH/bin/uf -> unbound-force (symlink)
```

### Running Tests

```bash
make test
# or: go test -race -count=1 ./...
```

### Checking for Stale References

After the rename, verify no stale `unbound` CLI
references remain in living docs:

```bash
grep -rn 'unbound init\|unbound doctor\|unbound setup\|unbound version\|cmd/unbound/\|scaffolded by unbound' \
  AGENTS.md README.md .opencode/ internal/ .github/ Makefile .goreleaser.yaml
```

This should return zero matches. All completed specs
under `specs/` (except `specs/013-binary-rename/`) and
archived OpenSpec changes under
`openspec/changes/archive/` are excluded as historical
records.

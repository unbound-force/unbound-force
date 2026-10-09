# Sibling Repository Schema

Version 1 defines the closed `.uf/sibling-repos.yaml` declaration
used by the Divisor review policy. It permits at most 32 uniquely
named siblings, bounded candidate paths, credential-free GitHub
HTTPS origins, and bounded repository-relative contract globs.
Runtime validation also enforces UTF-8 byte limits and rejects
duplicate sibling names.

The `acquire_sibling_evidence` tool takes an empty object. It reads the
project declaration and returns version 1 structured JSON with:

- one lexical-by-name result per sibling;
- availability, immutable commit, source mode, evidence metadata, and
  structured rejection or unavailable reasons; and
- a `prompt` string containing accepted file content only between exact
  `<!-- uf-untrusted-sibling-evidence:v1 -->` and
  `<!-- /uf-untrusted-sibling-evidence -->` delimiters.

The boundary never interprets sibling content as instructions. It
verifies the repository root, clean state, normalized origin,
commit, symlink confinement, file hashes, and the 20-file,
256-KiB-per-file, and 1-MiB-total limits before returning prompt
evidence. `contract-only` never uses the network. `clone-on-review`
disables credential prompts and records `offline-cache` when a
verified existing candidate is used after fetch or clone failure.

Positive and negative fixtures exercise the structural schema.
Dedicated Go and TypeScript tests cover semantic uniqueness, byte
bounds, acquisition, and rendering behavior that JSON Schema cannot
express by itself.

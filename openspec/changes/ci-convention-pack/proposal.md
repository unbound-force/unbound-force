## Why

AI agents and humans can write GitHub Actions `uses:` references
with invalid or hallucinated commit SHAs. These pass all local
validation (YAML lint, Go tests, formatting) but fail at runtime
when GitHub cannot resolve the SHA. This happened in
[complytime/.github#114](https://github.com/complytime/.github/pull/114)
where `actions/setup-node` was pinned to a nonexistent SHA.

No convention pack covers CI workflow authoring today. The Divisor
SRE agent references `[PACK]` CI/CD guidance that no pack provides,
causing it to fall back to "universal checks only" on every review.
Existing CI rules in `~/.config/opencode/coding-standards.md` are
invisible to the Divisor `[PACK]` mechanism -- they exist in the
instruction chain but not in the convention pack system.

This change centralizes all CI workflow conventions into a canonical
`ci.md` convention pack and trims the duplicated section from the
user-level `coding-standards.md` config.

## What Changes

1. **New `ci.md` canonical convention pack** -- always-deployed
   (language-agnostic), covering action pinning & supply chain,
   workflow structure, permissions & secrets, and reusable workflow
   design.
2. **New `ci-custom.md` companion** -- user-owned stub for
   project-specific CI rules, following the established `-custom`
   pattern.
3. **Scaffold engine updates** -- `shouldDeployPack()` and
   `collectDeployedPacks()` include the new pack in the
   always-deployed set.
4. **Pack validator exemption** -- `ci.md` uses a different H2
   structure than coding packs, exempt from the 6-section
   requirement (like `severity.md` and `content.md`).
5. **User config trim** -- `~/.config/opencode/coding-standards.md`
   "YAML / GitHub Actions Workflows" section replaced with a
   reference to the `ci.md` pack.

## Capabilities

### New Capabilities

- `ci.md convention pack`: Canonical CI workflow authoring rules
  including SHA verification (CI-002), action pinning (CI-001),
  version comments (CI-003), naming conventions, permissions,
  secrets, and reusable workflow design.
- `ci-custom.md`: Project-specific CI rule extension point.

### Modified Capabilities

- `shouldDeployPack()`: Recognizes `ci` and `ci-custom` as
  always-deployed packs.
- `collectDeployedPacks()`: Includes `ci.md` and `ci-custom.md`
  in the candidates list for all projects.

### Removed Capabilities

- None. The `coding-standards.md` CI section is trimmed, not
  removed -- replaced with a reference to the pack.

## Impact

- **Scaffold engine** (`internal/scaffold/scaffold.go`): Two
  functions gain `ci`/`ci-custom` entries.
- **Scaffold tests** (`internal/scaffold/scaffold_test.go`):
  Asset manifest, `isToolOwned`, `shouldDeployPack`, and
  `collectDeployedPacks` tests updated.
- **Pack validator tests** (`internal/schemas/packvalidator_test.go`):
  `ci.md` added to the skip list for structural validation.
- **Embedded assets** (`internal/scaffold/assets/opencode/uf/packs/`):
  Two new files.
- **Canonical sources** (`.opencode/uf/packs/`): Two new
  byte-identical copies for drift detection.
- **User config** (`~/.config/opencode/coding-standards.md`):
  CI section trimmed.
- **All consumer repos**: `uf init` will deploy the new pack
  on next run, auto-updating AGENTS.md and CLAUDE.md.
- **Divisor SRE agent**: `[PACK]` CI/CD references will now
  resolve to actual rules without code changes to the agent.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The convention pack is a self-describing artifact with
frontmatter metadata (pack_id, language, version). It is
consumed by Divisor agents through the established `[PACK]`
mechanism without requiring synchronous interaction. No
inter-hero communication changes.

### II. Composability First

**Assessment**: PASS

The pack is always-deployed but imposes no mandatory
dependencies. Projects without GitHub Actions workflows
simply have unused rules (same as `content.md` in non-content
projects). The pack follows the established `-custom`
extension pattern for project-specific overrides.

### III. Observable Quality

**Assessment**: PASS

The pack includes versioned YAML frontmatter (`pack_id`,
`language`, `version`) conforming to the established schema.
It supports machine-parseable consumption by the Divisor
`[PACK]` mechanism and maintains provenance metadata through
the scaffold version marker system.

### IV. Testability

**Assessment**: PASS

All scaffold changes are covered by existing test patterns:
drift detection (`TestEmbeddedAssets_MatchSource`,
`TestCanonicalSources_AreEmbedded`), deployment filtering
(`TestShouldDeployPack`), pack collection
(`TestCollectDeployedPacks_*`), and tool ownership
(`TestIsToolOwned`). The pack validator exemption follows the
established pattern for non-coding packs.

### V. Security by Default

**Assessment**: PASS

This change directly advances Principle V. The CI-002 rule
(verify SHAs at authoring time) enforces supply chain
integrity for CI pipelines. CI-001 (pin to commit SHAs)
prevents mutable tag attacks. CI-020/CI-022 codify least
privilege and secret hygiene. The pack makes these security
properties structural rather than review-time afterthoughts.

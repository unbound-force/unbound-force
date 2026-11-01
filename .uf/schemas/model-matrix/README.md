# Model Matrix Schema

The model matrix is the canonical configuration for capability-driven
content-agent model selection in `uf_delegate_command`. Version 1
defines capability profiles, convention-pack mappings, task-kind
defaults, agent pins, and the classifier configuration block. The
tracked file (`.uf/model-matrix.yaml`) carries structure only;
per-developer model identities arrive via the gitignored override.

The schema validates the closed structural contract. The runtime Zod
surface in `.opencode/lib/model-matrix.ts` enforces the identical
input plus the cross-reference rule, and `loadModelMatrix`
additionally enforces pack-file existence and the tracked-role
population rule — runtime checks beyond plain JSON Schema
expressiveness. The cross-validation harness
(`.opencode/test/model-matrix-schema.test.ts`) validates every
sample fixture against both surfaces with per-family pinned
verdicts, so the two representations cannot drift.

## Files

- `v1.schema.json` defines the closed version 1 contract
  (`additionalProperties: false` at every object level).
- `samples/valid-starter-matrix.json` mirrors the tracked starter
  matrix: the four starter capabilities with structure-only profiles,
  the starter pack mappings, task-kind defaults, and an empty
  classifier.
- `samples/valid-minimal-matrix.json` proves only `version` and
  `profiles` are required — every other root key is optional.
- `samples/invalid-unknown-root-field.json` carries an unknown root
  key (closed input).
- `samples/invalid-unknown-defaults-kind.json` carries a `defaults`
  key outside the closed set `plan`, `tasks`, `implement`.
- `samples/invalid-profile-count.json` declares 33 profiles, one
  over the bound of 32.
- `samples/invalid-xref-absent-profile.json` maps a pack to the
  undeclared capability `spec-authoring` — schema-valid, rejected by
  the Zod surface's cross-reference rule.
- `samples/invalid-loader-nonexistent-pack.json` maps the `ruby`
  pack, which has no `.opencode/uf/packs/ruby.md` file — valid on
  both parse surfaces, rejected by the loader's pack-existence
  check.
- `samples/invalid-override-structural-key.json` is an override
  document carrying the structural key `packs`.
- `samples/invalid-override-classifier-enabled.json` is an override
  document carrying `classifier.enabled`.
- `samples/loader-rejection-tracked-llm-values.json` is a tracked
  matrix with populated `classifier.llm` values — schema-valid and
  Zod-valid by design (the merged view legitimately populates those
  paths via the override), rejected by the loader's tracked-role
  rule.

Fixture families are pinned by name prefix in the harness: `valid-*`
passes both surfaces; `invalid-*` (bare) fails both; `invalid-xref-*`
passes `v1.schema.json` and fails the Zod surface only;
`invalid-loader-*` and `loader-rejection-*` pass both parse surfaces
and fail only `loadModelMatrix`; `invalid-override-*` must make
`validateModelMatrixOverride` throw.

## Closed Surface

- `version` — const `1`.
- `profiles` — 1 to 32 entries keyed `^[a-z][a-z0-9-]{0,31}$`; each
  entry MAY carry `model` (a `provider/model` identity string,
  3-256 characters) and `variant`.
- `packs` — pack name (`^[a-z][a-z0-9-]{0,63}$`) to capability; the
  loader requires a pack file at `.opencode/uf/packs/<pack>.md` for
  every key.
- `defaults` — the closed task-kind key set `plan`, `tasks`,
  `implement`, each mapping to a declared capability.
- `agents` — at most 32 agent pins; `divisor-*` names are rejected
  (reviewer-agent selection is governed by the review matrix).
- `classifier` — `enabled` (boolean), `agent`, `agent_timeout_ms`
  (bounded at 1,800,000), `llm` (`endpoint`, `model`, `apiKey`,
  `request_timeout_ms` bounded at 180,000), and
  `confidence_threshold` (0 to 1).

## Override File

Per-developer model and variant customization is supported via
`.uf/model-matrix.override.yaml`. This file is gitignored and allows
individual developers to supply model identities and classifier
engine values without modifying the tracked matrix.

**Constraints (whitelist):**

- Only `profiles.*.model`, `profiles.*.variant`, and
  `classifier.llm.*` paths are permitted.
- Any other key (e.g., `packs`, `defaults`, `agents`,
  `classifier.enabled`) causes a validation failure naming the
  whitelist, preventing silent governance weakening.
- The override is merged via nested-path merge — absent keys fall
  through to the tracked matrix, only capabilities already declared
  may be overridden, and the merged result re-validates against the
  identical schema.

## Loader-Level Rules

Three rules live beyond plain JSON Schema and are enforced at
runtime:

- **Cross-reference** (Zod surface, `parseModelMatrix`): every
  `packs`, `defaults`, and `agents` value must name a capability
  declared under `profiles`.
- **Pack-file existence** (`loadModelMatrix`): every `packs` key
  must have a `.opencode/uf/packs/<pack>.md` file.
- **Tracked-role population** (MM-FR-001, `loadModelMatrix`): the
  tracked matrix carries `classifier.llm` structure only — populated
  values belong in the gitignored override and fail the tracked-file
  load with the override path named as the remediation.

## Starter Matrix: `mechanical` Reachability

The `mechanical` capability is explicit-profile-only in the starter
matrix (design D9, MM-FR-001): no `packs` entry or `defaults` kind
maps to it, and the classifier resolves through `packs`, so it is
reachable only via an explicit `profile: mechanical` parameter or an
`agents:` pin naming it. Adding an implicit selection path to it is
a tracked-matrix change, never an override concern.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1 | 2026-11-01 | Initial closed capability contract: profiles, packs, defaults, agents, classifier. |

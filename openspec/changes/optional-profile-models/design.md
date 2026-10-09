## Context

PR #642 introduced `.uf/review-matrix.yaml` with a `profiles` section that maps tier names (lightweight, standard, heavy) to concrete provider/model strings. This forces all developers on a repository to use the same model provider, which is impractical when team members use different providers.

The review-dispatch plugin (`internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`) already has two code paths for model resolution:
- **Explicit**: model is set from a profile definition → `invoke_agent` uses that model directly
- **Host**: model is `null` → `invoke_agent` resolves the parent session's model at runtime

This change makes profile models optional (preserving the host fallback) and adds a gitignored override file for per-developer customization.

## Goals / Non-Goals

### Goals
- Allow the `model` field in each profile entry to be omitted; when absent, runs use the parent agent's model
- Provide a `.uf/review-matrix.override.yaml` mechanism for per-developer model customization (gitignored)
- Preserve full backward compatibility — existing matrices with explicit models continue to work
- Maintain provenance traceability in the HIC envelope

### Non-Goals
- Changing the tier structure (lightweight/standard/heavy) — tiers remain a team-wide agreement
- Allowing the `profiles` section itself to be entirely absent (keeps tier definitions required)
- A multi-file merge strategy beyond single-level nested-path merge
- Environment variable or CLI-based override mechanisms

## Decisions

### Decision 1: Make `ProfileSchema.model` optional rather than make entire `profiles` section optional

**Rationale**: The tier structure (lightweight/standard/heavy) and agent-to-tier mappings (`defaults`, `always`) are team-level policy decisions that belong in tracked configuration. Only the concrete model strings are provider-specific and need flexibility. Making just `model` optional preserves the team's tier agreement while allowing per-developer model choice.

**Alternative considered**: Making the entire `profiles` section optional. This would cascade through `defaults`, `always`, `advisor`, `fullPanelRuns`, and `addAugmentation` — each needing to handle the null case. Higher complexity for minimal gain.

### Decision 2: Nested-path merge for override file

**Rationale**: The override file is constrained to `profiles.*.model` and `profiles.*.variant` only (whitelist). This prevents silent weakening of team governance — developers cannot override `defaults`, `always`, tier structure, or other cross-profile policy. The surface is deliberately narrow: only the concrete model/variant strings that vary per developer/provider are overridable.

**Nested-path merge**: The override is parsed as YAML and validated against a whitelist before merging — non-whitelisted keys cause a validation failure. The override is walked for `profiles.<tier>.<profile>.[model, variant]` paths, and each matching leaf value replaces its counterpart in the base matrix. Only these specific nested key paths are overridable — the top-level `profiles` key in the override is never used as a wholesale replacement. Key paths absent in the override fall through to the base. The merged result is re-validated against the full schema.

### Decision 3: Override file is gitignored, not an environment variable

**Rationale**: A file-based override aligns with the existing pattern (`.uf/config.yaml` is already gitignored). Environment variables would be harder to discover and harder to tie to a specific repository. A file is also trivially testable.

### Decision 4: `profilePair()` returns `model: null` when model is not set in profile

**Rationale**: This is the existing signal for "use host model." No new sentinel values or union types needed. `invoke_agent` already handles `model: undefined` by resolving the parent session's model, which is exactly the desired behavior.

### Decision 5: Hike the review-matrix schema version

**Rationale**: Making `model` optional is a backward-compatible schema change, but the inference behavior changes. Bumping the version from `2` to `3` signals that a matrix written for v2 semantics (where all profiles had explicit models) is interpreted differently under v3. Existing v2 matrices continue to work because optional fields default to missing, and the v2→v3 inference is a no-op when `model` is present.

### Decision 6: Scaffold template and docs updated

**Rationale**: The scaffolded `review-matrix.yaml` in `internal/scaffold/assets/uf/` serves as the canonical template. It should demonstrate the optional-model pattern. The existing tracked `.uf/review-matrix.yaml` at repo root should also be updated to show the new optionality.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Developer forgets to create override file, runs use wrong provider | Validation warns when a profile has no model and no override is present (soft gate) |
| Override file drifts from base schema | Merge validates the result against the full schema; any drift causes a clear parse error |
| Nested-path merge is too narrow for future needs | Extensible by design — if broader merge is needed later, it can be added without breaking the nested-path merge contract |
| Schema version bump could cause confusion | Version change is documented in the matrix file comment; old v2 files are valid v3 files
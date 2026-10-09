## Why

The `.uf/review-matrix.yaml` file introduced in PR #642 hardcodes provider-specific model strings (e.g., `opencode-go/deepseek-v4-pro`) in the `profiles` section. Different developers working the same repository use different model providers, making committed model strings impractical. The `profiles` section needs to be flexible enough that a team can agree on the **tier structure** without forcing everyone onto the same provider.

## What Changes

1. **Model field becomes optional within each profile entry** — When `model` is omitted, review runs fall back to the parent agent's model (the existing `source: "host"` path with `model: null`). The tier structure (`lightweight`, `standard`, `heavy`) and agent-tier mappings (`defaults`, `always`) remain honored; only the concrete model resolution defers to the host.

2. **Add support for `.uf/review-matrix.override.yaml`** — A gitignored file that nested-path merges with the tracked `.uf/review-matrix.yaml`. Only whitelisted fields (`profiles.*.model` and `profiles.*.variant`) in the override replace their counterparts in the base. Non-whitelisted keys cause a validation failure, preventing silent governance weakening. This allows individual developers to supply their own model strings without modifying tracked configuration.

## Capabilities

### New Capabilities
- `profile-model-optional`: The `model` field within each profile entry is now optional; runs without an explicit model use the parent session's model.
- `review-matrix-override`: A `.uf/review-matrix.override.yaml` file, gitignored, allows per-developer overrides of profile model and variant fields (`profiles.*.model` and `profiles.*.variant`) only.

### Modified Capabilities
- `review-dispatch-planning`: `profilePair()` returns `model: null` when a profile has no model set, causing `invoke_agent` to resolve the parent model. Advisor agents without explicit models also fall back to host resolution.

### Removed Capabilities
- `ProfileSchema.model` is no longer required — this is a backward-compatible relaxation (existing matrices with models continue to work, new matrices may omit the field). No runtime capabilities are removed.

## Impact

### Documentation Impact
- `schemas/review-matrix/` — new `v3.schema.json`, update `README.md`
- `.uf/review-matrix.yaml` — tracked config with optional-model comments
- `internal/scaffold/assets/uf/review-matrix.yaml` — update scaffold template
- `docs/configuration.md` — add `.uf/review-matrix.override.yaml` to
  Review Council Configuration table
- `CHANGELOG.md` — add entry for optional model field and override file
- Cross-repo: website team should document override file workflow for
  multi-provider teams

### Code Impact
- **Schema change**: `ProfileSchema.model` becomes `.optional()` (`.opencode/plugins/review-dispatch/index.ts` and `internal/scaffold/assets/` equivalent)
- **Behavior change**: `profilePair()` returns `model: null` when model is not set
- **New file support**: `loadPolicies()` checks for override file and performs nested-path merge; override whitelist validates only `profiles.*.model`/`.variant`
- **.gitignore**: Add `.uf/review-matrix.override.yaml` to gitignored patterns
- **Backward compatible**: Existing matrices with explicit models continue to work unchanged

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.3.0).

### I. Autonomous Collaboration

**Assessment**: PASS

The override file is a well-defined artifact consumed by the review-dispatch plugin. It uses the same YAML schema as the base matrix. No synchronous inter-hero communication is introduced. The override mechanism is an extension point, not a coupling point.

### II. Composability First

**Assessment**: PASS

The override file is entirely optional — the review-dispatch plugin functions correctly without it. Each developer can deploy the plugin independently and configure their own models. No new mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

The review-dispatch HIC envelope already captures provenance (model, variant, run IDs). When models are resolved from the host session rather than an explicit profile, the envelope's `resolved_parent_model` field captures the actual model used — maintaining full traceability.

### IV. Testability

**Assessment**: PASS

`profilePair()` is a pure function testable in isolation. The override merge logic will be similarly testable. Coverage strategy: unit tests for `profilePair()` null-model return, merge function, override whitelist validation, override read/parse failure mode, and integration tests for the full dispatch flow with and without overrides. Target: >= 90% line coverage on changed plugin files, >= 80% on overall plugin suite (Constitution IV).

### V. Security by Default

**Assessment**: PASS

The override file is gitignored and never committed — no credentials or provider keys leak into version control. Input validation already exists for the matrix schema; the merge step will re-validate the merged result against the same schema.
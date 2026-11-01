import { z } from "zod"

import { parseYamlDocument, type YamlParser } from "./reviewer-manifest.js"

// ── Constants ────────────────────────────────────────────────

/** Repository-relative location of the tracked model matrix (design D1, D8). */
export const MODEL_MATRIX_PATH = ".uf/model-matrix.yaml"

/** Repository-relative location of the gitignored per-developer override (design D4, MM-FR-002). */
export const MODEL_MATRIX_OVERRIDE_PATH = ".uf/model-matrix.override.yaml"

/** Repository-relative directory whose `<pack>.md` files back `packs` keys (MM-FR-001). */
export const PACKS_DIRECTORY = ".opencode/uf/packs"

/** Closed task-kind key set the `defaults` mapping accepts (MM-FR-001). */
export const DEFAULT_TASK_KINDS = ["plan", "tasks", "implement"] as const

/** Inferred union of the closed `defaults` task-kind key set. */
export type DefaultTaskKind = (typeof DEFAULT_TASK_KINDS)[number]

/** Maximum number of capability profiles (MM-FR-001: 1-32 entries). */
const MAX_PROFILES = 32

/** Maximum number of agent pins, mirroring the review-matrix agent-map bound. */
const MAX_AGENT_PINS = 32

// ── Selection audit types (design D7, MM-FR-007) ──────────────

/** The eight-value selection provenance enum recorded on every delegation. */
export const SELECTION_SOURCE_VALUES = [
  "explicit-model",
  "explicit-profile",
  "explicit-pack",
  "agent-pin",
  "default-kind",
  "classifier-agent",
  "classifier-llm",
  "host",
] as const

/** Inferred union of the selection provenance enum. */
export type SelectionSource = (typeof SELECTION_SOURCE_VALUES)[number]

/**
 * Resolved model selection for one delegation, consumed by the selection
 * audit trail (MM-FR-007). `model` carries the combined `provider/model`
 * identity string (split into a `ModelIdentity` by the delegation tool);
 * a null `model` with `selection_source: "host"` records host fallback.
 */
export interface SelectionResult {
  /** Which precedence-chain step produced this selection (MM-FR-003). */
  readonly selection_source: SelectionSource
  /**
   * Selected, classified, or explicitly supplied convention pack, when one
   * participated. An explicitly supplied pack that passes validation rides
   * along even when a higher-precedence parameter supplies the model
   * (MM-FR-003: never silently dropped).
   */
  readonly pack: string | null
  /** Resolved capability name, when selection resolved a profile. */
  readonly capability: string | null
  /** Resolved `provider/model` identity, or null for host fallback. */
  readonly model: string | null
  /** Resolved model variant, when the profile or call carried one. */
  readonly variant: string | null
  /** Classifier confidence, when a classifier engine participated. */
  readonly confidence: number | null
  /** Why a fallback occurred (classifier failure, profile without a model). */
  readonly fallback_reason: string | null
}

// ── Closed v1 Zod schema (design D1/D3, MM-FR-001) ────────────
//
// Every object level is `.strict()` so unknown keys fail validation: the
// Zod surface must stay identical to `schemas/model-matrix/v1.schema.json`,
// and the cross-validation harness asserts identical verdicts on fixtures.

const CapabilityNameSchema = z.string().regex(/^[a-z][a-z0-9-]{0,31}$/)
const ProfileModelSchema = z
  .string()
  .min(3)
  .max(256)
  .regex(/^[A-Za-z0-9][A-Za-z0-9._-]*\/[A-Za-z0-9][A-Za-z0-9._@-]*(?:\/[A-Za-z0-9][A-Za-z0-9._@-]*)*$/)
const ProfileVariantSchema = z.string().regex(/^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/)
const PackNameSchema = z.string().regex(/^[a-z][a-z0-9-]{0,63}$/)
// Negative lookahead keeps reviewer agents out of `agents` pins: divisor-*
// selection is governed by the review matrix, not the model matrix.
const AgentPinNameSchema = z.string().regex(/^(?!divisor-)[a-z][a-z0-9-]{0,63}$/)

const ModelMatrixProfileSchema = z
  .object({
    model: ProfileModelSchema.optional(),
    variant: ProfileVariantSchema.optional(),
  })
  .strict()

const ProfilesSchema = z
  .record(CapabilityNameSchema, ModelMatrixProfileSchema)
  .refine((profiles) => Object.keys(profiles).length >= 1, "at least one profile is required")
  .refine((profiles) => Object.keys(profiles).length <= MAX_PROFILES, `at most ${MAX_PROFILES} profiles are allowed`)

const PacksSchema = z.record(PackNameSchema, CapabilityNameSchema)

// Closed task-kind key set: entries are individually optional so a matrix can
// omit static routing, but `plan`, `tasks`, and `implement` are the only keys
// ever accepted (MM-FR-001; unknown kinds fail with the set named).
const DefaultsSchema = z
  .object({
    plan: CapabilityNameSchema.optional(),
    tasks: CapabilityNameSchema.optional(),
    implement: CapabilityNameSchema.optional(),
  })
  .strict()

const AgentsSchema = z
  .record(AgentPinNameSchema, CapabilityNameSchema)
  .refine((pins) => Object.keys(pins).length <= MAX_AGENT_PINS, `at most ${MAX_AGENT_PINS} agent pins are allowed`)

// Classifier structure only (MM-FR-001): the tracked matrix carries no model
// strings — per-developer values under `classifier.llm` arrive exclusively
// via the gitignored override (design D4), and the merged result re-validates
// against this same schema, so the keys are closed while values stay strings.
// The `{env:VAR}`-only apiKey rule is enforced by the LLM engine (MM-FR-006).
const ClassifierLlmSchema = z
  .object({
    endpoint: z.string().min(1).max(2048).optional(),
    model: z.string().min(1).max(256).optional(),
    apiKey: z.string().min(1).max(512).optional(),
    request_timeout_ms: z.number().int().positive().max(180_000).optional(),
  })
  .strict()

const ClassifierSchema = z
  .object({
    enabled: z.boolean().optional(),
    agent: z
      .string()
      .regex(/^[a-z][a-z0-9-]{0,63}$/)
      .optional(),
    agent_timeout_ms: z.number().int().positive().max(1_800_000).optional(),
    llm: ClassifierLlmSchema.optional(),
    confidence_threshold: z.number().min(0).max(1).optional(),
  })
  .strict()

/** Closed Zod schema for the v1 model matrix surface (mirrors v1.schema.json). */
export const ModelMatrixSchema = z
  .object({
    version: z.literal(1),
    profiles: ProfilesSchema,
    packs: PacksSchema.optional(),
    defaults: DefaultsSchema.optional(),
    agents: AgentsSchema.optional(),
    classifier: ClassifierSchema.optional(),
  })
  .strict()
  .superRefine((matrix, context) => {
    // Cross-reference validation: every mapped capability must be an own
    // property of `profiles`. Object.hasOwn (not `in`) so prototype-chain
    // names such as "constructor" can never satisfy a reference.
    for (const [pack, capability] of Object.entries(matrix.packs ?? {})) {
      if (!Object.hasOwn(matrix.profiles, capability)) {
        context.addIssue({ code: "custom", path: ["packs", pack], message: `unknown profile ${capability}` })
      }
    }
    for (const [kind, capability] of Object.entries(matrix.defaults ?? {})) {
      if (!Object.hasOwn(matrix.profiles, capability)) {
        context.addIssue({ code: "custom", path: ["defaults", kind], message: `unknown profile ${capability}` })
      }
    }
    for (const [agent, capability] of Object.entries(matrix.agents ?? {})) {
      if (!Object.hasOwn(matrix.profiles, capability)) {
        context.addIssue({ code: "custom", path: ["agents", agent], message: `unknown profile ${capability}` })
      }
    }
  })

/** Parsed and closed-validated model matrix document. */
export type ModelMatrix = z.infer<typeof ModelMatrixSchema>

/** One capability-named model slot (`model` and `variant` both optional). */
export type ModelMatrixProfile = z.infer<typeof ModelMatrixProfileSchema>

// ── Canonical no-op matrix (MM-FR-004) ────────────────────────

/**
 * Canonical empty matrix for absent-content documents. A present-but-empty
 * or comments-only `.uf/model-matrix.yaml` is a VALID no-op (MM-FR-004), so
 * parsing resolves it to this value and every precedence step falls through
 * to the host model. Frozen because it is shared by reference across loads
 * (convention pack AP-004 immutability).
 */
export const EMPTY_MODEL_MATRIX: ModelMatrix = Object.freeze({
  version: 1,
  profiles: Object.freeze({}),
  packs: Object.freeze({}),
  defaults: Object.freeze({}),
  agents: Object.freeze({}),
  classifier: Object.freeze({}),
} satisfies ModelMatrix)

// ── Parsing ───────────────────────────────────────────────────

/** Extracts a message from an unknown rejection. */
function errorText(error: unknown): string {
  return error instanceof Error ? error.message : "unknown error"
}

/**
 * Formats model-matrix validation issues for the fail-closed error path.
 * Augments `defaults` unrecognized-key failures with the accepted task-kind
 * set: MM-FR-001 requires the failure to name the accepted key set, and
 * zod's default message names only the offending key.
 * @param error The Zod validation error.
 * @returns Sorted, "; "-joined issue descriptions.
 */
function formatModelMatrixIssues(error: z.ZodError): string {
  return error.issues
    .map((issue) => {
      const location = issue.path.join(".") || "<root>"
      if (issue.code === "unrecognized_keys" && issue.path[0] === "defaults") {
        return `${location}: ${issue.message} (accepted task kinds: ${DEFAULT_TASK_KINDS.join(", ")})`
      }
      return `${location}: ${issue.message}`
    })
    .sort()
    .join("; ")
}

/**
 * Parses and closed-validates a model matrix document. Empty and
 * comments-only documents (YAML null) resolve to the canonical no-op matrix
 * per MM-FR-004; any other document that fails the closed v1 surface throws.
 * @param text YAML or JSON matrix text.
 * @param parser Injected YAML parser.
 * @returns The validated matrix, or the frozen no-op matrix for empty input.
 * @throws When the document parses to a non-null value that fails validation.
 */
export function parseModelMatrix(text: string, parser: YamlParser): ModelMatrix {
  const document = parseYamlDocument(text, "model matrix", parser)
  if (document === null || document === undefined) {
    return EMPTY_MODEL_MATRIX
  }
  const result = ModelMatrixSchema.safeParse(document)
  if (!result.success) {
    throw new Error(`validate model matrix: ${formatModelMatrixIssues(result.error)}`)
  }
  return result.data
}

// ── Override (design D4, MM-FR-002) ──────────────────────────

/**
 * The fail-closed override whitelist (MM-FR-002). Every override validation
 * error names this set so a misconfigured override states its own remedy.
 */
const OVERRIDE_WHITELIST = "profiles.*.model, profiles.*.variant, classifier.llm.*"

/** Narrows a parsed YAML value to a non-array object (the override walk shape). */
function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

/**
 * Validates a parsed override document against the fail-closed whitelist
 * `profiles.*.model`, `profiles.*.variant`, `classifier.llm.*` (design D4,
 * MM-FR-002). Only per-developer identity paths are permitted: structural
 * keys (`packs`, `defaults`, `agents`, `classifier.enabled`, `classifier.agent`,
 * ...) are tracked-matrix surface, so accepting them would let a gitignored
 * file silently change governance. Any nested key under `classifier.llm` is
 * admitted here — that subtree is exclusively per-developer — while the
 * closed schema still bounds the accepted keys at merged-result
 * re-validation, a deliberate two-layer defense.
 * @param override Parsed override document (null/undefined for empty files).
 * @throws When any key or nested path is not whitelisted.
 */
export function validateModelMatrixOverride(override: unknown): void {
  // Empty and comments-only override documents parse to null/undefined; an
  // object without keys carries no paths. All three are valid no-ops.
  if (override === null || override === undefined) {
    return
  }
  if (!isRecord(override)) {
    throw new Error(`model matrix override must be an object — only ${OVERRIDE_WHITELIST} paths are allowed`)
  }
  if (Object.keys(override).length === 0) {
    return
  }
  for (const key of Object.keys(override)) {
    if (key !== "profiles" && key !== "classifier") {
      throw new Error(
        `model matrix override key "${key}" is not permitted — only ${OVERRIDE_WHITELIST} paths are allowed`,
      )
    }
  }
  validateOverrideProfiles(override.profiles)
  validateOverrideClassifier(override.classifier)
}

/**
 * Walks the override's `profiles` subtree: only `model` and `variant` leaves
 * are whitelisted per capability (MM-FR-002).
 * @param profiles The parsed `profiles` value (null/undefined when absent).
 * @throws When a profile entry or leaf key is not whitelisted.
 */
function validateOverrideProfiles(profiles: unknown): void {
  if (profiles === null || profiles === undefined) {
    return // absent profiles is a no-op
  }
  if (!isRecord(profiles)) {
    throw new Error(`model matrix override profiles must be an object — only ${OVERRIDE_WHITELIST} paths are allowed`)
  }
  for (const capability of Object.keys(profiles)) {
    const profile = profiles[capability]
    if (profile === null || profile === undefined) {
      continue // an empty capability entry is a no-op
    }
    if (!isRecord(profile)) {
      throw new Error(`model matrix override profiles.${capability} must be an object`)
    }
    for (const key of Object.keys(profile)) {
      if (key !== "model" && key !== "variant") {
        throw new Error(
          `model matrix override key "profiles.${capability}.${key}" is not permitted — only ${OVERRIDE_WHITELIST} paths are allowed`,
        )
      }
    }
  }
}

/**
 * Walks the override's `classifier` subtree: everything under `classifier`
 * is structural and tracked except `llm`, whose nested keys are exclusively
 * per-developer (design D4), so any key under `classifier.llm` is admitted
 * here and the closed schema bounds the accepted set at merged-result
 * re-validation.
 * @param classifier The parsed `classifier` value (null/undefined when absent).
 * @throws When a classifier key outside `llm` is present.
 */
function validateOverrideClassifier(classifier: unknown): void {
  if (classifier === null || classifier === undefined) {
    return // absent classifier is a no-op
  }
  if (!isRecord(classifier)) {
    throw new Error(`model matrix override classifier must be an object — only ${OVERRIDE_WHITELIST} paths are allowed`)
  }
  for (const key of Object.keys(classifier)) {
    if (key !== "llm") {
      throw new Error(
        `model matrix override key "classifier.${key}" is not permitted — only ${OVERRIDE_WHITELIST} paths are allowed`,
      )
    }
  }
  const llm = classifier.llm
  if (llm === null || llm === undefined) {
    return // absent llm is a no-op
  }
  if (!isRecord(llm)) {
    throw new Error("model matrix override classifier.llm must be an object")
  }
}

/**
 * Performs the nested-path merge of a whitelisted override into the base
 * matrix (design D4, MM-FR-002): each override leaf replaces its base
 * counterpart, absent paths fall through to the base, and an explicit
 * `null` model or variant becomes an absent key — host fallback for models.
 * The `profiles` map is never wholesale-replaced, and only capabilities
 * already defined in the base may be overridden: adding capabilities is a
 * tracked-matrix change, not a per-developer one. The merged result is
 * re-validated against the identical schema via a JSON round-trip, which
 * also returns a sanitized clone with plain prototypes. Callers MUST run
 * {@link validateModelMatrixOverride} first so non-whitelisted keys fail
 * closed instead of being silently skipped by the walk.
 * @param base The validated base matrix.
 * @param override The parsed, whitelist-validated override document.
 * @returns The merged, re-validated matrix, or `base` unchanged when the
 *   override carries no whitelisted entries.
 * @throws When the override names a capability absent from the base, or the
 *   merged result fails the closed schema.
 */
export function mergeOverride(base: ModelMatrix, override: Record<string, unknown>): ModelMatrix {
  const profileOverrides = isRecord(override.profiles) ? override.profiles : null
  const classifierOverride = isRecord(override.classifier) ? override.classifier : null
  const llmOverrides = classifierOverride !== null && isRecord(classifierOverride.llm) ? classifierOverride.llm : null

  const hasProfileEntries = profileOverrides !== null && Object.keys(profileOverrides).length > 0
  const hasLlmEntries = llmOverrides !== null && Object.keys(llmOverrides).length > 0
  if (!hasProfileEntries && !hasLlmEntries) {
    return base // no whitelisted entries — a no-op (MM-FR-002)
  }

  const merged: Record<string, unknown> = { ...base }

  if (profileOverrides !== null) {
    const mergedProfiles: Record<string, unknown> = Object.create(null)
    for (const capability of Object.keys(base.profiles)) {
      mergedProfiles[capability] = { ...base.profiles[capability] }
    }
    for (const capability of Object.keys(profileOverrides)) {
      const profileOverride = profileOverrides[capability]
      if (!isRecord(profileOverride)) {
        continue // null and non-object entries carry no leaves (validated upstream)
      }
      // Own-property check so prototype-chain names such as "constructor" can
      // never satisfy the capability reference (the loadPolicies pattern).
      if (!Object.hasOwn(base.profiles, capability)) {
        throw new Error(
          `model matrix override profiles.${capability} is not a defined capability in the base matrix — only existing capabilities may be overridden`,
        )
      }
      const mergedProfile: Record<string, unknown> = { ...base.profiles[capability] }
      if ("model" in profileOverride) {
        // Null propagates host fallback: optional Zod fields accept undefined,
        // never null (Bun.YAML null-for-empty semantics).
        mergedProfile.model = profileOverride.model ?? undefined
      }
      if ("variant" in profileOverride) {
        mergedProfile.variant = profileOverride.variant ?? undefined
      }
      mergedProfiles[capability] = mergedProfile
    }
    merged.profiles = mergedProfiles
  }

  if (llmOverrides !== null) {
    const baseClassifier = isRecord(base.classifier) ? base.classifier : {}
    const baseLlm = isRecord(baseClassifier.llm) ? baseClassifier.llm : {}
    const mergedLlm: Record<string, unknown> = Object.create(null)
    for (const key of Object.keys(baseLlm)) {
      mergedLlm[key] = baseLlm[key]
    }
    for (const key of Object.keys(llmOverrides)) {
      mergedLlm[key] = llmOverrides[key] ?? undefined
    }
    merged.classifier = { ...baseClassifier, llm: mergedLlm }
  }

  // Re-validate the merged result against the identical schema (MM-FR-002).
  // The JSON round-trip sanitizes the clone — null-prototype records become
  // plain objects and undefined-valued keys drop — before the closed surface
  // re-validates, so override typos fail closed at load.
  return parseModelMatrix(JSON.stringify(merged), (text: string) => JSON.parse(text) as unknown)
}

// ── Loader ────────────────────────────────────────────────────

/** Type guard for ENOENT filesystem errors (the loadPolicies convention). */
function isEnoent(error: unknown): boolean {
  return error instanceof Error && "code" in error && (error as NodeJS.ErrnoException).code === "ENOENT"
}

/**
 * Collects `packs` keys whose pack file is missing under `PACKS_DIRECTORY`.
 * @param matrix The validated matrix.
 * @param fileExists Injected existence checker for project-relative paths.
 * @returns Sorted names of packs without a `<pack>.md` file.
 */
function missingPackFiles(matrix: ModelMatrix, fileExists: (relativePath: string) => boolean): readonly string[] {
  const missing: string[] = []
  for (const pack of Object.keys(matrix.packs ?? {})) {
    if (!fileExists(`${PACKS_DIRECTORY}/${pack}.md`)) {
      missing.push(pack)
    }
  }
  return missing.sort()
}

/** `classifier.llm` leaves the tracked-role population rule governs (MM-FR-001). */
const TRACKED_LLM_KEYS = ["endpoint", "model", "apiKey", "request_timeout_ms"] as const

/**
 * Collects populated `classifier.llm` leaves of a parsed matrix document for
 * the tracked-role population rule (MM-FR-001): the tracked matrix carries
 * `classifier.llm` structure only, so any populated value in the tracked
 * file fails load with the override file named as the remediation. The
 * check is role-aware by necessity — the identical schema accepts populated
 * values in merged views (override values legitimately populate those
 * paths, MM-FR-002), so it runs only against the pre-merge tracked
 * document, never against a merged result.
 * @param matrix The parsed tracked-file matrix, before any override merge.
 * @returns Sorted names of populated `classifier.llm` leaves.
 */
function populatedTrackedLlmLeaves(matrix: ModelMatrix): readonly string[] {
  const llm = matrix.classifier?.llm
  if (llm === undefined) {
    return []
  }
  return TRACKED_LLM_KEYS.filter((key) => llm[key] !== undefined).sort()
}

/** Injectable dependencies for loading the model matrix (the loadPolicies pattern). */
export interface ModelMatrixDependencies {
  /** Reads a project-relative file; absence MUST reject with an ENOENT-coded error. */
  readonly readText: (relativePath: string) => Promise<string>
  /** Narrow YAML parsing boundary supplied by Bun in production, tests in isolation. */
  readonly parseYaml: YamlParser
  /** Reports whether a project-relative path exists (pack existence, MM-FR-001). */
  readonly fileExists: (relativePath: string) => boolean
}

/** Discriminated load outcome for the tracked matrix (design D8 zero-config probe). */
export type ModelMatrixLoadResult =
  | { readonly status: "absent" }
  | { readonly status: "ok"; readonly matrix: ModelMatrix }
  | { readonly status: "invalid"; readonly error: string }

/**
 * Loads the tracked model matrix and applies the gitignored per-developer
 * override when one is present (design D4). An absent matrix resolves to the
 * zero-config `absent` state (host model, no override read, MM-FR-004); a
 * present matrix that fails the closed schema, carries populated
 * `classifier.llm` values in the tracked file (the tracked-role population
 * rule, MM-FR-001), names packs without pack
 * files, or pairs with an override that violates the whitelist or fails
 * merged-result re-validation resolves to `invalid` with an error naming the
 * path, accepted key set, whitelist, or offending pack keys (fail-closed).
 * Never throws: every failure is carried in the result so the delegation
 * tool has one uniform failure path.
 * @param dependencies Injected read, parse, and existence seams.
 * @returns The discriminated tristate load result.
 */
export async function loadModelMatrix(dependencies: ModelMatrixDependencies): Promise<ModelMatrixLoadResult> {
  let text: string
  try {
    text = await dependencies.readText(MODEL_MATRIX_PATH)
  } catch (error: unknown) {
    if (isEnoent(error)) {
      return { status: "absent" }
    }
    return { status: "invalid", error: `cannot read model matrix ${MODEL_MATRIX_PATH}: ${errorText(error)}` }
  }

  let matrix: ModelMatrix
  try {
    matrix = parseModelMatrix(text, dependencies.parseYaml)
  } catch (error: unknown) {
    return { status: "invalid", error: `model matrix ${MODEL_MATRIX_PATH} is invalid: ${errorText(error)}` }
  }

  // Tracked-role population rule (MM-FR-001, task 2.4): the tracked file
  // carries `classifier.llm` structure only. The check runs against the
  // tracked document BEFORE the override merge — merged views legitimately
  // populate those paths (MM-FR-002) — and separately from Zod schema
  // validation, which accepts populated values by design so the merged
  // result can re-validate against the identical schema. The no-op matrix
  // (comments-only file) resolves to the canonical empty classifier and
  // stays valid (MM-FR-004).
  const populatedLlm = populatedTrackedLlmLeaves(matrix)
  if (populatedLlm.length > 0) {
    return {
      status: "invalid",
      error:
        `model matrix ${MODEL_MATRIX_PATH} is invalid: tracked classifier.llm carries populated values ` +
        `(${populatedLlm.join(", ")}) — per-developer engine configuration belongs in ` +
        `${MODEL_MATRIX_OVERRIDE_PATH}; keep the tracked file structure only`,
    }
  }

  const missing = missingPackFiles(matrix, dependencies.fileExists)
  if (missing.length > 0) {
    const details = missing.map((pack) => `${PACKS_DIRECTORY}/${pack}.md`).join(", ")
    return {
      status: "invalid",
      error: `model matrix ${MODEL_MATRIX_PATH} is invalid: packs key(s) ${missing.join(", ")} have no pack file (${details})`,
    }
  }

  // Per-developer override (design D4): an absent override keeps the base; a
  // present one must pass the whitelist, merge by nested path, and survive
  // re-validation against the identical schema. The override can never carry
  // `packs` keys, so the pack-existence check above stays valid for the merged
  // result. Structural override failures fail closed exactly like base-matrix
  // failures, and a whitelist violation throws before any merge runs, so no
  // merged configuration is produced (MM-FR-002).
  let overrideText: string | null = null
  try {
    overrideText = await dependencies.readText(MODEL_MATRIX_OVERRIDE_PATH)
  } catch (error: unknown) {
    if (!isEnoent(error)) {
      return {
        status: "invalid",
        error: `cannot read model matrix override ${MODEL_MATRIX_OVERRIDE_PATH}: ${errorText(error)}`,
      }
    }
    // Override file does not exist — the base matrix is used as-is.
  }

  if (overrideText !== null) {
    try {
      const document = parseYamlDocument(overrideText, "model matrix override", dependencies.parseYaml)
      if (document !== null && document !== undefined) {
        validateModelMatrixOverride(document)
        matrix = mergeOverride(matrix, document as Record<string, unknown>)
      }
    } catch (error: unknown) {
      return {
        status: "invalid",
        error: `model matrix override ${MODEL_MATRIX_OVERRIDE_PATH} is invalid: ${errorText(error)}`,
      }
    }
  }

  return { status: "ok", matrix }
}

// ── Selection precedence (design D2, MM-FR-003/MM-FR-004) ──────

/** Default classifier confidence threshold when the matrix omits one (design D5). */
export const DEFAULT_CONFIDENCE_THRESHOLD = 0.6

/** Which classifier engine produced a winning cascade result (MM-FR-005 order). */
export type ClassifierEngine = "agent" | "llm"

/**
 * Pre-computed classifier cascade outcome supplied to {@link resolveSelection}.
 * Selection is a pure function (Constitution IV, design D5), so the cascade's
 * engines — the hidden-subagent spawn and the LLM fetch — run before
 * resolution and participate only as this value: the cascade (task 3.2)
 * supplies the winning engine, pack, and confidence, or a not-run/failed
 * state, and {@link resolveSelection} applies matrix-validity and threshold
 * gating to whatever it receives.
 */
export type ClassifierOutcome =
  /** The cascade did not run (static task kind, classifier disabled, or zero-config). */
  | { readonly status: "not-run"; readonly reason?: string }
  /** Every engine failed; `reason` names the engines attempted and why each failed. */
  | { readonly status: "failed"; readonly reason: string }
  /** A winning result: which engine produced it, its pack, and its confidence. */
  | {
      readonly status: "selected"
      readonly engine: ClassifierEngine
      readonly pack: string
      readonly confidence: number
    }

/**
 * Selection parameters as supplied by one delegation call (design D2).
 * Parameter shapes are validated at the tool boundary (the delegation
 * schema's model, variant, profile, and pack parameters); this module treats
 * any non-undefined value as present.
 */
export interface SelectionInputs {
  /** Explicit `model` parameter — a `provider/model` identity that wins outright (step 1). */
  readonly model?: string
  /** Explicit `variant` parameter, paired with `model` (step 1). */
  readonly variant?: string
  /** Explicit `profile` parameter naming a matrix capability (step 2). */
  readonly profile?: string
  /** Explicit `pack` parameter naming a `packs` key (step 3). */
  readonly pack?: string
  /** Delegation `checkpoint_step` task kind: `plan`, `tasks`, `implement`, or another/absent value. */
  readonly checkpoint_step?: string
  /** Delegated agent name, consulted for the `agents:` pin (step 4). */
  readonly agent?: string
}

/** Stable failure codes for non-retryable selection validation failures (design D2). */
export const SELECTION_FAILURE_CODES = ["MATRIX_NOT_CONFIGURED", "UNKNOWN_PROFILE", "INVALID_PACK"] as const

/** Inferred union of the selection failure codes. */
export type SelectionFailureCode = (typeof SELECTION_FAILURE_CODES)[number]

/**
 * Discriminated resolution outcome for one delegation's model selection: the
 * resolved {@link SelectionResult} on success, or a non-retryable failure
 * carrying a stable code and a message that names the offending value and its
 * remediation — the `SKILL_NOT_FOUND` envelope pattern. Validation failures
 * never resolve on retry, so the delegation tool fails the call pre-spawn
 * with zero child sessions (design D2).
 */
export type SelectionResolution =
  | { readonly ok: true; readonly selection: SelectionResult }
  | {
      readonly ok: false
      readonly code: SelectionFailureCode
      readonly error: string
      readonly retryable: false
    }

/** Attempt outcome for a selected classifier result (internal to step 6). */
type ClassifierAttempt =
  | { readonly resolved: true; readonly result: SelectionResolution }
  | { readonly resolved: false; readonly fallbackReason: string }

/**
 * Builds a host-fallback selection (MM-FR-003 step 7): no matrix model
 * resolved, so the delegation tool introspects the parent session. The pack
 * and capability stay recorded when a precedence step resolved them, keeping
 * the fallback auditable (MM-FR-007).
 * @param pack Selected or classified pack, when one participated.
 * @param capability Resolved capability, when a step resolved one.
 * @param fallbackReason Why host fallback occurred, when applicable.
 * @param confidence Classifier confidence, when an accepted classifier
 *   result resolved a model-less profile.
 * @returns The host-fallback selection resolution.
 */
function hostSelection(
  pack: string | null,
  capability: string | null,
  fallbackReason: string | null,
  confidence: number | null = null,
): SelectionResolution {
  return {
    ok: true,
    selection: {
      selection_source: "host",
      pack,
      capability,
      model: null,
      variant: null,
      confidence,
      fallback_reason: fallbackReason,
    },
  }
}

/**
 * Resolves a capability through its profile (MM-FR-003: steps 2-6 resolve a
 * capability first, then its profile). A profile carrying a model yields the
 * given `source`; a model-less profile falls to host with the capability and
 * pack still recorded for the audit trail. Closed-schema validation
 * guarantees every mapped capability is an own property of `profiles`, so
 * the lookup cannot miss.
 * @param matrix The configured (non-empty) matrix.
 * @param capability The capability a precedence step resolved.
 * @param source The selection source recorded when the profile carries a model.
 * @param pack The pack that produced the capability, when one participated.
 * @param fallbackReason Prior fallback context (classifier failure) to
 *   record alongside the profile outcome.
 * @returns The resolved or host-fallback selection.
 */
function capabilitySelection(
  matrix: ModelMatrix,
  capability: string,
  source: SelectionSource,
  pack: string | null,
  fallbackReason: string | null,
): SelectionResolution {
  const profile = matrix.profiles[capability]
  if (profile.model !== undefined) {
    return {
      ok: true,
      selection: {
        selection_source: source,
        pack,
        capability,
        model: profile.model,
        variant: profile.variant ?? null,
        confidence: null,
        fallback_reason: fallbackReason,
      },
    }
  }
  const noModelReason = `profile "${capability}" carries no model`
  return hostSelection(
    pack,
    capability,
    fallbackReason !== null ? `${fallbackReason}; ${noModelReason}` : noModelReason,
  )
}

/**
 * Applies classifier gating to a selected cascade outcome (MM-FR-005): the
 * pack must be a `packs` key of the configured matrix, and the confidence
 * must meet the configured threshold (default 0.6). An accepted result
 * resolves through the pack's profile — the same double duty as an explicit
 * pack parameter; a discarded result reports the fallback reason for the
 * static-default path so the fallback stays auditable (MM-FR-007).
 * @param matrix The configured matrix supplying `packs` and the threshold.
 * @param outcome The winning cascade outcome.
 * @returns The resolved selection, or the fallback reason when discarded.
 */
function attemptClassifierSelection(
  matrix: ModelMatrix,
  outcome: Extract<ClassifierOutcome, { status: "selected" }>,
): ClassifierAttempt {
  const packs = matrix.packs ?? {}
  if (!Object.hasOwn(packs, outcome.pack)) {
    return {
      resolved: false,
      fallbackReason: `classifier pack "${outcome.pack}" is not a packs key in the configured matrix`,
    }
  }
  const threshold = matrix.classifier?.confidence_threshold ?? DEFAULT_CONFIDENCE_THRESHOLD
  if (outcome.confidence < threshold) {
    return {
      resolved: false,
      fallbackReason: `classifier confidence ${outcome.confidence} is below the threshold ${threshold}`,
    }
  }
  const capability = packs[outcome.pack]
  const profile = matrix.profiles[capability]
  if (profile.model !== undefined) {
    return {
      resolved: true,
      result: {
        ok: true,
        selection: {
          selection_source: outcome.engine === "agent" ? "classifier-agent" : "classifier-llm",
          pack: outcome.pack,
          capability,
          model: profile.model,
          variant: profile.variant ?? null,
          confidence: outcome.confidence,
          fallback_reason: null,
        },
      },
    }
  }
  return {
    resolved: true,
    result: hostSelection(outcome.pack, capability, `profile "${capability}" carries no model`, outcome.confidence),
  }
}

/**
 * Renders the auditable fallback reason for a classifier outcome that did
 * not yield a selection (MM-FR-007: fallbacks are recorded, never silent).
 * @param outcome The not-run or failed cascade outcome.
 * @returns The fallback reason recorded with the static fallback.
 */
function classifierFallbackReason(outcome: ClassifierOutcome): string {
  if (outcome.status === "failed") {
    return `classifier cascade failed: ${outcome.reason}`
  }
  if (outcome.status === "not-run" && outcome.reason !== undefined) {
    return `classifier not run: ${outcome.reason}`
  }
  return "classifier not run"
}

/**
 * Builds a non-retryable selection failure (design D2). The message names
 * the offending value and its remediation — the `SKILL_NOT_FOUND` envelope
 * pattern — so the delegation tool can fail the call pre-spawn with zero
 * child sessions.
 * @param code Stable failure code.
 * @param error Message naming the failure and its remediation.
 * @returns The discriminated failure resolution.
 */
function selectionFailure(code: SelectionFailureCode, error: string): SelectionResolution {
  return { ok: false, code, error, retryable: false }
}

/**
 * Renders a sorted, comma-joined name list for failure remediation text.
 * @param names Capability or pack names to list.
 * @returns The joined list, or "(none)" when empty.
 */
function nameList(names: readonly string[]): string {
  return names.length > 0 ? [...names].sort().join(", ") : "(none)"
}

/**
 * Resolves one delegation's model selection through the fixed precedence
 * chain (design D2, MM-FR-003), stopping at the first step that yields:
 *
 * 1. an explicit `model` (with optional `variant`) — it needs no matrix;
 * 2. an explicit `profile` naming a declared capability;
 * 3. an explicit `pack` that is a `packs` key of the configured matrix;
 * 4. an `agents:` pin for the delegated agent;
 * 5. static `defaults:` by task kind — `plan`/`tasks` resolve statically
 *    and never reach the classifier;
 * 6. the classifier outcome for `implement` and unknown kinds, falling back
 *    to the static `defaults` entry for the kind, else host;
 * 7. host — a resolved profile without a model, or no matrix configured.
 *
 * Steps 2-6 resolve a capability first, then its profile; a resolved profile
 * without a model falls to host (step 7) rather than continuing down the
 * chain. Validation of explicit selection parameters is unconditional, not
 * reach-based: `profile` and `pack` are validated in full, in chain order,
 * before the chain runs, so a winning higher-precedence parameter never
 * shadows an invalid or unknown companion. An explicitly supplied `pack`
 * that passes validation is never silently dropped — it rides along for
 * directive rendering even when `model` or `profile` supplies the child
 * model — and an explicit `model`, `profile`, or `pack` bypasses the
 * classifier cascade entirely (the cascade participates only in chains
 * that reach step 6). The function is pure — no
 * filesystem, clock, network, or classifier invocation; the classifier
 * participates only as the pre-computed {@link ClassifierOutcome}. An
 * unknown explicit profile, an explicit pack that is not a `packs` key, or
 * an explicit `profile`/`pack` with no matrix configured fails closed with
 * a non-retryable error naming the remediation (MM-FR-003/MM-FR-004).
 * @param matrix The loaded matrix, or null when none is configured. The
 *   canonical no-op matrix (comments-only file) counts as not configured.
 * @param inputs The delegation call's selection parameters.
 * @param classifier The pre-computed classifier cascade outcome.
 * @returns The discriminated selection resolution.
 */
export function resolveSelection(
  matrix: ModelMatrix | null,
  inputs: SelectionInputs,
  classifier: ClassifierOutcome,
): SelectionResolution {
  // Normalize the not-configured state (MM-FR-004): an absent matrix (null)
  // and the canonical no-op matrix (comments-only file) behave identically,
  // and the closed schema requires at least one profile for any other
  // matrix — so an empty `profiles` map exactly identifies "no matrix is
  // configured" regardless of which shape the caller passed.
  const configured = matrix !== null && Object.keys(matrix.profiles).length > 0 ? matrix : null

  // Unconditional validation (MM-FR-003): explicit profile and pack
  // parameters are validated in full, in chain order, before the chain
  // runs — a winning higher-precedence parameter never shadows an invalid
  // or unknown companion (model + unknown profile, model + invalid pack,
  // and profile + invalid pack all fail closed pre-spawn, zero child
  // sessions).
  if (inputs.profile !== undefined) {
    if (configured === null) {
      return selectionFailure(
        "MATRIX_NOT_CONFIGURED",
        `profile "${inputs.profile}" cannot be used: no model matrix is configured at ${MODEL_MATRIX_PATH}. ` +
          `Remediation: declare the profile under profiles in ${MODEL_MATRIX_PATH}, or omit the profile parameter; ` +
          "retrying will not fix this.",
      )
    }
    if (!Object.hasOwn(configured.profiles, inputs.profile)) {
      return selectionFailure(
        "UNKNOWN_PROFILE",
        `profile "${inputs.profile}" is not a capability declared in the model matrix. ` +
          `Configured capabilities: ${nameList(Object.keys(configured.profiles))}. ` +
          "Remediation: pass a declared capability name, or omit the profile parameter; retrying will not fix this.",
      )
    }
  }
  if (inputs.pack !== undefined) {
    if (configured === null) {
      return selectionFailure(
        "MATRIX_NOT_CONFIGURED",
        `pack "${inputs.pack}" cannot be used: no model matrix is configured at ${MODEL_MATRIX_PATH}. ` +
          `Remediation: map the pack under packs in ${MODEL_MATRIX_PATH}, or omit the pack parameter; ` +
          "retrying will not fix this.",
      )
    }
    const packs = configured.packs ?? {}
    if (!Object.hasOwn(packs, inputs.pack)) {
      return selectionFailure(
        "INVALID_PACK",
        `pack "${inputs.pack}" is not a packs key in the model matrix. ` +
          `Configured packs: ${nameList(Object.keys(packs))}. ` +
          `Remediation: pass a pack mapped under packs in ${MODEL_MATRIX_PATH}, or omit the pack parameter; ` +
          "retrying will not fix this.",
      )
    }
  }

  // Step 1: an explicit model wins outright and needs no matrix (MM-FR-004).
  // An explicitly supplied, validated pack rides along for directive
  // rendering — never silently dropped (MM-FR-003, UF-FR-002).
  if (inputs.model !== undefined) {
    return {
      ok: true,
      selection: {
        selection_source: "explicit-model",
        pack: inputs.pack ?? null,
        capability: null,
        model: inputs.model,
        variant: inputs.variant ?? null,
        confidence: null,
        fallback_reason: null,
      },
    }
  }

  // Step 2: an explicit profile (validated above) resolves through its
  // profile; an explicitly supplied pack rides along.
  if (inputs.profile !== undefined && configured !== null) {
    return capabilitySelection(configured, inputs.profile, "explicit-profile", inputs.pack ?? null, null)
  }

  // Step 3: an explicit pack (validated above) resolves through its mapped
  // profile. Pack-file existence was already enforced per key at load time
  // (the loader's missingPackFiles check, MM-FR-001), so validation reduced
  // to the key-membership test above.
  if (inputs.pack !== undefined && configured !== null) {
    const packs = configured.packs ?? {}
    return capabilitySelection(configured, packs[inputs.pack], "explicit-pack", inputs.pack, null)
  }

  // Step 4: an `agents:` pin for the delegated agent.
  if (configured !== null && inputs.agent !== undefined) {
    const pins = configured.agents ?? {}
    if (Object.hasOwn(pins, inputs.agent)) {
      return capabilitySelection(configured, pins[inputs.agent], "agent-pin", null, null)
    }
  }

  const kind = inputs.checkpoint_step

  // Step 5: `plan`/`tasks` resolve statically from `defaults` and never
  // reach the classifier (MM-FR-003 scenario: classifier not invoked).
  if (kind === "plan" || kind === "tasks") {
    if (configured !== null) {
      const capability = configured.defaults?.[kind]
      if (capability !== undefined) {
        return capabilitySelection(configured, capability, "default-kind", null, null)
      }
    }
    return hostSelection(null, null, null)
  }

  // Step 6: the classifier cascade for `implement` and unknown kinds
  // (MM-FR-005). With no matrix there is no static default and no classifier
  // participation — plain host, zero-config semantics (MM-FR-004).
  if (configured === null) {
    return hostSelection(null, null, null)
  }
  const attempt: ClassifierAttempt =
    classifier.status === "selected"
      ? attemptClassifierSelection(configured, classifier)
      : { resolved: false, fallbackReason: classifierFallbackReason(classifier) }
  if (attempt.resolved) {
    return attempt.result
  }
  // Static fallback: the `defaults` entry for the kind when present (only
  // `implement` can have one — unknown kinds are outside the closed key set).
  if (kind === "implement") {
    const staticCapability = configured.defaults?.implement
    if (staticCapability !== undefined) {
      return capabilitySelection(configured, staticCapability, "default-kind", null, attempt.fallbackReason)
    }
  }
  return hostSelection(null, null, `${attempt.fallbackReason}; no static default for task kind "${kind ?? "unknown"}"`)
}

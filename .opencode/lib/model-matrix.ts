import { z } from "zod"

import { parseYamlDocument, type YamlParser } from "./reviewer-manifest.js"

// ── Constants ────────────────────────────────────────────────

/** Repository-relative location of the tracked model matrix (design D1, D8). */
export const MODEL_MATRIX_PATH = ".uf/model-matrix.yaml"

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
  /** Selected or classified convention pack, when one participated. */
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
 * Loads the tracked model matrix. An absent file resolves to the zero-config
 * `absent` state (host model, MM-FR-004); a present file that fails the
 * closed schema, or that names packs without pack files, resolves to
 * `invalid` with an error naming the path, accepted key set, or offending
 * pack keys (design D4 fail-closed). Never throws: every failure is carried
 * in the result so the delegation tool has one uniform failure path.
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

  const missing = missingPackFiles(matrix, dependencies.fileExists)
  if (missing.length > 0) {
    const details = missing.map((pack) => `${PACKS_DIRECTORY}/${pack}.md`).join(", ")
    return {
      status: "invalid",
      error: `model matrix ${MODEL_MATRIX_PATH} is invalid: packs key(s) ${missing.join(", ")} have no pack file (${details})`,
    }
  }

  return { status: "ok", matrix }
}

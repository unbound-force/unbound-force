import { Buffer } from "node:buffer"
import { createHash, randomUUID } from "node:crypto"
import { chmod, link, mkdir, open, readFile, readdir, unlink, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join, resolve } from "node:path"

import { tool, type PluginInput, type PluginModule, type ToolContext } from "@opencode-ai/plugin"
import { z } from "zod"

import {
  DEFAULT_TIMEOUT_MILLISECONDS,
  MAX_PROMPT_FILE_BYTES,
  PromptSchema,
  PromptFileSchema,
  TimeoutSchema,
  directModelIdentity,
  errorText,
  executeAgentSession,
  failedResult,
  sanitizeInvocationError,
  validateTimeout,
  type ExecutorDependencies,
  type InvocationProvenance,
  type InvokeAgentResult,
  type ModelIdentity,
} from "../../lib/agent-executor.js"
import { createPrepareLessonLearningTool } from "../../lib/review-dispatch-lesson-proposal.js"
export {
  createPrepareLessonLearningTool,
  lessonDedupeSHA256,
  LESSON_PROPOSAL_BEGIN,
  LESSON_PROPOSAL_END,
  normalizeLessonInformation,
  prepareLessonLearning,
  type DeweyStoreLearningPayload,
  type LessonProposal,
  type LessonProvenance,
  type LessonProvenanceSource,
  type LessonSkipReason,
  type PreparedLessonLearning,
} from "../../lib/review-dispatch-lesson-proposal.js"
import {
  createAcquireSiblingEvidenceTool,
  createSiblingAcquisitionDependencies,
} from "../../lib/review-dispatch-sibling-evidence.js"
export {
  acquireSiblingEvidence,
  createAcquireSiblingEvidenceTool,
  createSiblingAcquisitionDependencies,
  parseSiblingRepositories,
  UNTRUSTED_SIBLING_EVIDENCE_BEGIN,
  UNTRUSTED_SIBLING_EVIDENCE_END,
  type AcquisitionFileInfo,
  type AcquireSiblingEvidenceInput,
  type AcquireSiblingEvidenceResult,
  type CommandOptions,
  type CommandResult,
  type SiblingAcquisitionDependencies,
  type SiblingEvidenceFile,
  type SiblingEvidenceRejection,
  type SiblingEvidenceRepositoryResult,
} from "../../lib/review-dispatch-sibling-evidence.js"
import {
  AgentNameSchema,
  CategorySchema,
  createBunYamlParser,
  formatValidationError,
  parseReviewerManifest,
  parseYamlDocument,
  uniqueArray,
  type ReviewCategory,
  type Reviewer,
  type ReviewerManifest,
  type YamlParser,
} from "../../lib/reviewer-manifest.js"
export { createBunYamlParser, parseReviewerManifest, type YamlParser } from "../../lib/reviewer-manifest.js"

const MATRIX_PATH = ".uf/review-matrix.yaml"
const MATRIX_OVERRIDE_PATH = ".uf/review-matrix.override.yaml"
const MANIFEST_PATH = ".uf/reviewer-capabilities.yaml"
const PLAN_VERSION = 1
const REVIEW_DISPATCH_SCHEMA_VERSION = "1.0.0"
const REVIEW_VERDICT_SCHEMA_VERSION = "2.0.0"
const PRODUCER_VERSION = "1.0.0"
const DISPATCH_ARTIFACT_DIRECTORY = ".uf/artifacts/dispatch"
const MAX_CORRELATION_ATTEMPTS = 32

const DEFAULT_TIER_CAPS = {
  lightweight: 2,
  standard: null,
  heavy: null,
} as const

const DEFAULT_LIMITS = {
  max_personas: 16,
  max_runs_per_persona: 3,
  max_total_runs: 24,
  max_parallel_runs: 4,
  per_run_timeout_seconds: 600,
  max_reported_cost_usd: 25,
  tier_caps: DEFAULT_TIER_CAPS,
} as const

const TIER_VALUES = ["lightweight", "standard", "heavy"] as const
const SOURCE_VALUES = ["explicit", "advisor", "host"] as const
const GENERIC_VERDICT_VALUES = [
  "APPROVE",
  "APPROVE WITH ADVISORIES",
  "REQUEST CHANGES",
  "INCONCLUSIVE",
  "UNAVAILABLE",
] as const
const TERMINAL_STATUS_VALUES = [
  "success",
  "failed",
  "skipped",
  "budget_skipped",
  "limit_skipped",
  "cancelled",
] as const

const ProfileNameSchema = z.string().regex(/^[a-z][a-z0-9-]{0,31}$/)
const ModelSchema = z
  .string()
  .min(3)
  .max(256)
  .regex(
    /^[A-Za-z0-9][A-Za-z0-9._-]*\/[A-Za-z0-9][A-Za-z0-9._@-]*(?:\/[A-Za-z0-9][A-Za-z0-9._@-]*)*$/,
  )
const VariantSchema = z.string().regex(/^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/)
const TierSchema = z.enum(TIER_VALUES)


const ProfileSchema = z
  .object({
    model: ModelSchema.optional(),
    variant: VariantSchema.optional(),
  })
  .strict()

const ProfilesSchema = z
  .record(ProfileNameSchema, ProfileSchema)
  .refine((profiles) => Object.keys(profiles).length >= 1, "at least one profile is required")
  .refine((profiles) => Object.keys(profiles).length <= 32, "at most 32 profiles are allowed")
  .refine(
    (profiles) => TIER_VALUES.every((tier) => profiles[tier] !== undefined),
    "lightweight, standard, and heavy profiles are required",
  )

const DefaultsSchema = z
  .object({
    code: ProfileNameSchema,
    spec: ProfileNameSchema,
    triage: ProfileNameSchema,
    feedback: ProfileNameSchema,
  })
  .strict()

const AgentProfileMapSchema = z
  .record(AgentNameSchema, ProfileNameSchema)
  .refine((mapping) => Object.keys(mapping).length <= 32, "at most 32 agent overrides are allowed")

const AgentArraySchema = uniqueArray(AgentNameSchema, 32)
const ModeAgentMapSchema = z
  .object({
    code: AgentArraySchema.optional(),
    spec: AgentArraySchema.optional(),
    triage: AgentArraySchema.optional(),
    feedback: AgentArraySchema.optional(),
  })
  .strict()

const ProfileRunSchema = z
  .object({
    profile: ProfileNameSchema,
    variant: VariantSchema.optional(),
  })
  .strict()
const ModelRunSchema = z
  .object({
    model: ModelSchema,
    variant: VariantSchema.optional(),
  })
  .strict()
const RunSchema = z.union([ProfileRunSchema, ModelRunSchema])
const AgentRunsSchema = z.array(RunSchema).min(1).max(5)
const ModeRunsSchema = z
  .record(AgentNameSchema, AgentRunsSchema)
  .refine((mapping) => Object.keys(mapping).length <= 32, "at most 32 agents may declare runs")
const RunsSchema = z
  .object({
    code: ModeRunsSchema.optional(),
    spec: ModeRunsSchema.optional(),
    triage: ModeRunsSchema.optional(),
    feedback: ModeRunsSchema.optional(),
  })
  .strict()

const RiskAugmentationSchema = z
  .object({
    enabled: z.boolean(),
    agents: AgentArraySchema,
  })
  .strict()

const TierCapSchema = z.number().int().positive().nullable().optional()
const TierCapsSchema = z
  .object({
    lightweight: TierCapSchema,
    standard: TierCapSchema,
    heavy: TierCapSchema,
  })
  .strict()
  .optional()

const LimitsSchema = z
  .object({
    max_personas: z.number().int().min(1).max(32).optional(),
    max_runs_per_persona: z.number().int().min(1).max(5).optional(),
    max_total_runs: z.number().int().min(1).max(64).optional(),
    max_parallel_runs: z.number().int().min(1).max(8).optional(),
    per_run_timeout_seconds: z.number().int().min(30).max(1800).optional(),
    max_reported_cost_usd: z.number().min(1).max(100).optional(),
    tier_caps: TierCapsSchema,
  })
  .strict()

const ReviewMatrixSchema = z
  .object({
    version: z.union([z.literal(2), z.literal(3)]),
    profiles: ProfilesSchema,
    defaults: DefaultsSchema,
    always: AgentProfileMapSchema.optional(),
    advisor: ModeAgentMapSchema.optional(),
    runs: RunsSchema.optional(),
    risk_augmentation: RiskAugmentationSchema.optional(),
    limits: LimitsSchema.optional(),
  })
  .strict()
  .superRefine((matrix, context) => {
    // Version 3 makes model optional; version 2 requires it on every profile.
    if (matrix.version === 2) {
      for (const [profile, entry] of Object.entries(matrix.profiles)) {
        if (entry.model === undefined) {
          context.addIssue({
            code: "custom",
            path: ["profiles", profile, "model"],
            message: `version 2 requires model on every profile; omit for version 3`,
          })
        }
      }
    }
    for (const [mode, profile] of Object.entries(matrix.defaults)) {
      if (matrix.profiles[profile] === undefined) {
        context.addIssue({
          code: "custom",
          path: ["defaults", mode],
          message: `unknown profile ${profile}`,
        })
      }
    }
    for (const [agent, profile] of Object.entries(matrix.always ?? {})) {
      if (matrix.profiles[profile] === undefined) {
        context.addIssue({
          code: "custom",
          path: ["always", agent],
          message: `unknown profile ${profile}`,
        })
      }
    }
  })


const ChangedFileSchema = z
  .object({
    path: z.string().min(1).max(1024),
    additions: z.number().int().nonnegative(),
    deletions: z.number().int().nonnegative(),
  })
  .strict()

const IssueCommentIDSchema = z.union([
  z.number().int().positive().safe(),
  z.string().regex(/^[0-9]+$/),
])
const IssueCommentSchema = z
  .object({
    id: IssueCommentIDSchema,
    created_at: z.string().min(1).max(64),
    body: z.string(),
  })
  .strict()
const IssueContentSchema = z
  .object({
    title: z.string(),
    body: z.string().nullable(),
    comments: z.array(IssueCommentSchema).max(10_000),
  })
  .strict()

const PlanInputSchema = z
  .object({
    mode: z.enum(["code", "specs", "triage", "feedback", "test"]),
    discovered_agents: uniqueArray(AgentNameSchema, 64),
    full: z.boolean().optional().default(false),
    augment: z.boolean().optional().default(false),
    changed_files: z.array(ChangedFileSchema).max(100_000).optional(),
    issue: IssueContentSchema.optional(),
  })
  .strict()
  .superRefine((input, context) => {
    if (input.mode === "triage" && input.issue === undefined) {
      context.addIssue({ code: "custom", path: ["issue"], message: "triage mode requires issue content" })
    }
    if (input.mode !== "triage" && input.changed_files === undefined) {
      context.addIssue({
        code: "custom",
        path: ["changed_files"],
        message: `${input.mode} mode requires changed files`,
      })
    }
    if (input.mode === "triage" && input.changed_files !== undefined) {
      context.addIssue({
        code: "custom",
        path: ["changed_files"],
        message: "triage mode must not synthesize a diff",
      })
    }
  })

const UUIDSchema = z
  .string()
  .regex(/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/)
const SHA1Schema = z.string().regex(/^[0-9a-f]{40}$/)
const SHA256Schema = z.string().regex(/^[0-9a-f]{64}$/)
const RefSchema = z.string().min(1).max(255).regex(/^[ -~]+$/)
const TimestampSchema = z
  .string()
  .min(20)
  .max(64)
  .refine(
    (value) =>
      /^\d{4}-\d{2}-\d{2}T.*(?:Z|[+-]\d{2}:\d{2})$/.test(value) && Number.isFinite(Date.parse(value)),
    "must be an RFC3339 timestamp with an explicit offset",
  )
const NullableTimestampSchema = TimestampSchema.nullable()
const NullableModelSchema = ModelSchema.nullable()
const NullableVariantSchema = VariantSchema.nullable()
const GenericVerdictSchema = z.enum(GENERIC_VERDICT_VALUES)
const SeveritySchema = z.enum(["CRITICAL", "HIGH", "MEDIUM", "LOW"])
const WorkflowRunVerdictSchema = z.enum([
  "APPROVE",
  "APPROVE WITH ADVISORIES",
  "REQUEST CHANGES",
  "VALID",
  "INVALID",
  "NEEDS-CLARIFICATION",
  "ACCEPT",
  "AUTHOR-DECIDES",
])

function sortedUniqueArray<T extends z.ZodTypeAny>(item: T, minimum: number, maximum: number): z.ZodArray<T> {
  return z
    .array(item)
    .min(minimum)
    .max(maximum)
    .superRefine((values, context) => {
      if (new Set(values).size !== values.length) {
        context.addIssue({ code: "custom", message: "items must be unique" })
      }
      const sorted = [...values].sort((left, right) => String(left).localeCompare(String(right)))
      if (values.some((value, index) => value !== sorted[index])) {
        context.addIssue({ code: "custom", message: "items must be sorted" })
      }
    })
}

const PullRequestContextSchema = z
  .object({
    kind: z.literal("pr"),
    pr_number: z.number().int().min(1).max(999_999),
    base_ref: RefSchema,
    base_sha: SHA1Schema,
    head_ref: RefSchema,
    head_sha: SHA1Schema,
  })
  .strict()
const LocalContextSchema = z
  .object({
    kind: z.literal("local"),
    pr_number: z.null(),
    base_ref: RefSchema,
    base_sha: SHA1Schema,
    head_ref: RefSchema,
    head_sha: SHA1Schema,
    uncommitted: z.boolean().optional(),
  })
  .strict()
const IssueContextSchema = z
  .object({
    kind: z.literal("issue"),
    issue_number: z.number().int().min(1).max(999_999),
    issue_url: z
      .string()
      .max(512)
      .regex(/^https:\/\/github\.com\/[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+\/issues\/[1-9][0-9]{0,5}$/),
    content_sha256: SHA256Schema,
  })
  .strict()
const InputContextSchema = z.discriminatedUnion("kind", [
  PullRequestContextSchema,
  LocalContextSchema,
  IssueContextSchema,
])

const ProfileCategoriesSchema = sortedUniqueArray(CategorySchema, 1, 7)
const DiffChangeProfileSchema = z
  .object({
    kind: z.literal("diff"),
    lines: z.number().int().nonnegative(),
    files: z.number().int().nonnegative(),
    components: z.number().int().nonnegative(),
    security_sensitive: z.boolean(),
    user_facing: z.boolean(),
    categories: ProfileCategoriesSchema,
    tier: TierSchema,
  })
  .strict()
const IssueChangeProfileSchema = z
  .object({
    kind: z.literal("issue"),
    text_bytes: z.number().int().nonnegative(),
    comment_count: z.number().int().nonnegative(),
    content_sha256: SHA256Schema,
    matched_rules: sortedUniqueArray(z.string().min(1).max(128), 0, 128),
    security_sensitive: z.boolean(),
    user_facing: z.boolean(),
    categories: ProfileCategoriesSchema,
    tier: TierSchema,
  })
  .strict()
const ChangeProfileSchema = z.discriminatedUnion("kind", [DiffChangeProfileSchema, IssueChangeProfileSchema])

const DispatchPlanEntrySchema = z
  .object({
    agent: AgentNameSchema,
    decision: z.enum(["include", "skip"]),
    reason_code: z.string().min(1).max(128),
    reason: z.string().min(1).max(4096),
    source: z.enum(SOURCE_VALUES),
    sequence: z.number().int().positive(),
    read_only: z.boolean(),
    tier: TierSchema.nullable(),
    model: NullableModelSchema,
    variant: NullableVariantSchema,
    validation_errors: uniqueArray(z.string().min(1).max(4096), 64),
  })
  .strict()

const UsageSchema = z
  .object({
    cost_usd: z.number().nonnegative().nullable(),
    tokens: z
      .object({
        input: z.number().int().nonnegative(),
        output: z.number().int().nonnegative(),
      })
      .strict()
      .nullable(),
  })
  .strict()
const RunErrorSchema = z
  .object({
    code: z.string().min(1).max(128).regex(/^[a-z][a-z0-9_]*$/),
    message: z.string().min(1).max(4096),
    retryable: z.boolean(),
  })
  .strict()
const RunFindingSchema = z
  .object({
    severity: SeveritySchema,
    category: z.string().min(1).max(128),
    description: z.string().min(1).max(16_384),
    root_cause: z.string().min(1).max(4096),
    file: z.string().min(1).max(1024).nullable(),
    line: z.number().int().positive().nullable(),
  })
  .strict()
const DispatchRunSchema = z
  .object({
    run_id: UUIDSchema,
    agent: AgentNameSchema,
    source: z.enum(SOURCE_VALUES),
    requested_model: NullableModelSchema,
    provider: z.string().min(1).max(128).regex(/^[A-Za-z0-9][A-Za-z0-9._-]*$/).nullable(),
    model_id: z.string().min(1).max(256).regex(/^[A-Za-z0-9][A-Za-z0-9._@\/-]*$/).nullable(),
    variant: NullableVariantSchema,
    resolved_parent_model: NullableModelSchema,
    resolved_parent_variant: NullableVariantSchema,
    sequence: z.number().int().positive(),
    status: z.enum(["pending", "running", ...TERMINAL_STATUS_VALUES]),
    started_at: NullableTimestampSchema,
    finished_at: NullableTimestampSchema,
    usage: UsageSchema.nullable(),
    error: RunErrorSchema.nullable(),
    reported_model: NullableModelSchema,
    model_mismatch: z.boolean(),
    workflow_verdict: WorkflowRunVerdictSchema.nullable(),
    findings: z.array(RunFindingSchema).max(512),
  })
  .strict()

const ConsolidatedFindingSchema = RunFindingSchema.extend({
  run_ids: uniqueArray(UUIDSchema, 64).min(1),
}).strict()
const AdvisorySchema = z
  .object({
    severity: SeveritySchema,
    description: z.string().min(1).max(16_384),
    run_ids: uniqueArray(UUIDSchema, 64).min(1),
  })
  .strict()
const CoverageSchema = z
  .object({
    preflight_verdict: z.enum(["PASS", "FAIL", "SOFT_GATE", "NOT_RUN"]),
    checks_total: z.number().int().nonnegative(),
    checks_passed: z.number().int().nonnegative(),
  })
  .strict()
// ── Submission store for child tool call handoff ─────────────

/** Lesson proposal submitted by a child Divisor agent via submit_lesson_proposal. */
interface LessonProposal {
  readonly information: string
  readonly tag: string
  readonly category?: string
}

/** Data accumulated from child tool calls during a single dispatch_agent_run. */
interface SubmittedData {
  findings: z.infer<typeof RunFindingSchema>[]
  proposals: LessonProposal[]
}

/**
 * Module-level in-process store keyed by child session ID.
 * Populated by submit_review_findings / submit_lesson_proposal tools,
 * harvested by dispatch_agent_run after the child session completes.
 */
const submissionStore = new Map<string, SubmittedData>()

/**
 * Module-level map from parent session ID to correlation ID.
 * dispatch_agent_run auto-generates a correlation_id on the first call from a
 * given parent session and reuses it for all subsequent calls in that session.
 * consolidate_dispatch and dispatch_status look up the correlation_id by
 * context.sessionID so the agent never has to track or pass correlation IDs.
 */
const sessionCorrelationMap = new Map<string, string>()

/** Exported for testing only. */
export function _getSubmissionStore(): Map<string, SubmittedData> {
  return submissionStore
}

/** Exported for testing only. */
export function _getSessionCorrelationMap(): Map<string, string> {
  return sessionCorrelationMap
}

/** Persisted run data written to dispatch session directory. */
interface PersistedRunData {
  readonly correlation_id: string
  readonly run_id: string
  readonly agent: string
  readonly source: string
  readonly sequence: number
  readonly status: string
  readonly started_at: string
  readonly finished_at: string
  readonly requested_model: string | null
  readonly provider: string | null
  readonly model_id: string | null
  readonly variant: string | null
  readonly resolved_parent_model: string | null
  readonly resolved_parent_variant: string | null
  readonly reported_model: string | null
  readonly model_mismatch: boolean
  readonly usage: z.infer<typeof UsageSchema> | null
  readonly error: z.infer<typeof RunErrorSchema> | null
  readonly workflow_verdict: string | null
  readonly findings: z.infer<typeof RunFindingSchema>[]
  readonly proposals: LessonProposal[]
  readonly text: string
  readonly read_only: boolean
}

/** Structured result returned by dispatch_agent_run for agent consumption. */
interface DispatchAgentRunResult extends InvokeAgentResult {
  readonly correlation_id: string
  readonly run_id: string
}

/** Compute filesystem path for a dispatch session directory. */
function dispatchSessionDir(correlationId: string): string {
  return join(tmpdir(), "opencode", `dispatch-${correlationId}`)
}

const SEVERITY_RANK: Record<string, number> = { LOW: 1, MEDIUM: 2, HIGH: 3, CRITICAL: 4 }

const RunCountsSchema = z
  .object({
    total: z.number().int().nonnegative(),
    success: z.number().int().nonnegative(),
    failed: z.number().int().nonnegative(),
    skipped: z.number().int().nonnegative(),
    budget_skipped: z.number().int().nonnegative(),
    limit_skipped: z.number().int().nonnegative(),
    cancelled: z.number().int().nonnegative(),
  })
  .strict()

const CouncilResultSchema = z
  .object({ kind: z.literal("council"), value: GenericVerdictSchema })
  .strict()
const TriageResultSchema = z
  .object({
    kind: z.literal("triage"),
    value: z.enum(["VALID", "INVALID", "NEEDS-CLARIFICATION", "INCONCLUSIVE", "UNAVAILABLE"]),
  })
  .strict()
const FeedbackResultSchema = z
  .object({
    kind: z.literal("feedback"),
    value: z.enum(["ACCEPT", "AUTHOR-DECIDES", "INCONCLUSIVE", "UNAVAILABLE"]),
  })
  .strict()
const TestReviewResultSchema = z
  .object({ kind: z.literal("test-review"), value: GenericVerdictSchema })
  .strict()
const WorkflowResultSchema = z.discriminatedUnion("kind", [
  CouncilResultSchema,
  TriageResultSchema,
  FeedbackResultSchema,
  TestReviewResultSchema,
])

const ReviewDispatchPayloadSchema = z
  .object({
    command: z.enum(["review-council", "triage-issue", "address-feedback", "speckit-testreview"]),
    mode: z.enum(["code", "specs", "triage", "feedback", "test"]),
    full: z.boolean(),
    input_context: InputContextSchema,
    change_profile: ChangeProfileSchema,
    plan_version: z.literal(1),
    plan: z.array(DispatchPlanEntrySchema).max(512),
    runs: z.array(DispatchRunSchema).max(64),
    coverage: CoverageSchema,
    findings: z.array(ConsolidatedFindingSchema).max(512),
    advisories: z.array(AdvisorySchema).max(512),
    verdict: GenericVerdictSchema,
    verdict_reason: z.string().min(1).max(4096),
    run_counts: RunCountsSchema,
    workflow_result: WorkflowResultSchema,
    correlation_id: UUIDSchema,
  })
  .strict()
  .superRefine((payload, context) => {
    const expected = {
      "review-council": { modes: ["code", "specs"], kind: "council" },
      "triage-issue": { modes: ["triage"], kind: "triage" },
      "address-feedback": { modes: ["feedback"], kind: "feedback" },
      "speckit-testreview": { modes: ["test"], kind: "test-review" },
    } as const
    const command = expected[payload.command]
    if (!(command.modes as readonly string[]).includes(payload.mode)) {
      context.addIssue({ code: "custom", path: ["mode"], message: `invalid mode for ${payload.command}` })
    }
    if (payload.workflow_result.kind !== command.kind) {
      context.addIssue({
        code: "custom",
        path: ["workflow_result", "kind"],
        message: `invalid workflow result kind for ${payload.command}`,
      })
    }
    if (payload.coverage.checks_passed > payload.coverage.checks_total) {
      context.addIssue({
        code: "custom",
        path: ["coverage", "checks_passed"],
        message: "checks_passed must not exceed checks_total",
      })
    }
  })

const ArtifactProvenanceSchema = z
  .object({
    branch: RefSchema,
    commit: SHA1Schema,
    workflow_id: z.string().min(1).max(255).regex(/^[ -~]+$/),
  })
  .strict()
const FinalizeInputSchema = z
  .object({
    payload: ReviewDispatchPayloadSchema,
    provenance: ArtifactProvenanceSchema,
  })
  .strict()

type ReviewMatrix = z.infer<typeof ReviewMatrixSchema>
type Tier = z.infer<typeof TierSchema>
type Source = (typeof SOURCE_VALUES)[number]
type MatrixMode = "code" | "spec" | "triage" | "feedback"
type CommandMode = z.infer<typeof PlanInputSchema>["mode"]
type IssueComment = z.infer<typeof IssueCommentSchema>

interface TierCaps {
  readonly lightweight: number | null
  readonly standard: number | null
  readonly heavy: number | null
}

interface Limits {
  readonly max_personas: number
  readonly max_runs_per_persona: number
  readonly max_total_runs: number
  readonly max_parallel_runs: number
  readonly per_run_timeout_seconds: number
  readonly max_reported_cost_usd: number
  readonly tier_caps: TierCaps
}

/** Input accepted by the deterministic review-dispatch planner. */
export type PlanReviewDispatchInput = z.input<typeof PlanInputSchema>


/** Injectable planner dependencies for policy reads and YAML parsing. */
export interface PlannerDependencies {
  readonly readText: (relativePath: string) => Promise<string>
  readonly parseYaml: YamlParser
}

/** Closed review-dispatch payload accepted by the finalization boundary. */
export type ReviewDispatchPayload = z.output<typeof ReviewDispatchPayloadSchema>

/** Untrusted input accepted by finalize_review_dispatch. */
export type FinalizeReviewDispatchInput = z.input<typeof FinalizeInputSchema>

/** Generic verdict shared by review-dispatch and review-verdict version 2. */
export type GenericVerdict = z.output<typeof GenericVerdictSchema>

/** Native workflow result retained losslessly beside its generic projection. */
export type NativeWorkflowResult = ReviewDispatchPayload["workflow_result"]

/** Minimal file handle contract required for durable artifact writes. */
export interface ArtifactFileHandle {
  readonly writeFile: (data: string, encoding: "utf8") => Promise<void>
  readonly sync: () => Promise<void>
  readonly close: () => Promise<void>
}

/** Injectable filesystem, clock, and UUID operations used by finalization. */
export interface FinalizationDependencies {
  readonly projectRoot: string
  readonly mkdir: (path: string, options: { readonly recursive: true; readonly mode: number }) => Promise<unknown>
  readonly chmod: (path: string, mode: number) => Promise<void>
  readonly open: (path: string, flags: string, mode?: number) => Promise<ArtifactFileHandle>
  readonly link: (existingPath: string, newPath: string) => Promise<void>
  readonly unlink: (path: string) => Promise<void>
  readonly now: () => Date
  readonly uuid: () => string
}

/** Hero Interface Contract envelope emitted for review-dispatch provenance. */
export interface ReviewDispatchEnvelope {
  readonly hero: "the-divisor"
  readonly version: "1.0.0"
  readonly timestamp: string
  readonly artifact_type: "review-dispatch"
  readonly schema_version: "1.0.0"
  readonly context: {
    readonly branch: string
    readonly commit: string
    readonly correlation_id: string
    readonly workflow_id: string
  }
  readonly payload: ReviewDispatchPayload
}

/** Canonical review-verdict version 2 decision projection returned for dual emission. */
export interface ReviewVerdictDecisionData {
  readonly schema_version: "2.0.0"
  readonly council_decision: "APPROVED" | "CHANGES_REQUESTED" | "ESCALATED" | "INCONCLUSIVE" | "UNAVAILABLE"
  readonly persona_verdicts: readonly {
    readonly persona: string
    readonly verdict: "APPROVED" | "CHANGES_REQUESTED"
    readonly findings: readonly {
      readonly severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW"
      readonly category: string
      readonly description: string
      readonly file?: string
      readonly line?: number
    }[]
    readonly summary: string
  }[]
  readonly unresolved_findings: readonly {
    readonly severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW"
    readonly category: string
    readonly description: string
    readonly file?: string
    readonly line?: number
  }[]
}

/** Calculated review state retained even when required persistence fails. */
export interface CalculatedAssessment {
  readonly authoritative: boolean
  readonly payload: ReviewDispatchPayload
  readonly review_verdict: ReviewVerdictDecisionData
}

/** Deterministic structured response returned by finalize_review_dispatch. */
export interface FinalizeReviewDispatchResult {
  readonly status: "success" | "failed"
  readonly operation_verdict: GenericVerdict
  readonly correlation_id: string | null
  readonly artifact_path: string | null
  readonly review_dispatch: ReviewDispatchEnvelope | null
  readonly review_verdict: ReviewVerdictDecisionData | null
  readonly calculated_assessment: CalculatedAssessment | null
  readonly errors: readonly string[]
}

interface DiffProfile {
  readonly kind: "diff"
  readonly lines: number
  readonly files: number
  readonly components: number
  readonly security_sensitive: boolean
  readonly user_facing: boolean
  readonly categories: readonly ReviewCategory[]
  readonly tier: Tier
}

interface IssueProfile {
  readonly kind: "issue"
  readonly text_bytes: number
  readonly comment_count: number
  readonly content_sha256: string
  readonly matched_rules: readonly string[]
  readonly security_sensitive: boolean
  readonly user_facing: boolean
  readonly categories: readonly ReviewCategory[]
  readonly tier: Tier
}

type ChangeProfile = DiffProfile | IssueProfile

interface PlanEntry {
  readonly agent: string
  readonly decision: "include" | "skip"
  readonly reason_code: string
  readonly reason: string
  readonly source: Source
  readonly sequence: number
  readonly read_only: true
  readonly tier: Tier | null
  readonly model: string | null
  readonly variant: string | null
  readonly validation_errors: readonly string[]
}

interface PlanOmission {
  readonly agent: string
  readonly reason_code: "duplicate-model-variant"
  readonly model: string
  readonly variant: string | null
}

interface LimitState {
  readonly required_personas: number
  readonly configured_personas: number
  readonly required_runs: number
  readonly configured_runs: number
}

/** Plan-level advisory emitted during dispatch planning (before runs execute). */
interface PlanAdvisory {
  readonly severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW"
  readonly description: string
}

/** Byte-stable, versioned output returned by plan_review_dispatch. */
export interface DispatchPlan {
  readonly plan_version: 1
  readonly status: "ready" | "inconclusive"
  readonly workflow_result: "INCONCLUSIVE" | null
  readonly mode: CommandMode
  readonly matrix_mode: MatrixMode
  readonly full: boolean
  readonly augmentation_requested: boolean
  readonly change_profile: ChangeProfile
  readonly limits: Limits
  readonly entries: readonly PlanEntry[]
  readonly omissions: readonly PlanOmission[]
  readonly limit_state: LimitState | null
  readonly advisories: readonly PlanAdvisory[]
  readonly errors: readonly string[]
}


const ISSUE_KEYWORDS: Readonly<Record<Exclude<ReviewCategory, "standard">, readonly string[]>> = {
  security: [
    "security",
    "vulnerability",
    "cve",
    "secret",
    "credential",
    "token",
    "auth",
    "permission",
    "injection",
  ],
  "cli-ux": ["cli", "command", "flag", "argument", "output", "prompt", "terminal", "ux"],
  "test-quality": ["test", "testing", "coverage", "assertion", "flaky", "regression", "fixture"],
  documentation: ["documentation", "docs", "readme", "guide", "tutorial", "changelog"],
  "ci-cd": ["ci", "workflow", "pipeline", "action", "deploy", "release", "build"],
  dependencies: ["dependency", "dependencies", "package", "module", "version", "upgrade", "supply-chain"],
}



/**
 * Parses and closed-validates a review matrix.
 * @param text YAML or JSON policy text.
 * @param parser Injected YAML parser.
 * @returns The validated matrix.
 */
export function parseReviewMatrix(text: string, parser: YamlParser): ReviewMatrix {
  const result = ReviewMatrixSchema.safeParse(parseYamlDocument(text, "review matrix", parser))
  if (!result.success) {
    throw new Error(`validate review matrix: ${formatValidationError(result.error)}`)
  }
  return result.data
}



function normalizePath(input: string): string {
  const normalized = input.replaceAll("\\", "/")
  const segments = normalized.split("/")
  if (
    normalized.startsWith("/") ||
    normalized.includes("\0") ||
    segments.some((segment) => segment === "" || segment === "." || segment === "..")
  ) {
    throw new Error(`invalid changed-file path ${JSON.stringify(input)}`)
  }
  return normalized
}

function classifyPath(path: string): ReviewCategory {
  const lower = path.toLowerCase()
  const base = lower.slice(lower.lastIndexOf("/") + 1)
  if (
    /_test\.(go|py)$/.test(base) ||
    lower.includes("/__tests__/") ||
    /_spec\.[^/]+$/.test(base)
  ) {
    return "test-quality"
  }
  if (/(^|\/)(cmd|cli)(\/|$)/.test(lower)) {
    return "cli-ux"
  }
  if (/(^|\/)(api|handler|middleware|routes)(\/|$)/.test(lower)) {
    return "security"
  }
  if (base.endsWith(".md") || lower.startsWith("docs/") || lower.includes("/docs/")) {
    return "documentation"
  }
  if (lower.startsWith(".github/workflows/") || base.startsWith("dockerfile")) {
    return "ci-cd"
  }
  if (["go.mod", "package.json", "requirements.txt"].includes(base)) {
    return "dependencies"
  }
  return "standard"
}

function componentForPath(path: string): string {
  const separator = path.lastIndexOf("/")
  if (separator < 0) {
    return "<root>"
  }
  const parent = path.slice(0, separator)
  if (parent.endsWith("/__tests__")) {
    return parent.slice(0, -"/__tests__".length) || "<root>"
  }
  return parent
}

function diffTier(lines: number, files: number, components: number, securitySensitive: boolean): Tier {
  if (securitySensitive || lines > 300 || files > 10) {
    return "heavy"
  }
  if (lines > 50 || files > 3 || components > 1) {
    return "standard"
  }
  return "lightweight"
}

function profileDiff(files: readonly z.infer<typeof ChangedFileSchema>[]): DiffProfile {
  const seenPaths = new Set<string>()
  const categories = new Set<ReviewCategory>()
  const components = new Set<string>()
  let lines = 0
  let securitySensitive = false
  let userFacing = false

  for (const file of files) {
    const path = normalizePath(file.path)
    if (seenPaths.has(path)) {
      throw new Error(`duplicate changed-file path ${JSON.stringify(path)}`)
    }
    seenPaths.add(path)
    const category = classifyPath(path)
    const lower = path.toLowerCase()
    categories.add(category)
    components.add(componentForPath(path))
    lines += file.additions + file.deletions
    if (!Number.isSafeInteger(lines)) {
      throw new Error("changed line total exceeds the safe integer range")
    }
    securitySensitive ||=
      category === "security" ||
      lower.startsWith(".github/workflows/") ||
      /(^|\/)(cmd|internal)(\/|$)/.test(lower) ||
      /(secret|token|key)/.test(lower)
    userFacing ||=
      category === "cli-ux" ||
      category === "documentation" ||
      /(^|\/)(readme|changelog)\.md$/.test(lower) ||
      lower.startsWith(".opencode/agents/") ||
      lower.startsWith(".opencode/commands/")
  }

  const orderedCategories = [...categories].sort()
  if (orderedCategories.length === 0) {
    orderedCategories.push("standard")
  }
  return {
    kind: "diff",
    lines,
    files: seenPaths.size,
    components: components.size,
    security_sensitive: securitySensitive,
    user_facing: userFacing,
    categories: orderedCategories,
    tier: diffTier(lines, seenPaths.size, components.size, securitySensitive),
  }
}

function normalizeText(text: string): string {
  return text.replace(/\r\n?/g, "\n").normalize("NFC")
}

function canonicalCommentID(id: string | number): string {
  const value = String(id).replace(/^0+(?=[0-9])/, "")
  if (value === "0") {
    throw new Error("comment id must be positive")
  }
  return value
}

function canonicalTimestamp(value: string): string {
  if (!/^\d{4}-\d{2}-\d{2}T/.test(value) || !/(?:Z|[+-]\d{2}:\d{2})$/.test(value)) {
    throw new Error(`invalid comment timestamp ${JSON.stringify(value)}`)
  }
  const milliseconds = Date.parse(value)
  if (!Number.isFinite(milliseconds)) {
    throw new Error(`invalid comment timestamp ${JSON.stringify(value)}`)
  }
  return new Date(milliseconds).toISOString().replace(/\.\d{3}Z$/, "Z")
}

function compareDecimalIDs(left: string, right: string): number {
  return left.length - right.length || left.localeCompare(right)
}

function issueFrame(title: string, body: string | null, comments: readonly IssueComment[]): {
  readonly frame: string
  readonly textBytes: number
  readonly normalizedText: string
} {
  const normalizedTitle = normalizeText(title)
  const normalizedBody = body === null ? null : normalizeText(body)
  const normalizedComments = comments
    .map((comment) => ({
      id: canonicalCommentID(comment.id),
      createdAt: canonicalTimestamp(comment.created_at),
      body: normalizeText(comment.body),
    }))
    .sort(
      (left, right) =>
        left.createdAt.localeCompare(right.createdAt) || compareDecimalIDs(left.id, right.id),
    )

  let frame = `uf-issue-content-v1\ntitle:${Buffer.byteLength(normalizedTitle, "utf8")}\n${normalizedTitle}\n`
  frame +=
    normalizedBody === null
      ? "body:null\n"
      : `body:${Buffer.byteLength(normalizedBody, "utf8")}\n${normalizedBody}\n`
  frame += `comments:${normalizedComments.length}\n`
  for (const comment of normalizedComments) {
    frame += `comment-id:${Buffer.byteLength(comment.id, "utf8")}\n${comment.id}\n`
    frame += `created-at:${Buffer.byteLength(comment.createdAt, "utf8")}\n${comment.createdAt}\n`
    frame += `comment-body:${Buffer.byteLength(comment.body, "utf8")}\n${comment.body}\n`
  }

  const bodies = normalizedComments.map((comment) => comment.body)
  const normalizedText = [normalizedTitle, ...(normalizedBody === null ? [] : [normalizedBody]), ...bodies].join(
    "\n",
  )
  const textBytes =
    Buffer.byteLength(normalizedTitle, "utf8") +
    (normalizedBody === null ? 0 : Buffer.byteLength(normalizedBody, "utf8")) +
    normalizedComments.reduce((total, comment) => total + Buffer.byteLength(comment.body, "utf8"), 0)
  return { frame, textBytes, normalizedText }
}

function classifyIssueText(text: string): {
  readonly categories: readonly ReviewCategory[]
  readonly matchedRules: readonly string[]
} {
  const tokens = text.toLowerCase().match(/[a-z0-9]+/g) ?? []
  const tokenSet = new Set(tokens)
  const categories = new Set<ReviewCategory>()
  const matchedRules = new Set<string>()

  for (const [category, keywords] of Object.entries(ISSUE_KEYWORDS) as ReadonlyArray<
    readonly [Exclude<ReviewCategory, "standard">, readonly string[]]
  >) {
    for (const keyword of keywords) {
      const matched =
        keyword === "supply-chain"
          ? tokens.some((token, index) => token === "supply" && tokens[index + 1] === "chain")
          : tokenSet.has(keyword)
      if (matched) {
        categories.add(category)
        matchedRules.add(`${category}:${keyword}`)
      }
    }
  }

  if (categories.size === 0) {
    categories.add("standard")
  }
  return { categories: [...categories].sort(), matchedRules: [...matchedRules].sort() }
}

function profileIssue(issue: z.infer<typeof IssueContentSchema>): IssueProfile {
  const framed = issueFrame(issue.title, issue.body, issue.comments)
  const classified = classifyIssueText(framed.normalizedText)
  const securitySensitive = classified.categories.includes("security")
  const tier: Tier =
    securitySensitive || framed.textBytes > 32_768 || issue.comments.length > 20
      ? "heavy"
      : framed.textBytes > 4_096 || issue.comments.length > 5 || classified.categories.length > 1
        ? "standard"
        : "lightweight"
  return {
    kind: "issue",
    text_bytes: framed.textBytes,
    comment_count: issue.comments.length,
    content_sha256: createHash("sha256").update(framed.frame, "utf8").digest("hex"),
    matched_rules: classified.matchedRules,
    security_sensitive: securitySensitive,
    user_facing:
      classified.categories.includes("documentation") || classified.categories.includes("cli-ux"),
    categories: classified.categories,
    tier,
  }
}

function matrixMode(mode: CommandMode): MatrixMode {
  if (mode === "specs") {
    return "spec"
  }
  if (mode === "test") {
    return "code"
  }
  return mode
}

function limitsFor(matrix: ReviewMatrix): Limits {
  const { tier_caps: matrixTierCaps, ...matrixScalars } = matrix.limits ?? {}
  const tier_caps: TierCaps = {
    ...DEFAULT_TIER_CAPS,
    ...matrixTierCaps,
  }
  return { ...DEFAULT_LIMITS, ...matrixScalars, tier_caps }
}

function profilePair(matrix: ReviewMatrix, profileName: string, variant?: string): {
  readonly model: string | null
  readonly variant: string | null
  readonly tier: Tier | null
} {
  const profile = matrix.profiles[profileName]
  if (profile === undefined) {
    throw new Error(`unknown profile ${profileName}`)
  }
  return {
    model: profile.model ?? null,
    variant: variant ?? profile.variant ?? null,
    tier: TierSchema.safeParse(profileName).success ? (profileName as Tier) : null,
  }
}

function sourceFor(matrix: ReviewMatrix, mode: MatrixMode, agent: string): Source {
  if (matrix.runs?.[mode]?.[agent] !== undefined) {
    return "explicit"
  }
  if ((matrix.advisor?.[mode] ?? []).includes(agent)) {
    return "advisor"
  }
  return "host"
}

function relevance(
  reviewer: Reviewer,
  profile: ChangeProfile,
  mode: CommandMode,
  full: boolean,
): { readonly include: boolean; readonly code: string; readonly reason: string } {
  if (reviewer.capability === "content") {
    return { include: false, code: "content-capability", reason: "content personas do not assess reviews" }
  }
  if (mode === "test") {
    return reviewer.agent === "divisor-testing"
      ? { include: true, code: "test-review-scope", reason: "test review is restricted to divisor-testing" }
      : { include: false, code: "test-review-scope", reason: "outside the testing-only review scope" }
  }
  if (full) {
    return { include: true, code: "full-panel", reason: "full panel includes each review-capable persona" }
  }
  if (reviewer.agent === "divisor-adversary" || reviewer.agent === "divisor-guard") {
    return { include: true, code: "always-required", reason: "required baseline reviewer" }
  }
  if (reviewer.agent === "divisor-curator") {
    return profile.categories.includes("documentation") || profile.user_facing
      ? { include: true, code: "curator-relevant", reason: "documentation or user-facing change" }
      : { include: false, code: "curator-pruned", reason: "no documentation or user-facing change" }
  }
  const matchingScope = reviewer.scopes.find((scope) => profile.categories.includes(scope))
  return matchingScope === undefined
    ? { include: false, code: "scope-miss", reason: "reviewer scopes do not intersect change categories" }
    : { include: true, code: "scope-match", reason: `matched ${matchingScope} scope` }
}

interface ResolvedRun {
  readonly source: Source
  readonly tier: Tier | null
  readonly model: string | null
  readonly variant: string | null
  readonly reasonCode: string
  readonly reason: string
  readonly validationErrors: readonly string[]
}

function explicitRuns(matrix: ReviewMatrix, mode: MatrixMode, agent: string): ResolvedRun[] | null {
  const configured = matrix.runs?.[mode]?.[agent]
  if (configured === undefined) {
    return null
  }
  return configured.map((run) => {
    if ("model" in run) {
      return {
        source: "explicit",
        tier: null,
        model: run.model,
        variant: run.variant ?? null,
        reasonCode: "explicit-run",
        reason: "authoritative matrix run",
        validationErrors: [],
      }
    }
    try {
      const resolved = profilePair(matrix, run.profile, run.variant)
      return {
        source: "explicit",
        ...resolved,
        reasonCode: "explicit-run",
        reason: "authoritative matrix run",
        validationErrors: [],
      }
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : "invalid explicit run"
      return {
        source: "explicit",
        tier: null,
        model: null,
        variant: null,
        reasonCode: "invalid-explicit-run",
        reason: "explicit run failed validation",
        validationErrors: [message],
      }
    }
  })
}

function baseRuns(matrix: ReviewMatrix, mode: MatrixMode, agent: string): ResolvedRun[] {
  const explicit = explicitRuns(matrix, mode, agent)
  if (explicit !== null) {
    return explicit
  }
  if ((matrix.advisor?.[mode] ?? []).includes(agent)) {
    const selectedProfile = matrix.always?.[agent] ?? matrix.defaults[mode]
    const resolved = profilePair(matrix, selectedProfile)
    return [
      {
        source: "advisor",
        ...resolved,
        reasonCode: "advisor-run",
        reason: `advisor selected ${selectedProfile} profile`,
        validationErrors: [],
      },
    ]
  }
  return [
    {
      source: "host",
      tier: null,
      model: null,
      variant: null,
      reasonCode: "host-fallback",
      reason: "no explicit or advisor run; use the active host model",
      validationErrors: [],
    },
  ]
}

function adjacentTier(tier: Tier): Tier {
  return tier === "lightweight" ? "standard" : tier === "standard" ? "heavy" : "standard"
}

function shouldAugment(agent: string, profile: ChangeProfile): boolean {
  return (
    (agent === "divisor-adversary" && profile.security_sensitive) ||
    (agent === "divisor-architect" && profile.kind === "diff" && profile.lines > 500)
  )
}

function addAugmentation(
  matrix: ReviewMatrix,
  agent: string,
  profile: ChangeProfile,
  requested: boolean,
  full: boolean,
  runs: readonly ResolvedRun[],
): ResolvedRun[] {
  if (
    full ||
    !requested ||
    matrix.risk_augmentation?.enabled !== true ||
    !matrix.risk_augmentation.agents.includes(agent) ||
    !shouldAugment(agent, profile)
  ) {
    return [...runs]
  }
  const tier = adjacentTier(profile.tier)
  const resolved = profilePair(matrix, tier)
  return [
    ...runs,
    {
      source: "advisor",
      ...resolved,
      reasonCode: "risk-augmentation",
      reason: `opt-in risk augmentation selected adjacent ${tier} profile`,
      validationErrors: [],
    },
  ]
}

function deduplicateRuns(
  agent: string,
  runs: readonly ResolvedRun[],
): { readonly runs: readonly ResolvedRun[]; readonly omissions: readonly PlanOmission[] } {
  const seen = new Set<string>()
  const deduplicated: ResolvedRun[] = []
  const omissions: PlanOmission[] = []
  for (const run of runs) {
    if (run.model === null) {
      deduplicated.push(run)
      continue
    }
    const key = `${run.model}\0${run.variant ?? ""}`
    if (seen.has(key)) {
      omissions.push({
        agent,
        reason_code: "duplicate-model-variant",
        model: run.model,
        variant: run.variant,
      })
      continue
    }
    seen.add(key)
    deduplicated.push(run)
  }
  return { runs: deduplicated, omissions }
}

function fullPanelRuns(matrix: ReviewMatrix): ResolvedRun[] {
  const resolved = profilePair(matrix, "standard")
  return [
    {
      source: "advisor",
      ...resolved,
      reasonCode: "full-panel",
      reason: "full panel forces one standard-profile run",
      validationErrors: [],
    },
  ]
}

function failedPlan(
  input: z.output<typeof PlanInputSchema>,
  profile: ChangeProfile,
  errors: readonly string[],
): DispatchPlan {
  return {
    plan_version: PLAN_VERSION,
    status: "inconclusive",
    workflow_result: "INCONCLUSIVE",
    mode: input.mode,
    matrix_mode: matrixMode(input.mode),
    full: input.full,
    augmentation_requested: input.augment,
    change_profile: profile,
    limits: { ...DEFAULT_LIMITS },
    entries: [],
    omissions: [],
    limit_state: null,
    advisories: [],
    errors: [...errors].sort(),
  }
}

/**
 * Own-property check that rejects prototype-chain lookups (__proto__, constructor, etc.).
 */
function hasOwn(obj: Record<string, unknown>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(obj, key)
}

/**
 * Validates that an override object only contains whitelisted keys:
 * `profiles.<profileName>.[model, variant]` paths.
 * Any other key or nested path causes a validation failure.
 * @param override Parsed override YAML object.
 * @throws Error if non-whitelisted keys are present.
 */
function validateOverride(override: unknown): void {
  // Empty or comments-only override documents parse to null/undefined/empty-object.
  // These are valid no-ops — use the base matrix as-is.
  if (override === null || override === undefined) {
    return
  }
  if (typeof override !== "object" || Array.isArray(override)) {
    throw new Error(
      "override validation: override must be an object — only profiles.*.model and profiles.*.variant paths are allowed",
    )
  }
  const obj = override as Record<string, unknown>
  if (Object.keys(obj).length === 0) {
    return // empty object override is a valid no-op
  }
  const allowedTopLevel = new Set(["profiles"])
  for (const key of Object.keys(obj)) {
    if (!allowedTopLevel.has(key)) {
      throw new Error(
        `override validation: key "${key}" is not permitted — only profiles.*.model and profiles.*.variant paths are allowed`,
      )
    }
  }
  const profiles = override.profiles
  if (profiles === undefined || profiles === null) {
    return // empty override is valid
  }
  if (typeof profiles !== "object" || Array.isArray(profiles)) {
    throw new Error("override validation: profiles must be an object")
  }
  const profilesObj = profiles as Record<string, unknown>
  for (const profileName of Object.keys(profilesObj)) {
    const profileValue = profilesObj[profileName]
    if (profileValue === undefined || profileValue === null) {
      continue
    }
    if (typeof profileValue !== "object" || Array.isArray(profileValue)) {
      throw new Error(`override validation: profiles.${profileName} must be an object`)
    }
    const profileObj = profileValue as Record<string, unknown>
    const allowedProfileKeys = new Set(["model", "variant"])
    for (const profileKey of Object.keys(profileObj)) {
      if (!allowedProfileKeys.has(profileKey)) {
        throw new Error(
          `override validation: profiles.${profileName}.${profileKey} is not permitted — only model and variant are allowed`,
        )
      }
    }
  }
}

/**
 * Performs nested-path merge of an override into the base review matrix.
 * Walks the override for `profiles.<profileName>.[model, variant]` paths
 * and replaces each matching leaf in the base. The top-level `profiles` key
 * is never used as a wholesale replacement.
 * @param base The base review matrix loaded from .uf/review-matrix.yaml.
 * @param override Parsed override YAML object.
 * @returns A new ReviewMatrix with override values merged in.
 */
function mergeOverride(base: ReviewMatrix, override: Record<string, unknown>): ReviewMatrix {
  const profiles = override.profiles
  if (profiles === undefined || profiles === null || typeof profiles !== "object" || Array.isArray(profiles)) {
    return base // no profiles to merge
  }
  const profilesObj = profiles as Record<string, unknown>
  const mergedProfiles: Record<string, unknown> = Object.create(null)
  for (const profileName of Object.keys(base.profiles)) {
    mergedProfiles[profileName] = { ...base.profiles[profileName] }
  }
  for (const profileName of Object.keys(profilesObj)) {
    const profileValue = profilesObj[profileName]
    if (profileValue === undefined || profileValue === null || typeof profileValue !== "object" || Array.isArray(profileValue)) {
      continue
    }
    const profileObj = profileValue as Record<string, unknown>
    // Use own-property check to reject __proto__, constructor, toString etc.
    // which would match via the prototype chain but are not base matrix tiers.
    if (!hasOwn(base.profiles, profileName)) {
      throw new Error(
        `override validation: profiles.${profileName} is not a defined tier in the base matrix — only existing profile tiers may be overridden`,
      )
    }
    const existing = base.profiles[profileName]
    const merged: Record<string, unknown> = { ...existing }
    if ("model" in profileObj) {
      // Convert null to undefined — ModelSchema.optional() allows undefined but not null.
      merged.model = profileObj.model ?? undefined
    }
    if ("variant" in profileObj) {
      merged.variant = profileObj.variant ?? undefined
    }
    mergedProfiles[profileName] = merged as typeof existing
  }
  return { ...base, profiles: mergedProfiles }
}

async function loadPolicies(dependencies: PlannerDependencies): Promise<{
  readonly matrix: ReviewMatrix
  readonly manifest: ReviewerManifest
}> {
  const [matrixText, manifestText] = await Promise.all([
    dependencies.readText(MATRIX_PATH),
    dependencies.readText(MANIFEST_PATH),
  ])
  const baseMatrix = parseReviewMatrix(matrixText, dependencies.parseYaml)

  // Attempt to load override file for per-developer model/variant customization.
  let overrideText: string | null = null
  try {
    overrideText = await dependencies.readText(MATRIX_OVERRIDE_PATH)
  } catch (error: unknown) {
    const isNotFound = error instanceof Error && "code" in error && (error as NodeJS.ErrnoException).code === "ENOENT"
    if (!isNotFound) {
      const message = error instanceof Error ? error.message : "unknown error"
      throw new Error(`INCONCLUSIVE: cannot read override file ${MATRIX_OVERRIDE_PATH}: ${message}`)
    }
    // Override file does not exist — use base as-is.
  }

  let matrix = baseMatrix
  if (overrideText !== null) {
    try {
      const override = dependencies.parseYaml(overrideText)
      validateOverride(override)
      if (override === null || override === undefined || typeof override !== "object" || Array.isArray(override)) {
        // Empty/comments-only override — use base as-is (valid no-op).
        matrix = baseMatrix
      } else {
        const merged = mergeOverride(baseMatrix, override as Record<string, unknown>)
        // Re-validate merged result against the schema and use the sanitized clone.
        matrix = parseReviewMatrix(
          JSON.stringify(merged),
          (text: string) => JSON.parse(text) as unknown,
        )
      }
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : "unknown override failure"
      throw new Error(`INCONCLUSIVE: override file ${MATRIX_OVERRIDE_PATH} is invalid: ${message}`)
    }
  }

  return {
    matrix,
    manifest: parseReviewerManifest(manifestText, dependencies.parseYaml),
  }
}

/** Priority bucket for tier-cap agent selection. Lower index = higher priority. */
const TIER_CAP_PRIORITY: readonly string[] = ["always-required", "scope-match", "curator-relevant"]

/**
 * Applies a tier-based agent count cap to the set of included agents.
 * Agents beyond the cap receive skip entries with reason_code "tier-cap".
 * Priority order: always-required > scope-match (alphabetical) > curator-relevant.
 * Returns the set of agents to keep and the skip entries for excluded agents.
 */
function applyTierCap(
  includedAgents: ReadonlyMap<string, string>,
  tier: Tier,
  cap: number,
): { readonly keep: ReadonlySet<string>; readonly skipped: readonly PlanEntry[] } {
  // Floor enforcement: clamp cap to max(cap, always_required_count)
  // so that always-required agents are never dropped by the tier cap.
  const alwaysRequiredCount = [...includedAgents.values()].filter((code) => code === "always-required").length
  const effectiveCap = Math.max(cap, alwaysRequiredCount)

  if (includedAgents.size <= effectiveCap) {
    return { keep: new Set(includedAgents.keys()), skipped: [] }
  }

  // Sort agents into priority buckets, alphabetical within each bucket
  const buckets = new Map<string, string[]>()
  for (const code of TIER_CAP_PRIORITY) {
    buckets.set(code, [])
  }
  buckets.set("other", [])

  for (const [agent, reasonCode] of includedAgents) {
    const bucket = TIER_CAP_PRIORITY.includes(reasonCode) ? reasonCode : "other"
    buckets.get(bucket)!.push(agent)
  }

  // Sort each bucket alphabetically for determinism
  for (const agents of buckets.values()) {
    agents.sort()
  }

  // Fill keep set up to the effective cap, in priority order
  const keep = new Set<string>()
  const ordered = [...TIER_CAP_PRIORITY, "other"]
  for (const bucket of ordered) {
    for (const agent of buckets.get(bucket) ?? []) {
      if (keep.size < effectiveCap) {
        keep.add(agent)
      }
    }
  }

  // Build skip entries for agents beyond the cap
  const skipped: PlanEntry[] = []
  for (const bucket of ordered) {
    for (const agent of buckets.get(bucket) ?? []) {
      if (!keep.has(agent)) {
        skipped.push({
          agent,
          decision: "skip",
          reason_code: "tier-cap",
          reason: `tier cap of ${effectiveCap} exceeded for ${tier} tier`,
          source: "host",
          sequence: 1,
          read_only: true,
          tier: null,
          model: null,
          variant: null,
          validation_errors: [],
        })
      }
    }
  }

  return { keep, skipped }
}

function buildPlan(
  input: z.output<typeof PlanInputSchema>,
  profile: ChangeProfile,
  matrix: ReviewMatrix,
  manifest: ReviewerManifest,
): DispatchPlan {
  const mode = matrixMode(input.mode)
  const limits = limitsFor(matrix)
  const byAgent = new Map(manifest.reviewers.map((reviewer) => [reviewer.agent, reviewer]))
  const entries: PlanEntry[] = []
  const omissions: PlanOmission[] = []
  const errors: string[] = []
  const includedAgents = new Set<string>()

  // Pass 1: Evaluate relevance for all agents.
  const relevanceIncluded = new Map<string, string>() // agent → reason_code
  for (const agent of [...input.discovered_agents].sort()) {
    const reviewer = byAgent.get(agent)
    if (reviewer === undefined) {
      errors.push(`discovered agent ${agent} has no reviewer-manifest entry`)
      continue
    }
    const selection = relevance(reviewer, profile, input.mode, input.full)
    if (!selection.include) {
      entries.push({
        agent,
        decision: "skip",
        reason_code: selection.code,
        reason: selection.reason,
        source: input.full ? "advisor" : sourceFor(matrix, mode, agent),
        sequence: 1,
        read_only: true,
        tier: null,
        model: null,
        variant: null,
        validation_errors: [],
      })
      continue
    }
    relevanceIncluded.set(agent, selection.code)
  }

  // Apply tier cap between passes. Guard conditions: skip when full flag is set,
  // when command mode is triage, feedback, or test, or when the tier's cap is null.
  const tierCap = limits.tier_caps[profile.tier]
  const capApplies = !input.full && (input.mode === "code" || input.mode === "specs") && tierCap !== null
  const capResult = capApplies
    ? applyTierCap(relevanceIncluded, profile.tier, tierCap)
    : { keep: new Set(relevanceIncluded.keys()), skipped: [] as readonly PlanEntry[] }
  entries.push(...capResult.skipped)

  // Pass 2: Run resolution for agents that survived the tier cap.
  for (const agent of [...capResult.keep].sort()) {
    includedAgents.add(agent)
    const initialRuns = input.full ? fullPanelRuns(matrix) : baseRuns(matrix, mode, agent)
    const augmented = addAugmentation(matrix, agent, profile, input.augment, input.full, initialRuns)
    const deduplicated = deduplicateRuns(agent, augmented)
    omissions.push(...deduplicated.omissions)
    if (agent === "divisor-curator" && deduplicated.runs.length > 1) {
      errors.push("divisor-curator may have at most one assessment run")
    }
    if (deduplicated.runs.length > limits.max_runs_per_persona) {
      errors.push(
        `${agent} requires ${deduplicated.runs.length} runs but max_runs_per_persona is ${limits.max_runs_per_persona}`,
      )
    }
    deduplicated.runs.forEach((run, index) => {
      if (run.validationErrors.length > 0) {
        errors.push(`${agent} run ${index + 1}: ${run.validationErrors.join("; ")}`)
      }
      entries.push({
        agent,
        decision: run.validationErrors.length === 0 ? "include" : "skip",
        reason_code: run.reasonCode,
        reason: run.reason,
        source: run.source,
        sequence: index + 1,
        read_only: true,
        tier: run.tier,
        model: run.model,
        variant: run.variant,
        validation_errors: run.validationErrors,
      })
    })
  }

  const includedRuns = entries.filter((entry) => entry.decision === "include").length
  const limitExceeded =
    includedAgents.size > limits.max_personas || includedRuns > limits.max_total_runs
  const limitState: LimitState | null = limitExceeded
    ? {
        required_personas: includedAgents.size,
        configured_personas: limits.max_personas,
        required_runs: includedRuns,
        configured_runs: limits.max_total_runs,
      }
    : null
  if (limitExceeded) {
    errors.push(
      `panel requires ${includedAgents.size} personas/${includedRuns} runs; ` +
        `configured limits are ${limits.max_personas} personas/${limits.max_total_runs} runs`,
    )
  }
  if (includedRuns === 0) {
    errors.push("dispatch plan contains no runnable review assessment")
  }

  // Emit advisory when heavy tier cap is configured and the diff is security-sensitive.
  const advisories: PlanAdvisory[] = []
  if (
    profile.kind === "diff" &&
    profile.security_sensitive &&
    limits.tier_caps.heavy !== null
  ) {
    advisories.push({
      severity: "HIGH",
      description:
        `tier_caps.heavy is set to ${limits.tier_caps.heavy} for a security-sensitive change. ` +
        `This may reduce the number of review agents below the full security review panel. ` +
        `Consider removing the heavy tier cap or setting it to null for security-sensitive changes.`,
    })
  }

  const orderedErrors = [...new Set(errors)].sort()
  const status = orderedErrors.length === 0 ? "ready" : "inconclusive"
  return {
    plan_version: PLAN_VERSION,
    status,
    workflow_result: status === "ready" ? null : "INCONCLUSIVE",
    mode: input.mode,
    matrix_mode: mode,
    full: input.full,
    augmentation_requested: input.augment,
    change_profile: profile,
    limits,
    entries,
    omissions,
    limit_state: limitState,
    advisories,
    errors: orderedErrors,
  }
}

/**
 * Produces a deterministic dispatch plan from raw change inputs and injected policy dependencies.
 * @param rawInput Untrusted tool input.
 * @param dependencies Policy I/O and YAML parser dependencies.
 * @returns A ready plan or a fail-closed INCONCLUSIVE plan.
 */
export async function planReviewDispatch(
  rawInput: PlanReviewDispatchInput,
  dependencies: PlannerDependencies,
): Promise<DispatchPlan> {
  const input = PlanInputSchema.parse(rawInput)
  const profile = input.mode === "triage" ? profileIssue(input.issue!) : profileDiff(input.changed_files!)
  try {
    const policies = await loadPolicies(dependencies)
    return buildPlan(input, profile, policies.matrix, policies.manifest)
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : "unknown planning failure"
    return failedPlan(input, profile, [`review dispatch policy failure: ${message}`])
  }
}

/**
 * Serializes a dispatch plan with stable property and array ordering.
 * @param plan Validated dispatch plan.
 * @returns Canonical pretty-printed JSON with a terminal newline.
 */
export function serializeDispatchPlan(plan: DispatchPlan): string {
  return `${JSON.stringify(plan, null, 2)}\n`
}

type DispatchRun = ReviewDispatchPayload["runs"][number]
type CanonicalFinding = ReviewVerdictDecisionData["unresolved_findings"][number]

function uniqueSorted(values: readonly string[]): string[] {
  return [...new Set(values)].sort()
}

function planIdentity(value: { readonly agent: string; readonly source: string; readonly sequence: number }): string {
  return `${value.agent}\0${value.source}\0${value.sequence}`
}

function normalizedFindingKey(finding: { readonly file: string | null; readonly root_cause: string }): string {
  const file = (finding.file ?? "<none>")
    .normalize("NFC")
    .replaceAll("\\", "/")
    .replace(/^\.\//, "")
    .toLowerCase()
  const rootCause = finding.root_cause.normalize("NFC").trim().replace(/\s+/g, " ").toLowerCase()
  return `${file}\0${rootCause}`
}

function equalSets(left: readonly string[], right: readonly string[]): boolean {
  const sortedRight = uniqueSorted(right)
  return left.length === right.length && uniqueSorted(left).every((value, index) => value === sortedRight[index])
}

function triageMajority(values: readonly string[]): "VALID" | "INVALID" | "NEEDS-CLARIFICATION" {
  const clarification = values.filter((value) => value === "NEEDS-CLARIFICATION").length
  if (clarification > values.length / 2) {
    return "NEEDS-CLARIFICATION"
  }
  const valid = values.filter((value) => value === "VALID").length
  const invalid = values.filter((value) => value === "INVALID").length
  return valid === invalid ? "NEEDS-CLARIFICATION" : valid > invalid ? "VALID" : "INVALID"
}

function availabilityError(code: string): boolean {
  const parts = code.split("_")
  return (
    parts.some((part) => part === "provider" || part === "model" || part === "runtime") &&
    (parts.some((part) => ["unavailable", "availability", "timeout", "unsupported"].includes(part)) ||
      code.includes("_not_found"))
  )
}

function noSuccessResult(runs: readonly DispatchRun[]): "INCONCLUSIVE" | "UNAVAILABLE" {
  if (runs.length === 0) {
    return "INCONCLUSIVE"
  }
  const availabilityOnly = runs.every(
    (run) =>
      (run.status === "failed" || run.status === "cancelled") &&
      run.error !== null &&
      availabilityError(run.error.code),
  )
  return availabilityOnly ? "UNAVAILABLE" : "INCONCLUSIVE"
}

function successfulPersonaVotes(payload: ReviewDispatchPayload): ReadonlyMap<string, string> {
  const byAgent = new Map<string, string[]>()
  for (const run of payload.runs) {
    if (run.status !== "success" || run.workflow_verdict === null) {
      continue
    }
    const values = byAgent.get(run.agent) ?? []
    values.push(run.workflow_verdict)
    byAgent.set(run.agent, values)
  }

  const votes = new Map<string, string>()
  for (const [agent, values] of byAgent) {
    if (payload.workflow_result.kind === "triage") {
      votes.set(agent, triageMajority(values))
    } else if (payload.workflow_result.kind === "feedback") {
      votes.set(agent, values.includes("ACCEPT") ? "ACCEPT" : "AUTHOR-DECIDES")
    } else if (values.includes("REQUEST CHANGES")) {
      votes.set(agent, "REQUEST CHANGES")
    } else if (values.includes("APPROVE WITH ADVISORIES")) {
      votes.set(agent, "APPROVE WITH ADVISORIES")
    } else {
      votes.set(agent, "APPROVE")
    }
  }
  return votes
}

function calculatedWorkflowValue(payload: ReviewDispatchPayload): NativeWorkflowResult["value"] {
  const votes = [...successfulPersonaVotes(payload).values()]
  if (votes.length === 0) {
    return noSuccessResult(payload.runs)
  }
  if (payload.workflow_result.kind === "triage") {
    return triageMajority(votes)
  }
  if (payload.workflow_result.kind === "feedback") {
    return votes.includes("ACCEPT") ? "ACCEPT" : "AUTHOR-DECIDES"
  }
  if (votes.includes("REQUEST CHANGES")) {
    return "REQUEST CHANGES"
  }
  if (votes.includes("APPROVE WITH ADVISORIES") || payload.advisories.length > 0) {
    return "APPROVE WITH ADVISORIES"
  }
  return "APPROVE"
}

/**
 * Projects an authoritative native workflow result to the generic dispatch verdict.
 * @param result Native workflow result retained in the review-dispatch payload.
 * @returns The lossless generic verdict required by RD-FR-008.
 */
export function projectWorkflowVerdict(result: NativeWorkflowResult): GenericVerdict {
  return projectNativeValue(result.value)
}

function projectNativeValue(value: NativeWorkflowResult["value"]): GenericVerdict {
  switch (value) {
    case "VALID":
    case "ACCEPT":
      return "APPROVE"
    case "INVALID":
      return "REQUEST CHANGES"
    case "NEEDS-CLARIFICATION":
    case "AUTHOR-DECIDES":
      return "APPROVE WITH ADVISORIES"
    default:
      return value
  }
}

function canonicalDecision(verdict: GenericVerdict): ReviewVerdictDecisionData["council_decision"] {
  switch (verdict) {
    case "APPROVE":
      return "APPROVED"
    case "APPROVE WITH ADVISORIES":
      return "ESCALATED"
    case "REQUEST CHANGES":
      return "CHANGES_REQUESTED"
    default:
      return verdict
  }
}

function canonicalFinding(finding: ReviewDispatchPayload["findings"][number]): CanonicalFinding {
  return {
    severity: finding.severity,
    category: finding.category,
    description: finding.description,
    ...(finding.file === null ? {} : { file: finding.file }),
    ...(finding.line === null ? {} : { line: finding.line }),
  }
}

/**
 * Produces canonical review-verdict version 2 decision data from a validated dispatch payload.
 * @param payload Semantically valid review-dispatch payload.
 * @returns Canonical decision and persona projections for the dual-emission boundary.
 */
export function projectReviewVerdict(payload: ReviewDispatchPayload): ReviewVerdictDecisionData {
  const votes = successfulPersonaVotes(payload)
  const personaVerdicts = [...votes.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([persona, vote]) => {
      const runIDs = new Set(payload.runs.filter((run) => run.agent === persona).map((run) => run.run_id))
      const findings = payload.findings
        .filter((finding) => finding.run_ids.some((runID) => runIDs.has(runID)))
        .map(canonicalFinding)
      return {
        persona,
        verdict:
          vote === "REQUEST CHANGES" || vote === "INVALID"
            ? ("CHANGES_REQUESTED" as const)
            : ("APPROVED" as const),
        findings,
        summary: `Native verdict: ${vote}`,
      }
    })
  return {
    schema_version: REVIEW_VERDICT_SCHEMA_VERSION,
    council_decision: canonicalDecision(payload.verdict),
    persona_verdicts: personaVerdicts,
    unresolved_findings: payload.findings.map(canonicalFinding),
  }
}

function validateRunState(run: DispatchRun, errors: string[]): void {
  if (run.status === "pending" || run.status === "running") {
    errors.push(`run ${run.run_id} has nonterminal status ${run.status}`)
  }
  if (run.started_at !== null && run.finished_at !== null && Date.parse(run.finished_at) < Date.parse(run.started_at)) {
    errors.push(`run ${run.run_id} finished before it started`)
  }
  if (run.status === "success") {
    if (run.started_at === null || run.finished_at === null || run.error !== null || run.workflow_verdict === null) {
      errors.push(`successful run ${run.run_id} has inconsistent timestamps, error, or verdict`)
    }
  } else {
    if (run.workflow_verdict !== null || run.findings.length > 0) {
      errors.push(`non-success run ${run.run_id} must not vote or emit findings`)
    }
    if ((run.status === "failed" || run.status === "cancelled") && run.error === null) {
      errors.push(`${run.status} run ${run.run_id} requires an error`)
    }
    if (run.status === "failed" && run.finished_at === null) {
      errors.push(`failed run ${run.run_id} requires finished_at`)
    }
    if (["skipped", "budget_skipped", "limit_skipped"].includes(run.status)) {
      if (run.error !== null || run.usage !== null) {
        errors.push(`${run.status} run ${run.run_id} must not contain error or usage data`)
      }
    }
  }
}

function validateRunModel(run: DispatchRun, plan: ReviewDispatchPayload["plan"][number], errors: string[]): void {
  let expectedModel: string | null = null
  if (run.source === "host") {
    if (run.requested_model !== null || run.provider !== null || run.model_id !== null || run.variant !== null) {
      errors.push(`host run ${run.run_id} must retain null requested model provenance`)
    }
    if (run.status === "success" && (run.resolved_parent_model === null || run.resolved_parent_variant === null)) {
      errors.push(`successful host run ${run.run_id} requires resolved parent model and variant`)
    }
    expectedModel = run.resolved_parent_model
  } else {
    if (run.requested_model === null || run.provider === null || run.model_id === null) {
      errors.push(`${run.source} run ${run.run_id} requires requested model provenance`)
    } else if (run.requested_model !== `${run.provider}/${run.model_id}`) {
      errors.push(`run ${run.run_id} requested model does not match provider and model_id`)
    }
    if (run.resolved_parent_model !== null || run.resolved_parent_variant !== null) {
      errors.push(`${run.source} run ${run.run_id} must not contain resolved parent provenance`)
    }
    expectedModel = run.requested_model
  }

  if (plan.model !== run.requested_model || plan.variant !== run.variant) {
    errors.push(`run ${run.run_id} model provenance does not match its plan entry`)
  }
  const mismatch = run.reported_model !== null && (expectedModel === null || run.reported_model !== expectedModel)
  if (run.model_mismatch !== mismatch) {
    errors.push(`run ${run.run_id} model_mismatch does not match reported provenance`)
  }
  if (run.model_mismatch && run.status !== "failed") {
    errors.push(`model mismatch run ${run.run_id} must be failed`)
  }
  if (run.status === "success" && run.reported_model === null) {
    errors.push(`successful run ${run.run_id} requires reported model provenance`)
  }
}

function allowedRunVerdicts(payload: ReviewDispatchPayload): ReadonlySet<string> {
  switch (payload.workflow_result.kind) {
    case "triage":
      return new Set(["VALID", "INVALID", "NEEDS-CLARIFICATION"])
    case "feedback":
      return new Set(["ACCEPT", "AUTHOR-DECIDES"])
    default:
      return new Set(["APPROVE", "APPROVE WITH ADVISORIES", "REQUEST CHANGES"])
  }
}

function validatePlanAndRuns(payload: ReviewDispatchPayload, errors: string[]): void {
  const entries = new Map<string, ReviewDispatchPayload["plan"][number]>()
  for (const entry of payload.plan) {
    const identity = planIdentity(entry)
    if (entries.has(identity)) {
      errors.push(`duplicate plan entry ${entry.agent}/${entry.source}/${entry.sequence}`)
    }
    entries.set(identity, entry)
    if (entry.decision === "include" && entry.validation_errors.length > 0) {
      errors.push(`included plan entry ${entry.agent}/${entry.sequence} has validation errors`)
    }
  }

  const runIDs = new Set<string>()
  const observedEntries = new Set<string>()
  const allowedVerdicts = allowedRunVerdicts(payload)
  for (const run of payload.runs) {
    if (runIDs.has(run.run_id)) {
      errors.push(`duplicate run id ${run.run_id}`)
    }
    runIDs.add(run.run_id)
    const identity = planIdentity(run)
    const entry = entries.get(identity)
    if (entry === undefined || entry.decision !== "include") {
      errors.push(`run ${run.run_id} has no included plan entry`)
      continue
    }
    if (observedEntries.has(identity)) {
      errors.push(`plan entry ${entry.agent}/${entry.sequence} has more than one run`)
    }
    observedEntries.add(identity)
    validateRunState(run, errors)
    validateRunModel(run, entry, errors)
    if (run.workflow_verdict !== null && !allowedVerdicts.has(run.workflow_verdict)) {
      errors.push(`run ${run.run_id} has a verdict outside ${payload.workflow_result.kind}`)
    }
  }
  for (const entry of payload.plan) {
    if (entry.decision === "include" && !observedEntries.has(planIdentity(entry))) {
      errors.push(`included plan entry ${entry.agent}/${entry.sequence} has no run`)
    }
  }
}

function validateRunCounts(payload: ReviewDispatchPayload, errors: string[]): void {
  const actual: Record<(typeof TERMINAL_STATUS_VALUES)[number], number> = {
    success: 0,
    failed: 0,
    skipped: 0,
    budget_skipped: 0,
    limit_skipped: 0,
    cancelled: 0,
  }
  for (const run of payload.runs) {
    if (run.status in actual) {
      actual[run.status as keyof typeof actual] += 1
    }
  }
  for (const status of TERMINAL_STATUS_VALUES) {
    if (payload.run_counts[status] !== actual[status]) {
      errors.push(`run_counts.${status} is ${payload.run_counts[status]}; expected ${actual[status]}`)
    }
  }
  const sum = TERMINAL_STATUS_VALUES.reduce((total, status) => total + payload.run_counts[status], 0)
  if (payload.run_counts.total !== payload.runs.length || payload.run_counts.total !== sum) {
    errors.push(`run_counts.total must equal runs.length and the terminal count sum`)
  }
}

function validateFindings(payload: ReviewDispatchPayload, errors: string[]): void {
  const successfulIDs = new Set(payload.runs.filter((run) => run.status === "success").map((run) => run.run_id))
  const contributors = new Map<string, string[]>()
  for (const run of payload.runs) {
    if (run.status !== "success") {
      continue
    }
    for (const finding of run.findings) {
      const key = normalizedFindingKey(finding)
      contributors.set(key, [...(contributors.get(key) ?? []), run.run_id])
    }
  }

  const consolidatedKeys = new Set<string>()
  for (const finding of payload.findings) {
    const key = normalizedFindingKey(finding)
    if (consolidatedKeys.has(key)) {
      errors.push(`duplicate consolidated finding for normalized file and root cause`)
    }
    consolidatedKeys.add(key)
    const expected = uniqueSorted(contributors.get(key) ?? [])
    if (!equalSets(finding.run_ids, expected)) {
      errors.push(`consolidated finding has incorrect contributing run ids`)
    }
    if (finding.run_ids.some((runID) => !successfulIDs.has(runID))) {
      errors.push(`consolidated finding references a non-successful or unknown run`)
    }
  }
  for (const key of contributors.keys()) {
    if (!consolidatedKeys.has(key)) {
      errors.push(`run finding is missing from consolidated findings`)
    }
  }

  const advisoryKeys = new Set<string>()
  for (const advisory of payload.advisories) {
    const key = advisory.description.normalize("NFC").trim().replace(/\s+/g, " ").toLowerCase()
    if (advisoryKeys.has(key)) {
      errors.push(`duplicate advisory description`)
    }
    advisoryKeys.add(key)
    if (advisory.run_ids.some((runID) => !successfulIDs.has(runID))) {
      errors.push(`advisory references a non-successful or unknown run`)
    }
  }
}

function validateContext(
  payload: ReviewDispatchPayload,
  provenance: z.output<typeof ArtifactProvenanceSchema>,
  errors: string[],
): void {
  const issueWorkflow = payload.command === "triage-issue"
  if (issueWorkflow !== (payload.input_context.kind === "issue" && payload.change_profile.kind === "issue")) {
    errors.push(`input context and change profile do not match the workflow`)
  }
  if (payload.input_context.kind === "issue" && payload.change_profile.kind === "issue") {
    if (payload.input_context.content_sha256 !== payload.change_profile.content_sha256) {
      errors.push(`issue input context and change profile content hashes differ`)
    }
  }
  if (payload.input_context.kind !== "issue" && provenance.commit !== payload.input_context.head_sha) {
    errors.push(`artifact commit must equal the immutable reviewed head SHA`)
  }
}

function validatePayloadSemantics(
  payload: ReviewDispatchPayload,
  provenance: z.output<typeof ArtifactProvenanceSchema>,
): string[] {
  const errors: string[] = []
  validateContext(payload, provenance, errors)
  validatePlanAndRuns(payload, errors)
  validateRunCounts(payload, errors)
  validateFindings(payload, errors)
  const expectedWorkflowValue = calculatedWorkflowValue(payload)
  if (payload.workflow_result.value !== expectedWorkflowValue) {
    errors.push(`workflow_result.value is ${payload.workflow_result.value}; expected ${expectedWorkflowValue}`)
  }
  const expectedVerdict = projectNativeValue(expectedWorkflowValue)
  if (payload.verdict !== expectedVerdict) {
    errors.push(`verdict is ${payload.verdict}; expected ${expectedVerdict}`)
  }
  return uniqueSorted(errors)
}

function isErrno(error: unknown, code: string): boolean {
  return typeof error === "object" && error !== null && "code" in error && error.code === code
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "unknown persistence failure"
}

async function cleanupTemporary(path: string | null, dependencies: FinalizationDependencies): Promise<void> {
  if (path === null) {
    return
  }
  try {
    await dependencies.unlink(path)
  } catch (error: unknown) {
    if (!isErrno(error, "ENOENT")) {
      throw error
    }
  }
}

function nextUUID(dependencies: FinalizationDependencies): string {
  return UUIDSchema.parse(dependencies.uuid())
}

function envelopeFor(
  payload: ReviewDispatchPayload,
  provenance: z.output<typeof ArtifactProvenanceSchema>,
  timestamp: string,
): ReviewDispatchEnvelope {
  return {
    hero: "the-divisor",
    version: PRODUCER_VERSION,
    timestamp,
    artifact_type: "review-dispatch",
    schema_version: REVIEW_DISPATCH_SCHEMA_VERSION,
    context: {
      branch: provenance.branch,
      commit: provenance.commit,
      correlation_id: payload.correlation_id,
      workflow_id: provenance.workflow_id,
    },
    payload,
  }
}

async function persistEnvelope(
  input: z.output<typeof FinalizeInputSchema>,
  dependencies: FinalizationDependencies,
  collisionAttempt = 0,
): Promise<{ readonly envelope: ReviewDispatchEnvelope; readonly path: string }> {
  const directory = resolve(dependencies.projectRoot, DISPATCH_ARTIFACT_DIRECTORY)
  await dependencies.mkdir(directory, { recursive: true, mode: 0o750 })
  await dependencies.chmod(directory, 0o750)
  const correlationID = input.payload.correlation_id
  const payload = { ...input.payload, correlation_id: correlationID }
  const timestamp = dependencies.now().toISOString()
  const envelope = envelopeFor(payload, input.provenance, timestamp)
  const targetPath = join(directory, `${correlationID}.json`)
  let temporaryPath: string | null = null
  let handle: ArtifactFileHandle | null = null
  let published = false
  try {
    temporaryPath = join(directory, `.${correlationID}.${nextUUID(dependencies)}.tmp`)
    handle = await dependencies.open(temporaryPath, "wx", 0o600)
    await handle.writeFile(`${JSON.stringify(envelope, null, 2)}\n`, "utf8")
    await handle.sync()
    await handle.close()
    handle = null
    await dependencies.chmod(temporaryPath, 0o600)
    await dependencies.link(temporaryPath, targetPath)
    published = true
    await dependencies.unlink(temporaryPath)
    temporaryPath = null
    return { envelope, path: targetPath }
  } catch (error: unknown) {
    if (handle !== null) {
      try {
        await handle.close()
      } catch {
        // The original write failure remains the actionable persistence error.
      }
    }
    try {
      await cleanupTemporary(temporaryPath, dependencies)
    } catch (cleanupError: unknown) {
      throw new Error(`${errorMessage(error)}; temporary cleanup failed: ${errorMessage(cleanupError)}`)
    }
    if (published) {
      return { envelope, path: targetPath }
    }
    if (isErrno(error, "EEXIST") && collisionAttempt + 1 < MAX_CORRELATION_ATTEMPTS) {
      return persistEnvelope(
        { ...input, payload: { ...input.payload, correlation_id: nextUUID(dependencies) } },
        dependencies,
        collisionAttempt + 1,
      )
    }
    throw error
  }
}

function structuralFailure(error: z.ZodError): FinalizeReviewDispatchResult {
  return {
    status: "failed",
    operation_verdict: "INCONCLUSIVE",
    correlation_id: null,
    artifact_path: null,
    review_dispatch: null,
    review_verdict: null,
    calculated_assessment: null,
    errors: [`validate review dispatch input: ${formatValidationError(error)}`],
  }
}

/**
 * Closed-validates, semantically verifies, and atomically persists one finalized review dispatch.
 * @param rawInput Untrusted finalization input containing payload and envelope provenance.
 * @param dependencies Injectable filesystem, clock, UUID, and project-root operations.
 * @returns Structured dispatch and canonical decision output, or a fail-closed result.
 */
export async function finalizeReviewDispatch(
  rawInput: FinalizeReviewDispatchInput,
  dependencies: FinalizationDependencies,
): Promise<FinalizeReviewDispatchResult> {
  const parsed = FinalizeInputSchema.safeParse(rawInput)
  if (!parsed.success) {
    return structuralFailure(parsed.error)
  }
  const semanticErrors = validatePayloadSemantics(parsed.data.payload, parsed.data.provenance)
  if (semanticErrors.length > 0) {
    return {
      status: "failed",
      operation_verdict: "INCONCLUSIVE",
      correlation_id: parsed.data.payload.correlation_id,
      artifact_path: null,
      review_dispatch: null,
      review_verdict: null,
      calculated_assessment: null,
      errors: semanticErrors,
    }
  }

  const reviewVerdict = projectReviewVerdict(parsed.data.payload)
  try {
    const persisted = await persistEnvelope(parsed.data, dependencies)
    const persistedVerdict = projectReviewVerdict(persisted.envelope.payload)
    return {
      status: "success",
      operation_verdict: persisted.envelope.payload.verdict,
      correlation_id: persisted.envelope.payload.correlation_id,
      artifact_path: persisted.path,
      review_dispatch: persisted.envelope,
      review_verdict: persistedVerdict,
      calculated_assessment: {
        authoritative: true,
        payload: persisted.envelope.payload,
        review_verdict: persistedVerdict,
      },
      errors: [],
    }
  } catch (error: unknown) {
    return {
      status: "failed",
      operation_verdict: "INCONCLUSIVE",
      correlation_id: parsed.data.payload.correlation_id,
      artifact_path: null,
      review_dispatch: null,
      review_verdict: null,
      calculated_assessment: {
        authoritative: false,
        payload: parsed.data.payload,
        review_verdict: reviewVerdict,
      },
      errors: [`persist review dispatch artifact: ${errorMessage(error)}`],
    }
  }
}

/**
 * Creates production finalization dependencies rooted at one OpenCode worktree.
 * @param projectRoot Absolute project worktree path.
 * @returns Node-backed dependencies; callers may replace operations for isolated tests.
 */
export function createFinalizationDependencies(projectRoot: string): FinalizationDependencies {
  return {
    projectRoot,
    mkdir,
    chmod,
    open: async (path: string, flags: string, mode?: number): Promise<ArtifactFileHandle> => {
      const handle = await open(path, flags, mode)
      return {
        writeFile: async (data: string, encoding: "utf8"): Promise<void> => {
          await handle.writeFile(data, { encoding })
        },
        sync: async (): Promise<void> => handle.sync(),
        close: async (): Promise<void> => handle.close(),
      }
    },
    link,
    unlink,
    now: (): Date => new Date(),
    uuid: (): string => randomUUID(),
  }
}


/**
 * Creates the plan_review_dispatch OpenCode tool with injected dependencies.
 * @param dependencies Policy I/O and YAML parsing dependencies.
 * @returns An OpenCode tool definition.
 */
export function createPlanReviewDispatchTool(
  dependencies: PlannerDependencies,
): ReturnType<typeof tool> {
  return tool({
    description:
      "Create a deterministic, fail-closed Divisor review dispatch plan " +
      "from discovered agents and raw diff or issue inputs.",
    args: {
      mode: PlanInputSchema.shape.mode,
      discovered_agents: PlanInputSchema.shape.discovered_agents,
      full: z.boolean().optional().default(false),
      augment: z.boolean().optional().default(false),
      changed_files: PlanInputSchema.shape.changed_files,
      issue: PlanInputSchema.shape.issue,
    },
    async execute(args): Promise<string> {
      return serializeDispatchPlan(await planReviewDispatch(args, dependencies))
    },
  })
}

/**
 * Creates the finalize_review_dispatch OpenCode tool with injectable durable-write dependencies.
 * @param dependencies Filesystem, clock, UUID, and worktree dependencies.
 * @returns An OpenCode tool definition with closed bounded input schemas.
 */
export function createFinalizeReviewDispatchTool(
  dependencies: FinalizationDependencies,
): ReturnType<typeof tool> {
  return tool({
    description:
      "Validate a terminal Divisor review dispatch, persist its HIC artifact atomically, " +
      "and return canonical review-verdict version 2 decision data.",
    args: {
      payload: ReviewDispatchPayloadSchema,
      provenance: ArtifactProvenanceSchema,
    },
    async execute(args): Promise<string> {
      return `${JSON.stringify(await finalizeReviewDispatch(args, dependencies), null, 2)}\n`
    },
  })
}

// ── dispatch_agent_run ────────────────────────────────────────

/** Session metadata persisted on the first dispatch_agent_run call so
 * consolidate_dispatch can reconstruct the full finalization payload
 * without requiring the agent to re-supply plan, context, and coverage. */
const SessionMetadataSchema = z
  .object({
    command: z.enum(["review-council", "triage-issue", "address-feedback", "speckit-testreview"]),
    mode: z.enum(["code", "specs", "triage", "feedback", "test"]),
    full: z.boolean(),
    input_context: InputContextSchema,
    change_profile: ChangeProfileSchema,
    plan: z.array(DispatchPlanEntrySchema).max(512),
    coverage: CoverageSchema,
  })
  .strict()

const SESSION_METADATA_FILE = "session-metadata.json"

const DispatchAgentRunInputSchema = z
  .object({
    agent: AgentNameSchema,
    prompt: PromptSchema.optional(),
    promptFile: PromptFileSchema.optional(),
    tier: TierSchema.optional(),
    model: ModelSchema.optional(),
    variant: VariantSchema.optional(),
    read_only: z.boolean().optional(),
    timeout: TimeoutSchema.optional(),
    source: z.enum(SOURCE_VALUES).optional(),
    sequence: z.number().int().positive().optional(),
    session_metadata: SessionMetadataSchema.optional(),
  })
  .strict()
  .superRefine((input, context) => {
    if (input.prompt !== undefined && input.promptFile !== undefined) {
      context.addIssue({
        code: "custom",
        path: ["promptFile"],
        message: "prompt and promptFile are mutually exclusive",
      })
    }
    if (input.prompt === undefined && input.promptFile === undefined) {
      context.addIssue({
        code: "custom",
        path: ["prompt"],
        message: "one of prompt or promptFile is required",
      })
    }
    if (input.tier !== undefined && input.model !== undefined) {
      context.addIssue({
        code: "custom",
        path: ["model"],
        message: "tier and model are mutually exclusive; use tier for matrix-resolved models or model for direct bypass",
      })
    }
  })

/** Dependencies injected into the dispatch_agent_run tool handler. */
export interface DispatchAgentRunDependencies {
  readonly plannerDependencies: PlannerDependencies
  readonly client: PluginInput["client"]
  readonly directory: string
}

/**
 * Creates the dispatch_agent_run tool for tier-based or direct-model agent invocation.
 *
 * Resolves models via the review matrix (tier-based) or accepts direct provider/model-id
 * bypass. Validates agents against the reviewer manifest. Delegates session lifecycle
 * to the shared `executeAgentSession()` executor.
 *
 * @param deps Injected planner dependencies, client, session context, and project root.
 * @returns The OpenCode tool definition for dispatch_agent_run.
 */
export function createDispatchAgentRunTool(deps: DispatchAgentRunDependencies): ReturnType<typeof tool> {
  return tool({
    description:
      "Execute one already-planned Divisor review-agent run in a parented child session with explicit model provenance.",
    args: {
      agent: AgentNameSchema.describe("Manifest-declared divisor-* review agent."),
      prompt: PromptSchema.optional().describe("Bounded full prompt for this single planned run."),
      promptFile: PromptFileSchema.optional().describe(
        "Absolute path to a file containing the prompt text. Mutually exclusive with prompt. Max 1024 chars.",
      ),
      tier: TierSchema.optional().describe(
        'Review matrix profile tier. Mutually exclusive with model. Defaults to "standard" when neither is provided.',
      ),
      model: ModelSchema.optional().describe(
        "Optional direct provider/model-id bypassing tier resolution. Include the full slug with any @suffix (e.g. provider/model-id@default). Mutually exclusive with tier.",
      ),
      variant: VariantSchema.optional().describe("Optional direct-model runtime variant."),
      read_only: z.boolean().optional().describe("Invocation provenance only; does not alter permissions."),
      timeout: TimeoutSchema.optional().describe("Optional run timeout in milliseconds (bounded to 1_800_000)."),
      source: z.enum(SOURCE_VALUES).optional().describe('Run source from the dispatch plan (explicit, advisor, host). Defaults to "explicit".'),
      sequence: z.number().int().positive().optional().describe("Run sequence number from the dispatch plan. Defaults to 1."),
      session_metadata: SessionMetadataSchema.optional().describe(
        "Session metadata (command, mode, full, input_context, change_profile, plan, coverage) to persist " +
          "for consolidate_dispatch. Supply on the FIRST dispatch_agent_run call only; subsequent calls ignore it. " +
          "When omitted, consolidate_dispatch requires a prior call that stored metadata.",
      ),
    },
    async execute(args, context): Promise<{ readonly output: string; readonly metadata: Record<string, unknown> }> {
      const result = await dispatchAgentRun(args, context, deps)
      const summary =
        result.error === null
          ? `ok (correlation_id=${result.correlation_id}, run_id=${result.run_id})`
          : `dispatch_agent_run error: ${result.error.message}`
      return {
        output: summary,
        metadata: { ...result },
      }
    },
  })
}

/** Build a provenance placeholder for early-exit error paths. */
function makeProvenance(
  agent: string,
  requestedModel: string | null,
  requestedVariant: string | null,
  readOnly: boolean,
): InvocationProvenance {
  return {
    agent,
    requested_model: requestedModel,
    requested_variant: requestedVariant,
    resolved_parent_model: null,
    resolved_parent_variant: null,
    reported_child_model: null,
    model_mismatch: false,
    read_only: readOnly,
  }
}

/**
 * Execute a single Divisor review-agent run in a parented child session.
 *
 * Unlike invoke_agent (which resolves the host model from the parent session), dispatch_agent_run
 * resolves models from the review matrix profiles. This gives the dispatch planner
 * deterministic control over which model each agent receives.
 *
 * @param rawInput Untrusted invocation arguments from the OpenCode tool boundary.
 * @param context Calling tool context used for cancellation signal.
 * @param deps Injected planner dependencies, client, session context, and project root.
 * @returns Observable result with status, text, usage, error, provenance, correlation_id, and run_id.
 */
export async function dispatchAgentRun(
  rawInput: unknown,
  context: ToolContext,
  deps: DispatchAgentRunDependencies,
): Promise<DispatchAgentRunResult> {
  const parsed = DispatchAgentRunInputSchema.safeParse(rawInput)
  if (!parsed.success) {
    const placeholder = makeProvenance("invalid", null, null, true)
    return { ...failedResult(placeholder, "invalid_input", parsed.error.message, false), correlation_id: randomUUID(), run_id: randomUUID() }
  }

  const input = parsed.data
  const readOnly = input.read_only ?? true
  const timeoutMilliseconds = input.timeout ?? DEFAULT_TIMEOUT_MILLISECONDS

  // Resolve or auto-generate the correlation_id for this parent session.
  // All dispatch_agent_run calls from the same parent session share one ID.
  let correlationId = sessionCorrelationMap.get(context.sessionID)
  const isFirstCall = correlationId === undefined
  if (isFirstCall) {
    correlationId = randomUUID()
    sessionCorrelationMap.set(context.sessionID, correlationId)
  }

  // Persist session metadata on the first call so consolidate_dispatch can
  // reconstruct the finalization payload without the agent re-supplying it.
  if (isFirstCall && input.session_metadata !== undefined) {
    try {
      const dir = dispatchSessionDir(correlationId)
      await mkdir(dir, { recursive: true })
      await writeFile(join(dir, SESSION_METADATA_FILE), JSON.stringify(input.session_metadata, null, 2), "utf8")
    } catch {
      // Best-effort; consolidate_dispatch will report a clear error if missing.
    }
  }

  /** Wrap a failed InvokeAgentResult with auto-generated IDs for early exits. */
  const earlyExit = (result: InvokeAgentResult): DispatchAgentRunResult => ({
    ...result,
    correlation_id: correlationId,
    run_id: randomUUID(),
  })

  /** Build a provenance placeholder using the parsed input's common fields. */
  const earlyProvenance = (modelOverride?: string) =>
    makeProvenance(
      input.agent,
      modelOverride ?? input.model ?? input.tier ?? "standard",
      input.variant ?? null,
      readOnly,
    )

  try {
    validateTimeout(timeoutMilliseconds)
  } catch (error: unknown) {
    return earlyExit(failedResult(earlyProvenance(), "invalid_invocation", errorText(error), false))
  }

  // Resolve prompt text: inline prompt or file-backed prompt.
  let promptText: string
  if (input.promptFile !== undefined) {
    try {
      const fileContent = await deps.plannerDependencies.readText(input.promptFile)
      if (Buffer.byteLength(fileContent, "utf8") > MAX_PROMPT_FILE_BYTES) {
        return earlyExit(failedResult(earlyProvenance(), "invalid_input", "promptFile content exceeds 4 MiB UTF-8", false))
      }
      promptText = fileContent
    } catch (error: unknown) {
      return earlyExit(failedResult(earlyProvenance(), "prompt_file_read_failed", sanitizeInvocationError(errorText(error)), false))
    }
  } else {
    // superRefine guarantees exactly one of prompt/promptFile is present.
    promptText = input.prompt!
  }

  // Load policies once — provides both manifest (for agent validation) and matrix (for tier resolution).
  let policies: { readonly matrix: ReviewMatrix; readonly manifest: ReviewerManifest }
  try {
    policies = await loadPolicies(deps.plannerDependencies)
  } catch (error: unknown) {
    return earlyExit(failedResult(earlyProvenance(), "policy_load_failed", sanitizeInvocationError(errorText(error)), true))
  }

  // Validate the agent exists in the reviewer manifest.
  const reviewer = policies.manifest.reviewers.find((candidate: Reviewer) => candidate.agent === input.agent)
  if (reviewer === undefined) {
    return earlyExit(failedResult(
      earlyProvenance(),
      "unknown_agent",
      `agent ${input.agent} is absent from the reviewer manifest`,
      false,
    ))
  }

  // Resolve model: tier-based via review matrix or direct bypass.
  let selectedModel: ModelIdentity
  let selectedVariant: string | undefined
  let requestedModelLabel: string

  if (input.model !== undefined) {
    // Direct model bypass — skip matrix lookup.
    selectedModel = directModelIdentity(input.model)
    selectedVariant = input.variant
    requestedModelLabel = input.model
  } else {
    // Tier-based resolution via review matrix.
    const tier = input.tier ?? "standard"
    const resolved = profilePair(policies.matrix, tier, input.variant)
    if (resolved.model === null) {
      return earlyExit(failedResult(
        earlyProvenance(tier),
        "tier_model_unavailable",
        `profile "${tier}" has no model configured in the review matrix`,
        false,
      ))
    }

    selectedModel = directModelIdentity(resolved.model)
    selectedVariant = resolved.variant ?? undefined
    requestedModelLabel = resolved.model
  }

  const provenance = makeProvenance(
    input.agent,
    requestedModelLabel,
    selectedVariant ?? null,
    readOnly,
  )

  // Delegate session lifecycle to the shared executor.
  const startedAt = new Date().toISOString()
  const executorDeps: ExecutorDependencies = {
    client: deps.client,
    sessionID: context.sessionID,
    directory: deps.directory,
    parentAbort: context.abort,
    timeoutMilliseconds,
  }

  const result = await executeAgentSession(executorDeps, input.agent, promptText, selectedModel, selectedVariant, readOnly, provenance)
  const finishedAt = new Date().toISOString()

  // Harvest submitted findings/proposals from child session via the in-process Map.
  const childSessionID = result.childSessionID
  const submitted = childSessionID !== undefined ? submissionStore.get(childSessionID) : undefined
  if (childSessionID !== undefined) {
    submissionStore.delete(childSessionID)
  }

  const harvestedFindings = submitted?.findings ?? []
  const harvestedProposals = submitted?.proposals ?? []

  // Reuse the session-scoped correlation_id; generate run_id per invocation.
  const runId = randomUUID()
  const runData: PersistedRunData = {
    correlation_id: correlationId,
    run_id: runId,
    agent: input.agent,
    source: input.source ?? "explicit",
    sequence: input.sequence ?? 1,
    status: result.status,
    started_at: startedAt,
    finished_at: finishedAt,
    requested_model: requestedModelLabel,
    provider: selectedModel.providerID || null,
    model_id: selectedModel.modelID || null,
    variant: selectedVariant ?? null,
    resolved_parent_model: result.provenance.resolved_parent_model,
    resolved_parent_variant: result.provenance.resolved_parent_variant,
    reported_model: result.provenance.reported_child_model,
    model_mismatch: result.provenance.model_mismatch,
    usage: result.usage,
    error: result.error,
    workflow_verdict: null,
    findings: harvestedFindings,
    proposals: harvestedProposals,
    text: result.text,
    read_only: readOnly,
  }
  try {
    const dir = dispatchSessionDir(correlationId)
    await mkdir(dir, { recursive: true })
    await writeFile(join(dir, `run-${input.agent}.json`), JSON.stringify(runData, null, 2), "utf8")
  } catch {
    // Persistence is best-effort; the result is authoritative.
  }

  return { ...result, correlation_id: correlationId, run_id: runId }
}

// ── submit_review_findings ───────────────────────────────────

/**
 * Creates the submit_review_findings tool for child Divisor agent sessions.
 * Findings are Zod-validated and stored in the module-level Map for later
 * harvesting by dispatch_agent_run.
 */
function createSubmitReviewFindingsTool(): ReturnType<typeof tool> {
  return tool({
    description:
      "Submit structured review findings from a child Divisor agent run. " +
      "Findings are Zod-validated and stored for later consolidation. " +
      "Call once per run; multiple calls in the same session are additive.",
    args: {
      findings: z
        .array(RunFindingSchema)
        .min(1)
        .max(512)
        .describe("Array of review findings with severity, category, description, root_cause, file, and line."),
    },
    async execute(args, context): Promise<{ readonly output: string }> {
      const entry = submissionStore.get(context.sessionID) ?? { findings: [], proposals: [] }
      entry.findings.push(...args.findings)
      submissionStore.set(context.sessionID, entry)
      return { output: `${args.findings.length} finding(s) submitted` }
    },
  })
}

// ── submit_lesson_proposal ───────────────────────────────────

/**
 * Creates the submit_lesson_proposal tool for child Divisor agent sessions.
 * Proposals are stored in the module-level Map for later harvesting.
 */
function createSubmitLessonProposalTool(): ReturnType<typeof tool> {
  return tool({
    description:
      "Submit a lesson proposal (insight, pattern, gotcha) from a child Divisor agent run. " +
      "Proposals are Zod-validated and stored for later processing by consolidate_dispatch.",
    args: {
      information: z.string().min(1).max(16_384).describe("The learning text to store."),
      tag: z.string().min(1).max(128).describe("Required topic tag (e.g. authentication, review-dispatch)."),
      category: z
        .enum(["decision", "pattern", "gotcha", "context", "reference"])
        .optional()
        .describe("Optional category for the learning."),
    },
    async execute(args, context): Promise<{ readonly output: string }> {
      const entry = submissionStore.get(context.sessionID) ?? { findings: [], proposals: [] }
      entry.proposals.push({ information: args.information, tag: args.tag, category: args.category })
      submissionStore.set(context.sessionID, entry)
      return { output: `lesson proposal submitted (tag: ${args.tag})` }
    },
  })
}

// ── consolidate_dispatch ─────────────────────────────────────

const ConsolidateDispatchInputSchema = z
  .object({
    provenance: ArtifactProvenanceSchema.optional(),
  })
  .strict()

/**
 * Normalize a file path for deduplication: lowercase, forward-slash, trim whitespace.
 */
function normalizeFile(file: string | null): string {
  return file?.toLowerCase().replace(/\\/g, "/").trim() ?? ""
}

/**
 * Deduplicate findings across all runs by normalized file + root_cause.
 * When duplicates exist, keep the highest severity and merge run_ids.
 */
function deduplicateFindings(
  runs: readonly PersistedRunData[],
): z.infer<typeof ConsolidatedFindingSchema>[] {
  const groups = new Map<
    string,
    { finding: z.infer<typeof RunFindingSchema>; run_ids: Set<string> }
  >()

  for (const run of runs) {
    for (const finding of run.findings) {
      const key = `${normalizeFile(finding.file)}\0${finding.root_cause}`
      const existing = groups.get(key)
      if (existing !== undefined) {
        existing.run_ids.add(run.run_id)
        if ((SEVERITY_RANK[finding.severity] ?? 0) > (SEVERITY_RANK[existing.finding.severity] ?? 0)) {
          existing.finding = { ...finding }
        }
      } else {
        groups.set(key, { finding: { ...finding }, run_ids: new Set([run.run_id]) })
      }
    }
  }

  return [...groups.values()].map(({ finding, run_ids }) => ({
    ...finding,
    run_ids: [...run_ids].sort(),
  }))
}

/**
 * Compute council verdict from consolidated findings.
 */
function computeVerdict(findings: readonly z.infer<typeof ConsolidatedFindingSchema>[]): {
  readonly verdict: (typeof GENERIC_VERDICT_VALUES)[number]
  readonly reason: string
} {
  const hasCritical = findings.some((f) => f.severity === "CRITICAL")
  const hasHigh = findings.some((f) => f.severity === "HIGH")
  const hasMedium = findings.some((f) => f.severity === "MEDIUM")
  const hasLow = findings.some((f) => f.severity === "LOW")

  if (hasCritical || hasHigh) {
    const highest = hasCritical ? "CRITICAL" : "HIGH"
    const count = findings.filter((f) => f.severity === highest).length
    return {
      verdict: "REQUEST CHANGES",
      reason: `${count} ${highest} severity finding(s) require changes`,
    }
  }
  if (hasMedium) {
    const count = findings.filter((f) => f.severity === "MEDIUM").length
    return {
      verdict: "APPROVE WITH ADVISORIES",
      reason: `${count} MEDIUM severity finding(s) noted as advisories`,
    }
  }
  if (hasLow) {
    return {
      verdict: "APPROVE WITH ADVISORIES",
      reason: `${findings.length} LOW severity finding(s) noted as advisories`,
    }
  }
  return {
    verdict: "APPROVE",
    reason: "No findings requiring changes",
  }
}

/**
 * Compute a per-run workflow verdict from the run's findings.
 * Successful runs MUST have a non-null workflow_verdict for semantic validation.
 */
function runVerdict(findings: readonly { readonly severity: string }[]): string {
  if (findings.some((f) => f.severity === "CRITICAL" || f.severity === "HIGH")) {
    return "REQUEST CHANGES"
  }
  if (findings.some((f) => f.severity === "MEDIUM" || f.severity === "LOW")) {
    return "APPROVE WITH ADVISORIES"
  }
  return "APPROVE"
}

/**
 * Derive the workflow_result kind from the command name.
 */
function workflowKind(command: string): "council" | "triage" | "feedback" | "test-review" {
  switch (command) {
    case "review-council":
      return "council"
    case "triage-issue":
      return "triage"
    case "address-feedback":
      return "feedback"
    case "speckit-testreview":
      return "test-review"
    default:
      return "council"
  }
}

/**
 * Map a generic verdict to a workflow-specific verdict value.
 */
function workflowVerdictValue(
  kind: ReturnType<typeof workflowKind>,
  verdict: (typeof GENERIC_VERDICT_VALUES)[number],
): string {
  if (kind === "triage") {
    switch (verdict) {
      case "APPROVE":
        return "VALID"
      case "REQUEST CHANGES":
        return "INVALID"
      default:
        return "NEEDS-CLARIFICATION"
    }
  }
  if (kind === "feedback") {
    switch (verdict) {
      case "APPROVE":
        return "ACCEPT"
      case "APPROVE WITH ADVISORIES":
        return "AUTHOR-DECIDES"
      default:
        return verdict
    }
  }
  return verdict
}

/**
 * Create advisories from MEDIUM and LOW findings for APPROVE WITH ADVISORIES verdicts.
 */
function buildAdvisories(
  findings: readonly z.infer<typeof ConsolidatedFindingSchema>[],
): z.infer<typeof AdvisorySchema>[] {
  return findings
    .filter((f) => f.severity === "MEDIUM" || f.severity === "LOW")
    .map((f) => ({
      severity: f.severity,
      description: f.description,
      run_ids: f.run_ids,
    }))
}

/**
 * Creates the consolidate_dispatch tool that reads persisted run files,
 * deduplicates findings, computes verdict, and returns a finalization-ready payload.
 * When provenance is supplied, calls finalizeReviewDispatch internally to skip
 * the lossy LLM round-trip and returns the finalization result directly.
 */
export function createConsolidateDispatchTool(
  finalizationDependencies: FinalizationDependencies,
): ReturnType<typeof tool> {
  return tool({
    description:
      "Consolidate all dispatch_agent_run results from this session, deduplicate findings, compute verdict, " +
      "collect lesson proposals, and optionally finalize the dispatch artifact in one step. " +
      "Reads session metadata (command, mode, plan, coverage, etc.) from the file persisted by the first " +
      "dispatch_agent_run call. When provenance is provided, calls finalize_review_dispatch internally " +
      "(no LLM round-trip) and returns the finalization result directly. When omitted, returns the payload JSON. " +
      "Lesson proposals submitted by child agents are collected and returned in a top-level proposals array; " +
      "call dewey_store_learning for each. " +
      "Automatically resolves the dispatch session from prior dispatch_agent_run calls.",
    args: {
      provenance: ArtifactProvenanceSchema.optional().describe(
        "When provided, consolidate_dispatch calls finalize_review_dispatch internally " +
          "and returns the finalization result. Requires branch (e.g. opsx/foo), " +
          "commit (40-char SHA), and workflow_id (e.g. uf.review-council).",
      ),
    },
    async execute(args, context): Promise<{ readonly output: string }> {
      const parsed = ConsolidateDispatchInputSchema.safeParse(args)
      if (!parsed.success) {
        return { output: `consolidate_dispatch validation failed: ${parsed.error.message}` }
      }
      const input = parsed.data

      // Resolve correlation_id from the session map — set by prior dispatch_agent_run calls.
      const correlationId = sessionCorrelationMap.get(context.sessionID)
      if (correlationId === undefined) {
        return { output: "consolidate_dispatch error: no dispatch session found for this session; call dispatch_agent_run first" }
      }

      // Load session metadata persisted by the first dispatch_agent_run call.
      const dir = dispatchSessionDir(correlationId)
      let sessionMetadata: z.infer<typeof SessionMetadataSchema>
      try {
        const raw = await readFile(join(dir, SESSION_METADATA_FILE), "utf8")
        const parsed2 = SessionMetadataSchema.safeParse(JSON.parse(raw))
        if (!parsed2.success) {
          return { output: `consolidate_dispatch error: invalid session metadata: ${parsed2.error.message}` }
        }
        sessionMetadata = parsed2.data
      } catch {
        return {
          output:
            "consolidate_dispatch error: no session metadata found; pass session_metadata on the first dispatch_agent_run call",
        }
      }

      // Read all run files from the dispatch session directory.
      let runFiles: string[]
      try {
        const entries = await readdir(dir)
        runFiles = entries.filter((name) => name.startsWith("run-") && name.endsWith(".json")).sort()
      } catch {
        return { output: `consolidate_dispatch error: no dispatch session directory found for correlation_id ${correlationId}` }
      }

      if (runFiles.length === 0) {
        return { output: "consolidate_dispatch error: no run files found in dispatch session directory" }
      }

      // Parse all run files.
      const runs: PersistedRunData[] = []
      for (const file of runFiles) {
        try {
          const content = await readFile(join(dir, file), "utf8")
          runs.push(JSON.parse(content) as PersistedRunData)
        } catch (error: unknown) {
          return { output: `consolidate_dispatch error: failed to read ${file}: ${errorText(error)}` }
        }
      }

      // Collect lesson proposals from all runs, tagged with agent name.
      const collectedProposals = runs.flatMap((run) =>
        run.proposals.map((p) => ({ agent: run.agent, information: p.information, tag: p.tag, category: p.category })),
      )

      // Deduplicate findings across all runs.
      const consolidatedFindings = deduplicateFindings(runs)

      // Compute verdict.
      const { verdict, reason: verdictReason } = computeVerdict(consolidatedFindings)

      // Build advisories for advisory-level verdicts.
      const advisories = verdict === "APPROVE WITH ADVISORIES" ? buildAdvisories(consolidatedFindings) : []

      // Build run entries matching DispatchRunSchema.
      const dispatchRuns = runs.map((run) => ({
        run_id: run.run_id,
        agent: run.agent,
        source: run.source as (typeof SOURCE_VALUES)[number],
        requested_model: run.requested_model,
        provider: run.provider,
        model_id: run.model_id,
        variant: run.variant,
        resolved_parent_model: run.resolved_parent_model,
        resolved_parent_variant: run.resolved_parent_variant,
        sequence: run.sequence,
        status: run.status as "success" | "failed" | "cancelled",
        started_at: run.started_at,
        finished_at: run.finished_at,
        usage: run.usage,
        error: run.error,
        reported_model: run.reported_model,
        model_mismatch: run.model_mismatch,
        workflow_verdict: run.status === "success" ? runVerdict(run.findings) : null,
        findings: run.findings,
      }))

      // Compute run counts.
      const runCounts = {
        total: runs.length,
        success: runs.filter((r) => r.status === "success").length,
        failed: runs.filter((r) => r.status === "failed").length,
        skipped: runs.filter((r) => r.status === "skipped").length,
        budget_skipped: 0,
        limit_skipped: 0,
        cancelled: runs.filter((r) => r.status === "cancelled").length,
      }

      // Build workflow result.
      const kind = workflowKind(sessionMetadata.command)
      const workflowResult = {
        kind,
        value: workflowVerdictValue(kind, verdict),
      }

      // Assemble the full finalization payload.
      const payload = {
        command: sessionMetadata.command,
        mode: sessionMetadata.mode,
        full: sessionMetadata.full,
        input_context: sessionMetadata.input_context,
        change_profile: sessionMetadata.change_profile,
        plan_version: 1 as const,
        plan: sessionMetadata.plan,
        runs: dispatchRuns,
        coverage: sessionMetadata.coverage,
        findings: consolidatedFindings,
        advisories,
        verdict,
        verdict_reason: verdictReason,
        run_counts: runCounts,
        workflow_result: workflowResult,
        correlation_id: correlationId,
      }

      // When provenance is provided, finalize internally — no LLM round-trip.
      if (input.provenance !== undefined) {
        const result = await finalizeReviewDispatch(
          { payload, provenance: input.provenance },
          finalizationDependencies,
        )
        return {
          output: JSON.stringify(
            { ...result, proposals: collectedProposals.length > 0 ? collectedProposals : undefined },
            null,
            2,
          ),
        }
      }

      return {
        output: JSON.stringify(
          { ...payload, proposals: collectedProposals.length > 0 ? collectedProposals : undefined },
          null,
          2,
        ),
      }
    },
  })
}

// ── dispatch_status ──────────────────────────────────────────

/**
 * Creates the dispatch_status query tool for inspecting dispatch session state.
 */
function createDispatchStatusTool(): ReturnType<typeof tool> {
  return tool({
    description:
      "Query the status of this session's dispatch. Automatically resolves the dispatch session " +
      "from prior dispatch_agent_run calls. Returns completed runs, statuses, finding counts, " +
      "and any pending submissions in the store.",
    args: {},
    async execute(_args, context): Promise<{ readonly output: string }> {
      const correlationId = sessionCorrelationMap.get(context.sessionID)
      if (correlationId === undefined) {
        return { output: JSON.stringify({ status: "not_found", message: "no dispatch session for this session", runs: [] }) }
      }
      const dir = dispatchSessionDir(correlationId)

      let runFiles: string[]
      try {
        const entries = await readdir(dir)
        runFiles = entries.filter((name) => name.startsWith("run-") && name.endsWith(".json")).sort()
      } catch {
        return { output: JSON.stringify({ status: "not_found", correlation_id: correlationId, runs: [] }) }
      }

      const runSummaries: Array<{
        agent: string
        status: string
        findings: number
        proposals: number
      }> = []

      for (const file of runFiles) {
        try {
          const content = await readFile(join(dir, file), "utf8")
          const run = JSON.parse(content) as PersistedRunData
          runSummaries.push({
            agent: run.agent,
            status: run.status,
            findings: run.findings.length,
            proposals: run.proposals.length,
          })
        } catch {
          runSummaries.push({ agent: file.replace(/^run-|\.json$/g, ""), status: "unreadable", findings: 0, proposals: 0 })
        }
      }

      // Check for any pending submissions in the store.
      const pendingSubmissions = submissionStore.size

      return {
        output: JSON.stringify(
          {
            status: "ok",
            correlation_id: correlationId,
            runs: runSummaries,
            total_runs: runSummaries.length,
            pending_submissions: pendingSubmissions,
          },
          null,
          2,
        ),
      }
    },
  })
}

/** Auto-discovered OpenCode policy plugin for deterministic review planning and finalization. */
export const ReviewDispatchPlugin = {
  id: "review-dispatch",
  server: async (input) => {
    const projectRoot = input.worktree || input.directory
    const client = input.client
    const dependencies: PlannerDependencies = {
      readText: async (relativePath: string): Promise<string> =>
        readFile(resolve(projectRoot, relativePath), "utf8"),
      parseYaml: createBunYamlParser(),
    }
    return {
      tool: {
        plan_review_dispatch: createPlanReviewDispatchTool(dependencies),
        finalize_review_dispatch: createFinalizeReviewDispatchTool(createFinalizationDependencies(projectRoot)),
        acquire_sibling_evidence: createAcquireSiblingEvidenceTool(
          createSiblingAcquisitionDependencies(projectRoot, createBunYamlParser()),
        ),
        prepare_lesson_learning: createPrepareLessonLearningTool(),
        dispatch_agent_run: createDispatchAgentRunTool({
          plannerDependencies: dependencies,
          client,
          directory: projectRoot,
        }),
        submit_review_findings: createSubmitReviewFindingsTool(),
        submit_lesson_proposal: createSubmitLessonProposalTool(),
        consolidate_dispatch: createConsolidateDispatchTool(createFinalizationDependencies(projectRoot)),
        dispatch_status: createDispatchStatusTool(),
      },
    }
  },
} satisfies PluginModule

export default ReviewDispatchPlugin

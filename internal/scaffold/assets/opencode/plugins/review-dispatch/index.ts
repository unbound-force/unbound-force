import { Buffer } from "node:buffer"
import { createHash, randomUUID } from "node:crypto"
import { chmod, link, mkdir, open, readFile, unlink } from "node:fs/promises"
import { join, resolve } from "node:path"

import { tool, type PluginModule } from "@opencode-ai/plugin"
import { z } from "zod"

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
const MANIFEST_PATH = ".uf/reviewer-capabilities.yaml"
const PLAN_VERSION = 1
const REVIEW_DISPATCH_SCHEMA_VERSION = "1.0.0"
const REVIEW_VERDICT_SCHEMA_VERSION = "2.0.0"
const PRODUCER_VERSION = "1.0.0"
const DISPATCH_ARTIFACT_DIRECTORY = ".uf/artifacts/dispatch"
const MAX_CORRELATION_ATTEMPTS = 32

const DEFAULT_LIMITS = {
  max_personas: 16,
  max_runs_per_persona: 3,
  max_total_runs: 24,
  max_parallel_runs: 4,
  per_run_timeout_seconds: 600,
  max_reported_cost_usd: 25,
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
    /^[A-Za-z0-9][A-Za-z0-9._-]*\/[A-Za-z0-9][A-Za-z0-9._-]*(?:\/[A-Za-z0-9][A-Za-z0-9._-]*)*$/,
  )
const VariantSchema = z.string().regex(/^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/)
const TierSchema = z.enum(TIER_VALUES)


const ProfileSchema = z
  .object({
    model: ModelSchema,
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

const LimitsSchema = z
  .object({
    max_personas: z.number().int().min(1).max(32).optional(),
    max_runs_per_persona: z.number().int().min(1).max(5).optional(),
    max_total_runs: z.number().int().min(1).max(64).optional(),
    max_parallel_runs: z.number().int().min(1).max(8).optional(),
    per_run_timeout_seconds: z.number().int().min(30).max(1800).optional(),
    max_reported_cost_usd: z.number().min(1).max(100).optional(),
  })
  .strict()

const ReviewMatrixSchema = z
  .object({
    version: z.literal(2),
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
    model_id: z.string().min(1).max(256).regex(/^[A-Za-z0-9][A-Za-z0-9._\/-]*$/).nullable(),
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

interface Limits {
  readonly max_personas: number
  readonly max_runs_per_persona: number
  readonly max_total_runs: number
  readonly max_parallel_runs: number
  readonly per_run_timeout_seconds: number
  readonly max_reported_cost_usd: number
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
 * Parses and closed-validates a version 2 review matrix.
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
  return { ...DEFAULT_LIMITS, ...matrix.limits }
}

function profilePair(matrix: ReviewMatrix, profileName: string, variant?: string): {
  readonly model: string
  readonly variant: string | null
  readonly tier: Tier | null
} {
  const profile = matrix.profiles[profileName]
  if (profile === undefined) {
    throw new Error(`unknown profile ${profileName}`)
  }
  return {
    model: profile.model,
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
    errors: [...errors].sort(),
  }
}

async function loadPolicies(dependencies: PlannerDependencies): Promise<{
  readonly matrix: ReviewMatrix
  readonly manifest: ReviewerManifest
}> {
  const [matrixText, manifestText] = await Promise.all([
    dependencies.readText(MATRIX_PATH),
    dependencies.readText(MANIFEST_PATH),
  ])
  return {
    matrix: parseReviewMatrix(matrixText, dependencies.parseYaml),
    manifest: parseReviewerManifest(manifestText, dependencies.parseYaml),
  }
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

/** Auto-discovered OpenCode policy plugin for deterministic review planning and finalization. */
export const ReviewDispatchPlugin = {
  id: "review-dispatch",
  server: async (input) => {
    const projectRoot = input.worktree || input.directory
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
      },
    }
  },
} satisfies PluginModule

export default ReviewDispatchPlugin

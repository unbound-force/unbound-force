import type { PluginInput, PluginModule, ToolContext } from "@opencode-ai/plugin"
import { execFileSync, execSync } from "node:child_process"
import { createHash, randomUUID } from "node:crypto"
import { mkdir, mkdtemp, readFile, rm, stat, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import type { ToolResult } from "../lib/uf-workflow-types.js"
import { formatUnleashExit, type UnleashExitData, type UnleashExitVariant } from "../lib/unleash-templates.js"
import {
  LESSON_PROPOSAL_BEGIN,
  LESSON_PROPOSAL_END,
  ReviewDispatchPlugin,
  UNTRUSTED_SIBLING_EVIDENCE_BEGIN,
  UNTRUSTED_SIBLING_EVIDENCE_END,
  _getSessionCorrelationMap,
} from "../plugins/review-dispatch/index.js"
import UfFinalePlugin from "../plugins/uf-finale/index.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"
import { withBun } from "./helpers/bun-runtime.js"

// ---------------------------------------------------------------------------
// Golden output-shape contract tests (seam-verification task 2.2, DE-002)
//
// Fixture source: the task-2.1 site registry in
// openspec/changes/seam-verification/metrics.md. Every documented
// command-to-tool call site lists the output fields the command relies
// on; this suite executes each documented tool through the plugin
// registry -- the same `module.server()` registration and
// `tool.execute()` dispatch path the OpenCode host uses -- against real
// scratch-project filesystems with DI'd git/gh fakes, then asserts the
// documented fields on the returned envelope.
//
// Coverage: 43 of the 47 documented sites. The four dispatch_review_runs
// sites (uf.triage-issue.md:263, uf.review-council.md:276,
// uf.review-pr.md:394, uf.address-feedback.md:363) belong to task 2.3
// and are pinned in dispatch-review-runs.test.ts instead — one golden
// per documented input variant (triage, code, feedback).
//
// Per-plugin site counts (registered-tool attribution):
// - uf-workflow: 21 sites -- get_progress x4 (uf.unleash.md:55,
//   uf.finale.md:42, uf.review-council.md:51, uf.address-feedback.md:71),
//   uf_unleash_branch_gate (uf.unleash.md:81), uf_unleash_detect_resume
//   (uf.unleash.md:104), scan_clarifications (uf.unleash.md:169),
//   format_unleash_exit x8 (uf.unleash.md:235, 423, 558, 610, 666, 733,
//   862, 892), parse_tasks (uf.unleash.md:458), mark_task_complete
//   (uf.unleash.md:515), scan_conflict_markers (uf.unleash.md:598),
//   format_unleash_demo (uf.unleash.md:1010), resolve_base_ref
//   (uf.review-council.md:126), normalize_model_name
//   (uf.address-feedback.md:801)
// - uf-finale: 8 sites -- uf_finale_startup (uf.finale.md:51),
//   uf_finale_stage (71), uf_finale_commit_context (82),
//   uf_finale_push_state (131), uf_finale_pr_context (154),
//   uf_finale_create_pr (204), uf_finale_ci_poll (212),
//   uf_finale_return_to_main (291)
// - review-dispatch: 14 sites -- plan_review_dispatch x7 (uf.unleash.md:353
//   and 799, uf.review-council.md:231, uf.review-pr.md:353,
//   uf.triage-issue.md:190, uf.address-feedback.md:339,
//   speckit.testreview.md:154), consolidate_dispatch x3
//   (uf.review-council.md:361, uf.triage-issue.md:780,
//   uf.address-feedback.md:541), finalize_review_dispatch x2
//   (uf.review-pr.md:442, speckit.testreview.md:278),
//   acquire_sibling_evidence (speckit.testreview.md:134),
//   prepare_lesson_learning (speckit.testreview.md:268)
//
// Envelope contract under test: uf-workflow and uf-finale tools return
// `{ output: JSON.stringify(ToolResult) }` where ToolResult is the
// `{status: "ok", data}` / `{status: "error", message, retryable,
// code?}` discriminated union from lib/uf-workflow-types.ts;
// review-dispatch plan/finalize/sibling/lesson tools return canonical
// pretty-printed JSON strings; consolidate_dispatch returns the
// `{output}` envelope.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Module-level child_process mock (hoisted by vitest)
//
// The registry wires defaultGitExec (execFileSync) into uf-workflow and
// uf-finale tools and execSync (gh CLI) into uf-finale; the sibling
// acquisition path imports execFile. All three named exports must exist
// on the mock or plugin module loading fails. Tests configure responses
// per case via mockGit/mockGh; the beforeEach default fails closed on
// any unconfigured invocation so unexpected shell-outs surface as test
// failures instead of silent undefined-return errors.
// ---------------------------------------------------------------------------

vi.mock("node:child_process", () => ({
  execFileSync: vi.fn(),
  execSync: vi.fn(),
  execFile: vi.fn(),
}))

// ---------------------------------------------------------------------------
// Shared test infrastructure
// ---------------------------------------------------------------------------

/** Runtime view of one registered tool definition. */
interface RegisteredTool {
  readonly execute: (
    args: Record<string, unknown>,
    ctx: ToolContext,
  ) => Promise<{ readonly output: string } | string>
}

/** Tool registry as returned by a plugin `server()` call. */
type ToolRegistry = Record<string, RegisteredTool>

const scratchDirectories: string[] = []
const dispatchSessionDirs: string[] = []

afterEach(async () => {
  _getSessionCorrelationMap().clear()
  await Promise.all([
    ...scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })),
    ...dispatchSessionDirs.splice(0).map((directory) => rm(directory, { recursive: true, force: true })),
  ])
})

beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(execFileSync).mockImplementation((() => {
    throw new Error("unexpected execFileSync call -- configure mockGit first")
  }) as unknown as typeof execFileSync)
  vi.mocked(execSync).mockImplementation((() => {
    throw new Error("unexpected execSync call -- configure mockGh first")
  }) as unknown as typeof execSync)
})

/** Create an empty scratch directory tracked for afterEach cleanup. */
async function makeScratch(prefix = "uf-golden-"): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), prefix))
  scratchDirectories.push(directory)
  return directory
}

/**
 * Scratch project carrying the review-dispatch policy files (written as
 * JSON, parsed by the JSON-backed Bun shim) and a sibling declaration
 * whose only local candidate is missing -- the deterministic
 * unavailable-sibling path for acquire_sibling_evidence.
 */
async function makeScratchProject(): Promise<string> {
  const directory = await makeScratch("uf-golden-project-")
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(
    join(directory, ".uf", "sibling-repos.yaml"),
    JSON.stringify({
      version: 1,
      siblings: [
        {
          name: "gaze",
          url: "https://github.com/unbound-force/gaze.git",
          default_branch: "main",
          role: "sibling",
          fetch: "contract-only",
          contracts: ["README.md"],
          local_paths: ["missing-gaze-checkout"],
        },
      ],
    }),
  )
  return directory
}

/** Write one file inside a scratch root, creating parent directories. */
async function writeProjectFile(root: string, relativePath: string, contents: string): Promise<void> {
  const target = join(root, relativePath)
  await mkdir(dirname(target), { recursive: true })
  await writeFile(target, contents, "utf8")
}

/** Minimal fake PluginInput pointing at a scratch worktree. */
function pluginInput(directory: string): PluginInput {
  return {
    directory,
    worktree: directory,
    client: {} as PluginInput["client"],
  } as PluginInput
}

/** Minimal tool context; sessionID is the consolidate_dispatch correlation-map key. */
function toolContext(sessionID = "golden-session"): ToolContext {
  return {
    sessionID,
    messageID: "message-golden",
    agent: "build",
    directory: "/workspace",
    worktree: "/workspace",
    abort: new AbortController().signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("permission prompts are not used")
    }) as ToolContext["ask"],
  }
}

/** Load a plugin's tool registry through the SDK registration path. */
async function loadRegistry(module: PluginModule, directory: string): Promise<ToolRegistry> {
  const hooks = await withBun(() => module.server(pluginInput(directory)))
  return (hooks.tool ?? {}) as ToolRegistry
}

/** Execute one registered tool and return its raw result. */
async function callTool(
  registry: ToolRegistry,
  name: string,
  args: Record<string, unknown>,
  ctx: ToolContext = toolContext(),
): Promise<{ readonly output: string } | string> {
  const tool = registry[name]
  if (tool === undefined) {
    throw new Error(`tool ${name} is not registered`)
  }
  return tool.execute(args, ctx)
}

/** Execute one `{output}`-envelope tool and parse the ToolResult envelope. */
async function callEnvelopeTool<T = Record<string, unknown>>(
  registry: ToolRegistry,
  name: string,
  args: Record<string, unknown>,
  ctx?: ToolContext,
): Promise<ToolResult<T>> {
  const result = await callTool(registry, name, args, ctx)
  if (typeof result !== "object" || result === null || typeof result.output !== "string") {
    throw new Error(`tool ${name} did not return an {output} envelope`)
  }
  return JSON.parse(result.output) as ToolResult<T>
}

/** Execute one canonical-JSON-string tool and return the raw string. */
async function callStringToolRaw(
  registry: ToolRegistry,
  name: string,
  args: Record<string, unknown>,
  ctx?: ToolContext,
): Promise<string> {
  const result = await callTool(registry, name, args, ctx)
  if (typeof result !== "string") {
    throw new Error(`tool ${name} did not return a string payload`)
  }
  return result
}

/** Execute one canonical-JSON-string tool and parse the payload. */
async function callStringTool<T = Record<string, unknown>>(
  registry: ToolRegistry,
  name: string,
  args: Record<string, unknown>,
  ctx?: ToolContext,
): Promise<T> {
  return JSON.parse(await callStringToolRaw(registry, name, args, ctx)) as T
}

/** Assert the envelope succeeded and return its data payload. */
function expectOk<T>(envelope: ToolResult<T>): T {
  expect(envelope.status).toBe("ok")
  if (envelope.status !== "ok") {
    throw new Error("tool envelope reported an error")
  }
  return envelope.data as T
}

/**
 * Configure execFileSync responses keyed by the exact
 * `git <args joined by spaces>` invocation. Keys absent from the map
 * throw -- tools that probe candidates (resolve_base_ref) treat the
 * throw as a failed candidate and continue.
 */
function mockGit(responses: Readonly<Record<string, string>>): void {
  vi.mocked(execFileSync).mockImplementation(((command: string, args: readonly string[]) => {
    const key = `${command} ${args.join(" ")}`
    const output = responses[key]
    if (output === undefined) {
      throw new Error(`unexpected git invocation: ${key}`)
    }
    return output
  }) as unknown as typeof execFileSync)
}

/**
 * Configure execSync (gh / shell) responses as ordered regex-to-output
 * pairs. gh command strings embed unpredictable temp paths (PR body
 * files), so prefix regexes key the responses.
 */
function mockGh(responses: ReadonlyArray<readonly [RegExp, string]>): void {
  vi.mocked(execSync).mockImplementation(((command: string) => {
    for (const [pattern, output] of responses) {
      if (pattern.test(command)) return output
    }
    throw new Error(`unexpected shell invocation: ${command}`)
  }) as unknown as typeof execSync)
}

// Shared fake git outputs.
const GIT_BRANCH = "git rev-parse --abbrev-ref HEAD"
const GIT_HEAD = "git rev-parse HEAD"
const SHA_A = "1".repeat(40)
/** HEAD reported by resolve_base_ref in the golden fixture. */
const SHA_HEAD = "5".repeat(40)
const MERGE_BASE = "3".repeat(40)
const LESSON_COMMIT = "b".repeat(40)
/** Reviewed head SHA persisted in goldenSessionMetadata; consolidate provenance must match it. */
const SESSION_HEAD_SHA = "b".repeat(40)
/** Reviewed head SHA embedded in the sample review-dispatch payload; finalize provenance must match it. */
const SAMPLE_HEAD_SHA = "2".repeat(40)

// ---------------------------------------------------------------------------
// Dispatch-session fixtures (consolidate_dispatch through the registry)
// ---------------------------------------------------------------------------

/** PersistedRunData fixture for one successful no-findings run. */
function goldenDispatchRun(agent: string): Record<string, unknown> {
  return {
    correlation_id: randomUUID(),
    run_id: randomUUID(),
    agent,
    source: "explicit",
    sequence: 1,
    status: "success",
    started_at: "2026-10-08T10:00:00.000Z",
    finished_at: "2026-10-08T10:01:00.000Z",
    requested_model: "provider/standard",
    provider: "provider",
    model_id: "standard",
    variant: null,
    resolved_parent_model: null,
    resolved_parent_variant: null,
    reported_model: "provider/standard",
    model_mismatch: false,
    usage: { cost_usd: 0.5, tokens: { input: 1000, output: 500 } },
    error: null,
    workflow_verdict: null,
    findings: [],
    proposals: [],
    text: "golden review output",
    read_only: true,
  }
}

/** Plan entry matching goldenDispatchRun's agent identity and model. */
function goldenPlanEntry(agent: string): Record<string, unknown> {
  return {
    agent,
    decision: "include",
    reason_code: "explicit",
    reason: "Explicitly included",
    source: "explicit",
    sequence: 1,
    read_only: true,
    tier: "standard",
    model: "provider/standard",
    variant: null,
    validation_errors: [],
  }
}

/** Issue-shaped session-metadata overrides for triage-issue sessions. */
const TRIAGE_METADATA_OVERRIDES: Record<string, unknown> = {
  input_context: {
    kind: "issue",
    issue_number: 42,
    issue_url: "https://github.com/unbound-force/unbound-force/issues/42",
    content_sha256: "a".repeat(64),
  },
  change_profile: {
    kind: "issue",
    text_bytes: 500,
    comment_count: 3,
    content_sha256: "a".repeat(64),
    matched_rules: [],
    security_sensitive: false,
    user_facing: true,
    categories: ["standard"],
    tier: "standard",
  },
}

/** Session metadata persisted for the first dispatch_agent_run call. */
function goldenSessionMetadata(
  command: string,
  mode: string,
  overrides: Record<string, unknown>,
): Record<string, unknown> {
  return {
    command,
    mode,
    full: false,
    input_context: {
      kind: "pr",
      pr_number: 686,
      base_ref: "main",
      base_sha: "a".repeat(40),
      head_ref: "feature/golden",
      head_sha: SESSION_HEAD_SHA,
    },
    change_profile: {
      kind: "diff",
      lines: 100,
      files: 5,
      components: 2,
      security_sensitive: false,
      user_facing: true,
      categories: ["standard"],
      tier: "standard",
    },
    plan: [goldenPlanEntry("divisor-guard")],
    coverage: { preflight_verdict: "PASS", checks_total: 5, checks_passed: 5 },
    ...overrides,
  }
}

/**
 * Persist one dispatch session (metadata + run files) and bind it to the
 * golden-session correlation map entry, mirroring what the first
 * dispatch_agent_run call writes. Returns the correlation id.
 */
async function setupGoldenDispatchSession(
  command: string,
  mode: string,
  overrides: Record<string, unknown>,
): Promise<string> {
  const correlationId = randomUUID()
  const directory = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
  await mkdir(directory, { recursive: true })
  dispatchSessionDirs.push(directory)
  _getSessionCorrelationMap().set("golden-session", correlationId)
  await writeFile(
    join(directory, "session-metadata.json"),
    JSON.stringify(goldenSessionMetadata(command, mode, overrides), null, 2),
    "utf8",
  )
  await writeFile(
    join(directory, "run-divisor-guard.json"),
    JSON.stringify(goldenDispatchRun("divisor-guard"), null, 2),
    "utf8",
  )
  return correlationId
}

// ---------------------------------------------------------------------------
// uf-workflow registry -- 21 documented sites
// ---------------------------------------------------------------------------

describe("golden output shapes -- uf-workflow registry (21 documented sites)", () => {
  it("get_progress [uf.unleash.md:55, uf.finale.md:42, uf.review-council.md:51, uf.address-feedback.md:71] returns completed_steps and current_step", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const documentedCommands = [
      "uf.unleash",
      "uf.finale",
      "uf.review-council",
      "uf.address-feedback",
    ] as const
    for (const command of documentedCommands) {
      const step = `${command}-golden-step`
      const recorded = expectOk(
        await callEnvelopeTool(registry, "checkpoint", { command, step, key: "golden" }),
      )
      expect(recorded.completed_steps).toContain(step)

      const state = expectOk(
        await callEnvelopeTool(registry, "get_progress", { command, key: "golden" }),
      )
      expect(state.current_step).toBe(step)
      expect(state.completed_steps).toContain(step)
    }
  })

  it("uf_unleash_branch_gate [uf.unleash.md:81] returns the ok payload with workflow_tier, feature_dir, and change_name", async () => {
    const directory = await makeScratch()
    await writeProjectFile(directory, "openspec/changes/golden/tasks.md", "- [x] 1.1 Done\n")
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    mockGit({ [GIT_BRANCH]: "opsx/golden" })
    const data = expectOk(
      await callEnvelopeTool<{
        branch: string
        workflow_tier: string
        feature_dir: string
        change_name: string
      }>(registry, "uf_unleash_branch_gate", {}),
    )
    expect(data).toEqual({
      branch: "opsx/golden",
      workflow_tier: "openspec",
      feature_dir: "openspec/changes/golden/",
      change_name: "golden",
    })
  })

  it("uf_unleash_branch_gate [uf.unleash.md:81] carries the ON_MAIN, NO_TASKS, NO_SPEC, and UNRECOGNIZED_BRANCH failure codes", async () => {
    const directory = await makeScratch()
    await writeProjectFile(directory, "openspec/changes/golden/tasks.md", "- [x] 1.1 Done\n")
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const rejections: ReadonlyArray<readonly [string, string]> = [
      ["main", "ON_MAIN"],
      ["opsx/missing", "NO_TASKS"],
      ["018-golden", "NO_SPEC"],
      ["feature/golden", "UNRECOGNIZED_BRANCH"],
    ]
    for (const [branch, code] of rejections) {
      mockGit({ [GIT_BRANCH]: branch })
      const envelope = await callEnvelopeTool(registry, "uf_unleash_branch_gate", {})
      expect(envelope.status).toBe("error")
      expect(envelope.status === "error" ? envelope.code : undefined).toBe(code)
    }
  })

  it("uf_unleash_detect_resume [uf.unleash.md:104] returns step_statuses, resume_from, all_complete, unchecked_count, and checkpoint_claims", async () => {
    const directory = await makeScratch()
    await writeProjectFile(directory, "openspec/changes/golden/tasks.md", "## Phase 1\n\n- [x] 1.1 Done\n")
    const registry = await loadRegistry(UfWorkflowPlugin, directory)
    mockGit({ [GIT_BRANCH]: "opsx/golden", [GIT_HEAD]: SHA_A })

    const data = expectOk(
      await callEnvelopeTool<{
        step_statuses: Record<string, boolean>
        resume_from: string | null
        all_complete: boolean
        unchecked_count: number | null
        checkpoint_claims: { spec_review: boolean; code_review: boolean }
      }>(registry, "uf_unleash_detect_resume", {
        feature_dir: "openspec/changes/golden/",
        workflow_tier: "openspec",
      }),
    )
    expect(data.step_statuses).toEqual({
      clarify: true,
      plan: true,
      tasks: true,
      spec_review: false,
      implement: true,
      code_review: false,
    })
    expect(data.resume_from).toBe("spec_review")
    expect(data.all_complete).toBe(false)
    expect(data.unchecked_count).toBe(0)
    expect(data.checkpoint_claims).toEqual({ spec_review: false, code_review: false })
  })

  it("scan_clarifications [uf.unleash.md:169] returns questions with question, context, and line plus count", async () => {
    const directory = await makeScratch()
    await writeProjectFile(
      directory,
      "specs/018-golden/spec.md",
      [
        "Intro line one.",
        "Intro line two.",
        "The store is unclear. [NEEDS CLARIFICATION: which database?]",
        "Outro line one.",
        "Outro line two.",
      ].join("\n"),
    )
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const data = expectOk(
      await callEnvelopeTool<{
        questions: Array<{ question: string; context: string; line: number }>
        count: number
      }>(registry, "scan_clarifications", { feature_dir: "specs/018-golden/" }),
    )
    expect(data.count).toBe(1)
    expect(data.questions).toHaveLength(1)
    expect(data.questions[0]!.question).toBe("which database?")
    expect(data.questions[0]!.line).toBe(3)
    expect(data.questions[0]!.context).toContain("[NEEDS CLARIFICATION: which database?]")
    expect(data.questions[0]!.context).toContain("Intro line two.")
    expect(data.questions[0]!.context).toContain("Outro line one.")
  })

  it.each([
    ["clarify", { variant: "clarify", questions: [{ question: "Which database?", context: "The store is unclear." }] }],
    ["spec-review", { variant: "spec-review", tier: "speckit", findings: ["FR-001 lacks an acceptance scenario"] }],
    ["implement-worker-failed", { variant: "implement-worker-failed", phase: 1, failed_task: "2.1 implement the schema", error: "worker crashed" }],
    ["implement-merge-conflict", { variant: "implement-merge-conflict", phase: 2, conflicting_files: ["internal/store/store.go"] }],
    ["implement-phase-checkpoint", { variant: "implement-phase-checkpoint", phase: 1, failed_command: "make test", output: "FAIL internal/store/store_test.go" }],
    ["code-review-preflight", { variant: "code-review-preflight", findings: ["lint failed on the feature branch"] }],
    ["code-review-no-success", { variant: "code-review-no-success", reasons: ["divisor-guard failed to start"] }],
    ["code-review-iterations-exhausted", { variant: "code-review-iterations-exhausted", persistent_findings: ["HIGH finding persists after three iterations"] }],
  ] as ReadonlyArray<readonly [string, Record<string, unknown>]>)(
    "format_unleash_exit [%s variant] returns the exact paused-at template text",
    async (_variant, args) => {
      const directory = await makeScratch()
      const registry = await loadRegistry(UfWorkflowPlugin, directory)

      const data = expectOk(
        await callEnvelopeTool<{ text: string }>(registry, "format_unleash_exit", args),
      )
      // Tool==lib equivalence: the registered tool strips `variant` and
      // renders through the same pure lib function the renderer goldens
      // pin per variant (unleash-formatters.test.ts), so the full
      // variant-specific documented output — not just the shared skeleton
      // below — cannot drift between tool and renderer.
      const { variant, ...fields } = args
      expect(data).toStrictEqual({
        text: formatUnleashExit(variant as UnleashExitVariant, fields as UnleashExitData),
      })
      expect(typeof data.text).toBe("string")
      expect(data.text).toContain("## /uf.unleash paused at:")
      expect(data.text).toContain("**Reason**:")
      expect(data.text).toContain("### What to do next")
      expect(data.text).toContain("### Then resume")
    },
  )

  it("parse_tasks [uf.unleash.md:458] returns phases, total_tasks, done_count, and all_done", async () => {
    const directory = await makeScratch()
    await writeProjectFile(
      directory,
      "openspec/changes/golden/tasks.md",
      [
        "## Phase 1 - Setup",
        "",
        "- [ ] 1.1 Write the schema",
        "- [ ] 1.2 [P] Add migrations",
        "- [x] 1.3 Seed data",
        "",
        "## Phase 2 - Verify",
        "",
        "- [ ] 2.1 Run the suite",
      ].join("\n"),
    )
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const data = expectOk(
      await callEnvelopeTool<{
        phases: Array<{
          phase: string
          tasks: Array<{ id: string; description: string; parallel: boolean; done: boolean }>
        }>
        total_tasks: number
        done_count: number
        all_done: boolean
      }>(registry, "parse_tasks", { tasks_path: "openspec/changes/golden/tasks.md" }),
    )
    expect(data.total_tasks).toBe(4)
    expect(data.done_count).toBe(1)
    expect(data.all_done).toBe(false)
    expect(data.phases).toEqual([
      {
        phase: "Phase 1 - Setup",
        tasks: [
          { id: "1.1", description: "Write the schema", parallel: false, done: false },
          { id: "1.2", description: "Add migrations", parallel: true, done: false },
          { id: "1.3", description: "Seed data", parallel: false, done: true },
        ],
      },
      {
        phase: "Phase 2 - Verify",
        tasks: [{ id: "2.1", description: "Run the suite", parallel: false, done: false }],
      },
    ])
  })

  it("mark_task_complete [uf.unleash.md:515] flips the first checkbox and reports already_complete on the no-op rerun", async () => {
    const directory = await makeScratch()
    const tasksPath = "openspec/changes/golden/tasks.md"
    await writeProjectFile(directory, tasksPath, "## Phase 1\n\n- [ ] 1.1 Write the schema\n")
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const first = expectOk(
      await callEnvelopeTool<{
        tasks_path: string
        task_id: string
        already_complete: boolean
      }>(registry, "mark_task_complete", { tasks_path: tasksPath, task_id: "1.1" }),
    )
    expect(first).toEqual({ tasks_path: tasksPath, task_id: "1.1", already_complete: false })
    expect(await readFile(join(directory, tasksPath), "utf8")).toContain("- [x] 1.1 Write the schema")

    const second = expectOk(
      await callEnvelopeTool<{ already_complete: boolean }>(registry, "mark_task_complete", {
        tasks_path: tasksPath,
        task_id: "1.1",
      }),
    )
    expect(second.already_complete).toBe(true)
  })

  it("scan_conflict_markers [uf.unleash.md:598] returns conflicts with line and marker, clean, and warnings", async () => {
    const directory = await makeScratch()
    await writeProjectFile(
      directory,
      "merged.go",
      "const a = 1\n<<<<<<< HEAD\nconst b = 2\n=======\nconst b = 3\n>>>>>>> main\n",
    )
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const data = expectOk(
      await callEnvelopeTool<{
        conflicts: Array<{ path: string; markers: Array<{ line: number; marker: string }> }>
        clean: boolean
        warnings: string[]
      }>(registry, "scan_conflict_markers", { paths: ["merged.go"] }),
    )
    expect(data.clean).toBe(false)
    expect(data.warnings).toEqual([])
    expect(data.conflicts).toEqual([
      {
        path: "merged.go",
        markers: [
          { line: 2, marker: "<<<<<<<" },
          { line: 4, marker: "=======" },
          { line: 6, marker: ">>>>>>>" },
        ],
      },
    ])
  })

  it("format_unleash_demo [uf.unleash.md:1010] returns the Demo block text with the two Next Steps options", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfWorkflowPlugin, directory)
    mockGit({ [GIT_BRANCH]: "opsx/golden" })

    const data = expectOk(
      await callEnvelopeTool<{ text: string }>(registry, "format_unleash_demo", {
        tier: "openspec",
        whatWasBuilt: "Golden output-shape contract tests.",
        howToVerify: "npx vitest run test/golden-output-shape.test.ts",
        testResults: "43 documented sites covered.",
        changeName: "golden-change",
      }),
    )
    for (const heading of ["## What Was Built", "## How to Verify", "## Test Results", "## Next Steps"]) {
      expect(data.text).toContain(heading)
    }
    expect(data.text).toContain("Run `/uf.finale` to create PR and watch CI")
    expect(data.text).toContain("openspec/changes/golden-change/")
  })

  it("resolve_base_ref [uf.review-council.md:126] returns the ref, resolved SHA, source label, HEAD sha, and divergence flag", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfWorkflowPlugin, directory)
    // upstream/main is absent from the map, so its probe throws and the
    // tool falls through to the next candidate, exactly like a missing ref.
    mockGit({
      "git rev-parse --verify origin/main": SHA_A,
      "git rev-parse --verify HEAD": SHA_HEAD,
      [`git merge-base --is-ancestor ${SHA_A} HEAD`]: "",
    })

    const data = expectOk(
      await callEnvelopeTool<{ ref: string; sha: string; source: string; head_sha: string; diverged: boolean }>(
        registry,
        "resolve_base_ref",
        {},
      ),
    )
    expect(data).toEqual({ ref: "origin/main", sha: SHA_A, source: "origin/main", head_sha: SHA_HEAD, diverged: false })
  })

  it("normalize_model_name [uf.address-feedback.md:801] returns normalized and the unknown-model fallback", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfWorkflowPlugin, directory)

    const normalized = expectOk(
      await callEnvelopeTool<{ normalized: string }>(registry, "normalize_model_name", {
        rawModelId: "google-vertex-anthropic/claude-sonnet-4-20250514@default",
      }),
    )
    expect(normalized.normalized).toBe("claude-sonnet-4")

    const fallback = expectOk(
      await callEnvelopeTool<{ normalized: string; warning: string }>(
        registry,
        "normalize_model_name",
        { rawModelId: "???" },
      ),
    )
    expect(fallback.normalized).toBe("unknown-model")
    expect(fallback.warning).toContain("???")
  })
})

// ---------------------------------------------------------------------------
// uf-finale registry -- 8 documented sites
// ---------------------------------------------------------------------------

describe("golden output shapes -- uf-finale registry (8 documented sites)", () => {
  it("uf_finale_startup [uf.finale.md:51] returns branch, branchType, isMain, hasChanges, secretCandidates, and unpushedCommits", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGit({
      [GIT_BRANCH]: "opsx/golden",
      // defaultGitExec trims command output, so an " M " porcelain line
      // loses its leading space and shifts the XY-prefixed filename; the
      // untracked "?? " form survives trimming and parses to ".env".
      "git status --porcelain": "?? .env\n",
      "git log @{u}..HEAD --oneline": "abc1234 Add the golden tests\n",
    })

    const data = expectOk(
      await callEnvelopeTool<{
        branch: string
        isMain: boolean
        branchType: string
        changeName: string
        hasChanges: boolean
        secretCandidates: string[]
        unpushedCommits: Array<{ sha: string; subject: string }>
      }>(registry, "uf_finale_startup", {}),
    )
    expect(data).toMatchObject({
      branch: "opsx/golden",
      isMain: false,
      branchType: "opsx",
      changeName: "golden",
      hasChanges: true,
      secretCandidates: [".env"],
    })
    expect(data.unpushedCommits).toEqual([{ sha: "abc1234", subject: "Add the golden tests" }])
  })

  it("uf_finale_stage [uf.finale.md:71] returns fileCount and files", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGit({
      "git add .": "",
      "git diff --cached --name-only": "a.ts\nb.ts\n",
    })

    const data = expectOk(
      await callEnvelopeTool<{ fileCount: number; files: string[] }>(
        registry,
        "uf_finale_stage",
        {},
      ),
    )
    expect(data).toEqual({ fileCount: 2, files: ["a.ts", "b.ts"] })
  })

  it("uf_finale_commit_context [uf.finale.md:82] returns diffStat, recentLog, and normalizedModel", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGit({
      "git diff --cached --stat": " 2 files changed, 10 insertions(+)",
      "git log -5 --oneline": "abc1234 Add the golden tests\n",
    })

    const data = expectOk(
      await callEnvelopeTool<{ diffStat: string; recentLog: string; normalizedModel: string }>(
        registry,
        "uf_finale_commit_context",
        { rawModelId: "provider/golden-model-4-20260101@default" },
      ),
    )
    // defaultGitExec trims command output before it reaches the payload.
    expect(data.diffStat).toBe("2 files changed, 10 insertions(+)")
    expect(data.recentLog).toBe("abc1234 Add the golden tests")
    expect(data.normalizedModel).toBe("golden-model-4")
  })

  it("uf_finale_push_state [uf.finale.md:131] returns upstreamExists, diverged, aheadBy, and behindBy", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGit({
      "git rev-parse --abbrev-ref @{u}": "origin/main",
      "git rev-list @{u}..HEAD --count": "2",
      "git rev-list HEAD..@{u} --count": "1",
    })

    const data = expectOk(
      await callEnvelopeTool<{
        upstreamExists: boolean
        diverged: boolean
        aheadBy: number
        behindBy: number
      }>(registry, "uf_finale_push_state", {}),
    )
    expect(data).toEqual({ upstreamExists: true, diverged: true, aheadBy: 2, behindBy: 1 })
  })

  it("uf_finale_pr_context [uf.finale.md:154] returns existingPr, isFork, forkParent, templateSections, commitLog, diffStat, originatingIssue, and unleashSummary", async () => {
    const directory = await makeScratch()
    await writeProjectFile(directory, ".github/PULL_REQUEST_TEMPLATE.md", "## Summary\n\n## Changes\n")
    await writeProjectFile(directory, ".openspec.yaml", "change: golden\noriginating_issue: 732\n")
    await writeProjectFile(
      directory,
      ".uf/artifacts/unleash-summary/opsx-golden.json",
      JSON.stringify({ change_name: "golden", tier: "openspec" }),
    )
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGit({
      [GIT_BRANCH]: "opsx/golden",
      "git merge-base HEAD origin/main": MERGE_BASE,
      [`git log --oneline ${MERGE_BASE}..HEAD`]: "abc1234 Add the golden tests",
      [`git diff --stat ${MERGE_BASE}..HEAD`]: " 1 file changed",
    })
    mockGh([
      [
        /^gh pr view --json number,url,title$/,
        JSON.stringify({
          number: 42,
          url: "https://github.com/unbound-force/unbound-force/pull/42",
          title: "Golden tests",
        }),
      ],
      [/^gh repo view --json isFork --jq '\.isFork'$/, "false"],
    ])

    const data = expectOk(
      await callEnvelopeTool<{
        existingPr: { number: number; url: string; title: string } | null
        isFork: boolean
        forkParent: { owner: string; repo: string } | null
        templateSections: string[]
        commitLog: Array<{ sha: string; subject: string }>
        diffStat: string
        originatingIssue: number | null
        unleashSummary: Record<string, unknown> | null
      }>(registry, "uf_finale_pr_context", {}),
    )
    expect(data.existingPr).toEqual({
      number: 42,
      url: "https://github.com/unbound-force/unbound-force/pull/42",
      title: "Golden tests",
    })
    expect(data.isFork).toBe(false)
    expect(data.forkParent).toBeNull()
    expect(data.templateSections).toEqual(["Summary", "Changes"])
    expect(data.commitLog).toEqual([{ sha: "abc1234", subject: "Add the golden tests" }])
    expect(data.diffStat).toBe("1 file changed")
    expect(data.originatingIssue).toBe(732)
    expect(data.unleashSummary).toEqual({ change_name: "golden", tier: "openspec" })
  })

  it("uf_finale_create_pr [uf.finale.md:204] returns number and url", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGh([
      [
        /^gh pr create --title "Add golden tests" --body-file "/,
        "https://github.com/unbound-force/unbound-force/pull/43\n",
      ],
    ])

    const data = expectOk(
      await callEnvelopeTool<{ number: number; url: string }>(registry, "uf_finale_create_pr", {
        title: "Add golden tests",
        body: "## Summary\n\nGolden output-shape contract tests.",
      }),
    )
    expect(data).toEqual({
      number: 43,
      url: "https://github.com/unbound-force/unbound-force/pull/43",
    })
  })

  it("uf_finale_ci_poll [uf.finale.md:212] returns status, failedChecks, mergeable, and workflowsConfigured", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGh([
      [
        /^gh pr view 43 --json mergeable,mergeStateStatus$/,
        JSON.stringify({ mergeable: "MERGEABLE", mergeStateStatus: "CLEAN" }),
      ],
      [
        /^gh pr checks 43 --json name,state,conclusion,detailsUrl$/,
        JSON.stringify([
          { name: "ci", state: "COMPLETED", conclusion: "SUCCESS", detailsUrl: "https://example.com/ci" },
        ]),
      ],
    ])

    const data = expectOk(
      await callEnvelopeTool<{
        status: string
        failedChecks: unknown[]
        mergeable: boolean
        workflowsConfigured: boolean
      }>(registry, "uf_finale_ci_poll", { prNumber: 43 }),
    )
    expect(data).toEqual({
      status: "pass",
      failedChecks: [],
      mergeable: true,
      workflowsConfigured: true,
    })
  })

  it("uf_finale_return_to_main [uf.finale.md:291] returns success", async () => {
    const directory = await makeScratch()
    const registry = await loadRegistry(UfFinalePlugin, directory)
    mockGit({
      "git checkout main": "",
      "git pull": "Already up to date.",
    })

    const data = expectOk(
      await callEnvelopeTool<{ success: boolean }>(registry, "uf_finale_return_to_main", {}),
    )
    expect(data).toEqual({ success: true })
  })
})

// ---------------------------------------------------------------------------
// review-dispatch registry -- 14 documented sites
// ---------------------------------------------------------------------------

/** plan_review_dispatch inputs for the seven documented call sites. */
const PLAN_SITES: ReadonlyArray<{ readonly site: string; readonly args: Record<string, unknown> }> = [
  {
    site: "uf.unleash.md:353",
    args: {
      mode: "specs",
      discovered_agents: ["divisor-adversary", "divisor-architect", "divisor-guard"],
      changed_files: [{ path: "specs/018-golden/spec.md", additions: 20, deletions: 0 }],
    },
  },
  {
    site: "uf.unleash.md:799",
    args: {
      mode: "code",
      discovered_agents: ["divisor-adversary", "divisor-architect", "divisor-guard"],
      changed_files: [{ path: "internal/golden/golden.go", additions: 60, deletions: 2 }],
    },
  },
  {
    site: "uf.review-council.md:231",
    args: {
      mode: "code",
      discovered_agents: ["divisor-adversary", "divisor-architect", "divisor-guard"],
      changed_files: [{ path: "src/feature.ts", additions: 60, deletions: 2 }],
    },
  },
  {
    site: "uf.review-pr.md:353",
    args: {
      mode: "code",
      discovered_agents: ["divisor-adversary", "divisor-architect", "divisor-guard"],
      changed_files: [{ path: "cmd/golden/golden.go", additions: 60, deletions: 2 }],
    },
  },
  {
    site: "uf.triage-issue.md:190",
    args: {
      mode: "triage",
      discovered_agents: ["divisor-adversary", "divisor-architect", "divisor-guard"],
      issue: { title: "Coverage regression in the golden suite", body: null, comments: [] },
    },
  },
  {
    site: "uf.address-feedback.md:339",
    args: {
      mode: "feedback",
      discovered_agents: ["divisor-adversary", "divisor-architect", "divisor-guard"],
      changed_files: [{ path: "src/feature.ts", additions: 5, deletions: 0 }],
    },
  },
  {
    site: "speckit.testreview.md:154",
    args: {
      // test review is restricted to divisor-testing by the review matrix.
      mode: "test",
      discovered_agents: ["divisor-testing"],
      changed_files: [{ path: "internal/golden/golden_test.go", additions: 30, deletions: 0 }],
    },
  },
]

/** consolidate_dispatch sessions for the three documented call sites. */
const CONSOLIDATE_SITES: ReadonlyArray<{
  readonly site: string
  readonly command: "review-council" | "triage-issue" | "address-feedback"
  readonly mode: "code" | "triage" | "feedback"
  readonly metadataOverrides: Record<string, unknown>
}> = [
  { site: "uf.review-council.md:361", command: "review-council", mode: "code", metadataOverrides: {} },
  { site: "uf.triage-issue.md:780", command: "triage-issue", mode: "triage", metadataOverrides: TRIAGE_METADATA_OVERRIDES },
  { site: "uf.address-feedback.md:541", command: "address-feedback", mode: "feedback", metadataOverrides: {} },
]

describe("golden output shapes -- review-dispatch registry (15 documented sites)", () => {
  it.each(PLAN_SITES)(
    "plan_review_dispatch [$site] returns a ready plan with workflow_result null and empty errors",
    async ({ args }) => {
      const directory = await makeScratchProject()
      const registry = await loadRegistry(ReviewDispatchPlugin, directory)

      const raw = await callStringToolRaw(registry, "plan_review_dispatch", args)
      expect(raw.endsWith("\n")).toBe(true)
      const plan = JSON.parse(raw) as {
        plan_version: number
        status: string
        workflow_result: unknown
        mode: string
        matrix_mode: string
        full: boolean
        augmentation_requested: boolean
        change_profile: Record<string, unknown>
        limits: Record<string, unknown>
        entries: unknown[]
        omissions: unknown[]
        limit_state: unknown
        advisories: unknown[]
        errors: string[]
      }
      expect(plan.plan_version).toBe(1)
      expect(plan.status).toBe("ready")
      expect(plan.workflow_result).toBeNull()
      expect(plan.errors).toEqual([])
      expect(Array.isArray(plan.entries)).toBe(true)
      expect(plan.entries.length).toBeGreaterThan(0)
      expect(plan.change_profile).toBeTypeOf("object")
      expect(plan.limits).toBeTypeOf("object")
      expect(Array.isArray(plan.omissions)).toBe(true)
      expect(Array.isArray(plan.advisories)).toBe(true)
      expect(plan.mode).toBe(args.mode)
    },
  )

  it.each(CONSOLIDATE_SITES)(
    "consolidate_dispatch [$site] finalizes with provenance and returns the verdict artifact data",
    async ({ command, mode, metadataOverrides }) => {
      const directory = await makeScratchProject()
      const registry = await loadRegistry(ReviewDispatchPlugin, directory)
      const correlationId = await setupGoldenDispatchSession(command, mode, metadataOverrides)

      const result = await callEnvelopeTool<{
        status: string
        operation_verdict: string
        artifact_path: string | null
        review_verdict: { schema_version: string; council_decision: string } | null
        calculated_assessment: {
          authoritative: boolean
          payload: {
            verdict: string
            verdict_reason: string
            run_counts: Record<string, number>
            findings: unknown[]
            advisories: unknown[]
          }
        } | null
        correlation_id: string
      }>(
        registry,
        "consolidate_dispatch",
        {
          provenance: {
            branch: "opsx/golden",
            // finalization fails closed unless the provenance commit equals
            // the immutable reviewed head SHA persisted in the session.
            commit: SESSION_HEAD_SHA,
            workflow_id: `uf.${command}`,
          },
        },
        toolContext("golden-session"),
      )
      expect(result.status).toBe("success")
      expect(result.operation_verdict).toBe("APPROVE")
      expect(result.artifact_path).toBe(join(directory, ".uf/artifacts/dispatch", `${correlationId}.json`))
      await expect(stat(result.artifact_path!)).resolves.toBeTruthy()
      expect(result.review_verdict?.schema_version).toBe("2.0.0")
      expect(result.review_verdict?.council_decision).toBe("APPROVED")
      expect(result.calculated_assessment?.authoritative).toBe(true)
      // With provenance, the tool spreads FinalizeReviewDispatchResult; the
      // consolidated verdict fields ride inside calculated_assessment.payload.
      expect(result.calculated_assessment?.payload.verdict).toBe("APPROVE")
      expect(result.calculated_assessment?.payload.verdict_reason).toBe("No findings requiring changes")
      expect(result.calculated_assessment?.payload.run_counts).toMatchObject({ total: 1, success: 1 })
      expect(result.calculated_assessment?.payload.findings).toEqual([])
      expect(result.calculated_assessment?.payload.advisories).toEqual([])
      expect(result.correlation_id).toBe(correlationId)
    },
  )

  it.each([
    { site: "uf.review-pr.md:442", command: "review-council", mode: "code", workflowResult: { kind: "council", value: "APPROVE" } },
    { site: "speckit.testreview.md:278", command: "speckit-testreview", mode: "test", workflowResult: { kind: "test-review", value: "APPROVE" } },
  ] as const)(
    "finalize_review_dispatch [$site] persists the artifact and returns canonical review-verdict v2 data",
    async ({ command, mode, workflowResult }) => {
      const directory = await makeScratchProject()
      const registry = await loadRegistry(ReviewDispatchPlugin, directory)

      const payload = structuredClone(samplePayload) as unknown as Record<string, unknown>
      const run = (payload.runs as Record<string, unknown>[])[0]!
      run.workflow_verdict = "APPROVE"
      run.findings = []
      payload.command = command
      payload.mode = mode
      payload.findings = []
      payload.advisories = []
      payload.verdict = "APPROVE"
      payload.verdict_reason = "Successful assessment approved."
      payload.workflow_result = workflowResult

      const raw = await callStringToolRaw(
        registry,
        "finalize_review_dispatch",
        {
          payload,
          provenance: {
            branch: "opsx/golden",
            // must equal the sample payload's reviewed head SHA; the
            // finalizer rejects mismatched provenance commits.
            commit: SAMPLE_HEAD_SHA,
            workflow_id: "uf.golden",
          },
        },
      )
      expect(raw.endsWith("\n")).toBe(true)
      const result = JSON.parse(raw) as {
        status: string
        operation_verdict: string
        artifact_path: string | null
        review_verdict: {
          schema_version: string
          council_decision: string
          persona_verdicts: unknown[]
          unresolved_findings: unknown[]
        } | null
        calculated_assessment: { authoritative: boolean } | null
        errors: string[]
      }
      expect(result.status).toBe("success")
      expect(result.operation_verdict).toBe("APPROVE")
      expect(result.review_verdict?.schema_version).toBe("2.0.0")
      expect(result.review_verdict?.council_decision).toBe("APPROVED")
      expect(Array.isArray(result.review_verdict?.persona_verdicts)).toBe(true)
      expect(Array.isArray(result.review_verdict?.unresolved_findings)).toBe(true)
      expect(result.calculated_assessment?.authoritative).toBe(true)
      expect(typeof result.artifact_path).toBe("string")
      await expect(stat(result.artifact_path!)).resolves.toBeTruthy()
      expect(result.errors).toEqual([])
    },
  )

  it("acquire_sibling_evidence [speckit.testreview.md:134] returns structured provenance with commit, path, SHA256, source mode, and rejection fields", async () => {
    const directory = await makeScratchProject()
    const registry = await loadRegistry(ReviewDispatchPlugin, directory)

    const result = await callStringTool<{
      version: number
      status: string
      prompt: string
      siblings: Array<{
        sibling: string
        status: string
        commit: string | null
        source_mode: string | null
        source_candidate: string | null
        evidence: unknown[]
        rejections: Array<{ scope: string; subject: string; reason_code: string; reason: string }>
        unavailable_reason: string | null
      }>
      errors: string[]
    }>(registry, "acquire_sibling_evidence", {})
    expect(result.version).toBe(1)
    expect(result.status).toBe("ok")
    expect(result.errors).toEqual([])
    expect(result.prompt).toBe("")
    expect(result.siblings).toHaveLength(1)
    expect(result.siblings[0]).toMatchObject({
      sibling: "gaze",
      status: "unavailable",
      commit: null,
      source_mode: null,
      source_candidate: null,
      evidence: [],
      unavailable_reason: "no-verified-candidate",
    })
    expect(result.siblings[0]!.rejections).toEqual([
      {
        scope: "candidate",
        subject: "local_paths[0]",
        reason_code: "candidate-missing",
        reason: "candidate path is unavailable",
      },
    ])
  })

  it("prepare_lesson_learning [speckit.testreview.md:268] returns the ready Dewey learning payload with generated tag and reference category", async () => {
    const directory = await makeScratchProject()
    const registry = await loadRegistry(ReviewDispatchPlugin, directory)

    const content = "Contract coverage checks observable behavior."
    const fileSha256 = createHash("sha256").update(content, "utf8").digest("hex")
    const evidenceMetadata = {
      sibling: "gaze",
      commit: LESSON_COMMIT,
      path: "README.md",
      sha256: fileSha256,
      source_mode: "local",
    }
    const evidence = {
      version: 1,
      status: "ok",
      prompt: [
        UNTRUSTED_SIBLING_EVIDENCE_BEGIN,
        JSON.stringify(evidenceMetadata),
        content,
        UNTRUSTED_SIBLING_EVIDENCE_END,
      ].join("\n"),
      siblings: [
        {
          sibling: "gaze",
          status: "available",
          commit: LESSON_COMMIT,
          source_mode: "local",
          source_candidate: "/tmp/gaze",
          evidence: [evidenceMetadata],
          rejections: [],
          unavailable_reason: null,
        },
      ],
      errors: [],
    }
    const proposal = {
      schema_version: "1.0.0",
      sibling: "gaze",
      commit: LESSON_COMMIT,
      sources: [{ path: "README.md", sha256: fileSha256, excerpt: "observable behavior" }],
      information: "Contract coverage checks observable behavior.",
    }
    const childOutput = `Review completed.\n${LESSON_PROPOSAL_BEGIN}\n${JSON.stringify(proposal)}\n${LESSON_PROPOSAL_END}\nDone.`

    const result = await callStringTool<{
      status: string
      reason_code: string | null
      learning: { information: string; tag: string; category: string }
      dedupe_sha256: string
      provenance: {
        sibling: string
        commit: string
        sources: Array<{ path: string; sha256: string; excerpt: string }>
      }
    }>(registry, "prepare_lesson_learning", {
      child_output: childOutput,
      sibling_evidence: evidence,
      known_dedupe_sha256: [],
    })
    expect(result.status).toBe("ready")
    expect(result.reason_code).toBeNull()
    expect(result.learning.tag).toBe("sibling-gaze")
    expect(result.learning.category).toBe("reference")
    expect(result.learning.information).toContain("observable behavior")
    expect(result.dedupe_sha256).toMatch(/^[0-9a-f]{64}$/)
    expect(result.provenance).toMatchObject({ sibling: "gaze", commit: LESSON_COMMIT })
    expect(result.provenance.sources[0]).toMatchObject({
      path: "README.md",
      sha256: fileSha256,
      excerpt: "observable behavior",
    })
  })

  it("capture_review_diff [uf.unleash.md:539] returns mode, base, bounded diff path, per-file counts, and writes a readable patch", async () => {
    const directory = await makeScratchProject()
    const registry = await loadRegistry(ReviewDispatchPlugin, directory)
    await writeProjectFile(directory, "openspec/changes/golden/proposal.md", "alpha\nbeta\n")
    // upstream/main is absent from the map, so its probe throws and the
    // candidate loop falls through to origin/main, exactly like a missing ref.
    mockGit({
      "git rev-parse --verify origin/main": SHA_A,
      "git rev-parse --verify HEAD": SHA_HEAD,
      [`git merge-base --is-ancestor ${SHA_A} HEAD`]: "",
      [`git diff --no-renames ${SHA_A}`]:
        "diff --git a/src/a.ts b/src/a.ts\n--- a/src/a.ts\n+++ b/src/a.ts\n@@ -1 +1 @@\n-old\n+new\n",
      [`git diff --no-renames --numstat ${SHA_A}`]: "1\t1\tsrc/a.ts\n",
      "git ls-files --others --exclude-standard -- openspec/changes/golden": "",
      "git ls-files --others --ignored --exclude-standard -- openspec/changes/golden":
        "openspec/changes/golden/proposal.md\n",
    })

    const result = await callEnvelopeTool<{
      mode: string
      base: { ref: string; sha: string; source: string; head_sha: string; diverged: boolean } | null
      diff_path: string
      diff_bytes: number
      files: Array<{ path: string; additions: number; deletions: number }>
      total_files: number
      total_additions: number
      total_deletions: number
      excluded: Array<{ path: string; reason: string }>
    }>(registry, "capture_review_diff", { feature_dir: "openspec/changes/golden" })

    expect(result.mode).toBe("base")
    expect(result.base).toEqual({
      ref: "origin/main",
      sha: SHA_A,
      source: "origin/main",
      head_sha: SHA_HEAD,
      diverged: false,
    })
    expect(result.files).toEqual([
      { path: "src/a.ts", additions: 1, deletions: 1 },
      { path: "openspec/changes/golden/proposal.md", additions: 2, deletions: 0 },
    ])
    expect(result.total_files).toBe(2)
    expect(result.total_additions).toBe(3)
    expect(result.total_deletions).toBe(1)
    expect(result.excluded).toEqual([])
    // Write -> read round-trip: the saved patch exists and is readable.
    await expect(stat(result.diff_path)).resolves.toBeTruthy()
    const written = await readFile(result.diff_path, "utf8")
    expect(written).toContain("+++ b/openspec/changes/golden/proposal.md")
    expect(written).toContain("+alpha")
    expect(result.diff_bytes).toBe(Buffer.byteLength(written, "utf8"))
  })
})

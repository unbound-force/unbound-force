import { Buffer } from "node:buffer"
import { existsSync, lstatSync, mkdirSync, readFileSync, statSync, writeFileSync } from "node:fs"
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join, resolve } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { DEFAULT_TIMEOUT_MILLISECONDS, MAX_TIMEOUT_MILLISECONDS } from "../lib/agent-executor.js"
import { registerDeadline, unregisterDeadline } from "../lib/delegation-budget.js"
import { DEFAULT_STORM_THRESHOLD } from "../lib/model-matrix.js"
import { appendHumanDecisions } from "../lib/progress-state.js"
import { derivePerTaskTimeoutMs } from "../lib/run-timeout.js"
import type { HumanDecisionEntry, ToolResult } from "../lib/uf-workflow-types.js"
import {
  createDelegateCommandTool,
  createGetProgressTool,
  type DelegateCommandData,
  type DelegateCommandDependencies,
  type FileExists,
  type FootprintEstimatorDependencies,
  type GitExec,
  type ProgressState,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Module-level mocks (hoisted by vitest)
// ---------------------------------------------------------------------------

// The tool's auto-checkpoint (D5) persists through readState/writeState,
// which hit node:fs directly. Mocking the four functions that path uses (the
// unleash-branch-gate.test.ts recipe, per Dewey learning
// uf-workflow-testing-20261015T201843) keeps every delegate test off the real
// filesystem: state I/O is observed through the recorded calls instead of
// performed. existsSync is mocked so tests can prove artifact verification
// probes existence only through the injected FileExists dependency, never
// through the default existsSync-backed checker. The command-file reader is
// injected separately (readText), so node:fs/promises — a different specifier
// used only by the plugin registration — stays untouched.
vi.mock("node:fs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:fs")>()
  return {
    ...actual,
    existsSync: vi.fn(),
    lstatSync: vi.fn(),
    readFileSync: vi.fn(),
    mkdirSync: vi.fn(),
    writeFileSync: vi.fn(),
  }
})

// The human-decisions relay envelope (SYNTH-001/006,
// synthetic-compaction-continue-prompt-rec task 2.2) persists through
// appendHumanDecisions — real filesystem I/O in progress-state. The
// store contract itself has a dedicated suite (human-decisions-store
// .test.ts, real temp roots); this suite pins the ENVELOPE wiring: what
// the plugin passes to the store, when (spawn time, before the child
// session runs), and how persist failures surface. Mocking the seam
// keeps every delegate test off the real filesystem — the same recipe
// as the node:fs mock above — while the spread of actual preserves
// readState/writeState for the auto-checkpoint path.
vi.mock("../lib/progress-state.js", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../lib/progress-state.js")>()
  return {
    ...actual,
    appendHumanDecisions: vi.fn(),
  }
})

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

/** Fake repository root — never a real path on any filesystem. */
const FAKE_CWD = "/fake/project"

/**
 * The zero-config model-matrix probe: every delegation reads
 * .uf/model-matrix.yaml through the shared readText seam before host
 * introspection (model-matrix-delegation D8), so successful runs record
 * it in reads.calls even when no matrix exists.
 */
const MATRIX_PROBE = resolve(FAKE_CWD, ".uf/model-matrix.yaml")

/** Host model identity the fake parent session reports. */
const HOST_MODEL = "host-provider/host-model"

type Client = PluginInput["client"]
type CreateOptions = Parameters<Client["session"]["create"]>[0]
type MessageOptions = Parameters<Client["session"]["message"]>[0]
type PromptOptions = Parameters<Client["session"]["prompt"]>[0]
type AbortOptions = Parameters<Client["session"]["abort"]>[0]
type CreateResponse = Awaited<ReturnType<Client["session"]["create"]>>
type MessageResponse = Awaited<ReturnType<Client["session"]["message"]>>
type PromptResponse = Awaited<ReturnType<Client["session"]["prompt"]>>
type AbortResponse = Awaited<ReturnType<Client["session"]["abort"]>>

interface FakeCalls {
  readonly creates: CreateOptions[]
  readonly messages: MessageOptions[]
  readonly prompts: PromptOptions[]
  readonly aborts: AbortOptions[]
}

interface FakeClientOptions {
  /** Variant reported by the parent assistant message (default "default"; explicit undefined = absent). */
  readonly hostVariant?: string
  /** Parent assistant message reports no model identity (structural no-model resolution failure). */
  readonly hostModelMissing?: boolean
  /** Error payload returned by session.message (transient host-resolution failure). */
  readonly messageError?: unknown
  /** Error payload returned by session.create (child_create_failed path). */
  readonly createError?: unknown
  /** Parts reported by the child prompt response (default one text part). */
  readonly parts?: readonly unknown[]
  /** Hang session.prompt on its abort signal (timeout/cancellation paths). */
  readonly hangPrompt?: boolean
  /** Hang session.message on its abort signal (host-resolution cancellation). */
  readonly hangMessage?: boolean
  /** Expose the activity-probe endpoints with a never-changing child (idle paths). */
  readonly stalledProbe?: boolean
  /**
   * Expose the activity-probe endpoints whose latest part is a running
   * question tool call (question-gate wedge).
   */
  readonly questionProbe?: boolean
  /**
   * Expose the activity-probe endpoints with a growing text part
   * (never idle-aborts; long wall-clock paths).
   */
  readonly activeProbe?: boolean
  /**
   * Expose the activity-probe endpoints serving a fresh compaction marker
   * on every probe (the storm-detector path): each sample observes a new
   * marker id, so the storm accounting climbs toward the threshold with
   * no substantive progress between samples.
   */
  readonly compactionProbe?: boolean
  /**
   * Expose the activity-probe endpoints with a quiet probe root and one
   * unmanaged descendant session serving a fresh compaction marker on
   * every probe (the named-storm surfacing path): the root never
   * compacts, so the per-session accounting accrues wholly on the
   * descendant — no registered deadline marks it executor-managed, so it
   * counts per session — and the storm envelope must name that
   * descendant rather than the delegation's child session.
   */
  readonly descendantStormProbe?: boolean
  /**
   * Callback fired when the fake client successfully creates the child
   * session (SYNTH-001/006 task 2.2): spawn-ordering probes assert the
   * human-decision relay persist precedes the child session's creation.
   */
  readonly onCreate?: () => void
}

function fieldsResponse<T>(data: T): T {
  return data
}

/** Assistant message info carrying only the fields the executor reads. */
function assistantInfo(providerID: string, modelID: string): Record<string, unknown> {
  return {
    id: "message-child",
    sessionID: "child-session",
    role: "assistant",
    time: { created: 1, completed: 2 },
    parentID: "message-parent",
    providerID,
    modelID,
    mode: "primary",
    path: { cwd: FAKE_CWD, root: FAKE_CWD },
    cost: 0,
    tokens: { input: 0, output: 0, reasoning: 0, cache: { read: 0, write: 0 } },
  }
}

function textPart(text: string): Record<string, unknown> {
  return {
    id: `text-${text}`,
    sessionID: "child-session",
    messageID: "message-child",
    type: "text",
    text,
  }
}

function finishPart(cost: number, input: number, output: number): Record<string, unknown> {
  return {
    id: `finish-${cost}-${input}-${output}`,
    sessionID: "child-session",
    messageID: "message-child",
    type: "step-finish",
    reason: "stop",
    cost,
    tokens: { input, output, reasoning: 0, cache: { read: 0, write: 0 } },
  }
}

/**
 * Fake OpenCode client (invoke-agent.test.ts precedent) driving the REAL
 * executeAgentSession executor. The child session lifecycle — host
 * introspection, create, prompt, abort cleanup — runs for real against
 * recorded fakes, so timeout and cancellation propagation is exercised
 * through the same code path production uses, not a stubbed result.
 */
function fakeClient(options: FakeClientOptions = {}): { readonly client: Client; readonly calls: FakeCalls } {
  const calls: FakeCalls = { creates: [], messages: [], prompts: [], aborts: [] }
  // Probe counter backing the activeProbe and compactionProbe options:
  // each probe reports a new part (a text part or a compaction marker) so
  // the activity fingerprint changes on every sample.
  let probeCount = 0
  const session = {
    create: async (request: CreateOptions): Promise<CreateResponse> => {
      calls.creates.push(request)
      if (options.createError !== undefined) {
        return fieldsResponse({ data: undefined, error: options.createError } as unknown as CreateResponse)
      }
      // Fires only on successful creation — a createError run created no
      // child session, so an ordering probe must not observe one.
      options.onCreate?.()
      return fieldsResponse({ data: { id: "child-session" }, error: undefined } as unknown as CreateResponse)
    },
    message: async (request: MessageOptions): Promise<MessageResponse> => {
      calls.messages.push(request)
      if (options.messageError !== undefined) {
        return fieldsResponse({ data: undefined, error: options.messageError } as unknown as MessageResponse)
      }
      if (options.hangMessage === true) {
        return await new Promise<MessageResponse>((_resolve, reject) => {
          request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
        })
      }
      return fieldsResponse({
        data: {
          info: {
            // Empty provider/model identity drives reportedModel to null —
            // the structural no-model resolution failure path.
            ...assistantInfo(
              options.hostModelMissing === true ? "" : "host-provider",
              options.hostModelMissing === true ? "" : "host-model",
            ),
            variant: "hostVariant" in options ? options.hostVariant : "default",
          },
          parts: [],
        },
        error: undefined,
      } as unknown as MessageResponse)
    },
    prompt: async (request: PromptOptions): Promise<PromptResponse> => {
      calls.prompts.push(request)
      if (options.hangPrompt === true) {
        return await new Promise<PromptResponse>((_resolve, reject) => {
          request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
        })
      }
      // The child reports the model it was asked to run, so successful
      // delegations never trip the executor's model-mismatch guard.
      const bodyModel = request.body?.model
      const info = assistantInfo(bodyModel?.providerID ?? "provider", bodyModel?.modelID ?? "model")
      return fieldsResponse({
        data: { info, parts: options.parts ?? [textPart("delegated output")] },
        error: undefined,
      } as unknown as PromptResponse)
    },
    abort: async (request: AbortOptions): Promise<AbortResponse> => {
      calls.aborts.push(request)
      return fieldsResponse({ data: true, error: undefined } as unknown as AbortResponse)
    },
    // Activity-probe endpoints (delegation-timeout-budgets D2), present only
    // for idle-path tests: the child never changes and has no descendants.
    ...(options.stalledProbe === true
      ? {
          children: async () => ({ data: [] }),
          messages: async () => ({
            data: [{ info: { id: "message-child" }, parts: [textPart("thinking")] }],
          }),
        }
      : {}),
    // Question-gate wedge: the latest part of the latest message is a
    // running question tool call, so the idle path reports
    // pending_question instead of a plain stall.
    ...(options.questionProbe === true
      ? {
          children: async () => ({ data: [] }),
          messages: async () => ({
            data: [
              {
                info: { id: "message-child" },
                parts: [
                  {
                    id: "tool-question",
                    sessionID: "child-session",
                    messageID: "message-child",
                    type: "tool",
                    tool: "question",
                    state: { status: "running" },
                  },
                ],
              },
            ],
          }),
        }
      : {}),
    // Long wall-clock paths: the probed message grows every probe, so the
    // fingerprint keeps changing and the idle clock keeps resetting.
    ...(options.activeProbe === true
      ? {
          children: async () => ({ data: [] }),
          messages: async () => ({
            data: [{ info: { id: "message-child" }, parts: [textPart(`thinking-${++probeCount}`)] }],
          }),
        }
      : {}),
    // Compaction-storm path: every probe observes a fresh compaction
    // marker (distinct ids are distinct counted markers), so the storm
    // accounting climbs toward the threshold with no substantive
    // progress between samples.
    ...(options.compactionProbe === true
      ? {
          children: async () => ({ data: [] }),
          messages: async () => ({
            data: [
              {
                info: { id: `compaction-${++probeCount}` },
                parts: [{ type: "compaction", auto: true }],
              },
            ],
          }),
        }
      : {}),
    // Named-storm surfacing path: the probe root stays quiet (its latest
    // message is plain text — neither a marker nor substantive progress)
    // while one unmanaged descendant serves a fresh marker on every
    // probe, so the per-session storm signal accrues wholly on the
    // descendant and the abort envelope names it, distinct from the
    // child session the delegation created (issue #1125: recovery must
    // be able to tell which session was thrashing).
    ...(options.descendantStormProbe === true
      ? {
          children: async (request: { path: { id: string } }) => ({
            data:
              request.path.id === "child-session"
                ? [{ id: "looper-session", time: { created: 1 } }]
                : [],
          }),
          messages: async (request: { path: { id: string } }) => ({
            data:
              request.path.id === "child-session"
                ? [{ info: { id: "message-child" }, parts: [textPart("dispatching reviewers")] }]
                : [
                    {
                      info: { id: `compaction-${++probeCount}` },
                      parts: [{ type: "compaction", auto: true }],
                    },
                  ],
          }),
        }
      : {}),
  }
  return { client: { session } as unknown as Client, calls }
}

/**
 * Recording GitExec answering the auto-checkpoint key derivation with a
 * canned branch. The delegate tool runs git only when checkpoint_step is
 * supplied on a successful run, so tests assert on the recorded calls to
 * prove exactly that.
 */
function recordingExec(
  branch = "opsx/unleash-deterministic-tools",
): { exec: GitExec; calls: Array<{ args: string[]; cwd: string }> } {
  const calls: Array<{ args: string[]; cwd: string }> = []
  const exec: GitExec = (args, cwd) => {
    calls.push({ args, cwd })
    if (args.join(" ") === "rev-parse --abbrev-ref HEAD") return branch
    throw new Error(`unexpected git command: ${args.join(" ")}`)
  }
  return { exec, calls }
}

/**
 * Sequenced GitExec for per-task resume-hint tests (PTB-FR-003): answers
 * `status --porcelain -z` from a queue of canned outputs (delegation-start
 * before-state, then abort-time after-state), answers the checkpoint branch
 * read, and throws on anything else. Records every call for order
 * assertions; an exhausted queue throws so an unexpected capture fails the
 * test instead of silently returning stale state.
 */
function sequencedStatusExec(
  statuses: string[],
  branch = "opsx/unleash-deterministic-tools",
): { exec: GitExec; calls: Array<{ args: string[]; cwd: string }> } {
  const calls: Array<{ args: string[]; cwd: string }> = []
  let statusIndex = 0
  const exec: GitExec = (args, cwd) => {
    calls.push({ args, cwd })
    const command = args.join(" ")
    if (command === "status --porcelain -z") {
      const status = statuses[statusIndex]
      statusIndex++
      if (status === undefined) throw new Error(`no canned status for call ${statusIndex}`)
      return status
    }
    if (command === "rev-parse --abbrev-ref HEAD") return branch
    throw new Error(`unexpected git command: ${command}`)
  }
  return { exec, calls }
}

/**
 * Fake readText keyed on absolute paths (the tool resolves command_path
 * against the cwd before reading). Records every read so tests can assert
 * both repo-relative resolution and that the reviewer manifest is never
 * consulted through this dependency.
 */
function fakeReadText(files: Record<string, string>): {
  readText: DelegateCommandDependencies["readText"]
  calls: string[]
} {
  const calls: string[] = []
  const readText = async (path: string): Promise<string> => {
    calls.push(path)
    if (!(path in files)) {
      throw Object.assign(new Error(`ENOENT: no such file or directory, ${path}`), { code: "ENOENT" })
    }
    return files[path]
  }
  return { readText, calls }
}

/**
 * Sequenced readText for per-task resume-hint tests (PTB-FR-003): serves
 * one tasks.md path from a queue of canned bodies (delegation-start
 * before-state, then abort-time after-state) and throws ENOENT for every
 * other path — the model-matrix probe included — so unknown reads fail
 * closed the way production fs reads do. An exhausted queue throws rather
 * than serving stale state.
 */
function sequencedTasksReadText(
  tasksAbs: string,
  bodies: string[],
): DelegateCommandDependencies["readText"] {
  let bodyIndex = 0
  return async (path: string): Promise<string> => {
    if (path !== tasksAbs) {
      throw Object.assign(new Error(`ENOENT: no such file or directory, ${path}`), { code: "ENOENT" })
    }
    const body = bodies[bodyIndex]
    bodyIndex++
    if (body === undefined) throw new Error(`no canned tasks.md body for read ${bodyIndex}`)
    return body
  }
}

/**
 * Footprint-estimator seam (DELEG-FR-001, compaction-storm task 1.1)
 * reading from the same fake files map as fakeReadText: the estimator
 * measures command bodies through its own guarded read (deliberately
 * not the tool-level readText seam, whose calls the tests assert
 * exactly), so command-path tests inject this to keep the measurement
 * deterministic. realpath is the identity — the lexical path already
 * sits inside the fake repository root.
 */
function fakeFootprintSeam(files: Record<string, string>): FootprintEstimatorDependencies {
  return {
    readText: async (path) => {
      if (!(path in files)) {
        throw Object.assign(new Error(`ENOENT: no such file or directory, ${path}`), { code: "ENOENT" })
      }
      return files[path]
    },
    realpath: (path) => path,
  }
}

/** Fake FileExists keyed on repo-relative paths resolved against FAKE_CWD. */
function fakeFileExists(existing: readonly string[]): { fileExists: FileExists; calls: string[] } {
  const existingPaths = new Set(existing.map((rel) => resolve(FAKE_CWD, rel)))
  const calls: string[] = []
  const fileExists: FileExists = (path: string) => {
    calls.push(path)
    return existingPaths.has(path)
  }
  return { fileExists, calls }
}

/**
 * Parser stub mirroring Bun.YAML for the shapes under test (the
 * model-matrix.test.ts / review-dispatch convention): comments-only and
 * empty documents parse to null; everything else is JSON (a superset of
 * the YAML the fixtures use).
 */
function testParseYaml(text: string): unknown {
  const content = text
    .split(/\r?\n/)
    .filter((line) => line.trim() !== "" && !line.trim().startsWith("#"))
    .join("\n")
  if (content === "") {
    return null
  }
  return JSON.parse(text) as unknown
}

/**
 * Matrix-free model-matrix selection dependencies shared by the
 * zero-config tests: the YAML stub is never reached (fakeReadText reports
 * .uf/model-matrix.yaml absent, so loadModelMatrix resolves before
 * parsing), and every classifier seam fails the test when reached — the
 * bypass predicate must keep the cascade dark when no matrix is
 * configured.
 */
const matrixFreeDeps: DelegateCommandDependencies["modelMatrix"] = {
  parseYaml: testParseYaml,
  classifierFetch: (() => {
    throw new Error("classifier fetch must not run without a configured matrix")
  }) as typeof fetch,
  classifierReadEnv: () => undefined,
  spawnClassifierAgent: async () => {
    throw new Error("classifier spawn must not run without a configured matrix")
  },
}

/**
 * Recording model-matrix selection dependencies for matrix-driven tests:
 * the spawn seam records its prompt and returns a canned response
 * (default: a classifier JSON response naming no pack, exercising the
 * fall-through), and the fetch/env seams fail the test when reached (the
 * LLM engine only runs when the agent engine fails first).
 */
function fakeModelMatrixDeps(options: {
  readonly classifierResponse?: string
}): {
  modelMatrix: DelegateCommandDependencies["modelMatrix"]
  spawnCalls: Array<{ agent: string; prompt: string }>
} {
  const spawnCalls: Array<{ agent: string; prompt: string }> = []
  const modelMatrix: DelegateCommandDependencies["modelMatrix"] = {
    parseYaml: testParseYaml,
    classifierFetch: (() => {
      throw new Error("classifier fetch must not run when the agent engine answers")
    }) as typeof fetch,
    classifierReadEnv: () => undefined,
    spawnClassifierAgent: async (_context, agent, prompt) => {
      spawnCalls.push({ agent, prompt })
      return options.classifierResponse ?? '{"pack": null, "confidence": 0.9}'
    },
  }
  return { modelMatrix, spawnCalls }
}

/** Prompt text delivered to the child session (exactly one text part). */
function deliveredPrompt(calls: FakeCalls): string {
  const parts = calls.prompts[0]?.body?.parts as unknown as Array<{ type: string; text: string }>
  if (parts === undefined || parts.length !== 1 || parts[0].type !== "text") {
    throw new Error(`expected exactly one text part, got ${JSON.stringify(parts)}`)
  }
  return parts[0].text
}

/** Node-shaped ENOENT error so readState/writeState treat a path as absent. */
function enoent(what: string): NodeJS.ErrnoException {
  return Object.assign(new Error(`ENOENT: no such file or directory, ${what}`), { code: "ENOENT" })
}

/**
 * Invoke the delegate tool with a full parent-session context and parse its
 * JSON output. cwd defaults to the fake repository root; tests exercising
 * real-filesystem defaults (the statSync-backed isDirectory) pass a real
 * scratch directory instead.
 */
async function invokeDelegate(
  exec: GitExec,
  deps: DelegateCommandDependencies,
  args: Record<string, unknown>,
  abort: AbortSignal = new AbortController().signal,
  cwd: string = FAKE_CWD,
): Promise<ToolResult<DelegateCommandData>> {
  const delegate = createDelegateCommandTool(exec, cwd, deps)
  const result = await (
    delegate as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string; messageID: string; directory: string; abort: AbortSignal },
      ) => Promise<{ output: string }>
    }
  ).execute(args, {
    sessionID: "parent-session",
    messageID: "message-parent",
    directory: cwd,
    abort,
  })
  return JSON.parse(result.output) as ToolResult<DelegateCommandData>
}

beforeEach(() => {
  vi.resetAllMocks()
  // Default state-store simulation: no prior state file or directory, so
  // readState returns null (first checkpoint) and writeState takes the
  // create-directory branch.
  vi.mocked(lstatSync).mockImplementation(() => {
    throw enoent("lstat state path")
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — command-file read and prompt composition
// ---------------------------------------------------------------------------

describe("uf_delegate_command command-file read and prompt composition", () => {
  it("reads the command file repo-relative and composes command text, feature-dir context, then extra prompt", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.plan.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Create the implementation plan." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      prompt: "Focus on the parser.",
    })

    expect(result.status).toBe("ok")

    // The command file is resolved against the injected cwd before reading;
    // the zero-config matrix probe follows (task 5.2 selection wiring).
    expect(reads.calls).toEqual([commandAbs, MATRIX_PROBE])

    // Host model resolution introspects the parent session's message.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.messages[0]).toMatchObject({
      path: { id: "parent-session", messageID: "message-parent" },
      query: { directory: FAKE_CWD },
    })

    // The child session is parented, titled with agent + resolved model, and
    // prompted with the default agent (cobalt-crush-dev).
    expect(fake.calls.creates[0]?.body).toMatchObject({
      parentID: "parent-session",
      title: `cobalt-crush-dev @ ${HOST_MODEL} (default)`,
    })
    expect(fake.calls.prompts[0]?.body).toMatchObject({
      agent: "cobalt-crush-dev",
      model: { providerID: "host-provider", modelID: "host-model" },
      variant: "default",
    })

    // Prompt composition order (D2): command text, then feature-dir
    // context, then extra instructions.
    expect(deliveredPrompt(fake.calls)).toBe(
      "Create the implementation plan.\n\nFeature directory: specs/018-unleash-command/\n\nFocus on the parser.",
    )
  })

  it("delegates prompt-only calls without command file or feature-dir sections", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.5 exactly.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    // Prompt-only delegation (per-task implement and code-fix calls) reads
    // no command file and prepends no feature-dir section; the only read
    // is the zero-config matrix probe.
    expect(reads.calls).toEqual([MATRIX_PROBE])
    expect(deliveredPrompt(fake.calls)).toBe("Implement task 2.5 exactly.")
  })

  it("accepts a command file at exactly the 1 MiB bound", async () => {
    const atBound = "x".repeat(1024 * 1024)
    const fake = fakeClient()
    const reads = fakeReadText({ [resolve(FAKE_CWD, "cmd.md")]: atBound })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, { command_path: "cmd.md" })

    // The bound is inclusive: exactly 1 MiB passes through to the prompt.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(deliveredPrompt(fake.calls)).toBe(atBound)
  })

  it("fails with COMMAND_FILE_READ_FAILED when the command file cannot be read", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({}) // nothing mapped — every read throws ENOENT
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/missing.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("COMMAND_FILE_READ_FAILED")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        `Cannot read command file: ENOENT: no such file or directory, ${resolve(FAKE_CWD, ".opencode/commands/missing.md")}`,
      )
    }
    // Failed before any session activity.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — input validation
// ---------------------------------------------------------------------------

describe("uf_delegate_command input validation", () => {
  it("rejects command files over 1 MiB with COMMAND_FILE_TOO_LARGE", async () => {
    const over = "x".repeat(1024 * 1024 + 1)
    const fake = fakeClient()
    const reads = fakeReadText({ [resolve(FAKE_CWD, "big.md")]: over })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, { command_path: "big.md" })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("COMMAND_FILE_TOO_LARGE")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("Command file exceeds 1 MiB limit (1048577 bytes)")
    }
    // Rejected before any session activity.
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("rejects divisor-* reviewer agents with INVALID_AGENT directing to invoke_agent", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      agent: "divisor-guard",
      prompt: "Review this change.",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_AGENT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "reviewer agents (divisor-*) are dispatched via invoke_agent; uf_delegate_command is for content delegation",
      )
    }
    // The delegation/review boundary is structural (D2): no manifest read
    // and no session activity precede the rejection.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects inputs with neither command_path nor prompt", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {})

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      // The empty call violates both the instruction rule and the
      // prompt-only classification rule; both issues are reported.
      expect(result.message).toBe(
        "prompt: one of command_path or prompt is required; " +
          "task_classification: prompt-only delegation (no command_path) must declare task_classification -- " +
          "classify the work as deterministic or judgment before delegating",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([0, 1.5, "3"])("rejects invalid phase %p with INVALID_INPUT before any child spawn", async (phase) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Implement task 5.2 of the change.",
      task_classification: "judgment",
      checkpoint_step: "implement",
      phase,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("phase")
    }
    // Rejected in zod validation before any child spawn or state I/O.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — task classification and skill contract
// ---------------------------------------------------------------------------

describe("uf_delegate_command task classification and skill contract", () => {
  it("requires task_classification on prompt-only delegation", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 1.1 of the change.",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "task_classification: prompt-only delegation (no command_path) must declare task_classification -- " +
          "classify the work as deterministic or judgment before delegating",
      )
    }
    // Rejected in zod validation before any file read or session activity.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("proceeds with prompt-only delegation classified judgment (judgment pass-through)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Evaluate two design approaches for the parser.",
      task_classification: "judgment",
    })

    // Judgment work carries no skill requirement and passes through.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
  })

  it("returns SKILL_REFERENCE_REQUIRED when deterministic classification supplies neither skill nor reason", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Implement task 3.2 of the change.",
      task_classification: "deterministic",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_REFERENCE_REQUIRED")
      expect(result.retryable).toBe(false)
      // The envelope names every remediation the contract allows: pass a
      // skill reference, record the absence reason, or classify as judgment.
      expect(result.message).toBe(
        "Deterministic-classified delegation requires a skill reference or a recorded absence reason. " +
          "Remediation: pass a skill reference (skill), record the absence reason (skill_absent_reason), " +
          "or classify the work as judgment; retrying will not fix this.",
      )
    }
    // Fail-closed before host-model resolution and child spawn: no parent
    // introspection, no session activity, no state I/O, no command-file read.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
    expect(reads.calls).toEqual([])
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("returns SKILL_REFERENCE_REQUIRED for a command-path call classified deterministic without a reference", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({ [resolve(FAKE_CWD, "cmd.md")]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      task_classification: "deterministic",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_REFERENCE_REQUIRED")
      expect(result.retryable).toBe(false)
    }
    // The requirement binds any call declaring deterministic work --
    // command-path included -- and fires before the command file is read.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("proceeds with deterministic classification carrying a recorded absence reason", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: "no repo script exists for this one-off rename",
    })

    // The recorded reason satisfies the reference requirement, so the run
    // proceeds with no skill lookup (spec scenario: reason recorded)...
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(artifact.calls).toEqual([])
    // ...and no skill directive is injected: only a resolved reference
    // renders one, so the composed prompt is the bare instruction.
    expect(deliveredPrompt(fake.calls)).toBe("Rename the variable across the module.")
    // The success envelope records the classification and the reason (the
    // checkpoint-data half of the scenario is pinned by the audit-trail
    // suite's verbatim-recording test).
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBe("no repo script exists for this one-off rename")
    }
  })

  it("accepts deterministic classification carrying a skill reference", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    // The supplied skill satisfies the deterministic reference requirement
    // (task 1.2) and resolves under the first skill root (task 1.3), so
    // the run proceeds after the existence probe.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, ".opencode/skills/pre-flight/SKILL.md")])
  })

  it.each([
    ["skill", "pre-flight"],
    ["skill_absent_reason", "no repo script exists for this one-off rename"],
  ] as const)("rejects judgment-classified %s with INVALID_INPUT", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      prompt: "Evaluate two design approaches for the parser.",
      task_classification: "judgment",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        `${field}: skill references apply to deterministic work only -- ` +
          "classify the work as deterministic or omit the field",
      )
    }
    // The skill directive asserts deterministic work: no session spawned.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    ["skill", "pre-flight"],
    ["skill_absent_reason", "no repo script exists for this one-off rename"],
  ] as const)("rejects unclassified command-path calls supplying %s", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      command_path: "cmd.md",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        `${field}: skill references apply to deterministic work only -- ` +
          "classify the work as deterministic or omit the field",
      )
    }
    // Rejected before the command file was read or any session spawned.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects skill and skill_absent_reason supplied together", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
      skill_absent_reason: "no repo script exists for this one-off rename",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "skill: skill and skill_absent_reason are mutually exclusive -- " +
          "pass a skill reference or an absence reason, never both",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    ["skill", "x".repeat(257)],
    ["skill_absent_reason", "x".repeat(513)],
  ] as const)("rejects over-length %s values", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      // The path prefix names the offending field (zod's bound-message
      // wording is version-specific; the field identity is the contract).
      expect(result.message).toContain(field)
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    ["skill", "x".repeat(256)],
    ["skill_absent_reason", "x".repeat(512)],
  ] as const)("accepts %s at exactly its length bound", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    // The at-bound skill name must also resolve (task 1.3): it exists under
    // the first skill root, so the length check and resolution both pass.
    const existing = field === "skill" ? [`.opencode/skills/${value}/SKILL.md`] : []
    const artifact = fakeFileExists(existing)
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    // The bounds are inclusive: exactly 256 skill / 512 reason characters
    // pass input validation and the run proceeds.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
  })

  it("rejects an empty-string skill, instructing how absence is expressed", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "skill: empty string is invalid -- pass a skill reference or record skill_absent_reason instead",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects an empty-string skill_absent_reason, instructing how absence is expressed", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill_absent_reason: "",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "skill_absent_reason: empty string is invalid -- absence is expressed by omitting the field",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each(["Deterministic", "DETERMINISTIC", "mechanical", "deterministic "])(
    "rejects non-enum task_classification %p with INVALID_INPUT naming the accepted values",
    async (taskClassification) => {
      const fake = fakeClient()
      const reads = fakeReadText({})
      const artifact = fakeFileExists([])
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        readText: reads.readText,
        fileExists: artifact.fileExists,
      }

      const result = await invokeDelegate(recordingExec().exec, deps, {
        prompt: "Run the repository's pre-flight procedure.",
        task_classification: taskClassification,
      })

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("INVALID_INPUT")
        expect(result.retryable).toBe(false)
        expect(result.message).toContain("task_classification")
        // The failure names the accepted enum values.
        expect(result.message).toContain("deterministic")
        expect(result.message).toContain("judgment")
      }
      expect(fake.calls.creates).toHaveLength(0)
    },
  )

  it("rejects a non-string task_classification value with INVALID_INPUT", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: 42,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("task_classification")
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("proceeds with command-path delegation supplying no classification (command-path exemption)", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, { command_path: "cmd.md" })

    // The Steps 2/3 call shape is unaffected: no classification and no
    // skill fields; the run proceeds exactly as before this change (the
    // matrix probe is the one new read, zero-config D8).
    expect(result.status).toBe("ok")
    expect(reads.calls).toEqual([commandAbs, MATRIX_PROBE])
    expect(fake.calls.prompts).toHaveLength(1)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — skill resolution (delegation-skill-enforcement D2)
// ---------------------------------------------------------------------------

describe("uf_delegate_command skill resolution", () => {
  it("does not resolve a skill present only under the legacy .opencode/skill/ root (#933)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skill/unbound-force-heroes/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Load the hero routing skill.",
      task_classification: "deterministic",
      skill: "unbound-force-heroes",
    })

    // uf init migrates the legacy root into .opencode/skills/, so the
    // plugin never probes it: a legacy-only copy fails closed.
    expect(result.status === "error" && result.code).toBe("SKILL_NOT_FOUND")
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/unbound-force-heroes/SKILL.md"),
      resolve(FAKE_CWD, "unbound-force-heroes"),
    ])
    expect(fake.calls.creates).toHaveLength(0)
    expect(existsSync).not.toHaveBeenCalled()
  })

  it("resolves a bare repo-root file name when no skill root holds it (#934)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists(["Makefile"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
      // Same host-independence concern as the bare `..` test: the
      // fileExists fake claims FAKE_CWD/Makefile exists, so isDirectory
      // is pinned rather than left to a real statSync against the host.
      isDirectory: () => false,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the plugin-test target.",
      task_classification: "deterministic",
      skill: "Makefile",
    })

    expect(result.status).toBe("ok")
    // Skill roots are probed first; the root file is the fallback.
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/Makefile/SKILL.md"),
      resolve(FAKE_CWD, "Makefile"),
    ])
    expect(deliveredPrompt(fake.calls)).toContain("Required skill: Makefile\n\n")
    expect(deliveredPrompt(fake.calls)).toContain("run the script at Makefile")
  })

  it("does not resolve a directory as a root-file or path reference", async () => {
    const artifact = fakeFileExists(["docs", "scripts/tools"])
    const directories = new Set([resolve(FAKE_CWD, "docs"), resolve(FAKE_CWD, "scripts/tools")])
    const deps: DelegateCommandDependencies = {
      client: fakeClient().client,
      modelMatrix: matrixFreeDeps,
      readText: fakeReadText({}).readText,
      fileExists: artifact.fileExists,
      isDirectory: (path: string) => directories.has(path),
    }

    for (const skill of ["docs", "scripts/tools"]) {
      const result = await invokeDelegate(recordingExec().exec, deps, {
        prompt: "Run it.",
        task_classification: "deterministic",
        skill,
      })
      expect(result.status === "error" && result.code).toBe("SKILL_NOT_FOUND")
    }
  })

  it("filters a real directory and accepts a real file through the default isDirectory (#948)", async () => {
    // The default isDirectory (statSync-backed) has no positive-path coverage
    // elsewhere: every other test injects a fake, so a regression returning
    // isFile() -- or always false -- would pass the suite while letting a
    // real directory resolve as a script reference. statSync, mkdtemp, and
    // node:fs/promises stay real under the node:fs mock (only existsSync,
    // lstatSync, readFileSync, mkdirSync, and writeFileSync are replaced),
    // so this exercise runs against a real scratch tree.
    const directory = await mkdtemp(join(tmpdir(), "uf-delegate-isdir-"))
    try {
      await mkdir(join(directory, "docs"))
      await writeFile(join(directory, "run.sh"), "#!/bin/sh\n", { mode: 0o755 })
      const realExists: FileExists = (path) => {
        try {
          statSync(path)
          return true
        } catch {
          return false
        }
      }
      const deps: DelegateCommandDependencies = {
        client: fakeClient().client,
        modelMatrix: matrixFreeDeps,
        readText: fakeReadText({}).readText,
        fileExists: realExists,
        // isDirectory deliberately omitted: resolution must fall back to the
        // production default rather than a test fake.
      }

      // A real directory is not a root-file reference.
      const dirResult = await invokeDelegate(
        recordingExec().exec,
        deps,
        { prompt: "Run it.", task_classification: "deterministic", skill: "docs" },
        new AbortController().signal,
        directory,
      )
      expect(dirResult.status === "error" && dirResult.code).toBe("SKILL_NOT_FOUND")

      // A real file resolves through the same default.
      const fileResult = await invokeDelegate(
        recordingExec().exec,
        deps,
        { prompt: "Run it.", task_classification: "deterministic", skill: "run.sh" },
        new AbortController().signal,
        directory,
      )
      expect(fileResult.status).toBe("ok")
      expect(fileResult.status === "ok" && fileResult.data.skill).toBe("run.sh")
    } finally {
      await rm(directory, { recursive: true, force: true })
    }
  })

  it("prefers a skill over a same-name repo-root file", async () => {
    const fake = fakeClient()
    const artifact = fakeFileExists([".opencode/skills/build/SKILL.md", "build"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: fakeReadText({}).readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Build it.",
      task_classification: "deterministic",
      skill: "build",
    })

    expect(result.status).toBe("ok")
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, ".opencode/skills/build/SKILL.md")])
  })

  it("does not resolve a bare `..` outside the repository root", async () => {
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fakeClient().client,
      modelMatrix: matrixFreeDeps,
      readText: fakeReadText({}).readText,
      fileExists: (path: string) => path === resolve(FAKE_CWD, "..") || artifact.fileExists(path),
      // isDirectory must be injected, not left at the statSync-backed
      // production default: the fileExists fake claims the parent of
      // FAKE_CWD exists, and a host that really has that directory
      // would let the directory filter mask a traversal-guard
      // regression (the guard, not the filter, must trip here).
      isDirectory: () => false,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Escape.",
      task_classification: "deterministic",
      skill: "..",
    })

    expect(result.status === "error" && result.code).toBe("SKILL_NOT_FOUND")
  })

  it("rejects a command or make target as SKILL_NOT_FOUND without probing", async () => {
    const fake = fakeClient()
    const artifact = fakeFileExists(["Makefile"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: fakeReadText({}).readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the plugin tests.",
      task_classification: "deterministic",
      skill: "make plugin-test",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("commands and make targets are not skill references")
      expect(result.message).toContain("skill_absent_reason")
    }
    expect(artifact.calls).toEqual([])
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("resolves the .opencode/skills/ copy when a stale legacy copy also exists", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([
      ".opencode/skills/colliding-skill/SKILL.md",
      ".opencode/skill/colliding-skill/SKILL.md",
    ])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the colliding skill.",
      task_classification: "deterministic",
      skill: "colliding-skill",
    })

    // The .opencode/skills/ entry wins: probing stops there and the
    // legacy .opencode/skill/ copy is never consulted...
    expect(result.status).toBe("ok")
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/colliding-skill/SKILL.md"),
    ])
    // ...and that first-root entry is the resolved reference recorded in
    // the audit trail (spec scenario): the success envelope carries it.
    if (result.status === "ok") {
      expect(result.data.skill).toBe(".opencode/skills/colliding-skill/SKILL.md")
    }
  })

  it("resolves a script path without any skill-directory lookup", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".specify/scripts/bash/check-prerequisites.sh"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the prerequisites check.",
      task_classification: "deterministic",
      skill: ".specify/scripts/bash/check-prerequisites.sh",
    })

    // The path form is accepted as a traversal-protected repo-relative path
    // and the run proceeds (spec scenario: script-path reference).
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    // Only the repo-relative script path is probed — no skill-directory
    // lookup occurs.
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".specify/scripts/bash/check-prerequisites.sh"),
    ])
  })

  it("rejects a skill path escaping the repository root with SKILL_TRAVERSAL", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the outside script.",
      task_classification: "deterministic",
      skill: "../outside/repo-tool.sh",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_TRAVERSAL")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Path traversal blocked: skill resolves outside the repository root",
      )
    }
    // The traversal guard fires before any existence probe or session
    // activity (the command_path / expected_artifact precedent).
    expect(artifact.calls).toEqual([])
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("returns SKILL_NOT_FOUND naming both searched skill roots and spawns no child session", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([]) // nothing exists
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "does-not-exist",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
      // The envelope names every searched location (the skill root
      // and the repo-root file fallback).
      expect(result.message).toBe(
        'Skill reference "does-not-exist" does not resolve. ' +
          "Searched: .opencode/skills/does-not-exist/SKILL.md, does-not-exist (repo-root file). " +
          "Remediation: pass an existing skill name or a repo-relative script path, " +
          "or record skill_absent_reason when no repo skill applies; retrying will not fix this.",
      )
    }
    // The skill root, then the root file, were probed before failing closed.
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/does-not-exist/SKILL.md"),
      resolve(FAKE_CWD, "does-not-exist"),
    ])
    // Fail-closed before host-model resolution and the child spawn: no
    // parent introspection, no session activity, no aborts, no state I/O,
    // no state-key derivation (spec scenario: spawns no child session).
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
    expect(reads.calls).toEqual([])
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("returns SKILL_NOT_FOUND naming the searched path for a missing script", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the missing script.",
      task_classification: "deterministic",
      skill: "scripts/missing-tool.sh",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        'Skill reference "scripts/missing-tool.sh" does not resolve. ' +
          "Searched: scripts/missing-tool.sh (repo-relative path). " +
          "Remediation: pass an existing skill name or a repo-relative script path, " +
          "or record skill_absent_reason when no repo skill applies; retrying will not fix this.",
      )
    }
    // The repo-relative path was probed through the traversal guard first.
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, "scripts/missing-tool.sh")])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("resolves the skill before reading the command file on command-path delegation", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      task_classification: "deterministic",
      skill: "does-not-exist",
    })

    // The reference fails closed before the command file is read or any
    // session spawns (mirrors the task 1.2 gate ordering: contract
    // validation completes before I/O).
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
    }
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — skill directive injection (delegation-skill-enforcement D3)
// ---------------------------------------------------------------------------

describe("uf_delegate_command skill directive injection", () => {
  it("renders the directive after command text and feature-dir context, before the extra prompt", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.plan.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Create the implementation plan." })
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      prompt: "Focus on the parser.",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    // Golden shape (design D3): the directive is a fixed, tool-rendered
    // section — a header line naming the resolved reference, a blank
    // line, then the fixed paragraph — inserted after command text and
    // feature-dir context, immediately before the extra prompt. The
    // resolved reference headlines the directive and names the script
    // path; the `skill`-tool name is derived from the skill root.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(deliveredPrompt(fake.calls)).toBe(
      "Create the implementation plan.\n\n" +
        "Feature directory: specs/018-unleash-command/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Focus on the parser.",
    )
  })

  it("renders the directive on prompt-only delegation without command text", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // Per-task implement call shape: prompt-only with feature-dir context.
    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    // Without command text the directive holds its fixed position: after
    // feature-dir context, immediately before the extra prompt.
    expect(result.status).toBe("ok")
    expect(reads.calls).toEqual([MATRIX_PROBE])
    expect(deliveredPrompt(fake.calls)).toBe(
      "Feature directory: openspec/changes/delegation-skill-enforcement/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Run the repository's pre-flight procedure.",
    )
  })

  it("renders the directive for a script-path reference with the path in every slot", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".specify/scripts/bash/check-prerequisites.sh"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the prerequisites check.",
      task_classification: "deterministic",
      skill: ".specify/scripts/bash/check-prerequisites.sh",
    })

    // A script reference resolves to itself: the resolved path headlines
    // the directive and fills the script slot, and — carrying no skill
    // name — also fills the `skill`-tool slot, where the "or run the
    // script" clause is the operative instruction (spec scenario:
    // script-path skill).
    expect(result.status).toBe("ok")
    expect(deliveredPrompt(fake.calls)).toBe(
      "Required skill: .specify/scripts/bash/check-prerequisites.sh\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name ".specify/scripts/bash/check-prerequisites.sh") or run the script at ' +
        ".specify/scripts/bash/check-prerequisites.sh, and execute its full procedure. " +
        "Do not substitute an improvised procedure.\n\n" +
        "Run the prerequisites check.",
    )
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — expected-artifact verification
// ---------------------------------------------------------------------------

describe("uf_delegate_command expected-artifact verification", () => {
  it("verifies the expected artifact after a successful run and returns the full success envelope", async () => {
    const fake = fakeClient({
      parts: [textPart("plan written"), finishPart(0.25, 11, 7)],
    })
    const reads = fakeReadText({})
    const artifact = fakeFileExists(["specs/018-unleash-command/plan.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Write the plan.",
      task_classification: "judgment",
      expected_artifact: "specs/018-unleash-command/plan.md",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        agent: "cobalt-crush-dev",
        status: "success",
        child_session_id: "child-session",
        artifact_created: true,
        checkpoint_recorded: false,
        // SYNTH-001/006 (synthetic-compaction-continue-prompt-rec task
        // 2.2): no human_decisions envelope on this call, so no stamped
        // decision ids — null, never an empty array.
        human_decision_ids: null,
        // Delegation-skill-enforcement D5 (task 1.5): the audit fields
        // surface in the envelope -- judgment classification, no skill
        // reference, no absence reason.
        task_classification: "judgment",
        skill: null,
        skill_absent_reason: null,
        // Model-matrix-delegation D7 (task 5.4): the selection audit
        // fields surface in every success envelope — zero-config calls
        // record the host fallback with no resolved values.
        selection_source: "host",
        selection_pack: null,
        selection_capability: null,
        selection_model: null,
        selection_confidence: null,
        selection_fallback_reason: null,
        // DELEG-FR-001 (compaction-storm task 1.1): headroom provenance —
        // prompt-only footprint = cached prefix 56,000 + "Write the plan."
        // (15 bytes = 4 tokens at 4 bytes/token); clean measurement and no
        // matrix, so the documented default margin applies.
        selection_footprint: 56_004,
        selection_footprint_fallback: null,
        selection_margin_tokens: 24_000,
        selection_margin_warning: null,
        // DELEG-FR-002 (compaction-storm): storm provenance — no matrix,
        // so the documented default threshold applies with no warning.
        storm_threshold: DEFAULT_STORM_THRESHOLD,
        storm_threshold_warning: null,
        // PTB-FR-001/PTB-FR-004 (per-task-delegation-budgets): non-per-task
        // delegations carry no derivation provenance and no remaining
        // budget.
        timeout_provenance: null,
        remaining_ms: null,
        usage: { cost_usd: 0.25, tokens: { input: 11, output: 7 } },
        provenance: {
          agent: "cobalt-crush-dev",
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: HOST_MODEL,
          resolved_parent_variant: "default",
          reported_child_model: HOST_MODEL,
          model_mismatch: false,
          // Content delegation writes files, unlike review invocations.
          read_only: false,
        },
        text: "plan written",
      })
    }
    // Verification probes the artifact through the injected FileExists only.
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, "specs/018-unleash-command/plan.md")])
    expect(existsSync).not.toHaveBeenCalled()
  })

  it("returns the exact plan.md STOP message and records no checkpoint when the artifact is missing", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Write the plan.",
      task_classification: "judgment",
      expected_artifact: "specs/018-unleash-command/plan.md",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_MISSING")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Plan generation failed -- plan.md was not created. Check the agent output for errors.",
      )
    }
    // The child session ran; verification failed only after it finished.
    expect(fake.calls.creates).toHaveLength(1)
    expect(fake.calls.prompts).toHaveLength(1)
    // An unverified run never checkpoints (D5: success-only) and never
    // derives a state key.
    expect(execCalls).toEqual([])
    expect(lstatSync).not.toHaveBeenCalled()
    expect(readFileSync).not.toHaveBeenCalled()
    expect(mkdirSync).not.toHaveBeenCalled()
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("returns the exact tasks.md STOP message when tasks.md is missing", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Generate the task list.",
      task_classification: "judgment",
      expected_artifact: "openspec/changes/unleash-deterministic-tools/tasks.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_MISSING")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Task generation failed -- tasks.md was not created. Check the agent output for errors.",
      )
    }
  })

  it("returns the generic artifact STOP message naming the missing path", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Write the design record.",
      task_classification: "judgment",
      expected_artifact: "specs/018-unleash-command/design.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_MISSING")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Artifact generation failed -- specs/018-unleash-command/design.md was not created. " +
          "Check the agent output for errors.",
      )
    }
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — child-session failure propagation
// ---------------------------------------------------------------------------

describe("uf_delegate_command child-session propagation", () => {
  it("propagates a timed-out child session as DELEGATE_FAILED with a retryable error", async () => {
    const fake = fakeClient({ hangPrompt: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      timeout: 5,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.message).toBe("timeout: invoke_agent run timed out")
      // The failure envelope threads the child session id so the timed-out
      // child subtree can be inspected directly from the delegation result.
      expect(result.child_session_id).toBe("child-session")
    }
    // The executor aborted the hanging child session.
    expect(fake.calls.aborts).toHaveLength(1)
  })

  it("propagates parent cancellation of the child session as DELEGATE_CANCELLED", async () => {
    const controller = new AbortController()
    const fake = fakeClient({ hangPrompt: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const pending = invokeDelegate(
      recordingExec().exec,
      deps,
      { prompt: "Run the delegation.", task_classification: "judgment" },
      controller.signal,
    )
    await new Promise((resolve) => setTimeout(resolve, 0))
    controller.abort(new Error("parent stopped"))
    const result = await pending

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_CANCELLED")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("cancelled: invoke_agent run was cancelled")
      // DELEGATE_CANCELLED carries the child session id too: the child
      // existed before the parent abort reached it.
      expect(result.child_session_id).toBe("child-session")
    }
    // Cancellation reached the child session with an independent cleanup
    // signal (invoke-agent precedent).
    expect(fake.calls.aborts).toHaveLength(1)
    expect(fake.calls.aborts[0]?.path).toEqual({ id: "child-session" })
    expect(fake.calls.aborts[0]?.signal).not.toBe(controller.signal)
  })

  it("returns DELEGATE_CANCELLED before spawning when cancellation hits host resolution", async () => {
    const controller = new AbortController()
    const fake = fakeClient({ hangMessage: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const pending = invokeDelegate(
      recordingExec().exec,
      deps,
      { prompt: "Run the delegation.", task_classification: "judgment" },
      controller.signal,
    )
    await new Promise((resolve) => setTimeout(resolve, 0))
    controller.abort(new Error("parent stopped"))
    const result = await pending

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_CANCELLED")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("uf_delegate_command run was cancelled")
      // No child session ever existed, so the envelope omits the field.
      expect(result.child_session_id).toBeUndefined()
    }
    // Cancelled before any child session existed.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
  })

  it("propagates child creation failures as DELEGATE_FAILED with sanitized errors and no checkpoint", async () => {
    const fake = fakeClient({ createError: "credential=abc /home/alice/private" })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("child_create_failed:")
      // No child session ever existed, so the envelope omits the field
      // (the DELEGATE_FAILED counterpart of the cancelled-before-child case).
      expect(result.child_session_id).toBeUndefined()
      // The executor's sanitizer redacted the credential and home path.
      expect(result.message).not.toMatch(/abc|\/home\/alice/)
    }
    // Failed runs never checkpoint (D5) and never derive a state key.
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("fails with HOST_MODEL_UNAVAILABLE, non-retryable, when the host variant is structurally invalid", async () => {
    const fake = fakeClient({ hostVariant: "not valid" })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "current assistant active variant is invalid. Remediation: run the command from a provider whose active " +
          "variant is a valid slug; retrying will not fix this.",
      )
    }
    // No child session was spawned for an unresolvable host model.
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("fails with HOST_MODEL_UNAVAILABLE, retryable, when host introspection hits a transient error", async () => {
    const fake = fakeClient({ messageError: "service unavailable" })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      expect(result.retryable).toBe(true)
    }
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("delegates without a variant override when the host assistant message reports no variant", async () => {
    const fake = fakeClient({ hostVariant: undefined })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    expect(fake.calls.creates[0]?.body).toMatchObject({
      title: `cobalt-crush-dev @ ${HOST_MODEL}`,
    })
    expect(fake.calls.prompts[0]?.body).not.toHaveProperty("variant")
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — manifest isolation
// ---------------------------------------------------------------------------

describe("uf_delegate_command manifest isolation", () => {
  it("never consults .uf/reviewer-capabilities.yaml through any dependency or filesystem surface", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.tasks.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Generate the task list." })
    const artifact = fakeFileExists(["openspec/changes/unleash-deterministic-tools/tasks.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/speckit.tasks.md",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      expected_artifact: "openspec/changes/unleash-deterministic-tools/tasks.md",
      checkpoint_step: "tasks",
    })

    expect(result.status).toBe("ok")

    // The only text reads are the command file and the zero-config matrix
    // probe — the reviewer manifest is not an input to content delegation
    // (D2 boundary is agent validation).
    expect(reads.calls).toEqual([commandAbs, MATRIX_PROBE])

    // No filesystem probe touches the manifest, and existence checks flow
    // exclusively through the injected FileExists.
    const probed = [
      ...vi.mocked(existsSync).mock.calls.map((call) => String(call[0])),
      ...vi.mocked(lstatSync).mock.calls.map((call) => String(call[0])),
      ...vi.mocked(readFileSync).mock.calls.map((call) => String(call[0])),
      ...vi.mocked(writeFileSync).mock.calls.map((call) => String(call[0])),
    ]
    expect(probed.length).toBeGreaterThan(0)
    for (const path of probed) {
      expect(path.includes("reviewer-capabilities")).toBe(false)
    }
    expect(existsSync).not.toHaveBeenCalled()
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — checkpoint recording semantics
// ---------------------------------------------------------------------------

describe("uf_delegate_command checkpoint recording", () => {
  it("auto-checkpoints checkpoint_step on verified success with the delegation facts as data", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.plan.md")
    const fake = fakeClient()
    const commandFiles: Record<string, string> = { [commandAbs]: "Create the implementation plan." }
    const reads = fakeReadText(commandFiles)
    const artifact = fakeFileExists(["specs/018-unleash-command/plan.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      // DELEG-FR-001 (task 1.1): the footprint seam reads the same fake
      // files map, so the estimator measures the command body cleanly
      // instead of hitting the real filesystem.
      modelMatrix: { ...matrixFreeDeps, footprint: fakeFootprintSeam(commandFiles) },
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/unleash-deterministic-tools")

    const result = await invokeDelegate(exec, deps, {
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      prompt: "Focus on the parser.",
      expected_artifact: "specs/018-unleash-command/plan.md",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.checkpoint_recorded).toBe(true)
      expect(result.data?.artifact_created).toBe(true)
    }

    // Exactly one git invocation — the state-key branch read — against the
    // injected cwd.
    expect(execCalls).toEqual([{ args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD }])

    // D5: the checkpoint persists through the writeState path —
    // .uf/progress/<sanitized-branch>/uf.unleash.json with 0o600 file mode.
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw, writeOptions] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
    )
    expect(writeOptions).toMatchObject({ mode: 0o600 })
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.command).toBe("uf.unleash")
    expect(state.branch).toBe("opsx-unleash-deterministic-tools")
    expect(state.key).toBe("opsx-unleash-deterministic-tools")
    expect(state.current_step).toBe("plan")
    expect(state.completed_steps).toEqual(["plan"])
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      expected_artifact: "specs/018-unleash-command/plan.md",
      selection_source: "host",
    })

    expect(mkdirSync).toHaveBeenCalledTimes(1)
    const [mkdirPathArg, mkdirOptions] = vi.mocked(mkdirSync).mock.calls[0]
    expect(mkdirPathArg).toBe(resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools"))
    expect(mkdirOptions).toMatchObject({ recursive: true, mode: 0o700 })
  })

  it("records implement checkpoint data with the phase for per-task delegation", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/unleash-deterministic-tools")

    // Per-task implement call shape (command Step 5): prompt-only, no
    // command_path or expected_artifact, with the 1-based phase number.
    const result = await invokeDelegate(exec, deps, {
      prompt: "Implement task 5.2 of the change.",
      task_classification: "judgment",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      checkpoint_step: "implement",
      phase: 3,
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.checkpoint_recorded).toBe(true)
    }
    // The per-task wrapper snapshots the worktree before the run (D3
    // before-state capture), so the status read precedes the checkpoint
    // branch read.
    expect(execCalls).toEqual([
      { args: ["status", "--porcelain", "-z"], cwd: FAKE_CWD },
      { args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD },
    ])

    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.current_step).toBe("implement")
    expect(state.completed_steps).toEqual(["implement"])
    // D5: "implement" with phase data — the per-task call merges the
    // phase into the checkpoint data alongside the delegation facts; the
    // declared classification is recorded and judgment calls add no skill
    // keys (task 1.5 audit trail).
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      phase: 3,
      task_classification: "judgment",
      selection_source: "host",
    })
  })

  it("records no phase key in checkpoint data when phase is omitted", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 5.2 of the change.",
      task_classification: "judgment",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // Existing checkpoint shapes are preserved: absent phase adds no key;
    // the declared classification is the one audit key a judgment call
    // records (task 1.5).
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      task_classification: "judgment",
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("phase")
  })

  it("appends checkpoint_step idempotently to existing checkpoint state", async () => {
    // Simulate a re-run: prior state exists with the branch gate recorded,
    // a later current step, and stale feature-dir data.
    const existing: ProgressState = {
      command: "uf.unleash",
      branch: "opsx-unleash-deterministic-tools",
      key: "opsx-unleash-deterministic-tools",
      current_step: "branch-safety-gate",
      completed_steps: ["branch-safety-gate"],
      data: { feature_dir: "specs/018-stale-dir/" },
      created_at: "2026-10-14T00:00:00.000Z",
      updated_at: "2026-10-14T00:00:00.000Z",
    }
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(existing))

    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Write the plan.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")

    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // The new step appends without duplicating earlier ones (FR-001)...
    expect(state.completed_steps).toEqual(["branch-safety-gate", "plan"])
    expect(state.current_step).toBe("plan")
    // ...shallow data merge adds this run's delegation facts while the stale
    // feature_dir survives (this run supplied none), and the judgment
    // classification is recorded (task 1.5)...
    expect(state.data).toEqual({
      feature_dir: "specs/018-stale-dir/",
      agent: "cobalt-crush-dev",
      task_classification: "judgment",
      selection_source: "host",
    })
    // ...and created_at survives while updated_at is refreshed.
    expect(state.created_at).toBe("2026-10-14T00:00:00.000Z")
    expect(state.updated_at).not.toBe("2026-10-14T00:00:00.000Z")
  })

  it("records no checkpoint when checkpoint_step is omitted", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.checkpoint_recorded).toBe(false)
    }
    // No state-key derivation and no state I/O happened.
    expect(execCalls).toEqual([])
    expect(lstatSync).not.toHaveBeenCalled()
    expect(readFileSync).not.toHaveBeenCalled()
    expect(mkdirSync).not.toHaveBeenCalled()
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("fails closed with DELEGATE_ERROR when the checkpoint cannot be written", async () => {
    vi.mocked(writeFileSync).mockImplementation(() => {
      throw new Error("EACCES: permission denied writing state file")
    })
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    // recordAutoCheckpoint throws on state I/O failure so a delegation never
    // reports success without a durable checkpoint (fail closed, D5).
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_ERROR")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("EACCES: permission denied writing state file")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — audit trail (delegation-skill-enforcement D5)
// ---------------------------------------------------------------------------

describe("uf_delegate_command audit trail", () => {
  it("records task_classification and the resolved skill reference in checkpoint data", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/delegation-skill-enforcement")

    // Per-task implement call shape: prompt-only, classified deterministic
    // with a skill reference, checkpointed with the phase number.
    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
      phase: 1,
    })

    expect(result.status).toBe("ok")
    // D3 before-state capture first, then the checkpoint branch read.
    expect(execCalls).toEqual([
      { args: ["status", "--porcelain", "-z"], cwd: FAKE_CWD },
      { args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD },
    ])
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const state = JSON.parse(stateRaw as string) as ProgressState
    // D5 (task 1.5): the checkpoint records the declared classification and
    // the RESOLVED reference (.opencode/skills/pre-flight/SKILL.md, not the
    // raw "pre-flight" input); a skill-carrying call records no reason.
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      phase: 1,
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("skill_absent_reason")
  })

  it("records task_classification and skill_absent_reason verbatim in checkpoint data", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const reason =
      'searched .opencode/skills/ and the repo root for a "rename" skill or script; none exists for this one-off'

    const result = await invokeDelegate(recordingExec("opsx/delegation-skill-enforcement").exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: reason,
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // D5 (task 1.5): the reason is recorded verbatim -- the checkpoint
    // store's JSON serialization is the sanitization boundary, no content
    // filtering -- and a reason-carrying call records no skill reference.
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      task_classification: "deterministic",
      skill_absent_reason: reason,
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("skill")
  })

  it("round-trips the audit fields through the progress store", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const commandFiles: Record<string, string> = { [commandAbs]: "Run the delegated command." }
    const reads = fakeReadText(commandFiles)
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      // DELEG-FR-001 (task 1.1): the footprint seam reads the same fake
      // files map (both runs share these deps), keeping the command-path
      // measurement deterministic.
      modelMatrix: { ...matrixFreeDeps, footprint: fakeFootprintSeam(commandFiles) },
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const exec = recordingExec("opsx/delegation-skill-enforcement").exec

    // Run 1: a deterministic delegation writes the audit fields.
    const first = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
    })
    expect(first.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const written = JSON.parse(stateRaw as string) as ProgressState
    expect(written.data).toMatchObject({
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
    })

    // Round-trip: the store's read path (the same readState get_progress
    // reads through) parses run 1's serialization...
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => stateRaw)

    // ...and run 2, a command-path call supplying no classification, merges
    // over the parsed state: run 1's audit fields survive the round-trip.
    const second = await invokeDelegate(exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })
    expect(second.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(2)
    expect(writeFileSync).toHaveBeenCalledTimes(2)
    const [, mergedRaw] = vi.mocked(writeFileSync).mock.calls[1]
    const merged = JSON.parse(mergedRaw as string) as ProgressState
    expect(merged.completed_steps).toEqual(["implement", "plan"])
    expect(merged.current_step).toBe("plan")
    expect(merged.data).toEqual({
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      agent: "cobalt-crush-dev",
      command_path: "cmd.md",
      selection_source: "host",
    })
  })

  it("surfaces the classification and resolved skill reference in the success envelope", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // D5 (task 1.5): machine-parseable envelope fields -- the declared
      // classification, the RESOLVED reference, and a null absence reason.
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBe(".opencode/skills/pre-flight/SKILL.md")
      expect(result.data.skill_absent_reason).toBeNull()
    }
  })

  it("surfaces the recorded absence reason in the success envelope", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: "no repo script exists for this one-off rename",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBe("no repo script exists for this one-off rename")
    }
  })

  it("surfaces judgment classification with no skill reference or absence reason in the success envelope", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Evaluate two design approaches for the parser.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // Spec: judgment-classified calls MUST NOT carry a skill reference or
      // absence reason -- the envelope reports the classification with
      // null skill fields.
      expect(result.data.task_classification).toBe("judgment")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBeNull()
    }
  })

  it("leaves checkpoint data and envelope audit fields absent for unclassified command-path calls", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const commandFiles: Record<string, string> = { [commandAbs]: "Run the delegated command." }
    const reads = fakeReadText(commandFiles)
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      // DELEG-FR-001 (task 1.1): the footprint seam reads the same fake
      // files map, so the estimator's clean measurement keeps the
      // conditional audit keys out of the pre-change checkpoint shape.
      modelMatrix: { ...matrixFreeDeps, footprint: fakeFootprintSeam(commandFiles) },
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })

    // Backward compatibility (D5): a call omitting every new field behaves
    // exactly as before this change -- the run proceeds, the envelope
    // surfaces the additive fields as null...
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.task_classification).toBeNull()
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBeNull()
    }
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // ...and the checkpoint data keeps the pre-change shape plus the one
    // always-present model-matrix audit field: selection_source records
    // "host" for every zero-config call (task 5.4); no other new keys are
    // added when the call supplied none.
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      command_path: "cmd.md",
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("task_classification")
    expect(state.data).not.toHaveProperty("skill")
    expect(state.data).not.toHaveProperty("skill_absent_reason")
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — seam verification (DE-007, delegation-skill-enforcement
// task 4.2)
//
// This change binds a prose procedure to a tool contract, so DE-007 requires
// the increment to verify the seams it creates. Layer 1 (registration
// surface: the registered args match the documented contract) lives in
// uf-workflow.test.ts and scratch-smoke.test.ts, which pin the registered
// argument set including task_classification, skill, and skill_absent_reason.
// The tests below cover the remaining two layers: the documented output
// shape (golden assertions on the DelegateCommandData envelope and the
// checkpoint fields) and write-then-read round-trips (checkpoint data a
// delegated run records, read back through the progress store's get_progress
// read surface — complementing the audit-trail round-trip test, which
// verifies the delegate's own readState merge path).
// ---------------------------------------------------------------------------

describe("uf_delegate_command seam verification", () => {
  it("golden output shape: deterministic delegation with a skill (valued audit fields in envelope and checkpoint)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/delegation-skill-enforcement")

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
      phase: 2,
    })

    // Layer 2 (DE-007): the documented DelegateCommandData output shape,
    // golden end to end — every field, exact values, no extras or missing
    // keys. The audit triple takes its valued shape: the declared
    // classification, the RESOLVED reference, a null absence reason.
    // (The judgment-shaped full golden lives in the expected-artifact
    // suite; this is the valued-shape counterpart.)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        agent: "cobalt-crush-dev",
        status: "success",
        child_session_id: "child-session",
        artifact_created: false,
        checkpoint_recorded: true,
        // SYNTH-001/006 (task 2.2): no envelope supplied on this call.
        human_decision_ids: null,
        task_classification: "deterministic",
        skill: ".opencode/skills/pre-flight/SKILL.md",
        skill_absent_reason: null,
        // Model-matrix-delegation D7 (task 5.4): zero-config selection
        // audit — host fallback, no resolved values.
        selection_source: "host",
        selection_pack: null,
        selection_capability: null,
        selection_model: null,
        selection_confidence: null,
        selection_fallback_reason: null,
        // DELEG-FR-001 (compaction-storm task 1.1): headroom provenance —
        // prompt-only footprint = cached prefix 56,000 + the in-memory
        // material (feature-dir line + skill directive + prompt, 409
        // bytes = 103 tokens at 4 bytes/token); no matrix, so the
        // documented default margin applies.
        selection_footprint: 56_103,
        selection_footprint_fallback: null,
        selection_margin_tokens: 24_000,
        selection_margin_warning: null,
        // DELEG-FR-002 (compaction-storm): storm provenance — no matrix,
        // so the documented default threshold applies with no warning.
        storm_threshold: DEFAULT_STORM_THRESHOLD,
        storm_threshold_warning: null,
        // PTB-FR-001 (per-task-delegation-budgets): the per-task
        // implement shape derives its wall clock. No tasks.md is served
        // here, so the phase task count fails closed to zero and the
        // 409-byte composed prompt adds no prompt term — the unclamped
        // 600_000 base window. Unbudgeted run, so no remaining-budget
        // report (PTB-FR-004).
        timeout_provenance: {
          effective_ms: 600_000,
          source: "derived",
          base_ms: 600_000,
          phase_task_count: 0,
          phase_task_term_ms: 0,
          phase_file_count: 0,
          phase_file_term_ms: 0,
          prompt_bytes: 409,
          prompt_term_ms: 0,
          tier: null,
          tier_term_ms: 0,
          contention_multiplier: 1,
          max_parallel_runs: 1,
          derived_ms: 600_000,
          ceiling_ms: 1_800_000,
          clamped: false,
        },
        remaining_ms: null,
        usage: null,
        provenance: {
          agent: "cobalt-crush-dev",
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: HOST_MODEL,
          resolved_parent_variant: "default",
          reported_child_model: HOST_MODEL,
          model_mismatch: false,
          read_only: false,
        },
        text: "delegated output",
      })
    }

    // Layer 2 (DE-007): the documented checkpoint shape — the full
    // ProgressState field set a delegated run persists (phase-scoped
    // records belong to the checkpoint tool's surface, not the
    // delegate's auto-checkpoint)...
    // D3 before-state capture first, then the checkpoint branch read.
    expect(execCalls).toEqual([
      { args: ["status", "--porcelain", "-z"], cwd: FAKE_CWD },
      { args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD },
    ])
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(Object.keys(state).sort()).toEqual([
      "branch",
      "command",
      "completed_steps",
      "created_at",
      "current_step",
      "data",
      "key",
      "updated_at",
    ])
    // ...with the audit fields recorded in the checkpoint data's valued
    // shape: classification plus resolved reference, no reason key. The
    // model-matrix audit adds selection_source only — zero-config calls
    // keep the pre-change checkpoint shape plus that one field (task 5.4).
    expect(state.current_step).toBe("implement")
    expect(state.completed_steps).toEqual(["implement"])
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      phase: 2,
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
  })

  it("golden output shape: deterministic delegation with a recorded absence reason (valued audit fields in envelope)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const reason =
      'searched .opencode/skills/ and the repo root for a rename skill or script; none exists for this one-off'

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: reason,
    })

    // Layer 2 (DE-007): the same documented envelope shape with the audit
    // triple's other valued form — a null skill, the verbatim absence
    // reason — and no checkpoint (checkpoint_step omitted), so the golden
    // pins that shape's documented output too.
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        agent: "cobalt-crush-dev",
        status: "success",
        child_session_id: "child-session",
        artifact_created: false,
        checkpoint_recorded: false,
        // SYNTH-001/006 (task 2.2): no envelope supplied on this call.
        human_decision_ids: null,
        task_classification: "deterministic",
        skill: null,
        skill_absent_reason: reason,
        // Model-matrix-delegation D7 (task 5.4): zero-config selection
        // audit — host fallback, no resolved values.
        selection_source: "host",
        selection_pack: null,
        selection_capability: null,
        selection_model: null,
        selection_confidence: null,
        selection_fallback_reason: null,
        // DELEG-FR-001 (compaction-storm task 1.1): headroom provenance —
        // prompt-only footprint = cached prefix 56,000 + "Rename the
        // variable across the module." (38 bytes = 10 tokens at 4
        // bytes/token); no matrix, so the documented default margin.
        selection_footprint: 56_010,
        selection_footprint_fallback: null,
        selection_margin_tokens: 24_000,
        selection_margin_warning: null,
        // DELEG-FR-002 (compaction-storm): storm provenance — no matrix,
        // so the documented default threshold applies with no warning.
        storm_threshold: DEFAULT_STORM_THRESHOLD,
        storm_threshold_warning: null,
        // PTB-FR-001/PTB-FR-004 (per-task-delegation-budgets): no
        // per-task shape, so no derivation provenance or remaining
        // budget.
        timeout_provenance: null,
        remaining_ms: null,
        usage: null,
        provenance: {
          agent: "cobalt-crush-dev",
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: HOST_MODEL,
          resolved_parent_variant: "default",
          reported_child_model: HOST_MODEL,
          model_mismatch: false,
          read_only: false,
        },
        text: "delegated output",
      })
    }
  })

  it("reads a delegated run's recorded audit fields back through the get_progress progress store", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec } = recordingExec("opsx/delegation-skill-enforcement")

    // Write: a delegated deterministic run auto-checkpoints the audit
    // fields through writeState's JSON serialization.
    const delegated = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
    })
    expect(delegated.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const written = JSON.parse(stateRaw as string) as ProgressState

    // Read: the progress store's dedicated read surface (get_progress)
    // resolves the same state key from the same branch and parses the
    // same bytes the delegated run wrote. The audit-trail round-trip
    // test above exercises the delegate's own readState merge path;
    // this is the store's read side of the seam.
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => stateRaw)

    const progress = createGetProgressTool(exec, FAKE_CWD)
    const readResult = await (
      progress as {
        execute: (
          args: Record<string, unknown>,
          ctx: { sessionID: string },
        ) => Promise<{ output: string }>
      }
    ).execute({ command: "uf.unleash" }, { sessionID: "parent-session" })
    const readback = JSON.parse(readResult.output) as ToolResult<ProgressState>

    expect(readback.status).toBe("ok")
    if (readback.status === "ok") {
      // Exactly what the delegated run wrote is what the store returns.
      expect(readback.data).toStrictEqual(written)
    }
    // The read went through the store's state path for the same key...
    expect(vi.mocked(readFileSync).mock.calls[0][0]).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    // ...and the audit fields survived the round-trip: the declared
    // classification, the RESOLVED reference, and the recorded step.
    expect(written.current_step).toBe("implement")
    expect(written.completed_steps).toEqual(["implement"])
    expect(written.data).toEqual({
      agent: "cobalt-crush-dev",
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — spec scenario coverage (delegation-skill-enforcement)
// ---------------------------------------------------------------------------

// End-to-end mirrors of the change's spec-delta scenarios. Tasks 1.1-1.5
// landed the rule-level tests (schema validation, failure envelopes,
// resolution, directive, audit fields); this block consolidates each spec
// scenario whose THEN clause those tests split across separate flows, and
// names the scenarios no earlier test covered: deterministic-with-skill in
// one resolved+injected+recorded flow, the quick-tier whole-change
// classification pass-through, and the structural no-model resolution
// failure. Scenarios already pinned end to end by a single named test
// (SKILL_REFERENCE_REQUIRED, SKILL_NOT_FOUND, judgment pass-through, the
// prompt-only classification requirement, over-length / empty-string /
// non-enum rejection, script-path skill, command-path exemption, and the
// no-variant spawn) are not duplicated here; the deterministic-with-reason
// scenario's checkpoint half lives in the audit-trail suite.
describe("uf_delegate_command spec scenario coverage", () => {
  it("deterministic delegation carries a skill end to end: resolved, injected, recorded, run proceeds", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // Spec scenario: a per-task delegation classified deterministic with
    // skill "pre-flight", which exists at .opencode/skills/pre-flight/SKILL.md.
    const result = await invokeDelegate(recordingExec("opsx/delegation-skill-enforcement").exec, deps, {
      prompt: "Run the repository's pre-flight procedure for this task.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
      phase: 4,
    })

    // The run proceeds: exactly one child session spawned and prompted.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)

    // Resolved: the reference resolved under the first skill root through
    // the injected existence probe.
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, ".opencode/skills/pre-flight/SKILL.md")])

    // Injected: the composed prompt carries the tool-rendered directive
    // naming the resolved skill in its fixed position — after feature-dir
    // context, immediately before the extra prompt.
    expect(deliveredPrompt(fake.calls)).toBe(
      "Feature directory: openspec/changes/delegation-skill-enforcement/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Run the repository's pre-flight procedure for this task.",
    )

    // Recorded: the checkpoint data and the success envelope carry the
    // classification and the resolved reference — never the raw input.
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.current_step).toBe("implement")
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      phase: 4,
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBe(".opencode/skills/pre-flight/SKILL.md")
      expect(result.data.skill_absent_reason).toBeNull()
    }
  })

  it("classifies the quick-tier whole-change implement delegation; an unclassified call fails closed", async () => {
    // Quick-tier call shape (command Step 5): one prompt-only delegation
    // for the whole routed change — no command file, no feature directory
    // (the quick tier has none), no phase — auto-checkpointing "implement"
    // on success. The declared answer's deterministic branch (skill or
    // reason) is enforced identically to the per-task shape and is covered
    // by the end-to-end test above; this scenario pins the quick-tier
    // shape itself.
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // A docs-authorship change is judgment work — no repo skill or script
    // performs it end-to-end — so the declaration is "judgment" and the
    // run proceeds with the classification recorded.
    const result = await invokeDelegate(recordingExec("quick/add-docs-section").exec, deps, {
      prompt: "Implement the routed change completely on the current branch: add the usage docs section.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("judgment")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBeNull()
    }
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(resolve(FAKE_CWD, ".uf/progress/quick-add-docs-section/uf.unleash.json"))
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.current_step).toBe("implement")
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      task_classification: "judgment",
      selection_source: "host",
    })

    // An unclassified quick-tier delegation fails closed like any other
    // prompt-only call — the tier earns no exemption from the rule.
    const unclassified = await invokeDelegate(recordingExec("quick/add-docs-section").exec, deps, {
      prompt: "Implement the routed change completely on the current branch: add the usage docs section.",
    })
    expect(unclassified.status).toBe("error")
    if (unclassified.status === "error") {
      expect(unclassified.code).toBe("INVALID_INPUT")
      expect(unclassified.retryable).toBe(false)
      expect(unclassified.message).toBe(
        "task_classification: prompt-only delegation (no command_path) must declare task_classification -- " +
          "classify the work as deterministic or judgment before delegating",
      )
    }
    // The rejected call spawned no child session and wrote no state.
    expect(fake.calls.prompts).toHaveLength(1)
    expect(writeFileSync).toHaveBeenCalledTimes(1)
  })

  it("fails with HOST_MODEL_UNAVAILABLE, non-retryable, when no model can be resolved from the parent session at all", async () => {
    // Spec scenario (carried forward from pipeline hardening): structural
    // host-model resolution failure. The parent assistant message reports
    // no model identity, so no child session can replay it — the sibling
    // tests cover the invalid-variant and transient-failure causes.
    const fake = fakeClient({ hostModelMissing: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      // Structural, not transient: non-retryable, naming the remediation.
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "current assistant model identity is unavailable. Remediation: run the command from a provider " +
          "that reports the active model identity; retrying will not fix this.",
      )
    }
    // Failed before the child spawn: host introspection ran, nothing else.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — path traversal
// ---------------------------------------------------------------------------

describe("uf_delegate_command path traversal", () => {
  it.each(["../secrets.md", "/etc/passwd"])(
    "rejects command_path %s escaping the repository root with COMMAND_PATH_TRAVERSAL",
    async (commandPath) => {
      const fake = fakeClient()
      const reads = fakeReadText({})
      const artifact = fakeFileExists([])
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        readText: reads.readText,
        fileExists: artifact.fileExists,
      }

      const result = await invokeDelegate(recordingExec().exec, deps, { command_path: commandPath })

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("COMMAND_PATH_TRAVERSAL")
        expect(result.retryable).toBe(false)
        expect(result.message).toBe(
          "Path traversal blocked: command_path resolves outside the repository root",
        )
      }
      // Rejected before any file read, existence probe, or session activity.
      expect(reads.calls).toEqual([])
      expect(artifact.calls).toEqual([])
      expect(fake.calls.messages).toHaveLength(0)
      expect(fake.calls.creates).toHaveLength(0)
    },
  )

  it("rejects expected_artifact escaping the repository root with EXPECTED_ARTIFACT_TRAVERSAL before spawning", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      expected_artifact: "../../outside/plan.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_TRAVERSAL")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Path traversal blocked: expected_artifact resolves outside the repository root",
      )
    }
    // Pre-resolved before spawning: no host introspection, no child session,
    // no artifact existence probe.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(artifact.calls).toEqual([])
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — timeout budgets (delegation-timeout-budgets)
// ---------------------------------------------------------------------------

describe("uf_delegate_command timeout budgets", () => {
  const baseDeps = (client: Client, files: Record<string, string> = {}): DelegateCommandDependencies => ({
    client,
    modelMatrix: matrixFreeDeps,
    readText: fakeReadText(files).readText,
    fileExists: fakeFileExists([]).fileExists,
  })

  it("rejects a prompt-only timeout above the executor bound before spawning", async () => {
    const fake = fakeClient()
    const result = await invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      timeout: 1_800_001,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.message).toContain("prompt-only delegation timeout must not exceed 1800000 ms")
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects a command-path timeout above the 4-hour ceiling and an idle_timeout below 60 s", async () => {
    const fake = fakeClient()
    const tooLong = await invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
      command_path: "cmd.md",
      timeout: 14_400_001,
    })
    const tooIdle = await invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
      command_path: "cmd.md",
      idle_timeout: 59_999,
    })

    expect(tooLong.status === "error" && tooLong.code).toBe("INVALID_INPUT")
    expect(tooIdle.status === "error" && tooIdle.code).toBe("INVALID_INPUT")
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("accepts a 4-hour command-path timeout", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const result = await invokeDelegate(recordingExec().exec, baseDeps(fake.client, { [commandAbs]: "Run." }), {
      command_path: "cmd.md",
      timeout: 14_400_000,
    })

    expect(result.status).toBe("ok")
  })

  it("threads timeout details and durable resume state into a command-path timeout", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/uf.unleash.md")
    const fake = fakeClient({ hangPrompt: true })
    const state: ProgressState = {
      command: "uf.unleash",
      branch: "opsx-unleash-deterministic-tools",
      key: "opsx-unleash-deterministic-tools",
      current_step: "plan",
      completed_steps: ["spec-review", "plan"],
      data: {},
      created_at: "2026-11-03T00:00:00.000Z",
      updated_at: "2026-11-03T00:00:00.000Z",
    }
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(state))

    const result = (await invokeDelegate(
      recordingExec().exec,
      baseDeps(fake.client, { [commandAbs]: "Run the pipeline." }),
      { command_path: ".opencode/commands/uf.unleash.md", timeout: 5 },
    )) as unknown as Record<string, unknown>

    expect(result.status).toBe("error")
    expect(result.code).toBe("DELEGATE_FAILED")
    expect(result.child_session_id).toBe("child-session")
    // Non-storm kinds carry a null storm identity (design D3), threaded
    // through the same pass-through as the storm kind.
    expect(result.timeout).toMatchObject({
      kind: "wall_clock",
      effective_ms: 5,
      idle_timeout_ms: 600_000,
      storm_session_id: null,
    })
    expect(result.resume).toEqual({
      command: "uf.unleash",
      key: "opsx-unleash-deterministic-tools",
      last_checkpoint_step: "plan",
      completed_steps: ["spec-review", "plan"],
    })
    // The resume read goes through the progress state store path.
    expect(String(vi.mocked(readFileSync).mock.calls[0]?.[0])).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
    )
  })

  it("reports null resume state when no checkpoint exists or the key cannot be derived", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient({ hangPrompt: true })
    const deps = baseDeps(fake.client, { [commandAbs]: "Run." })

    const fresh = (await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      timeout: 5,
    })) as unknown as Record<string, unknown>
    expect(fresh.resume).toMatchObject({ command: "cmd", last_checkpoint_step: null, completed_steps: [] })

    const detached = (await invokeDelegate(recordingExec("HEAD").exec, deps, {
      command_path: "cmd.md",
      timeout: 5,
    })) as unknown as Record<string, unknown>
    expect(detached.resume).toBeNull()
  })

  it("omits resume state for prompt-only timeouts", async () => {
    const fake = fakeClient({ hangPrompt: true })
    const result = (await invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      timeout: 5,
    })) as unknown as Record<string, unknown>

    expect(result.timeout).toMatchObject({ kind: "wall_clock" })
    expect("resume" in result).toBe(false)
    // Prompt-only without checkpoint_step "implement" + phase is not a
    // per-task delegation, so the timeout envelope carries no advisory
    // hint either (PTB-FR-003 routes only the per-task call shape).
    expect((result.timeout as Record<string, unknown>).resume).toBeUndefined()
  })

  it("aborts a stalled child on the default 10-minute idle timeout long before the wall clock", async () => {
    vi.useFakeTimers()
    try {
      const fake = fakeClient({ hangPrompt: true, stalledProbe: true })
      const pending = invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
        prompt: "Run the delegation.",
        task_classification: "judgment",
        timeout: 1_800_000,
      })
      await vi.advanceTimersByTimeAsync(700_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.message).toBe("idle_timeout: invoke_agent run made no progress for 600000 ms")
      expect(result.timeout).toMatchObject({
        kind: "idle",
        idle_timeout_ms: 600_000,
        child_active: false,
        pending_question: false,
      })
      expect(fake.calls.aborts).toHaveLength(1)
    } finally {
      vi.useRealTimers()
    }
  })

  it("threads resume state into an idle-timeout failure on a checkpointed command", async () => {
    vi.useFakeTimers()
    try {
      const commandAbs = resolve(FAKE_CWD, ".opencode/commands/uf.unleash.md")
      const fake = fakeClient({ hangPrompt: true, stalledProbe: true })
      const state: ProgressState = {
        command: "uf.unleash",
        branch: "opsx-unleash-deterministic-tools",
        key: "opsx-unleash-deterministic-tools",
        current_step: "plan",
        completed_steps: ["spec-review", "plan"],
        data: {},
        created_at: "2026-11-03T00:00:00.000Z",
        updated_at: "2026-11-03T00:00:00.000Z",
      }
      vi.mocked(lstatSync).mockImplementation(
        () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
      )
      vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(state))

      const pending = invokeDelegate(
        recordingExec().exec,
        baseDeps(fake.client, { [commandAbs]: "Run the pipeline." }),
        { command_path: ".opencode/commands/uf.unleash.md", timeout: 1_800_000 },
      )
      await vi.advanceTimersByTimeAsync(700_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.status).toBe("error")
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.child_session_id).toBe("child-session")
      expect(result.timeout).toMatchObject({ kind: "idle", idle_timeout_ms: 600_000 })
      expect(result.resume).toMatchObject({
        command: "uf.unleash",
        key: "opsx-unleash-deterministic-tools",
        last_checkpoint_step: "plan",
        completed_steps: ["spec-review", "plan"],
      })
    } finally {
      vi.useRealTimers()
    }
  })

  it("reports pending_question on the envelope when the stalled child holds a running question", async () => {
    vi.useFakeTimers()
    try {
      const fake = fakeClient({ hangPrompt: true, questionProbe: true })
      const pending = invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
        prompt: "Run the delegation.",
        task_classification: "judgment",
        timeout: 1_800_000,
      })
      // The running question tool call is a running part, so DT-FR-014
      // defers the idle abort until idle_timeout (600 s) plus the
      // IDLE_DEFERRAL_CAP_MS bound measured from the first probe's
      // fingerprint change (30 s + 600 s + 900 s = 1,530 s) -- still
      // before the 1,800 s wall clock.
      await vi.advanceTimersByTimeAsync(1_540_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.timeout).toMatchObject({
        kind: "idle",
        idle_timeout_ms: 600_000,
        pending_question: true,
        running_tool: "question",
      })
    } finally {
      vi.useRealTimers()
    }
  })

  it("rejects an idle_timeout above the 1,800,000 ms bound", async () => {
    const fake = fakeClient()
    const tooIdle = await invokeDelegate(recordingExec().exec, baseDeps(fake.client), {
      command_path: "cmd.md",
      idle_timeout: 1_800_001,
    })

    expect(tooIdle.status === "error" && tooIdle.code).toBe("INVALID_INPUT")
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("threads a 4-hour command-path timeout through to the executor wall clock", async () => {
    vi.useFakeTimers()
    try {
      const commandAbs = resolve(FAKE_CWD, "cmd.md")
      const fake = fakeClient({ hangPrompt: true, activeProbe: true })
      const pending = invokeDelegate(
        recordingExec().exec,
        baseDeps(fake.client, { [commandAbs]: "Run." }),
        { command_path: "cmd.md", timeout: 14_400_000 },
      )
      await vi.advanceTimersByTimeAsync(14_400_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.status).toBe("error")
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.timeout).toMatchObject({
        kind: "wall_clock",
        effective_ms: 14_400_000,
        requested_ms: 14_400_000,
        budget_clamped: false,
      })
    } finally {
      vi.useRealTimers()
    }
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — per-task delegation budgets
// (per-task-delegation-budgets PTB-FR-001..004)
// ---------------------------------------------------------------------------

// The per-task implement call shape (prompt-only, checkpoint_step
// "implement", positive phase, feature_dir) routes through the per-task
// wrapper, so these rows verify the tool's own surfaces: the schema bound
// retained for the per-task shape, the derivation wired from tasks.md and
// the measured prompt, the wall clock actually running the derived window,
// the advisory resume hint on timeout envelopes, and the budget registry's
// fail-fast and remaining-budget reporting. The wrapper's executor-level
// semantics are pinned in delegation-budget.test.ts.
describe("uf_delegate_command per-task delegation budgets (per-task-delegation-budgets)", () => {
  const featureDir = "openspec/changes/demo"
  const tasksAbs = resolve(FAKE_CWD, `${featureDir}/tasks.md`)

  const baseDeps = (client: Client, files: Record<string, string> = {}): DelegateCommandDependencies => ({
    client,
    modelMatrix: matrixFreeDeps,
    readText: fakeReadText(files).readText,
    fileExists: fakeFileExists([]).fileExists,
  })

  /** tasks.md before-state: phase 1 with four tasks, none complete. */
  const tasksBefore = [
    "## Phase 1: Build the feature",
    "",
    "- [ ] 1.1 First task",
    "- [ ] 1.2 Second task",
    "- [ ] 1.3 Third task",
    "- [ ] 1.4 Fourth task",
  ].join("\n")

  /** tasks.md after-state: task 1.1 newly complete. */
  const tasksAfter = tasksBefore.replace("- [ ] 1.1 First task", "- [x] 1.1 First task")

  // The budget registry is module-global and invokeDelegate runs every
  // delegation from the fixed "parent-session" id, so a registered budget
  // must be cleared after every test to keep it from leaking into the
  // rest of the file's delegations.
  afterEach(() => {
    unregisterDeadline("parent-session")
  })

  it("retains the prompt-only timeout bound for the per-task implement call shape", async () => {
    const fake = fakeClient()
    const result = await invokeDelegate(recordingExec().exec, baseDeps(fake.client, { [tasksAbs]: tasksBefore }), {
      prompt: "Implement task 1.1.",
      task_classification: "judgment",
      checkpoint_step: "implement",
      phase: 1,
      feature_dir: featureDir,
      timeout: 1_800_001,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.message).toContain("prompt-only delegation timeout must not exceed 1800000 ms")
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("accepts an explicit timeout at the retained ceiling and bypasses derivation (PTB-FR-001)", async () => {
    const fake = fakeClient()
    const result = await invokeDelegate(recordingExec().exec, baseDeps(fake.client, { [tasksAbs]: tasksBefore }), {
      prompt: "Implement task 1.1.",
      task_classification: "judgment",
      checkpoint_step: "implement",
      phase: 1,
      feature_dir: featureDir,
      timeout: 1_800_000,
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // The per-task call site always supplies a derivation, so only the
      // explicitTimeoutMs wiring makes the caller's argument win: the
      // served four-task phase would derive ~780_000 ms, but the explicit
      // request passes through as the bare explicit provenance — no
      // derivation term modifies it. The value also sits exactly at the
      // retained 1_800_000 prompt-only ceiling (AE-FR-010), the accepted
      // boundary complementing the above-ceiling rejection.
      expect(result.data.timeout_provenance).toEqual({ effective_ms: 1_800_000, source: "explicit" })
      // Unbudgeted success reports no remaining budget (PTB-FR-004).
      expect(result.data.remaining_ms).toBeNull()
    }
  })

  it("derives the per-task wall clock from the phase task count and prompt bytes (PTB-FR-001)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({ [tasksAbs]: tasksBefore })
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: fakeFileExists([]).fileExists,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 1.1.",
      task_classification: "judgment",
      checkpoint_step: "implement",
      phase: 1,
      feature_dir: featureDir,
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // Oracle: the same pure derivation over the observed inputs — four
      // phase-1 tasks parsed from the served tasks.md and the byte length
      // of the prompt the child actually received.
      const expected = derivePerTaskTimeoutMs({
        baseMs: DEFAULT_TIMEOUT_MILLISECONDS,
        phaseTaskCount: 4,
        phaseFileCount: 0,
        promptBytes: Buffer.byteLength(deliveredPrompt(fake.calls), "utf8"),
        tier: null,
        maxParallelRuns: 1,
        configuredCeilingMs: MAX_TIMEOUT_MILLISECONDS,
      })
      expect(result.data.timeout_provenance).toEqual(expected.provenance)
      // Unbudgeted success reports no remaining budget (PTB-FR-004).
      expect(result.data.remaining_ms).toBeNull()
    }
    // The derivation read the served tasks.md through the shared seam.
    expect(reads.calls).toContain(tasksAbs)
  })

  it("applies the derived window as the wall clock, not the flat default", async () => {
    vi.useFakeTimers()
    try {
      // Six phase-1 tasks: (6-1) x 60_000 = 300_000 ms of sibling
      // overhead, so the derived window is ~900_000 ms — past the
      // 600_000 flat default the wall clock must no longer use.
      const sixTasks = [
        "## Phase 1: Build the feature",
        "",
        "- [ ] 1.1 First task",
        "- [ ] 1.2 Second task",
        "- [ ] 1.3 Third task",
        "- [ ] 1.4 Fourth task",
        "- [ ] 1.5 Fifth task",
        "- [ ] 1.6 Sixth task",
      ].join("\n")
      const fake = fakeClient({ hangPrompt: true, activeProbe: true })
      const pending = invokeDelegate(recordingExec().exec, baseDeps(fake.client, { [tasksAbs]: sixTasks }), {
        prompt: "Implement task 1.6.",
        task_classification: "judgment",
        checkpoint_step: "implement",
        phase: 1,
        feature_dir: featureDir,
      })
      // Past the flat default: the derived window keeps the run alive.
      await vi.advanceTimersByTimeAsync(700_000)
      expect(fake.calls.aborts).toHaveLength(0)
      const expected = derivePerTaskTimeoutMs({
        baseMs: DEFAULT_TIMEOUT_MILLISECONDS,
        phaseTaskCount: 6,
        phaseFileCount: 0,
        promptBytes: Buffer.byteLength(deliveredPrompt(fake.calls), "utf8"),
        tier: null,
        maxParallelRuns: 1,
        configuredCeilingMs: MAX_TIMEOUT_MILLISECONDS,
      })
      await vi.advanceTimersByTimeAsync(expected.effectiveMs - 700_000 + 10_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.timeout).toMatchObject({
        kind: "wall_clock",
        requested_ms: expected.effectiveMs,
      })
      expect(result.timeout_provenance).toMatchObject({ source: "derived", phase_task_count: 6 })
    } finally {
      vi.useRealTimers()
    }
  })

  it("threads the resume hint into an idle-timeout per-task envelope (PTB-FR-003)", async () => {
    vi.useFakeTimers()
    try {
      const status = sequencedStatusExec(["", "M  lib/agent-executor.ts\0"])
      const fake = fakeClient({ hangPrompt: true, stalledProbe: true })
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        // Three tasks.md reads in order: the derivation read, the
        // delegation-start snapshot, and the abort-time snapshot.
        readText: sequencedTasksReadText(tasksAbs, [tasksBefore, tasksBefore, tasksAfter]),
        fileExists: fakeFileExists([]).fileExists,
      }
      const pending = invokeDelegate(status.exec, deps, {
        prompt: "Implement task 1.1.",
        task_classification: "judgment",
        checkpoint_step: "implement",
        phase: 1,
        feature_dir: featureDir,
      })
      // Four phase tasks derive a ~780_000 ms wall clock, so the 600_000
      // idle window fires first on the stalled child.
      await vi.advanceTimersByTimeAsync(700_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.timeout).toMatchObject({ kind: "idle", idle_timeout_ms: 600_000 })
      expect((result.timeout as Record<string, unknown>).resume).toEqual({
        files_modified: ["lib/agent-executor.ts"],
        tasks_completed: ["1.1"],
      })
      expect(result.timeout_provenance).toMatchObject({ source: "derived", phase_task_count: 4 })
    } finally {
      vi.useRealTimers()
    }
  })

  it("returns an empty resume hint when the timed-out child made no durable progress", async () => {
    vi.useFakeTimers()
    try {
      const status = sequencedStatusExec(["", ""])
      const fake = fakeClient({ hangPrompt: true, stalledProbe: true })
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        readText: sequencedTasksReadText(tasksAbs, [tasksBefore, tasksBefore, tasksBefore]),
        fileExists: fakeFileExists([]).fileExists,
      }
      const pending = invokeDelegate(status.exec, deps, {
        prompt: "Implement task 1.1.",
        task_classification: "judgment",
        checkpoint_step: "implement",
        phase: 1,
        feature_dir: featureDir,
      })
      await vi.advanceTimersByTimeAsync(700_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect((result.timeout as Record<string, unknown>).resume).toEqual({
        files_modified: [],
        tasks_completed: [],
      })
    } finally {
      vi.useRealTimers()
    }
  })

  it("returns a null resume hint when the abort-time state is unreadable", async () => {
    vi.useFakeTimers()
    try {
      const status = sequencedStatusExec(["", "M  lib/agent-executor.ts\0"])
      const fake = fakeClient({ hangPrompt: true, stalledProbe: true })
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        // Only two bodies: the abort-time read throws, so the hint fails
        // closed to null without masking the timeout it annotates.
        readText: sequencedTasksReadText(tasksAbs, [tasksBefore, tasksBefore]),
        fileExists: fakeFileExists([]).fileExists,
      }
      const pending = invokeDelegate(status.exec, deps, {
        prompt: "Implement task 1.1.",
        task_classification: "judgment",
        checkpoint_step: "implement",
        phase: 1,
        feature_dir: featureDir,
      })
      await vi.advanceTimersByTimeAsync(700_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect((result.timeout as Record<string, unknown>).resume).toBeNull()
    } finally {
      vi.useRealTimers()
    }
  })

  it("fails fast budget_exhausted without spawning under a spent ancestor budget (PTB-FR-002)", async () => {
    vi.useFakeTimers()
    try {
      // 20_000 ms left is under the 30_000 minimum once the 60_000
      // reserve applies, so the per-task wrapper refuses before spawning.
      registerDeadline("parent-session", Date.now() + 20_000)
      const status = sequencedStatusExec(["", ""])
      const fake = fakeClient()
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        readText: sequencedTasksReadText(tasksAbs, [tasksBefore, tasksBefore, tasksBefore]),
        fileExists: fakeFileExists([]).fileExists,
      }
      const result = (await invokeDelegate(status.exec, deps, {
        prompt: "Implement task 1.1.",
        task_classification: "judgment",
        checkpoint_step: "implement",
        phase: 1,
        feature_dir: featureDir,
      })) as unknown as Record<string, unknown>

      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(false)
      expect(String(result.message)).toContain("budget exhausted")
      expect(result.timeout).toMatchObject({ kind: "budget" })
      // The frozen fake clock makes the remaining-budget report exact.
      expect(result.remaining_ms).toBe(20_000)
      expect(result.timeout_provenance).toMatchObject({ source: "derived" })
      // Fail-fast happens after the delegation-start snapshot but before
      // any child session exists; the no-progress hint still reports.
      expect((result.timeout as Record<string, unknown>).resume).toEqual({
        files_modified: [],
        tasks_completed: [],
      })
      expect(fake.calls.creates).toHaveLength(0)
    } finally {
      vi.useRealTimers()
    }
  })

  it("reports the remaining budget on a budgeted success (PTB-FR-004)", async () => {
    vi.useFakeTimers()
    try {
      registerDeadline("parent-session", Date.now() + 500_000)
      const fake = fakeClient()
      const result = await invokeDelegate(recordingExec().exec, baseDeps(fake.client, { [tasksAbs]: tasksBefore }), {
        prompt: "Implement task 1.1.",
        task_classification: "judgment",
        checkpoint_step: "implement",
        phase: 1,
        feature_dir: featureDir,
      })

      expect(result.status).toBe("ok")
      if (result.status === "ok") {
        expect(result.data.timeout_provenance).toMatchObject({ source: "derived", phase_task_count: 4 })
        // The frozen clock pins the report to the registered deadline.
        expect(result.data.remaining_ms).toBe(500_000)
      }
    } finally {
      vi.useRealTimers()
    }
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — compaction-storm surfacing (compaction-storm
// task 3.4, DELEG-FR-001/002/003)
// ---------------------------------------------------------------------------

// The change's caller-facing surfaces through the delegation tool's own
// envelope and checkpoint store. The executor-level behavior — probe
// accounting, threshold boundaries, matcher structure — is pinned in
// agent-executor.test.ts; these rows verify uf_delegate_command surfaces
// each outcome to ITS caller: the headroom rejection chain riding
// selection_fallback_reason into the success envelope and the
// get_progress-readable checkpoint (DELEG-FR-001), the storm abort's
// context_thrash code with the storm timeout kind plus the command resume
// state (DELEG-FR-002), and the final-text guard's synthetic_final_text
// reclassification (DELEG-FR-003).
describe("uf_delegate_command compaction-storm surfacing", () => {
  it("surfaces the headroom rejection chain in the envelope and checkpoints it for the get_progress store", async () => {
    // A configured matrix where every gated candidate is tight: the
    // agents: pin and the plan default both declare models whose usable
    // window (context - output) cannot hold the measured footprint plus
    // the default 24k margin, so the chain falls through both and
    // terminates at host with every rejection recorded.
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const files: Record<string, string> = {
      [commandAbs]: "Run the pipeline.",
      [MATRIX_PROBE]: JSON.stringify({
        version: 1,
        profiles: {
          "pin-capability": { model: "tight-provider/pin-model" },
          "plan-capability": { model: "tight-provider/plan-model" },
        },
        defaults: { plan: "plan-capability" },
        agents: { "cobalt-crush-dev": "pin-capability" },
      }),
    }
    const fake = fakeClient()
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: {
        // plan resolves statically (chain step 5), so the classifier
        // cascade must stay dark for a plan-kind delegation.
        parseYaml: testParseYaml,
        classifierFetch: (() => {
          throw new Error("classifier fetch must not run — plan resolves statically")
        }) as typeof fetch,
        classifierReadEnv: () => undefined,
        spawnClassifierAgent: async () => {
          throw new Error("classifier spawn must not run — plan resolves statically")
        },
        // The estimator reads the same fake files so the command body
        // measures deterministically: 56k cached prefix + 5 tokens for
        // the 17-byte body ("Run the pipeline." / 4, rounded up).
        footprint: fakeFootprintSeam(files),
        // Declared limits make both candidates tight — usable windows of
        // 52k (pin) and 56k (plan default) against the required 80,005
        // (footprint 56,005 + default margin 24,000).
        modelLimits: async () => ({
          "tight-provider/pin-model": { context: 60_000, output: 8_000 },
          "tight-provider/plan-model": { context: 64_000, output: 8_000 },
        }),
      },
      readText: fakeReadText(files).readText,
      fileExists: artifact.fileExists,
    }
    const { exec } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })

    // Every gated candidate was rejected as tight, so the chain fell
    // through to host with both rejections recorded (DELEG-FR-001):
    // each entry names the rejected model, its usable window, and the
    // required footprint + margin; the decorate step closes with the
    // fallback destination.
    const chain =
      'headroom: agent-pin model "tight-provider/pin-model" usable 52000 < required 80005 ' +
      "(footprint 56005 + margin 24000); " +
      'headroom: default-kind model "tight-provider/plan-model" usable 56000 < required 80005 ' +
      "(footprint 56005 + margin 24000) — fell back to the host model"
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("host")
      expect(result.data.selection_model).toBeNull()
      expect(result.data.selection_fallback_reason).toBe(chain)
      // Headroom provenance riding the same envelope (task 1.1 audit).
      expect(result.data.selection_footprint).toBe(56_005)
      expect(result.data.selection_margin_tokens).toBe(24_000)
    }
    // Host fallback ran the child on the host model (the fake prompt
    // echoes the body model, so the mismatch guard stays quiet).
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.prompts[0]?.body).toMatchObject({
      model: { providerID: "host-provider", modelID: "host-model" },
    })

    // The auto-checkpoint persists the same chain through the state path
    // get_progress reads (the write side of the round-trip).
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
    )
    const written = JSON.parse(stateRaw as string) as ProgressState
    expect(written.current_step).toBe("plan")
    expect(written.completed_steps).toEqual(["plan"])
    expect(written.data).toEqual({
      agent: "cobalt-crush-dev",
      command_path: "cmd.md",
      selection_source: "host",
      selection_fallback_reason: chain,
    })

    // Read side: get_progress parses the same bytes and returns the
    // chain unchanged — the fallback reason survives the write→read
    // round-trip through the store.
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => stateRaw)
    const progress = createGetProgressTool(exec, FAKE_CWD)
    const readResult = await (
      progress as {
        execute: (
          args: Record<string, unknown>,
          ctx: { sessionID: string },
        ) => Promise<{ output: string }>
      }
    ).execute({ command: "uf.unleash" }, { sessionID: "parent-session" })
    const readback = JSON.parse(readResult.output) as ToolResult<ProgressState>
    expect(readback.status).toBe("ok")
    if (readback.status === "ok") {
      expect(readback.data).toStrictEqual(written)
    }
    expect(vi.mocked(readFileSync).mock.calls[0][0]).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
    )
  })

  it("surfaces a compaction-storm abort as DELEGATE_FAILED with the context_thrash code, storm kind, storming session, and resume state", async () => {
    vi.useFakeTimers()
    try {
      const commandAbs = resolve(FAKE_CWD, ".opencode/commands/uf.unleash.md")
      const fake = fakeClient({ hangPrompt: true, compactionProbe: true })
      const state: ProgressState = {
        command: "uf.unleash",
        branch: "opsx-unleash-deterministic-tools",
        key: "opsx-unleash-deterministic-tools",
        current_step: "plan",
        completed_steps: ["spec-review", "plan"],
        data: {},
        created_at: "2026-11-03T00:00:00.000Z",
        updated_at: "2026-11-03T00:00:00.000Z",
      }
      vi.mocked(lstatSync).mockImplementation(
        () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
      )
      vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(state))

      const deps: DelegateCommandDependencies = {
        client: fake.client,
        // Matrix-free selection keeps the row on the failure envelope:
        // the surface under test is the executor wiring, not model
        // selection (zero-config host default).
        modelMatrix: matrixFreeDeps,
        readText: fakeReadText({ [commandAbs]: "Run the pipeline." }).readText,
        fileExists: fakeFileExists([]).fileExists,
      }
      const pending = invokeDelegate(recordingExec().exec, deps, {
        command_path: ".opencode/commands/uf.unleash.md",
      })
      // Interval probes fire at the default 30 s poll; each observes a
      // new compaction marker, so the default storm threshold (3) is
      // reached at the 90 s tick — long before the 10-minute default
      // wall clock and idle timeout.
      await vi.advanceTimersByTimeAsync(100_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.status).toBe("error")
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.child_session_id).toBe("child-session")
      // The executor's failure code prefixes the message so the storm
      // classification is machine-parseable from the envelope, and the
      // message names the storming session — the probe root here — with
      // its consecutive count (design D3, DELEG-FR-002 "Storm envelope
      // names the storming session").
      expect(result.message).toBe(
        `context_thrash: invoke_agent run recorded ${DEFAULT_STORM_THRESHOLD} consecutive compaction events ` +
          `with no substantive progress in session child-session (storm threshold ${DEFAULT_STORM_THRESHOLD}); ` +
          "aborting so the caller can resume from durable state",
      )
      expect(result.timeout).toMatchObject({
        kind: "storm",
        storm_session_id: "child-session",
        effective_ms: 600_000,
        requested_ms: 600_000,
        budget_clamped: false,
        idle_timeout_ms: 600_000,
        elapsed_ms: 90_000,
        pending_question: false,
      })
      // Limit-ended command-path runs thread the durable resume state.
      expect(result.resume).toEqual({
        command: "uf.unleash",
        key: "opsx-unleash-deterministic-tools",
        last_checkpoint_step: "plan",
        completed_steps: ["spec-review", "plan"],
      })
      // The resume read went through the progress state store path.
      expect(String(vi.mocked(readFileSync).mock.calls[0]?.[0])).toBe(
        resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
      )
      // The storm aborted the wedged child exactly once.
      expect(fake.calls.aborts).toHaveLength(1)
    } finally {
      vi.useRealTimers()
    }
  })

  it("names the storming descendant session in the threaded envelope, distinct from the child session", async () => {
    // Issue #1125's recovery question — which session was thrashing? —
    // is answerable only when the envelope's storm identity is the
    // storming session itself, not merely the delegation's child. The
    // probe root stays quiet while an unmanaged descendant (no
    // registered deadline, so it is counted per session under
    // DELEG-FR-002) accrues the threshold of consecutive markers: the
    // code-prefixed message and the threaded timeout details must both
    // name that descendant, and the durable resume state threads
    // exactly as for a root storm (design D3).
    vi.useFakeTimers()
    try {
      const commandAbs = resolve(FAKE_CWD, ".opencode/commands/uf.unleash.md")
      const fake = fakeClient({ hangPrompt: true, descendantStormProbe: true })
      const state: ProgressState = {
        command: "uf.unleash",
        branch: "opsx-unleash-deterministic-tools",
        key: "opsx-unleash-deterministic-tools",
        current_step: "plan",
        completed_steps: ["spec-review", "plan"],
        data: {},
        created_at: "2026-11-03T00:00:00.000Z",
        updated_at: "2026-11-03T00:00:00.000Z",
      }
      vi.mocked(lstatSync).mockImplementation(
        () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
      )
      vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(state))

      const deps: DelegateCommandDependencies = {
        client: fake.client,
        // Matrix-free selection keeps the row on the failure envelope:
        // the surface under test is the executor wiring, not model
        // selection (zero-config host default).
        modelMatrix: matrixFreeDeps,
        readText: fakeReadText({ [commandAbs]: "Run the pipeline." }).readText,
        fileExists: fakeFileExists([]).fileExists,
      }
      const pending = invokeDelegate(recordingExec().exec, deps, {
        command_path: ".opencode/commands/uf.unleash.md",
      })
      // Each 30 s probe observes a fresh marker on the descendant only,
      // so the default storm threshold (3) is reached at the 90 s tick
      // with the root contributing nothing to the storm signal.
      await vi.advanceTimersByTimeAsync(100_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.status).toBe("error")
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.child_session_id).toBe("child-session")
      // The message names the storming descendant and its consecutive
      // count, not the delegation's child session.
      expect(result.message).toBe(
        `context_thrash: invoke_agent run recorded ${DEFAULT_STORM_THRESHOLD} consecutive compaction events ` +
          `with no substantive progress in session looper-session (storm threshold ${DEFAULT_STORM_THRESHOLD}); ` +
          "aborting so the caller can resume from durable state",
      )
      expect(result.timeout).toMatchObject({
        kind: "storm",
        storm_session_id: "looper-session",
        effective_ms: 600_000,
        requested_ms: 600_000,
        budget_clamped: false,
        idle_timeout_ms: 600_000,
        elapsed_ms: 90_000,
        pending_question: false,
      })
      // The threaded storm identity is the descendant — recovery can
      // re-probe exactly the thrashing session, distinguishing it from
      // the healthy child the delegation created.
      expect((result.timeout as Record<string, unknown>).storm_session_id).not.toBe(
        result.child_session_id,
      )
      // Resume state threading is unchanged by the named-session
      // envelope: the command-keyed durable state still rides the
      // failure.
      expect(result.resume).toEqual({
        command: "uf.unleash",
        key: "opsx-unleash-deterministic-tools",
        last_checkpoint_step: "plan",
        completed_steps: ["spec-review", "plan"],
      })
      expect(String(vi.mocked(readFileSync).mock.calls[0]?.[0])).toBe(
        resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
      )
      expect(fake.calls.aborts).toHaveLength(1)
    } finally {
      vi.useRealTimers()
    }
  })

  it("surfaces a reclassified final text as DELEGATE_FAILED with the synthetic_final_text code and no resume", async () => {
    // The child "finishes" with the compaction-summary wrapper as its
    // final text while the probe endpoints are absent, so the guard's
    // final probe cannot observe markers and the recorded window stays
    // below the storm threshold: the run is reclassified
    // synthetic_final_text (design D3, DELEG-FR-003). The command path
    // sharpens the no-resume assertion — no limit ended the run, so the
    // envelope carries neither timeout nor resume.
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const summaryText = "<system-reminder>\n## Objective\n- keep working\n</system-reminder>"
    const fake = fakeClient({ parts: [textPart(summaryText)] })
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: fakeReadText({ [commandAbs]: "Run the pipeline." }).readText,
      fileExists: fakeFileExists([]).fileExists,
    }
    const result = (await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
    })) as unknown as Record<string, unknown>

    expect(result.status).toBe("error")
    expect(result.code).toBe("DELEGATE_FAILED")
    expect(result.retryable).toBe(true)
    expect(result.child_session_id).toBe("child-session")
    expect(result.message).toContain("synthetic_final_text: ")
    expect(result.message).toContain("compaction-summary wrapper")
    expect(result.message).toContain(summaryText)
    // No limit ended the run: both keys are omitted — there is no storm
    // window to report and no durable state to thread, even on a
    // command-path delegation.
    expect(result.timeout).toBeUndefined()
    expect("resume" in result).toBe(false)
  })

  // Shared fixture for the storm-knob threading rows: the headroom
  // row's matrix shape (every gated candidate tight, so selection
  // terminates at host through static resolution with the classifier
  // dark) carrying a delegation.storm_threshold knob — the knob's
  // value is the only variable under test.
  const stormKnobMatrix = (threshold: number): string =>
    JSON.stringify({
      version: 1,
      profiles: {
        "pin-capability": { model: "tight-provider/pin-model" },
        "plan-capability": { model: "tight-provider/plan-model" },
      },
      defaults: { plan: "plan-capability" },
      agents: { "cobalt-crush-dev": "pin-capability" },
      delegation: { storm_threshold: threshold },
    })

  const stormKnobDeps = (
    fake: ReturnType<typeof fakeClient>,
    files: Record<string, string>,
  ): DelegateCommandDependencies => ({
    client: fake.client,
    modelMatrix: {
      parseYaml: testParseYaml,
      classifierFetch: (() => {
        throw new Error("classifier fetch must not run — plan resolves statically")
      }) as typeof fetch,
      classifierReadEnv: () => undefined,
      spawnClassifierAgent: async () => {
        throw new Error("classifier spawn must not run — plan resolves statically")
      },
      footprint: fakeFootprintSeam(files),
      modelLimits: async () => ({
        "tight-provider/pin-model": { context: 60_000, output: 8_000 },
        "tight-provider/plan-model": { context: 64_000, output: 8_000 },
      }),
    },
    readText: fakeReadText(files).readText,
    fileExists: fakeFileExists([]).fileExists,
  })

  it("threads a configured storm threshold into the audit surfaces (envelope and checkpoint)", async () => {
    // DELEG-FR-002 design D4: the knob must configure, not just validate.
    // A valid configured value reaches BOTH caller-facing audit
    // surfaces — the success envelope (always) and the auto-checkpoint
    // (the write side of the config-key round-trip seam).
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const files: Record<string, string> = {
      [commandAbs]: "Run the pipeline.",
      [MATRIX_PROBE]: stormKnobMatrix(2),
    }
    const fake = fakeClient()
    const deps = stormKnobDeps(fake, files)

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.storm_threshold).toBe(2)
      expect(result.data.storm_threshold_warning).toBeNull()
    }
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const written = JSON.parse(stateRaw as string) as ProgressState
    expect(written.data).toMatchObject({
      agent: "cobalt-crush-dev",
      command_path: "cmd.md",
      selection_source: "host",
      storm_threshold: 2,
    })
  })

  it("aborts at a configured storm threshold below the default", async () => {
    // The detector-side proof of threading: with the knob at 2, the
    // abort fires one 30 s probe poll earlier than the default-3 row —
    // only a threshold that actually reached the executor can produce
    // this timing and message.
    vi.useFakeTimers()
    try {
      const commandAbs = resolve(FAKE_CWD, ".opencode/commands/uf.unleash.md")
      const files: Record<string, string> = {
        [commandAbs]: "Run the pipeline.",
        [MATRIX_PROBE]: stormKnobMatrix(2),
      }
      const fake = fakeClient({ hangPrompt: true, compactionProbe: true })
      const state: ProgressState = {
        command: "uf.unleash",
        branch: "opsx-unleash-deterministic-tools",
        key: "opsx-unleash-deterministic-tools",
        current_step: "plan",
        completed_steps: ["spec-review", "plan"],
        data: {},
        created_at: "2026-11-03T00:00:00.000Z",
        updated_at: "2026-11-03T00:00:00.000Z",
      }
      vi.mocked(lstatSync).mockImplementation(
        () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
      )
      vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(state))

      const deps = stormKnobDeps(fake, files)
      const pending = invokeDelegate(recordingExec().exec, deps, {
        command_path: ".opencode/commands/uf.unleash.md",
      })
      await vi.advanceTimersByTimeAsync(70_000)
      const result = (await pending) as unknown as Record<string, unknown>

      expect(result.status).toBe("error")
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.message).toBe(
        "context_thrash: invoke_agent run recorded 2 consecutive compaction events " +
          "with no substantive progress in session child-session (storm threshold 2); " +
          "aborting so the caller can resume from durable state",
      )
      expect(result.timeout).toMatchObject({ kind: "storm", storm_session_id: "child-session", elapsed_ms: 60_000 })
      expect(fake.calls.aborts).toHaveLength(1)
    } finally {
      vi.useRealTimers()
    }
  })

  it("substitutes an invalid configured storm threshold with the default and surfaces the warning", async () => {
    // D4 substitution on the command path: 99 is outside the documented
    // 2-10 range, so the documented default applies and the warning
    // names the key and the rejected value — a bad value can never
    // disable the detector.
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const files: Record<string, string> = {
      [commandAbs]: "Run the pipeline.",
      [MATRIX_PROBE]: stormKnobMatrix(99),
    }
    const fake = fakeClient()
    const deps = stormKnobDeps(fake, files)

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.storm_threshold).toBe(DEFAULT_STORM_THRESHOLD)
      expect(result.data.storm_threshold_warning).toContain(
        "delegation.storm_threshold is invalid (99)",
      )
    }
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const written = JSON.parse(stateRaw as string) as ProgressState
    expect(written.data).toMatchObject({
      storm_threshold: DEFAULT_STORM_THRESHOLD,
      storm_threshold_warning: expect.stringContaining(
        "delegation.storm_threshold is invalid (99)",
      ),
    })
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — human-decisions relay envelope (SYNTH-001/006)
// ---------------------------------------------------------------------------

describe("uf_delegate_command human-decisions relay envelope", () => {
  it("persists relayed decisions at spawn time, before the child session runs, and returns the stamped ids", async () => {
    const order: string[] = []
    const fake = fakeClient({ onCreate: () => order.push("create") })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const stamped: readonly HumanDecisionEntry[] = [
      {
        decision_id: "11111111-2222-4333-8444-555555555555",
        source: "orchestrator-relay",
        question: "Proceed past the phase 3 quality gate?",
        answer: "Yes, proceed; the flaky test is tracked separately.",
        recorded_at: "2026-11-15T00:00:00.000Z",
      },
      {
        decision_id: "66666666-7777-4888-9aaa-bbbbbbbbbbbb",
        source: "orchestrator-relay",
        question: "Skip the slow integration profile for this run?",
        answer: "No, run the full profile.",
        recorded_at: "2026-11-15T00:00:00.000Z",
      },
    ]
    vi.mocked(appendHumanDecisions).mockImplementation(async () => {
      order.push("persist")
      return stamped
    })

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.2 of the change.",
      task_classification: "judgment",
      human_decisions: [
        {
          question: "Proceed past the phase 3 quality gate?",
          answer: "Yes, proceed; the flaky test is tracked separately.",
        },
        {
          question: "Skip the slow integration profile for this run?",
          answer: "No, run the full profile.",
        },
      ],
    })

    expect(result.status).toBe("ok")
    // The relay persists through the progress store under the sanitized
    // branch key, with source pinned structurally to the relay surface —
    // callers supply question and answer only (SYNTH-006).
    expect(appendHumanDecisions).toHaveBeenCalledTimes(1)
    expect(appendHumanDecisions).toHaveBeenCalledWith(
      FAKE_CWD,
      "opsx-unleash-deterministic-tools",
      [
        {
          source: "orchestrator-relay",
          question: "Proceed past the phase 3 quality gate?",
          answer: "Yes, proceed; the flaky test is tracked separately.",
        },
        {
          source: "orchestrator-relay",
          question: "Skip the slow integration profile for this run?",
          answer: "No, run the full profile.",
        },
      ],
    )
    // Spawn-time ordering (SYNTH-001): the record is written before the
    // delegated child session runs, so a child's human_decision_ref
    // citations can never race their own authorization into existence.
    expect(order).toEqual(["persist", "create"])
    // The stamped ids surface for the child's checkpoint citations.
    if (result.status === "ok") {
      expect(result.data.human_decision_ids).toEqual([
        "11111111-2222-4333-8444-555555555555",
        "66666666-7777-4888-9aaa-bbbbbbbbbbbb",
      ])
    }
  })

  it("returns human_decision_ids null and never touches the record when no envelope is supplied", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.3 of the change.",
      task_classification: "judgment",
    })

    // Pre-change call shape is unaffected: no envelope means no record
    // write and a null (not empty-array) ids field in the envelope.
    expect(result.status).toBe("ok")
    expect(appendHumanDecisions).not.toHaveBeenCalled()
    if (result.status === "ok") {
      expect(result.data.human_decision_ids).toBeNull()
    }
  })

  it("fails closed with HUMAN_DECISION_PERSIST_FAILED before spawning when the record append fails", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    vi.mocked(appendHumanDecisions).mockRejectedValue(
      new Error(
        "human-decision record for key opsx-unleash-deterministic-tools is corrupted (malformed JSON)",
      ),
    )

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.2 of the change.",
      task_classification: "judgment",
      human_decisions: [{ question: "Proceed?", answer: "Yes." }],
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HUMAN_DECISION_PERSIST_FAILED")
      expect(result.retryable).toBe(true)
      // Bounded identifiers only: the key and the store's error, plus the
      // remediation path — no raw question or answer text (SYNTH-006).
      expect(result.message).toContain("key opsx-unleash-deterministic-tools")
      expect(result.message).toContain("corrupted (malformed JSON)")
      expect(result.message).toContain("retry once the condition clears")
    }
    // The persist sits after every validation (host introspection ran, so
    // the failure is the persist, not an earlier gate) but before the
    // child spawn: no session created, prompted, or aborted, and no
    // checkpoint write — authorization that never reached the record does
    // not exist, so the delegation never runs.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("rejects an empty envelope array with INVALID_INPUT before any spawn or record write", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.2 of the change.",
      task_classification: "judgment",
      human_decisions: [],
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "human_decisions: empty array is invalid -- omit human_decisions when no decisions are relayed",
      )
    }
    // Schema bounds reject before any security-sensitive operation: no
    // host introspection, no session activity, no record write.
    expect(appendHumanDecisions).not.toHaveBeenCalled()
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects envelope entries carrying extraneous fields -- no provenance or identity smuggling", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // A caller cannot claim question-tool provenance (or stamp a
    // decision_id / recorded_at) through the relay: the entry object is
    // strict, so extraneous keys are a schema violation, not data.
    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.2 of the change.",
      task_classification: "judgment",
      human_decisions: [{ question: "Proceed?", answer: "Yes.", source: "question-tool" }],
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      // The offending entry and field are named (zod's strict-object
      // wording is version-specific; the path identity is the contract).
      expect(result.message).toContain("human_decisions.0")
      expect(result.message).toContain("source")
    }
    expect(appendHumanDecisions).not.toHaveBeenCalled()
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    [
      "an empty-string question",
      [{ question: "", answer: "A." }],
      "human_decisions.0.question: empty string is invalid -- relay the actual question text",
    ],
    [
      "an empty-string answer",
      [{ question: "Q?", answer: "" }],
      "human_decisions.0.answer: empty string is invalid -- relay the actual answer text",
    ],
    [
      "a question over the 16 KiB byte bound",
      [{ question: "x".repeat(16 * 1024 + 1), answer: "A." }],
      "human_decisions.0.question: question exceeds the 16 KiB UTF-8 byte limit (SYNTH-006)",
    ],
    [
      "an answer over the 16 KiB byte bound",
      [{ question: "Q?", answer: "x".repeat(16 * 1024 + 1) }],
      "human_decisions.0.answer: answer exceeds the 16 KiB UTF-8 byte limit (SYNTH-006)",
    ],
    [
      "more than 32 entries",
      Array.from({ length: 33 }, () => ({ question: "Q?", answer: "A." })),
      // zod's array-max wording is version-specific; the field identity
      // is the contract (the over-length skill precedent).
      "human_decisions",
    ],
  ] as const)("rejects %s with INVALID_INPUT before any spawn or record write", async (_label, envelope, expected) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.2 of the change.",
      task_classification: "judgment",
      human_decisions: envelope,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain(expected)
    }
    expect(appendHumanDecisions).not.toHaveBeenCalled()
    expect(fake.calls.creates).toHaveLength(0)
  })
})

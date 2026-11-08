/**
 * Shared Bun.spawn mock harness for the gh-CLI-backed review-dispatch
 * tool suites.
 *
 * Extracted from fetch-pr-review-context.test.ts's local harness per
 * extend-fetch-pr-review-context design D7 (the bun-runtime.ts
 * shared-extraction precedent; default.md CS-004 [MUST]): the phase-3
 * suites for fetch_pr_checks and fetch_failed_run_logs need the same
 * text-stream plumbing, gh-subcommand response routing, and setupTool
 * scratch-project registration that the fetch_pr_review_context suite
 * built privately. fetch-pr-review-context.test.ts imports it here; any
 * other local copies migrate opportunistically, not by force.
 *
 * MockResponse carries stdout, stderr, and exitCode as peers, so the
 * exit-code-tolerant execution path (runGhCommandTolerant, design D4 —
 * stdout JSON is data on ANY exit code, while gh still reports
 * zero-checks and not-found outcomes through stderr markers like "no
 * checks reported") is mockable exactly like the throw-on-non-zero
 * path: a route can pin exit 8 with a valid JSON array, exit 1 with
 * empty stdout plus marker stderr, or any other combination a suite
 * needs. Suites supply their own gh subcommand fixtures; only the
 * universal prologue (`which gh`, `gh auth status` — run by every
 * gh-backed tool before any fetch) carries harness-level defaults.
 *
 * Registry-parameterized cleanup (the aggregate-dispatch-fixtures.ts
 * precedent): every suite owns its scratch-directory registry, drained
 * in its own afterEach, so this harness keeps no module-global state.
 */

import { mkdir, mkdtemp, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { vi } from "vitest"

import ReviewDispatchPlugin from "../../plugins/review-dispatch/index.js"
import manifestFixture from "../fixtures/review-dispatch/manifest.json"
import matrixFixture from "../fixtures/review-dispatch/matrix.json"
import type { DispatchPluginHooks } from "./aggregate-dispatch-fixtures.js"

// ── Types ────────────────────────────────────────────────────

/** Mimics the subset of the Bun.spawn result the gh tools consume. */
export interface MockSpawnProcess {
  readonly stdout: ReadableStream<Uint8Array>
  readonly stderr: ReadableStream<Uint8Array>
  readonly exited: Promise<number>
}

/**
 * A canned gh process outcome: stream text plus the exit code.
 *
 * The three fields are peers — callers of the exit-code-tolerant path
 * (runGhCommandTolerant) read stdout as data regardless of exitCode,
 * so tests for that contract set both independently (e.g. exit 8 with
 * a pending-checks JSON array, or exit 1 with empty stdout and a "no
 * checks reported" stderr marker).
 */
export interface MockResponse {
  readonly stdout: string
  readonly stderr: string
  readonly exitCode: number
}

/**
 * Route keys the spawn router recognizes: the `which gh` and
 * `gh auth status` prologue plus the gh subcommands the
 * review-dispatch fetch tools invoke (`pr view`, `pr checks`,
 * `pr diff`, `repo view`, `run view`) and the `gh api` review/comment
 * endpoints.
 */
export type GhSpawnRoute =
  | "which"
  | "auth status"
  | "pr view"
  | "pr checks"
  | "pr diff"
  | "repo view"
  | "run view"
  | "api reviews"
  | "api comments"
  | "api user"
  | "api compare"

/** Per-route mock responses; suites supply fixtures, tests override. */
export type MockSpawnResponses = Partial<Record<GhSpawnRoute, MockResponse>>

/** A spawn implementation patchable onto globalThis.Bun. */
export type SpawnImplementation = (args: string[], options?: unknown) => MockSpawnProcess

/** The routed spawn mock returned by createMockSpawn (a vitest Mock). */
export type MockSpawnFn = ReturnType<typeof createMockSpawn>

// ── Stream and process builders ──────────────────────────────

/** Encode text as a single-chunk UTF-8 ReadableStream (the Bun.spawn stdout/stderr shape). */
export function textStream(text: string): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder()
  return new ReadableStream<Uint8Array>({
    start(controller) {
      controller.enqueue(encoder.encode(text))
      controller.close()
    },
  })
}

/** Build a fake Bun.spawn process emitting the canned response. */
export function mockProcess(response: MockResponse): MockSpawnProcess {
  return {
    stdout: textStream(response.stdout),
    stderr: textStream(response.stderr),
    exited: Promise.resolve(response.exitCode),
  }
}

// ── Routed spawn mock ────────────────────────────────────────

/**
 * Universal prologue defaults: every gh-backed tool runs `which gh`
 * and `gh auth status` before any fetch, so their success shapes are
 * harness state rather than per-suite fixtures. Tests override either
 * route to pin `gh_not_found` / `gh_auth_failed`.
 */
const UNIVERSAL_PROLOGUE_DEFAULTS: MockSpawnResponses = {
  which: { stdout: "/usr/local/bin/gh\n", stderr: "", exitCode: 0 },
  "auth status": { stdout: "Logged in\n", stderr: "", exitCode: 0 },
}

/** Fallback for unrouted invocations — surfaces as a failing gh call so unexpected spawns are loud in test output. */
const UNKNOWN_COMMAND_RESPONSE: MockResponse = { stdout: "", stderr: "unknown command", exitCode: 1 }

/**
 * Build a Bun.spawn mock that routes on the command's argument array.
 *
 * Routing: `which` checks match args[0]; gh commands route on the
 * first two args after "gh" (`auth status`, `pr view`, `pr checks`,
 * `pr diff`, `repo view`, `run view`); `gh api` calls route on the
 * endpoint suffix (`/reviews`, `/comments`). Unrouted invocations and
 * routes with no configured response fall back to a failing response.
 * Responses merge the universal prologue defaults, the suite's fixture
 * defaults, and the per-test overrides (last wins).
 *
 * @param suiteDefaults Suite fixture responses for the routes its tool fetches.
 * @param overrides Per-test responses, merged over the suite defaults.
 * @returns A vitest mock fn installed as Bun.spawn; assert on mock.calls for argv pinning.
 */
export function createMockSpawn(suiteDefaults: MockSpawnResponses = {}, overrides: MockSpawnResponses = {}): MockSpawnFn {
  const responses: MockSpawnResponses = { ...UNIVERSAL_PROLOGUE_DEFAULTS, ...suiteDefaults, ...overrides }
  const respond = (response: MockResponse | undefined): MockSpawnProcess =>
    mockProcess(response ?? UNKNOWN_COMMAND_RESPONSE)

  return vi.fn((args: string[], _options?: unknown): MockSpawnProcess => {
    // which gh
    if (args[0] === "which") {
      return respond(responses.which)
    }

    // gh commands — first two args after "gh" determine the route
    const sub = args.slice(1, 3).join(" ")

    if (sub === "auth status") return respond(responses["auth status"])
    if (sub === "pr view") return respond(responses["pr view"])
    if (sub === "pr checks") return respond(responses["pr checks"])
    if (sub === "pr diff") return respond(responses["pr diff"])
    if (sub === "repo view") return respond(responses["repo view"])
    if (sub === "run view") return respond(responses["run view"])

    // API calls for reviews and comments
    if (args[1] === "api") {
      const endpoint = args[2] ?? ""
      if (endpoint.endsWith("/reviews")) return respond(responses["api reviews"])
      if (endpoint.endsWith("/comments")) return respond(responses["api comments"])
      if (endpoint === "user") return respond(responses["api user"])
      if (endpoint.includes("/compare/")) return respond(responses["api compare"])
    }

    // Fallback — unknown command
    return mockProcess(UNKNOWN_COMMAND_RESPONSE)
  })
}

// ── Bun global patching ──────────────────────────────────────

/** Bun global subset the patch manipulates: YAML for server() registration, spawn for gh calls. */
type PatchableBunRuntime = {
  readonly YAML: { readonly parse: (text: string) => unknown }
  spawn: SpawnImplementation
}

/**
 * Patch globalThis.Bun with the given spawn implementation, preserving
 * any existing Bun.YAML (falling back to a JSON-backed parse) so
 * ReviewDispatchPlugin.server() registration — which constructs a
 * Bun-backed YAML parser for manifest dependencies — still succeeds
 * under the Node vitest environment.
 *
 * @param spawn The spawn implementation to install.
 * @returns A restore function reinstating the original global; invoke
 *   it in a finally block.
 */
export function patchBunSpawn(spawn: SpawnImplementation): () => void {
  const runtime = globalThis as typeof globalThis & { Bun?: PatchableBunRuntime }
  const originalBun = runtime.Bun
  runtime.Bun = {
    ...originalBun,
    YAML: originalBun?.YAML ?? { parse: (text: string): unknown => JSON.parse(text) as unknown },
    spawn,
  }
  return () => {
    runtime.Bun = originalBun
  }
}

// ── setupTool pattern ────────────────────────────────────────

/**
 * Write the .uf policy fixtures ReviewDispatchPlugin.server() reads
 * (review matrix, reviewer capabilities, sibling repos) into a scratch
 * directory.
 *
 * @param directory The scratch project directory.
 */
export async function writeReviewPolicyFiles(directory: string): Promise<void> {
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(join(directory, ".uf", "sibling-repos.yaml"), JSON.stringify({ version: 1, siblings: [] }))
}

/** Options for registering the plugin with a mocked Bun.spawn. */
export interface GhMockSetupOptions {
  /** Suite fixture responses, merged over the universal prologue defaults. */
  readonly spawnDefaults?: MockSpawnResponses
  /** Per-test responses, merged over the suite defaults (last wins). */
  readonly spawnOverrides?: MockSpawnResponses
  /**
   * The suite's scratch-directory registry — the harness pushes its
   * mkdtemp directory here, and the suite drains the registry in its
   * afterEach (the aggregate-dispatch-fixtures registry pattern; this
   * harness keeps no module-global state).
   */
  readonly cleanupRegistry: string[]
}

/** A registered review-dispatch plugin with a routed Bun.spawn mock installed. */
export interface GhMockedPluginHarness {
  /** Hooks from ReviewDispatchPlugin.server(); suites select their tool. */
  readonly hooks: DispatchPluginHooks
  /** The installed spawn mock; assert on mock.calls for argv pinning. */
  readonly mockSpawn: MockSpawnFn
  /** The scratch project directory holding the .uf policy fixtures. */
  readonly tmpDir: string
  /** Reinstates the original globalThis.Bun; invoke in a finally block. */
  restore(): void
}

/**
 * Set up ReviewDispatchPlugin.server() against a scratch project with
 * Bun.spawn mocked — the setupTool pattern the fetch_pr_review_context
 * suite established.
 *
 * Creates a mkdtemp scratch directory (tracked on the suite's cleanup
 * registry), writes the .uf policy fixtures registration reads, patches
 * globalThis.Bun with the routed spawn mock (preserving Bun.YAML), and
 * registers the plugin with directory and worktree pointed at the
 * scratch directory. The caller MUST invoke restore() when done (in a
 * finally block) and drain cleanupRegistry in an afterEach.
 *
 * @param options Suite defaults, per-test overrides, and the cleanup registry.
 * @returns The plugin hooks, the spawn mock, the scratch directory, and the restore function.
 */
export async function setupGhMockedPlugin(options: GhMockSetupOptions): Promise<GhMockedPluginHarness> {
  const tmpDir = await mkdtemp(join(tmpdir(), "uf-gh-mock-"))
  options.cleanupRegistry.push(tmpDir)

  await writeReviewPolicyFiles(tmpDir)

  const mockSpawn = createMockSpawn(options.spawnDefaults, options.spawnOverrides)
  const restore = patchBunSpawn(mockSpawn)

  const hooks = await ReviewDispatchPlugin.server({
    client: { session: {} },
    directory: tmpDir,
    worktree: tmpDir,
  } as unknown as PluginInput)

  return { hooks, mockSpawn, tmpDir, restore }
}

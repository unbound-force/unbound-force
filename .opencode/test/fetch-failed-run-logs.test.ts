import { mkdir, mkdtemp, readFile, rm, stat, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import { setupGhMockedPlugin, type MockSpawnFn, type MockSpawnResponses } from "./helpers/gh-spawn-mock.js"

// ── Fixtures ─────────────────────────────────────────────────

/** Default run id: a realistic 11-digit GitHub Actions run id. */
const RUN_ID = "40908825770"

/** Run id for the not-found family (verified gh 2.91.0 stderr shapes). */
const MISSING_RUN_ID = "99999999999"

/** Run id whose log path is pre-blocked by a directory (EISDIR test). */
const BLOCKED_RUN_ID = "98765432101"

/** Optional --repo anchor: the base repository PR checks report runs in. */
const RUN_REPO = "unbound-force/unbound-force"

/**
 * Multi-line CI failure-log stdout in the `gh run view --log-failed`
 * shape: timestamp-prefixed lines from a failed `make test` step, ending
 * with gh's trailing newline. Two deliberate properties:
 *
 * - The ✗ character is 3 UTF-8 bytes but 1 char, so byte length exceeds
 *   string length — pinning log_bytes as a byte count, not a char count.
 * - Nine log lines plus the trailing newline's empty split element give
 *   log_lines = 10 (the split("\n") formula counts the terminator).
 */
const MOCK_FAILED_LOGS = `2026-10-29T14:02:11.1234560Z ##[group]Run make test
2026-10-29T14:02:11.1234600Z [command]make test
2026-10-29T14:02:14.9876540Z --- FAIL: TestDivisorAdversary_Secrets ---
2026-10-29T14:02:14.9876600Z     adversary_test.go:212: ✗ hardcoded secret found in internal/gateway/proxy.go:88
2026-10-29T14:02:14.9876660Z --- FAIL: TestScaffoldDrift ---
2026-10-29T14:02:14.9876700Z     scaffold_test.go:54: embedded asset differs from canonical source
2026-10-29T14:02:15.0012000Z FAIL
2026-10-29T14:02:15.0012400Z exit status 1
2026-10-29T14:02:15.1040000Z ##[error]Process completed with exit code 1.
`

// ── Mock fixture responses (harness: ./helpers/gh-spawn-mock.ts) ──

/** Suite fixture responses for the route fetch_failed_run_logs owns; the which/auth prologue defaults live in the harness. */
const SPAWN_FIXTURES: MockSpawnResponses = {
  "run view": { stdout: MOCK_FAILED_LOGS, stderr: "", exitCode: 0 },
}

// ── Test infrastructure ──────────────────────────────────────

const scratchDirectories: string[] = []

/**
 * Sets up the plugin hooks with Bun.spawn mocked through the shared
 * gh-spawn-mock harness, selecting the fetch_failed_run_logs tool.
 */
async function setupTool(spawnOverrides: MockSpawnResponses = {}) {
  const { hooks, mockSpawn, tmpDir, restore } = await setupGhMockedPlugin({
    spawnDefaults: SPAWN_FIXTURES,
    spawnOverrides,
    cleanupRegistry: scratchDirectories,
  })
  return { tool: hooks.tool.fetch_failed_run_logs, mockSpawn, tmpDir, restore }
}

async function callTool(
  overrides: MockSpawnResponses = {},
  runId = RUN_ID,
  repo?: string,
): Promise<{ parsed: Record<string, unknown>; raw: string; mockSpawn: MockSpawnFn; tmpDir: string; restore: () => void }> {
  const { tool, mockSpawn, tmpDir, restore } = await setupTool(overrides)
  // repo is omitted entirely when absent — the unanchored spawn shape —
  // rather than passed as an explicit undefined.
  const input = repo === undefined ? { run_id: runId } : { run_id: runId, repo }
  const result = await tool.execute(input, {} as never)
  const parsed = JSON.parse(result.output) as Record<string, unknown>
  return { parsed, raw: result.output, mockSpawn, tmpDir, restore }
}

/**
 * Return the argv of the first recorded `gh <group> <verb>` spawn call
 * (e.g. group "run", verb "view"), or undefined when the tool never
 * issued that call.
 */
function findGhArgv(mockSpawn: MockSpawnFn, group: string, verb: string): string[] | undefined {
  const call = mockSpawn.mock.calls.find(
    (c: unknown[]) => (c[0] as string[])[1] === group && (c[0] as string[])[2] === verb,
  )
  return call === undefined ? undefined : (call[0] as string[])
}

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((dir) => rm(dir, { recursive: true, force: true })))
})

// ── Tests ────────────────────────────────────────────────────

describe("fetch_failed_run_logs", () => {
  describe("successful output", () => {
    it("returns run_id, log_path, log_bytes, log_lines, and empty warnings", async () => {
      const { parsed, restore } = await callTool()
      try {
        expect(parsed.error).toBeUndefined()
        // Field-for-field FR-014 output contract: exactly these keys, so
        // no log-content field can sneak into the output shape.
        expect(Object.keys(parsed).sort()).toEqual([
          "log_bytes",
          "log_lines",
          "log_path",
          "repo",
          "run_id",
          "warnings",
        ])
        expect(parsed.run_id).toBe(RUN_ID)
        // Unanchored fetch echoes the empty string, not undefined.
        expect(parsed.repo).toBe("")
        const logPath = parsed.log_path as string
        expect(logPath).toContain(`run-${RUN_ID}-failed-logs.txt`)
        expect(logPath).toContain(join(tmpdir(), "opencode"))
        // Byte count, not char count: the ✗ fixture character is 3 UTF-8
        // bytes but 1 char, so byteLength exceeds .length.
        const logBytes = parsed.log_bytes as number
        expect(logBytes).toBe(Buffer.byteLength(MOCK_FAILED_LOGS, "utf8"))
        expect(logBytes).toBeGreaterThan(MOCK_FAILED_LOGS.length)
        // The split formula: nine log lines plus the trailing newline's
        // empty split element.
        expect(parsed.log_lines).toBe(MOCK_FAILED_LOGS.split("\n").length)
        expect(parsed.log_lines).toBe(10)
        expect(parsed.warnings).toEqual([])
      } finally {
        restore()
      }
    })

    it("echoes the repo anchor when one is provided", async () => {
      const { parsed, restore } = await callTool({}, RUN_ID, RUN_REPO)
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.run_id).toBe(RUN_ID)
        expect(parsed.repo).toBe(RUN_REPO)
        expect(parsed.warnings).toEqual([])
      } finally {
        restore()
      }
    })

    it("never returns the log text inline", async () => {
      // Design D5: a big CI failure log can be megabytes, so the caller
      // gets the path plus counts. The exact-keys assertion above pins
      // the shape; this pins the values — no fixture text may appear in
      // the raw serialized output.
      const { raw, restore } = await callTool()
      try {
        expect(raw).not.toContain("hardcoded secret")
        expect(raw).not.toContain("adversary_test.go")
        expect(raw).not.toContain("--- FAIL")
      } finally {
        restore()
      }
    })
  })

  describe("temp file persistence", () => {
    it("writes the fetched stdout to the log file verbatim (write→read round-trip)", async () => {
      const { parsed, restore } = await callTool()
      try {
        const content = await readFile(parsed.log_path as string, "utf8")
        expect(content).toBe(MOCK_FAILED_LOGS)
      } finally {
        restore()
      }
    })

    it("creates the log file 0o600 under a 0o700 parent", async () => {
      // The real $TMPDIR/opencode usually pre-exists at 0o755 (the
      // opencode app itself creates it), and a recursive mkdir cannot
      // tighten an existing directory's mode — so this test points TMPDIR
      // at a fresh scratch directory (os.tmpdir() reads the env per call)
      // and lets the tool create the opencode/ child itself, where the
      // 0o700 mkdir mode actually applies.
      const sandbox = await mkdtemp(join(tmpdir(), "uf-ffrl-perms-"))
      scratchDirectories.push(sandbox)
      const originalTmpdir = process.env.TMPDIR
      process.env.TMPDIR = sandbox
      try {
        const { parsed, restore } = await callTool()
        try {
          const fileStats = await stat(parsed.log_path as string)
          // Owner-only read/write (0o600 = 0o100600 for regular files).
          expect(fileStats.mode & 0o777).toBe(0o600)
          const dirStats = await stat(join(tmpdir(), "opencode"))
          expect(dirStats.mode & 0o777).toBe(0o700)
        } finally {
          restore()
        }
      } finally {
        // process.env coerces an assigned undefined into the literal
        // string "undefined": on Linux, where TMPDIR normally starts
        // unset, a plain `process.env.TMPDIR = originalTmpdir` would
        // leave that poison value behind, make every later os.tmpdir()
        // in this process return the relative path "undefined", and
        // route fetch_failed_run_logs log files into
        // <cwd>/.opencode/undefined/ (issue #1051). delete restores
        // unset-ness; assignment restores a real saved value.
        if (originalTmpdir === undefined) {
          delete process.env.TMPDIR
        } else {
          process.env.TMPDIR = originalTmpdir
        }
      }
      // Regression guard (issue #1051): when TMPDIR started unset it
      // must be unset again after the test. The old plain-assignment
      // restore left the string "undefined" behind, which this
      // assertion catches — dormant on macOS (TMPDIR always set),
      // active on Linux where the regression manifests.
      if (originalTmpdir === undefined) {
        expect(process.env.TMPDIR).toBeUndefined()
      }
    })

    it("replaces a pre-existing stale file at the log path (unlink-then-O_EXCL)", async () => {
      // O_EXCL alone would fail with EEXIST on a stale file left by an
      // earlier fetch; the unlink-first step clears the path so the
      // exclusive create succeeds and the new fetch's content replaces
      // the old.
      const logPath = join(tmpdir(), "opencode", `run-${RUN_ID}-failed-logs.txt`)
      await mkdir(join(tmpdir(), "opencode"), { recursive: true })
      await writeFile(logPath, "STALE CONTENT FROM AN EARLIER FETCH\n", "utf8")

      const freshStdout = "fresh\nfailure-step\nlogs\n"
      const { parsed, restore } = await callTool({
        "run view": { stdout: freshStdout, stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        const content = await readFile(parsed.log_path as string, "utf8")
        expect(content).toBe(freshStdout)
        expect(content).not.toContain("STALE")
      } finally {
        restore()
      }
    })
  })

  describe("log file write failure", () => {
    it("returns api_error when the log path is blocked by a directory", async () => {
      // Pre-create a directory at the log path to force open("wx") to fail with EISDIR.
      const logPath = join(tmpdir(), "opencode", `run-${BLOCKED_RUN_ID}-failed-logs.txt`)
      await mkdir(logPath, { recursive: true })

      const { parsed, restore } = await callTool({}, BLOCKED_RUN_ID)
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.retryable).toBe(false)
      } finally {
        restore()
        await rm(logPath, { recursive: true, force: true }).catch(() => {})
      }
    })
  })

  describe("Bun.spawn invocation", () => {
    it("calls which gh as the first spawn", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const firstCall = mockSpawn.mock.calls[0]
        expect(firstCall[0]).toEqual(["which", "gh"])
      } finally {
        restore()
      }
    })

    it("calls gh auth status before gh run view", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const calls = mockSpawn.mock.calls.map((c: unknown[]) => (c[0] as string[]).slice(0, 3).join(" "))
        const authIndex = calls.indexOf("gh auth status")
        const runIndex = calls.findIndex((c: string) => c.startsWith("gh run view"))
        expect(authIndex).toBeGreaterThan(-1)
        expect(runIndex).toBeGreaterThan(authIndex)
      } finally {
        restore()
      }
    })

    it("invokes gh run view with the pinned argument vector when no repo is given", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const runArgs = findGhArgv(mockSpawn, "run", "view")
        expect(runArgs).toBeDefined()
        // No --repo flag at all: remote resolution must not depend on the
        // local checkout when the caller did not anchor it.
        expect(runArgs).toEqual(["gh", "run", "view", RUN_ID, "--log-failed"])
      } finally {
        restore()
      }
    })

    it("passes --repo as discrete spawn arguments when provided", async () => {
      // --repo and the anchor are separate array elements — no shell
      // interpolation of the repository slug (design D5).
      const { mockSpawn, restore } = await callTool({}, RUN_ID, RUN_REPO)
      try {
        const runArgs = findGhArgv(mockSpawn, "run", "view")
        expect(runArgs).toBeDefined()
        expect(runArgs).toEqual(["gh", "run", "view", RUN_ID, "--repo", RUN_REPO, "--log-failed"])
      } finally {
        restore()
      }
    })
  })

  describe("run_not_found (verified gh 2.91.0 shapes)", () => {
    it("classifies an HTTP 404 unknown run id as run_not_found", async () => {
      // Verified shape: an unknown run id exits 1 with "failed to get
      // run: HTTP 404: Not Found" plus the actions/runs API url on
      // stderr.
      const { parsed, restore } = await callTool(
        {
          "run view": {
            stdout: "",
            stderr: `failed to get run: HTTP 404: Not Found (https://api.github.com/repos/${RUN_REPO}/actions/runs/${MISSING_RUN_ID})`,
            exitCode: 1,
          },
        },
        MISSING_RUN_ID,
      )
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("run_not_found")
        expect(parsed.retryable).toBe(false)
        // The gh error text is surfaced verbatim, prefixed with the run id.
        expect(parsed.message as string).toContain(
          `Run ${MISSING_RUN_ID} not found or its logs are no longer available`,
        )
        expect(parsed.message as string).toContain("HTTP 404: Not Found")
      } finally {
        restore()
      }
    })

    it("classifies an HTTP 410 expired-logs run as run_not_found", async () => {
      // Verified shape: a run whose logs passed GitHub's retention window
      // exits 1 with "failed to get run log: HTTP 410: Server Error" plus
      // the runs/<id>/logs url — HTTP 410 Gone is the expired-logs signal.
      const { parsed, restore } = await callTool(
        {
          "run view": {
            stdout: "",
            stderr: `failed to get run log: HTTP 410: Server Error (https://api.github.com/repos/${RUN_REPO}/actions/runs/${MISSING_RUN_ID}/logs)`,
            exitCode: 1,
          },
        },
        MISSING_RUN_ID,
      )
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("run_not_found")
        expect(parsed.retryable).toBe(false)
        expect(parsed.message as string).toContain("HTTP 410: Server Error")
      } finally {
        restore()
      }
    })

    it("classifies generic 'could not find' phrasing as run_not_found", async () => {
      const { parsed, restore } = await callTool(
        {
          "run view": { stdout: "", stderr: `could not find any workflow run with id ${MISSING_RUN_ID}`, exitCode: 1 },
        },
        MISSING_RUN_ID,
      )
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("run_not_found")
        expect(parsed.retryable).toBe(false)
        expect(parsed.message as string).toContain("could not find")
      } finally {
        restore()
      }
    })
  })

  describe("zero-line fetch", () => {
    it("succeeds with log_lines 0, an empty log file, and the warning", async () => {
      // Verified shape: a run with no failed steps exits 0 with empty
      // output. The result stays a success with a warning so the
      // STOP-on-empty decision remains in command prose (/uf.fix-ci
      // Phase 2.1), and the empty fetch is special-cased to 0 lines —
      // split("") would report a phantom line.
      const { parsed, restore } = await callTool({
        "run view": { stdout: "", stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.run_id).toBe(RUN_ID)
        expect(parsed.log_lines).toBe(0)
        expect(parsed.log_bytes).toBe(0)
        expect(parsed.warnings).toEqual([`--log-failed returned no log lines for run ${RUN_ID}`])
        // The log file is still created — empty.
        const content = await readFile(parsed.log_path as string, "utf8")
        expect(content).toBe("")
      } finally {
        restore()
      }
    })
  })

  describe("error handling", () => {
    it("returns gh_not_found when the gh CLI is not available", async () => {
      const { parsed, restore } = await callTool({
        which: { stdout: "", stderr: "which: no gh in PATH", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("gh_not_found")
        expect(parsed.retryable).toBe(false)
        expect(typeof parsed.message).toBe("string")
      } finally {
        restore()
      }
    })

    it("returns gh_auth_failed when gh auth status fails", async () => {
      const { parsed, restore } = await callTool({
        "auth status": { stdout: "", stderr: "not logged in", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("gh_auth_failed")
        expect(parsed.retryable).toBe(false)
        expect(parsed.message as string).toContain("not logged in")
      } finally {
        restore()
      }
    })

    it("returns timeout when gh reports a timeout", async () => {
      const { parsed, restore } = await callTool({
        "run view": { stdout: "", stderr: "request timed out", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("timeout")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("returns rate_limited when GitHub rate limits", async () => {
      const { parsed, restore } = await callTool({
        "run view": { stdout: "", stderr: "API rate limit exceeded", exitCode: 4 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("rate_limited")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("classifies an HTTP 403 permission failure as api_error, not run_not_found", async () => {
      // The HTTP status codes discriminate the not-found family (404/410)
      // from permission failures (403), which classify through
      // classifyError instead (see RUN_NOT_FOUND_PATTERN).
      const { parsed, restore } = await callTool({
        "run view": { stdout: "", stderr: "failed to get run: HTTP 403: Forbidden", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.retryable).toBe(false)
        expect(parsed.message as string).toContain("HTTP 403")
      } finally {
        restore()
      }
    })
  })

  describe("input validation", () => {
    it("rejects non-numeric run_id via schema validation", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "abc" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.message as string).toContain("run_id")
        // The schema rejects the value before any gh invocation.
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a command-injection payload in run_id", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "40908825770; rm -rf /" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a path-traversal payload in run_id", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "../etc/passwd" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a shell-chain payload in run_id", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "42 && echo pwned" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects an empty-string run_id", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a non-string run_id (type confusion)", async () => {
      // The schema demands a digit STRING: a numeric 42 must not slip
      // through z.string() into the spawn argv or the temp-file name.
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: 42 } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.message as string).toContain("run_id")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a repo containing spaces via schema validation", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "42", repo: "not a repo" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.message as string).toContain("repo")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a repo URL via schema validation", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "42", repo: "http://evil/x" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects a repo with an empty repository segment", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "42", repo: "owner/" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects extra fields via .strict() schema", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ run_id: "42", evil: "payload" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("accepts valid edge-case run_id=0", async () => {
      // "0" satisfies ^[0-9]+$ — unlike fetch_pr_checks's positive-integer
      // pr_number, a zero run id is syntactically valid input. The happy
      // path's default RUN_ID covers the realistic 11-digit shape.
      const { parsed, restore } = await callTool({}, "0")
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.run_id).toBe("0")
        expect(parsed.log_lines).toBe(MOCK_FAILED_LOGS.split("\n").length)
        expect(parsed.log_path as string).toContain("run-0-failed-logs.txt")
      } finally {
        restore()
      }
    })
  })
})

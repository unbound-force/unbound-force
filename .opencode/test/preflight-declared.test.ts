import { describe, expect, it } from "vitest"

import {
  DEFAULT_CHECK_TIMEOUT_SECONDS,
  OUTPUT_TAIL_BYTES,
  PreflightConfigError,
  computeDiffScope,
  findOnPath,
  isInScope,
  parsePreflightConfig,
  renderDeclaredRows,
  runDeclaredChecks,
  tailOutput,
  type PreflightDeclaredDependencies,
  type ProcessResult,
} from "../lib/preflight-declared.js"
import type { GitExec } from "../lib/uf-workflow-types.js"

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const ROOT = "/repo"
const BASE_SHA = "a".repeat(40)
const HEAD_SHA = "b".repeat(40)
const MERGE_BASE = "c".repeat(40)

/** Git fake keyed by the joined argv; unknown keys throw. */
function gitFake(responses: Readonly<Record<string, string>>): GitExec {
  return (args) => {
    const key = args.join(" ")
    const out = responses[key]
    if (out === undefined) {
      throw new Error(`unexpected git ${key}`)
    }
    return out
  }
}

/** Git responses for a feature branch whose diff is `files`. */
function branchGit(files: readonly string[]): GitExec {
  return gitFake({
    "rev-parse --verify upstream/main": "",
    "rev-parse --verify origin/main": BASE_SHA,
    "rev-parse --verify HEAD": HEAD_SHA,
    [`merge-base --is-ancestor ${BASE_SHA} HEAD`]: "",
    "rev-parse --is-shallow-repository": "false",
    [`merge-base ${BASE_SHA} HEAD`]: MERGE_BASE,
    [`diff --name-only ${MERGE_BASE}`]: files.join("\n"),
  })
}

interface FakeOptions {
  readonly config?: unknown
  readonly executables?: readonly string[]
  readonly git?: GitExec
  readonly results?: Readonly<Record<string, ProcessResult>>
  /** Results for runs whose cwd is the baseline worktree, keyed by file. */
  readonly baseline?: Readonly<Record<string, ProcessResult>>
}

const WORKTREE = "/tmp/uf-preflight-baseline-x"

/** Build deps plus a log of process invocations. */
function fakeDeps(options: FakeOptions) {
  const calls: Array<{ file: string; args: readonly string[]; cwd: string; timeoutMs: number }> = []
  const envs: Array<Readonly<Record<string, string>>> = []
  const removed: string[] = []
  const executables = new Set(options.executables ?? [])
  const deps: PreflightDeclaredDependencies = {
    readText: async () => JSON.stringify(options.config),
    fileExists: (path) => path === `${ROOT}/.uf/preflight.yaml` && options.config !== undefined,
    isExecutable: (path) => executables.has(path),
    parseYaml: (text) => JSON.parse(text) as unknown,
    pathEnv: "/usr/bin:/opt/tools/",
    git: options.git ?? branchGit([]),
    runProcess: async (file, args, cwd, timeoutMs, env) => {
      calls.push({ file, args, cwd, timeoutMs })
      envs.push(env)
      const table = cwd === WORKTREE ? options.baseline : options.results
      return table?.[file] ?? { exitCode: 0, output: "ok", timedOut: false }
    },
    makeTempDir: async () => WORKTREE,
    removeDir: async (path) => {
      removed.push(path)
    },
  }
  return { deps, calls, envs, removed }
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

describe("parsePreflightConfig", () => {
  it("applies defaults to a minimal entry", () => {
    const config = parsePreflightConfig({ version: 1, checks: [{ name: "lint-sh", run: "scripts/lint.sh" }] })
    expect(config.checks).toEqual([
      {
        name: "lint-sh",
        run: "scripts/lint.sh",
        args: [],
        requires: [],
        paths: [],
        gate: "required",
        timeout_seconds: DEFAULT_CHECK_TIMEOUT_SECONDS,
      },
    ])
  })

  it("accepts the empty seeded template (checks: [] and null)", () => {
    expect(parsePreflightConfig({ version: 1, checks: [] }).checks).toEqual([])
    expect(parsePreflightConfig({ version: 1, checks: null }).checks).toEqual([])
  })

  it.each([
    ["absolute run", { name: "a", run: "/usr/bin/true" }],
    ["parent traversal", { name: "a", run: "../outside.sh" }],
    ["whitespace", { name: "a", run: "scripts/a.sh --flag" }],
    ["semicolon", { name: "a", run: "scripts/a.sh;rm" }],
    ["command substitution", { name: "a", run: "$(evil)" }],
    ["pipe", { name: "a", run: "a|b" }],
    ["bad name", { name: "Bad Name", run: "a.sh" }],
    ["bad gate", { name: "a", run: "a.sh", gate: "hard" }],
    ["unknown key", { name: "a", run: "a.sh", shell: "bash" }],
    ["requires with path", { name: "a", run: "a.sh", requires: ["/bin/sh"] }],
    ["zero timeout", { name: "a", run: "a.sh", timeout_seconds: 0 }],
    ["negative timeout", { name: "a", run: "a.sh", timeout_seconds: -1 }],
    ["timeout above max", { name: "a", run: "a.sh", timeout_seconds: 3601 }],
    ["non-integer timeout", { name: "a", run: "a.sh", timeout_seconds: 1.5 }],
  ])("rejects %s", (_label, entry) => {
    expect(() => parsePreflightConfig({ version: 1, checks: [entry] })).toThrow(PreflightConfigError)
  })

  it("rejects wrong version and duplicate names, listing every issue", () => {
    expect(() => parsePreflightConfig({ version: 2, checks: [] })).toThrow(PreflightConfigError)
    try {
      parsePreflightConfig({
        version: 1,
        checks: [
          { name: "dup", run: "a.sh" },
          { name: "dup", run: "b.sh" },
        ],
      })
      expect.unreachable()
    } catch (error) {
      expect((error as PreflightConfigError).issues).toEqual(['checks.1.name: duplicate check name "dup"'])
    }
  })
})

// ---------------------------------------------------------------------------
// Pure helpers
// ---------------------------------------------------------------------------

describe("isInScope", () => {
  it("treats empty paths or an unfiltered diff as in scope", () => {
    expect(isInScope([], ["x.go"])).toBe(true)
    expect(isInScope(["scripts/"], null)).toBe(true)
  })

  it("matches directory prefixes and suffixes (with or without a leading *)", () => {
    expect(isInScope(["scripts/"], ["scripts/lib/a.sh"])).toBe(true)
    expect(isInScope(["scripts/"], ["docs/scripts/a.md"])).toBe(false)
    expect(isInScope([".sh"], ["eval/run.sh"])).toBe(true)
    expect(isInScope(["*.sh"], ["eval/run.sh"])).toBe(true)
    expect(isInScope(["*.sh", "harness/"], ["README.md"])).toBe(false)
  })
})

describe("findOnPath", () => {
  it("resolves bare names against PATH entries, tolerating trailing slashes", () => {
    const exec = new Set(["/opt/tools/yq"])
    expect(findOnPath("yq", "/usr/bin:/opt/tools/", (p) => exec.has(p))).toBe(true)
    expect(findOnPath("jq", "/usr/bin:/opt/tools/", (p) => exec.has(p))).toBe(false)
    expect(findOnPath("yq", "", (p) => exec.has(p))).toBe(false)
  })
})

describe("computeDiffScope", () => {
  it("returns the sorted branch diff", () => {
    expect(computeDiffScope(branchGit(["z.sh", "a.sh"]), ROOT)).toEqual({
      files: ["a.sh", "z.sh"],
      fallback_reason: null,
      base_ref: "origin/main",
      merge_base: MERGE_BASE,
    })
  })

  it("falls back when no base ref resolves", () => {
    expect(computeDiffScope(gitFake({}), ROOT).files).toBeNull()
  })

  it("returns an empty, non-fallback diff on the base branch itself (merge-base == HEAD)", () => {
    const onBase = gitFake({
      "rev-parse --verify upstream/main": "",
      "rev-parse --verify origin/main": HEAD_SHA,
      "rev-parse --verify HEAD": HEAD_SHA,
      [`merge-base --is-ancestor ${HEAD_SHA} HEAD`]: "",
      "rev-parse --is-shallow-repository": "false",
      [`merge-base ${HEAD_SHA} HEAD`]: HEAD_SHA,
      [`diff --name-only ${HEAD_SHA}`]: "",
    })
    expect(computeDiffScope(onBase, ROOT)).toEqual({ files: [], fallback_reason: null, base_ref: "origin/main", merge_base: HEAD_SHA })
  })

  it("falls back when git throws mid-diff (merge-base or diff fails)", () => {
    const gitFailure = gitFake({
      "rev-parse --verify upstream/main": "",
      "rev-parse --verify origin/main": BASE_SHA,
      "rev-parse --verify HEAD": HEAD_SHA,
      [`merge-base --is-ancestor ${BASE_SHA} HEAD`]: "",
    })
    expect(computeDiffScope(gitFailure, ROOT)).toEqual({
      files: null,
      fallback_reason: "git diff failed",
      base_ref: "origin/main",
      merge_base: null,
    })
  })

  it("falls back on a diverged base, a shallow clone, and an empty feature-branch diff", () => {
    const diverged = gitFake({ "rev-parse --verify upstream/main": BASE_SHA, "rev-parse --verify HEAD": HEAD_SHA })
    expect(computeDiffScope(diverged, ROOT).fallback_reason).toMatch(/not an ancestor/)

    const shallow = gitFake({
      "rev-parse --verify upstream/main": BASE_SHA,
      "rev-parse --verify HEAD": HEAD_SHA,
      [`merge-base --is-ancestor ${BASE_SHA} HEAD`]: "",
      [`merge-base ${BASE_SHA} HEAD`]: MERGE_BASE,
      "rev-parse --is-shallow-repository": "true",
    })
    expect(computeDiffScope(shallow, ROOT).fallback_reason).toBe("shallow clone")

    expect(computeDiffScope(branchGit([]), ROOT).fallback_reason).toBe("empty diff on a feature branch")
  })
})

describe("tailOutput and renderDeclaredRows", () => {
  it("keeps short output and truncates long output to the tail", () => {
    expect(tailOutput("short")).toBe("short")
    const long = "x".repeat(OUTPUT_TAIL_BYTES + 10) + "END"
    const tail = tailOutput(long)
    expect(tail.startsWith("[… 13 bytes truncated]\n")).toBe(true)
    expect(tail.endsWith("END")).toBe(true)
  })

  it("renders one escaped table row per check", () => {
    expect(
      renderDeclaredRows([
        { name: "a", gate: "required", command: "x.sh --k a|b", status: "FAIL", exit_code: 1, timed_out: false, unavailable: [], output: "", causality: "pre-existing" },
        {
          name: "b",
          gate: "advisory",
          command: "y.sh",
          status: "SKIP — not available",
          exit_code: null,
          timed_out: false,
          unavailable: ["ripwire"],
          output: "",
          causality: null,
        },
      ], "soft-gate"),
    ).toBe("| a | `x.sh --k a\\|b` | 1 | FAIL | pre-existing |\n| b | `y.sh` | — | SKIP — not available: ripwire | — |")
  })

  it("drops the causality column outside soft-gate", () => {
    const row = { name: "a", gate: "required", command: "x.sh", status: "FAIL", exit_code: 1, timed_out: false, unavailable: [], output: "", causality: null } as const
    expect(renderDeclaredRows([row], "hard-gate")).toBe("| a | `x.sh` | 1 | FAIL |")
    expect(renderDeclaredRows([row], "ci-aware")).toBe("| a | `x.sh` | 1 | FAIL |")
  })

  it("keeps backticks inside the command code span", () => {
    const row = (command: string) =>
      ({ name: "a", gate: "required", command, status: "PASS", exit_code: 0, timed_out: false, unavailable: [], output: "", causality: null }) as const
    expect(renderDeclaredRows([row("x.sh a`b")], "hard-gate")).toBe("| a | ``x.sh a`b`` | 0 | PASS |")
    expect(renderDeclaredRows([row("x.sh ``c`")], "hard-gate")).toBe("| a | ``` x.sh ``c` ``` | 0 | PASS |")
  })
})

// ---------------------------------------------------------------------------
// runDeclaredChecks — the whole Phase 2b chain
// ---------------------------------------------------------------------------

describe("runDeclaredChecks", () => {
  it("reports config_present=false and runs nothing when the file is absent", async () => {
    const { deps, calls } = fakeDeps({})
    const data = await runDeclaredChecks(ROOT, deps)
    expect(data.config_present).toBe(false)
    expect(data.checks).toEqual([])
    expect(calls).toEqual([])
  })

  it("rejects an invalid file before running any check", async () => {
    const { deps, calls } = fakeDeps({
      config: { version: 1, checks: [{ name: "ok", run: "ok.sh" }, { name: "bad", run: "a.sh; rm -rf /" }] },
      executables: [`${ROOT}/ok.sh`],
    })
    await expect(runDeclaredChecks(ROOT, deps)).rejects.toThrow(PreflightConfigError)
    expect(calls).toEqual([])
  })

  it("runs in file order with literal argv, maps gates, and reports required failures", async () => {
    const { deps, calls } = fakeDeps({
      config: {
        version: 1,
        checks: [
          { name: "contract", run: ".uf/scripts/contract", args: ["--base", "$HOME"], requires: ["yq"], paths: ["scripts/"] },
          { name: "impact", run: ".uf/scripts/impact", gate: "advisory", timeout_seconds: 5 },
          { name: "docs", run: ".uf/scripts/docs", paths: ["docs/"] },
          { name: "needs-ripwire", run: ".uf/scripts/rw", requires: ["ripwire"] },
          { name: "missing-run", run: ".uf/scripts/absent" },
        ],
      },
      executables: [
        "/opt/tools/yq",
        `${ROOT}/.uf/scripts/contract`,
        `${ROOT}/.uf/scripts/impact`,
        `${ROOT}/.uf/scripts/docs`,
        `${ROOT}/.uf/scripts/rw`,
      ],
      git: branchGit(["scripts/pre-code.src.sh"]),
      results: {
        [`${ROOT}/.uf/scripts/contract`]: { exitCode: 1, output: "ERROR harness-unsupplied var=X", timedOut: false },
        [`${ROOT}/.uf/scripts/impact`]: { exitCode: 124, output: "partial", timedOut: true },
      },
    })

    const data = await runDeclaredChecks(ROOT, deps)

    expect(calls).toEqual([
      { file: `${ROOT}/.uf/scripts/contract`, args: ["--base", "$HOME"], cwd: ROOT, timeoutMs: DEFAULT_CHECK_TIMEOUT_SECONDS * 1000 },
      { file: `${ROOT}/.uf/scripts/impact`, args: [], cwd: ROOT, timeoutMs: 5000 },
    ])
    expect(data.checks.map((c) => [c.name, c.status, c.exit_code])).toEqual([
      ["contract", "FAIL", 1],
      ["impact", "ADVISORY FAIL", 124],
      ["docs", "SKIP — no in-scope files", null],
      ["needs-ripwire", "SKIP — not available", null],
      ["missing-run", "SKIP — not available", null],
    ])
    expect(data.checks[3]?.unavailable).toEqual(["ripwire"])
    expect(data.checks[4]?.unavailable).toEqual([".uf/scripts/absent"])
    expect(data.checks[1]?.output).toBe("partial\n[timed out after 5s]")
    expect(data.required_failures).toEqual(["contract"])
    expect(data.scope.files).toEqual(["scripts/pre-code.src.sh"])
    expect(data.rendered.split("\n")).toHaveLength(5)
  })

  it("is deterministic: identical inputs give byte-identical payloads", async () => {
    const config = { version: 1, checks: [{ name: "a", run: "a.sh", paths: [".sh"] }] }
    const first = await runDeclaredChecks(ROOT, fakeDeps({ config, executables: [`${ROOT}/a.sh`], git: branchGit(["b.sh", "a.sh"]) }).deps)
    const second = await runDeclaredChecks(ROOT, fakeDeps({ config, executables: [`${ROOT}/a.sh`], git: branchGit(["a.sh", "b.sh"]) }).deps)
    expect(JSON.stringify(first)).toBe(JSON.stringify(second))
  })

  it("skips git entirely when no check declares paths", async () => {
    const { deps } = fakeDeps({
      config: { version: 1, checks: [{ name: "a", run: "a.sh" }] },
      executables: [`${ROOT}/a.sh`],
      git: () => {
        throw new Error("git must not run")
      },
    })
    const data = await runDeclaredChecks(ROOT, deps)
    expect(data.checks[0]?.status).toBe("PASS")
    expect(data.scope.fallback_reason).toBe("no check declares paths")
  })

  it("passes UF_PREFLIGHT_PROJECT_ROOT to every check", async () => {
    const { deps, envs } = fakeDeps({ config: { version: 1, checks: [{ name: "a", run: "a.sh" }] }, executables: [`${ROOT}/a.sh`] })
    await runDeclaredChecks(ROOT, deps)
    expect(envs).toEqual([{ UF_PREFLIGHT_PROJECT_ROOT: ROOT }])
  })
})

// ---------------------------------------------------------------------------
// Modes and soft-gate classification
// ---------------------------------------------------------------------------

/** Branch git fake that also accepts worktree add/remove at the merge-base. */
function baselineGit(files: readonly string[], worktreeLog: string[]): GitExec {
  const inner = branchGit(files)
  return (args, cwd) => {
    if (args[0] === "worktree") {
      worktreeLog.push(args.join(" "))
      return ""
    }
    return inner(args, cwd)
  }
}

describe("runDeclaredChecks modes", () => {
  const config = {
    version: 1,
    checks: [
      { name: "new-break", run: "new.sh" },
      { name: "old-break", run: "old.sh" },
      { name: "advisory", run: "adv.sh", gate: "advisory" },
    ],
  }
  const executables = [`${ROOT}/new.sh`, `${ROOT}/old.sh`, `${ROOT}/adv.sh`]
  const results = {
    [`${ROOT}/new.sh`]: { exitCode: 1, output: "new", timedOut: false },
    [`${ROOT}/old.sh`]: { exitCode: 1, output: "old", timedOut: false },
    [`${ROOT}/adv.sh`]: { exitCode: 1, output: "adv", timedOut: false },
  }

  it("hard-gate and ci-aware fail on any required failure without a baseline", async () => {
    for (const mode of ["hard-gate", "ci-aware"] as const) {
      const log: string[] = []
      const { deps } = fakeDeps({ config, executables, results, git: baselineGit(["x"], log) })
      const data = await runDeclaredChecks(ROOT, deps, mode)
      expect(data.verdict).toBe("FAIL")
      expect(data.baseline_method).toBeNull()
      expect(data.checks.every((c) => c.causality === null)).toBe(true)
      expect(log).toEqual([])
    }
  })

  it("soft-gate re-runs only required failures at the merge-base and classifies them", async () => {
    const log: string[] = []
    const { deps, calls, envs, removed } = fakeDeps({
      config,
      executables,
      results,
      git: baselineGit(["x"], log),
      baseline: { [`${ROOT}/old.sh`]: { exitCode: 1, output: "still broken", timedOut: false } },
    })
    const data = await runDeclaredChecks(ROOT, deps, "soft-gate")

    expect(log).toEqual([`worktree add --detach ${WORKTREE} ${MERGE_BASE}`, `worktree remove --force ${WORKTREE}`])
    expect(removed).toEqual([WORKTREE])
    // Baseline runs execute the branch's script from the worktree cwd.
    expect(calls.filter((c) => c.cwd === WORKTREE).map((c) => c.file)).toEqual([`${ROOT}/new.sh`, `${ROOT}/old.sh`])
    // Every run, baseline included, can find the primary checkout.
    expect(envs).toHaveLength(calls.length)
    expect(envs.every((env) => env.UF_PREFLIGHT_PROJECT_ROOT === ROOT)).toBe(true)
    expect(data.checks.map((c) => [c.name, c.causality])).toEqual([
      ["new-break", "branch-caused"],
      ["old-break", "pre-existing"],
      ["advisory", null],
    ])
    expect(data.baseline_method).toBe("worktree")
    expect(data.verdict).toBe("FAIL")
  })

  it("soft-gate passes when every required failure is pre-existing", async () => {
    const log: string[] = []
    const { deps } = fakeDeps({
      config: { version: 1, checks: [{ name: "old-break", run: "old.sh" }] },
      executables,
      results,
      git: baselineGit(["x"], log),
      baseline: { [`${ROOT}/old.sh`]: { exitCode: 2, output: "", timedOut: false } },
    })
    const data = await runDeclaredChecks(ROOT, deps, "soft-gate")
    expect(data.verdict).toBe("PASS")
    expect(data.rendered).toBe("| old-break | `old.sh` | 1 | FAIL | pre-existing |")
  })

  it("soft-gate classifies as unknown (gating) when no baseline is available", async () => {
    const { deps } = fakeDeps({
      config: { version: 1, checks: [{ name: "old-break", run: "old.sh" }] },
      executables,
      results,
      git: gitFake({}),
    })
    const data = await runDeclaredChecks(ROOT, deps, "soft-gate")
    expect(data.baseline_method).toBe("unavailable")
    expect(data.checks[0]?.causality).toBe("unknown")
    expect(data.verdict).toBe("FAIL")
    // The recomputed scope reports the real reason, not the no-paths note.
    expect(data.scope.fallback_reason).toBe("no base ref resolved")
  })

  it("soft-gate removes the temp dir when worktree creation fails", async () => {
    const git: GitExec = (args, cwd) => {
      if (args[0] === "worktree") {
        throw new Error("worktree add failed")
      }
      return branchGit(["x"])(args, cwd)
    }
    const { deps, removed } = fakeDeps({ config: { version: 1, checks: [{ name: "old-break", run: "old.sh" }] }, executables, results, git })
    const data = await runDeclaredChecks(ROOT, deps, "soft-gate")
    expect(data.baseline_method).toBe("unavailable")
    expect(removed).toEqual([WORKTREE])
  })
})

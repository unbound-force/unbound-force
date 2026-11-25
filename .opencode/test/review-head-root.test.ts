import { readFileSync } from "node:fs"
import { resolve } from "node:path"
import { fileURLToPath } from "node:url"

import { describe, expect, it } from "vitest"
import { z } from "zod"

import type { ToolResult } from "../lib/uf-workflow-types.js"
import {
  REVIEW_HEAD_DIR,
  REVIEW_HEAD_ROOT_ERROR_CODES,
  REVIEW_HEAD_STALE_MS,
  createReviewHeadRootTool,
  parseRemoteRepo,
  validateReviewHeadRoot,
  type GhRunner,
  type GitExec,
  type ReviewHeadFs,
  type ReviewHeadRootData,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// review_head_root unit tests (OpenSpec change review-head-preflight,
// issues #1203/#1204). git, gh, and the filesystem are injected fakes over
// one mutable model, so fetches add objects, worktree add/remove change
// the worktree list, and directory removal is observable.
// ---------------------------------------------------------------------------

const CWD = "/fake/project"
const OPENCODE_ROOT = fileURLToPath(new URL("../", import.meta.url))
const HEAD_SHA = "a".repeat(40)
const BASE_SHA = "b".repeat(40)
const USER_SHA = "c".repeat(40)
const PR = 42
const NOW = 1_800_000_000_000
const DIR = `${CWD}/${REVIEW_HEAD_DIR}`

interface Model {
  head: string
  status: string
  objects: Set<string>
  remotes: Record<string, string>
  /** refspec → sha added on fetch, or an Error thrown by the fetch. */
  fetchable: Record<string, string | Error>
  worktrees: Set<string>
  worktreeAddError: Error | null
  worktreeRemoveError: Error | null
  diff: string | Error
  /** gh pr view payload, or an Error. */
  pull: Record<string, unknown> | Error
  /** gh permission answer, or an Error. */
  permission: string | Error
  /** Directories on disk. */
  dirs: Set<string>
  files: Map<string, string>
  /** Symlink-style overrides for realpath. */
  links: Map<string, string>
  mtimes: Map<string, number>
  ids: string[]
}

interface Calls {
  git: string[][]
  gh: string[][]
  removed: string[]
}

function model(overrides: Partial<Model> = {}): Model {
  return {
    head: USER_SHA,
    status: "",
    objects: new Set([HEAD_SHA, BASE_SHA, USER_SHA]),
    remotes: { origin: "git@github.com:me/project.git", upstream: "https://github.com/Org/Project.git" },
    fetchable: {},
    worktrees: new Set(),
    worktreeAddError: null,
    worktreeRemoveError: null,
    diff: "internal/b.go\ninternal/a.go\n",
    pull: pr({}),
    permission: "write",
    dirs: new Set([CWD]),
    files: new Map(),
    links: new Map(),
    mtimes: new Map(),
    ids: ["k3v9x2qa", "m4w0y3rb"],
    ...overrides,
  }
}

/** A gh pr view payload; same-repo by default. */
function pr(overrides: { head?: string | null; author?: string | null; sha?: string; cross?: boolean }): Record<string, unknown> {
  const head = overrides.head === undefined ? "org/project" : overrides.head
  const [owner, name] = head === null ? [null, null] : head.split("/")
  return {
    url: `https://github.com/org/project/pull/${PR}`,
    headRefOid: overrides.sha ?? HEAD_SHA,
    isCrossRepository: overrides.cross ?? (head === null || head.toLowerCase() !== "org/project"),
    headRepository: name === null ? null : { name },
    headRepositoryOwner: owner === null ? null : { login: owner },
    author: overrides.author === null ? null : { login: overrides.author ?? "dev" },
  }
}

function gitExit(status: number, stderr = ""): Error {
  return Object.assign(new Error(`Command failed with exit ${status}`), { status, stderr })
}

function fakeGit(m: Model, calls: Calls): GitExec {
  return (args: string[], cwd: string) => {
    expect(cwd).toBe(CWD)
    calls.git.push([...args])
    const key = args.join(" ")
    if (key === "rev-parse HEAD") return m.head
    if (key === "status --porcelain --untracked-files=normal --ignore-submodules=none") return m.status
    if (args[0] === "cat-file" && args[1] === "-e") {
      const sha = (args[2] ?? "").replace("^{commit}", "")
      if (m.objects.has(sha)) return ""
      throw gitExit(128)
    }
    if (key === "remote") return Object.keys(m.remotes).join("\n")
    if (args[0] === "remote" && args[1] === "get-url" && args[2] === "--") {
      const url = m.remotes[args[3] as string]
      if (url === undefined) throw gitExit(2)
      return url
    }
    if (args[0] === "fetch" && args[1] === "--no-tags" && args[2] === "--") {
      const result = m.fetchable[args[4] as string]
      if (result instanceof Error) throw result
      if (result !== undefined) m.objects.add(result)
      return ""
    }
    if (key === "worktree list --porcelain") {
      return [`worktree ${CWD}`, ...[...m.worktrees].map((path) => `worktree ${path}`)].join("\n")
    }
    if (args[0] === "worktree" && args[1] === "add") {
      expect(args.slice(0, 4)).toEqual(["worktree", "add", "--detach", "--"])
      if (m.worktreeAddError !== null) {
        m.dirs.add(args[4] as string)
        throw m.worktreeAddError
      }
      m.worktrees.add(args[4] as string)
      m.dirs.add(args[4] as string)
      return ""
    }
    if (args[0] === "worktree" && args[1] === "remove") {
      expect(args.slice(0, 4)).toEqual(["worktree", "remove", "--force", "--"])
      if (m.worktreeRemoveError !== null) throw m.worktreeRemoveError
      m.worktrees.delete(args[4] as string)
      return ""
    }
    if (args[0] === "diff" && args[1] === "--name-only") {
      expect(args[2]).toBe(`${BASE_SHA}...${HEAD_SHA}`)
      if (m.diff instanceof Error) throw m.diff
      return m.diff
    }
    throw new Error(`unexpected git command: ${key}`)
  }
}

function fakeGh(m: Model, calls: Calls): GhRunner {
  return (args: string[], cwd: string) => {
    expect(cwd).toBe(CWD)
    calls.gh.push([...args])
    if (args.slice(0, 3).join(" ") === `pr view ${PR}`) {
      expect(args.slice(3)).toEqual(["--json", "url,headRefOid,isCrossRepository,headRepository,headRepositoryOwner,author"])
      if (m.pull instanceof Error) throw m.pull
      return JSON.stringify(m.pull)
    }
    if (args[0] === "api" && /^repos\/[^/]+\/[^/]+\/collaborators\/[^/]+\/permission$/.test(args[1] ?? "")) {
      if (m.permission instanceof Error) throw m.permission
      return m.permission
    }
    throw new Error(`unexpected gh command: ${args.join(" ")}`)
  }
}

function fakeFs(m: Model, calls: Calls): ReviewHeadFs {
  return {
    ensureDir: (path) => {
      m.dirs.add(path)
    },
    writeFileIfMissing: (path, content) => {
      if (!m.files.has(path)) m.files.set(path, content)
    },
    removeDir: (path) => {
      calls.removed.push(path)
      m.dirs.delete(path)
    },
    realpath: (path) => {
      const link = m.links.get(path)
      if (link !== undefined) return link
      return m.dirs.has(path) ? path : null
    },
    mtimeMs: (path) => (m.dirs.has(path) ? (m.mtimes.get(path) ?? NOW) : null),
    listDir: (path) =>
      [...m.dirs]
        .filter((dir) => dir.startsWith(`${path}/`) && !dir.slice(path.length + 1).includes("/"))
        .map((dir) => dir.slice(path.length + 1)),
    now: () => NOW,
    randomId: () => m.ids.shift() ?? "zzzzzzzz",
  }
}

function setup(overrides: Partial<Model> = {}) {
  const m = model(overrides)
  const calls: Calls = { git: [], gh: [], removed: [] }
  const tool = createReviewHeadRootTool(fakeGit(m, calls), fakeGh(m, calls), CWD, fakeFs(m, calls))
  const run = async (args: Record<string, unknown>) => {
    const result = await (tool as unknown as { execute: (a: unknown) => Promise<{ output: string }> }).execute(args)
    return JSON.parse(result.output) as ToolResult<ReviewHeadRootData> & { code?: string; retryable?: boolean; message?: string }
  }
  const prepare = () => run({ action: "prepare", pr_number: PR, base_sha: BASE_SHA, head_sha: HEAD_SHA })
  return { m, calls, run, prepare, fs: fakeFs(m, calls), git: fakeGit(m, calls) }
}

function data(envelope: ToolResult<ReviewHeadRootData>): ReviewHeadRootData {
  expect(envelope.status).toBe("ok")
  return (envelope as { data: ReviewHeadRootData }).data
}

/** True when any recorded git call starts with the given argv prefix. */
function called(calls: Calls, ...prefix: string[]): boolean {
  return calls.git.some((args) => prefix.every((part, index) => args[index] === part))
}

// ---------------------------------------------------------------------------
// prepare: reuse and worktree paths
// ---------------------------------------------------------------------------

describe("review_head_root prepare", () => {
  it("reuses a clean checkout of the head without the trust gate or a worktree", async () => {
    const { calls, prepare } = setup({ head: HEAD_SHA, permission: new Error("must not be read") })
    const result = data(await prepare())
    expect(result).toEqual({
      root: CWD,
      head_sha: HEAD_SHA,
      base_sha: BASE_SHA,
      method: "checkout",
      created: false,
      config_sha: HEAD_SHA,
      config_dirty: false,
      changed_files: ["internal/a.go", "internal/b.go"],
    })
    expect(calls.gh).toEqual([])
    expect(called(calls, "worktree", "add")).toBe(false)
  })

  it.each([
    ["an untracked file", "?? scratch_test.go"],
    ["a modified tracked file", " M internal/a.go"],
  ])("does not reuse the head checkout with %s", async (_label, status) => {
    const { m, prepare } = setup({ head: HEAD_SHA, status })
    const result = data(await prepare())
    expect(result.method).toBe("worktree")
    expect(result.root).toBe(`${DIR}/k3v9x2qa`)
    expect(result.config_sha).toBe(HEAD_SHA)
    expect(result.config_dirty).toBe(true)
    expect(m.worktrees.has(`${DIR}/k3v9x2qa`)).toBe(true)
  })

  it("adds a detached worktree for a same-repo PR and self-ignores the directory", async () => {
    const { m, calls, prepare } = setup()
    const result = data(await prepare())
    expect(result).toMatchObject({
      root: `${DIR}/k3v9x2qa`,
      method: "worktree",
      created: true,
      config_sha: USER_SHA,
      config_dirty: false,
    })
    expect(m.files.get(`${DIR}/.gitignore`)).toBe("*\n")
    expect(calls.git).toContainEqual(["worktree", "add", "--detach", "--", `${DIR}/k3v9x2qa`, HEAD_SHA])
    // The user's checkout is never touched.
    expect(called(calls, "checkout")).toBe(false)
    expect(called(calls, "switch")).toBe(false)
    expect(called(calls, "stash")).toBe(false)
    expect(called(calls, "reset")).toBe(false)
    expect(calls.gh.some((args) => args[1]?.includes("/permission"))).toBe(false)
  })

  it("trusts a fork whose author has write permission on the base repo", async () => {
    const { calls, prepare } = setup({ pull: pr({ head: "contrib/project", author: "contrib" }), permission: "write" })
    expect(data(await prepare()).method).toBe("worktree")
    expect(calls.gh).toContainEqual(["api", "repos/org/project/collaborators/contrib/permission", "--jq", ".permission"])
  })

  it.each([
    ["a read-only author", { head: "contrib/project", author: "contrib" }, "read"],
    ["a deleted fork", { head: null, author: "contrib" }, "read"],
    ["an unreadable permission", { head: "contrib/project", author: "contrib" }, new Error("HTTP 403")],
    ["a deleted author", { head: "contrib/project", author: null }, "write"],
    ["an author login outside the charset", { head: "contrib/project", author: "x/../y" }, "write"],
    ["a same-named fork marked cross-repository", { head: "org/project", author: "contrib", cross: true }, "read"],
  ])("refuses an untrusted fork head (%s) before any fetch or worktree", async (_label, pull, permission) => {
    const { calls, prepare } = setup({
      pull: pr(pull as Parameters<typeof pr>[0]),
      permission: permission as string | Error,
      objects: new Set([USER_SHA]),
    })
    const envelope = await prepare()
    expect(envelope.status).toBe("error")
    expect(envelope.code).toBe("UNTRUSTED_HEAD")
    expect(envelope.message).toContain(`gh pr checkout ${PR}`)
    expect(called(calls, "fetch")).toBe(false)
    expect(called(calls, "worktree", "add")).toBe(false)
  })

  it("fails HEAD_MOVED when head_sha is not the vetted PR's head, before fetch or worktree", async () => {
    // A head_sha from another (possibly untrusted) PR is already local:
    // the trusted author of PR 42 must not vouch for it.
    const { calls, prepare } = setup({ pull: pr({ sha: "f".repeat(40) }) })
    const envelope = await prepare()
    expect(envelope.code).toBe("HEAD_MOVED")
    expect(called(calls, "fetch")).toBe(false)
    expect(called(calls, "worktree", "add")).toBe(false)
  })

  it("fails HEAD_UNAVAILABLE when gh returns an unexpected PR URL", async () => {
    const { prepare } = setup({ pull: { ...pr({}), url: "https://github.com/org/project/issues/42" } })
    const envelope = await prepare()
    expect(envelope.code).toBe("HEAD_UNAVAILABLE")
  })

  it("rejects missing prepare arguments and bad SHAs", async () => {
    const { run } = setup()
    const missing = await run({ action: "prepare", pr_number: PR })
    expect(missing.status).toBe("error")
    const tool = createReviewHeadRootTool(
      () => "",
      () => "",
      CWD,
    ) as unknown as { args: Record<string, z.ZodTypeAny> }
    const schema = z.object(tool.args)
    expect(schema.safeParse({ action: "prepare", head_sha: "ABC" }).success).toBe(false)
    expect(schema.safeParse({ action: "delete" }).success).toBe(false)
    expect(schema.safeParse({ action: "prepare", pr_number: 0 }).success).toBe(false)
    expect(schema.safeParse({ action: "prepare", pr_number: PR, base_sha: BASE_SHA, head_sha: HEAD_SHA }).success).toBe(true)
  })
})

// ---------------------------------------------------------------------------
// prepare: fetching missing commits
// ---------------------------------------------------------------------------

describe("review_head_root fetch", () => {
  it("fetches a missing head from the base-repo remote and verifies it with cat-file", async () => {
    const { m, calls, prepare } = setup({
      objects: new Set([BASE_SHA, USER_SHA]),
      fetchable: { [`refs/pull/${PR}/head`]: HEAD_SHA },
    })
    expect(data(await prepare()).method).toBe("worktree")
    expect(calls.git).toContainEqual(["fetch", "--no-tags", "--", "upstream", `refs/pull/${PR}/head`])
    expect(calls.git.some((args) => args.includes("FETCH_HEAD"))).toBe(false)
    expect(m.objects.has(HEAD_SHA)).toBe(true)
  })

  it("fetches a missing base commit by SHA", async () => {
    const { calls, prepare } = setup({ objects: new Set([HEAD_SHA, USER_SHA]), fetchable: { [BASE_SHA]: BASE_SHA } })
    expect(data(await prepare()).method).toBe("worktree")
    expect(calls.git).toContainEqual(["fetch", "--no-tags", "--", "upstream", BASE_SHA])
  })

  it("fails HEAD_MOVED when the fetch succeeds but the frozen head is gone", async () => {
    const { calls, prepare } = setup({
      objects: new Set([BASE_SHA, USER_SHA]),
      fetchable: { [`refs/pull/${PR}/head`]: "d".repeat(40) },
    })
    const envelope = await prepare()
    expect(envelope.code).toBe("HEAD_MOVED")
    expect(envelope.retryable).toBe(false)
    expect(called(calls, "worktree", "add")).toBe(false)
  })

  it("fails HEAD_UNAVAILABLE (retryable) when the fetch errors", async () => {
    const { prepare } = setup({
      objects: new Set([BASE_SHA, USER_SHA]),
      fetchable: { [`refs/pull/${PR}/head`]: gitExit(128, "Could not resolve host") },
    })
    const envelope = await prepare()
    expect(envelope.code).toBe("HEAD_UNAVAILABLE")
    expect(envelope.retryable).toBe(true)
    expect(envelope.message).toContain("Could not resolve host")
  })

  it("fails HEAD_UNAVAILABLE (not retryable) when no remote names the base repo", async () => {
    const { prepare } = setup({ objects: new Set([USER_SHA]), remotes: { origin: "git@github.com:me/project.git" } })
    const envelope = await prepare()
    expect(envelope.code).toBe("HEAD_UNAVAILABLE")
    expect(envelope.retryable).toBe(false)
  })

  it("fails HEAD_UNAVAILABLE (retryable) when the PR lookup fails", async () => {
    const { prepare } = setup({ pull: new Error("gh: HTTP 502") })
    const envelope = await prepare()
    expect(envelope.code).toBe("HEAD_UNAVAILABLE")
    expect(envelope.retryable).toBe(true)
  })

  it("fails HEAD_UNAVAILABLE when the base commit stays missing after a fetch", async () => {
    const { prepare } = setup({ objects: new Set([HEAD_SHA, USER_SHA]) })
    expect((await prepare()).code).toBe("HEAD_UNAVAILABLE")
  })
})

describe("parseRemoteRepo", () => {
  it.each([
    ["git@github.com:Org/Project.git", "github.com", "org/project"],
    ["github.com:org/project", "github.com", "org/project"],
    ["ssh://git@github.com/org/project.git", "github.com", "org/project"],
    ["ssh://git@GitHub.com:22/org/project", "github.com", "org/project"],
    ["https://github.com/org/project", "github.com", "org/project"],
    ["https://user:token@github.com/org/project.git/", "github.com", "org/project"],
    ["git://github.com/org/project.git", "github.com", "org/project"],
  ])("parses %s", (url, host, repo) => {
    expect(parseRemoteRepo(url)).toEqual({ host, repo })
  })

  it.each(["/local/path/repo", "file:///srv/repo.git", "https://github.com/org", "../sibling"])("rejects %s", (url) => {
    expect(parseRemoteRepo(url)).toBeNull()
  })
})

// ---------------------------------------------------------------------------
// prepare: worktree failure and pruning
// ---------------------------------------------------------------------------

describe("review_head_root worktree lifecycle", () => {
  it("fails WORKTREE_FAILED with git's message and leaves no directory", async () => {
    const { m, prepare } = setup({ worktreeAddError: gitExit(128, "fatal: could not lock config file") })
    const envelope = await prepare()
    expect(envelope.code).toBe("WORKTREE_FAILED")
    expect(envelope.message).toContain("could not lock config file")
    expect(m.dirs.has(`${DIR}/k3v9x2qa`)).toBe(false)
  })

  it("never leaves a worktree behind when the diff fails", async () => {
    const { m, calls, prepare } = setup({ diff: gitExit(128, "bad object") })
    const envelope = await prepare()
    expect(envelope.status).toBe("error")
    expect(called(calls, "worktree", "add")).toBe(false)
    expect(m.worktrees.size).toBe(0)
  })

  it("prunes only review-head worktrees older than the stale window", async () => {
    const old = `${DIR}/oldoldol`
    const young = `${DIR}/youngyou`
    const evil = `${CWD}/.uf/review-head-x-evil`
    const user = "/elsewhere/user-worktree"
    const { m, calls, prepare } = setup({
      dirs: new Set([CWD, DIR, old, young, evil, user, `${DIR}/notanid`]),
      worktrees: new Set([old, young, evil, user, `${DIR}/notanid`]),
      mtimes: new Map([
        [old, NOW - REVIEW_HEAD_STALE_MS - 1],
        [young, NOW - REVIEW_HEAD_STALE_MS + 1],
        [evil, 0],
        [user, 0],
        [`${DIR}/notanid`, 0],
      ]),
    })
    data(await prepare())
    expect(calls.git).toContainEqual(["worktree", "remove", "--force", "--", old])
    expect(m.dirs.has(old)).toBe(false)
    for (const kept of [young, evil, user, `${DIR}/notanid`]) {
      expect(m.dirs.has(kept)).toBe(true)
    }
  })

  it("never fails prepare because pruning failed", async () => {
    const old = `${DIR}/oldoldol`
    const { prepare } = setup({
      dirs: new Set([CWD, DIR, old]),
      worktrees: new Set([old]),
      mtimes: new Map([[old, 0]]),
      worktreeRemoveError: gitExit(128, "locked"),
    })
    expect(data(await prepare()).method).toBe("worktree")
  })
})

// ---------------------------------------------------------------------------
// release and root validation
// ---------------------------------------------------------------------------

describe("review_head_root release", () => {
  it("removes a created worktree, and a second release is INVALID_ROOT", async () => {
    const { m, run, prepare } = setup()
    const root = data(await prepare()).root
    const first = await run({ action: "release", root })
    expect(first.status).toBe("ok")
    expect(m.worktrees.has(root)).toBe(false)
    expect(m.dirs.has(root)).toBe(false)
    const second = await run({ action: "release", root })
    expect(second.code).toBe("INVALID_ROOT")
  })

  it("releases a worktree prepared by another process (no in-memory registry)", async () => {
    const root = `${DIR}/k3v9x2qa`
    const { m, run } = setup({ dirs: new Set([CWD, DIR, root]), worktrees: new Set([root]) })
    expect((await run({ action: "release", root })).status).toBe("ok")
    expect(m.dirs.has(root)).toBe(false)
  })

  it("treats releasing the reused checkout as a no-op", async () => {
    const { calls, run } = setup()
    const envelope = await run({ action: "release", root: CWD })
    expect(envelope.status).toBe("ok")
    expect(calls.removed).toEqual([])
  })

  it("still deletes the directory when git worktree remove fails", async () => {
    const root = `${DIR}/k3v9x2qa`
    const { m, run } = setup({
      dirs: new Set([CWD, DIR, root]),
      worktrees: new Set([root]),
      worktreeRemoveError: gitExit(128, "is locked"),
    })
    expect((await run({ action: "release", root })).status).toBe("ok")
    expect(m.dirs.has(root)).toBe(false)
  })

  it.each([
    ["a trailing slash", `${DIR}/k3v9x2qa/`],
    ["a dot-dot segment", `${DIR}/../review-head/k3v9x2qa`],
    ["an escaping path", `${DIR}/../../x`],
    ["a non-id name", `${DIR}/K3V9X2QA`],
    ["the review-head directory itself", DIR],
    ["an unrelated path", "/tmp/anything"],
    ["no root", undefined],
  ])("refuses %s and deletes nothing", async (_label, root) => {
    const live = `${DIR}/k3v9x2qa`
    const { calls, run } = setup({ dirs: new Set([CWD, DIR, live]), worktrees: new Set([live]) })
    const envelope = await run(root === undefined ? { action: "release" } : { action: "release", root })
    expect(envelope.code).toBe("INVALID_ROOT")
    expect(calls.removed).toEqual([])
    expect(called(calls, "worktree", "remove")).toBe(false)
  })

  it("refuses a symlinked review-head directory", async () => {
    const live = `${DIR}/k3v9x2qa`
    const { calls, run } = setup({
      dirs: new Set([CWD, DIR, live]),
      worktrees: new Set([live]),
      links: new Map([[live, "/elsewhere/k3v9x2qa"]]),
    })
    expect((await run({ action: "release", root: live })).code).toBe("INVALID_ROOT")
    expect(calls.removed).toEqual([])
  })

  it("refuses a review-head directory git does not list as a worktree", async () => {
    const stray = `${DIR}/k3v9x2qa`
    const { calls, run } = setup({ dirs: new Set([CWD, DIR, stray]) })
    expect((await run({ action: "release", root: stray })).code).toBe("INVALID_ROOT")
    expect(calls.removed).toEqual([])
  })
})

describe("validateReviewHeadRoot", () => {
  it("classifies the plugin directory and live worktrees", () => {
    const live = `${DIR}/k3v9x2qa`
    const { git, fs } = setup({ dirs: new Set([CWD, DIR, live]), worktrees: new Set([live]) })
    expect(validateReviewHeadRoot(git, fs, CWD, CWD)).toBe("checkout")
    expect(validateReviewHeadRoot(git, fs, CWD, live)).toBe("worktree")
    expect(validateReviewHeadRoot(git, fs, CWD, `${live}/`)).toBeNull()
  })

  it("returns null when git cannot list worktrees", () => {
    const live = `${DIR}/k3v9x2qa`
    const { fs } = setup({ dirs: new Set([CWD, DIR, live]) })
    const broken: GitExec = () => {
      throw gitExit(128)
    }
    expect(validateReviewHeadRoot(broken, fs, CWD, live)).toBeNull()
  })
})

// ---------------------------------------------------------------------------
// Command code-list parity (task 4.7)
// ---------------------------------------------------------------------------

describe("review_head_root code-list parity with the consuming commands", () => {
  const copies = [
    resolve(OPENCODE_ROOT, "commands/uf.review-pr.md"),
    resolve(OPENCODE_ROOT, "../internal/scaffold/assets/opencode/commands/uf.review-pr.md"),
    resolve(OPENCODE_ROOT, "commands/uf.review-council.md"),
    resolve(OPENCODE_ROOT, "../internal/scaffold/assets/opencode/commands/uf.review-council.md"),
  ]

  it.each(copies)("%s handles every prepare failure code", (path) => {
    const text = readFileSync(path, "utf8")
    for (const code of REVIEW_HEAD_ROOT_ERROR_CODES) {
      // INVALID_ROOT is a release/declared-checks code, handled as a release warning.
      if (code === "INVALID_ROOT") continue
      expect(text, `${path} must handle ${code}`).toContain(`\`${code}\``)
    }
    expect(text).toContain("`review_head_root`")
  })
})

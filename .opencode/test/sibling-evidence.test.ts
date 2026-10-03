import { createHash } from "node:crypto"
import { mkdir, mkdtemp, readFile, realpath, rm, stat, symlink, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { dirname, join, resolve } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import {
  acquireSiblingEvidence,
  createAcquireSiblingEvidenceTool,
  parseSiblingRepositories,
  UNTRUSTED_SIBLING_EVIDENCE_BEGIN,
  UNTRUSTED_SIBLING_EVIDENCE_END,
  type CommandOptions,
  type CommandResult,
  type SiblingAcquisitionDependencies,
} from "../plugins/review-dispatch/index.js"

const COMMIT_A = "a".repeat(40)
const COMMIT_B = "b".repeat(40)
const scratchDirectories: string[] = []

interface SiblingFixture {
  readonly name: string
  readonly url: string
  readonly default_branch: string
  readonly role: "downstream-consumer" | "upstream-source" | "sibling"
  readonly fetch: "clone-on-review" | "contract-only"
  readonly contracts: readonly string[]
  readonly local_paths?: readonly string[]
  readonly clone_path?: string
  readonly cache_path?: string
  readonly notes?: string
}

interface RepositoryState {
  readonly root: string
  readonly origin: string
  readonly commit: string
  readonly tracked: readonly string[]
  dirty?: boolean
}

interface FakeCommandState {
  readonly repositories: Map<string, RepositoryState>
  readonly calls: Array<{ readonly args: readonly string[]; readonly options: CommandOptions }>
  readonly cloneFiles?: Readonly<Record<string, string>>
  cloneFailure?: boolean
}

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

async function scratchDirectory(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-sibling-evidence-"))
  scratchDirectories.push(directory)
  return directory
}

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

function siblingFixture(overrides: Partial<SiblingFixture> = {}): SiblingFixture {
  return {
    name: "gaze",
    url: "https://github.com/unbound-force/gaze.git",
    default_branch: "main",
    role: "sibling",
    fetch: "contract-only",
    contracts: ["README.md", "docs/**/*.md"],
    ...overrides,
  }
}

function configuration(...siblings: readonly SiblingFixture[]): string {
  return JSON.stringify({ version: 1, siblings })
}

async function writeRepositoryFiles(root: string, files: Readonly<Record<string, string | Uint8Array>>): Promise<void> {
  await mkdir(root, { recursive: true })
  for (const [path, content] of Object.entries(files)) {
    const target = join(root, path)
    await mkdir(dirname(target), { recursive: true })
    await writeFile(target, content)
  }
}

async function addRepository(
  state: FakeCommandState,
  root: string,
  files: Readonly<Record<string, string | Uint8Array>>,
  options: { readonly origin?: string; readonly commit?: string; readonly dirty?: boolean } = {},
): Promise<void> {
  await writeRepositoryFiles(root, files)
  const canonicalRoot = await realpath(root)
  state.repositories.set(canonicalRoot, {
    root: canonicalRoot,
    origin: options.origin ?? "https://github.com/unbound-force/gaze.git",
    commit: options.commit ?? COMMIT_A,
    tracked: Object.keys(files),
    dirty: options.dirty,
  })
}

function fakeCommandRunner(state: FakeCommandState): SiblingAcquisitionDependencies["runCommand"] {
  return async (command: string, args: readonly string[], options: CommandOptions = {}): Promise<CommandResult> => {
    expect(command).toBe("git")
    expect(options.env).toMatchObject({ GIT_TERMINAL_PROMPT: "0", GCM_INTERACTIVE: "never" })
    state.calls.push({ args: [...args], options })
    if (args[0] === "clone") {
      if (state.cloneFailure === true) {
        throw new Error("offline")
      }
      const cloneRoot = args.at(-1)
      const origin = args.at(-2)
      if (cloneRoot === undefined || origin === undefined) {
        throw new Error("invalid clone call")
      }
      const files = state.cloneFiles ?? { "README.md": "cloned evidence" }
      await writeRepositoryFiles(cloneRoot, files)
      const canonicalRoot = await realpath(cloneRoot)
      state.repositories.set(canonicalRoot, {
        root: canonicalRoot,
        origin,
        commit: COMMIT_B,
        tracked: Object.keys(files),
      })
      return { stdout: "", stderr: "" }
    }
    const cwd = options.cwd === undefined ? undefined : await realpath(options.cwd)
    const repository = cwd === undefined ? undefined : state.repositories.get(cwd)
    if (repository === undefined) {
      throw new Error(`unknown fake repository ${cwd ?? "<none>"}`)
    }
    const operation = args.join(" ")
    if (operation === "rev-parse --show-toplevel") return { stdout: `${repository.root}\n`, stderr: "" }
    if (operation === "status --porcelain --untracked-files=normal") {
      return { stdout: repository.dirty === true ? " M README.md\n" : "", stderr: "" }
    }
    if (operation === "remote get-url origin") return { stdout: `${repository.origin}\n`, stderr: "" }
    if (operation === "rev-parse HEAD") return { stdout: `${repository.commit}\n`, stderr: "" }
    if (operation === "ls-files -z") return { stdout: `${repository.tracked.join("\0")}\0`, stderr: "" }
    if (args[0] === "fetch" || args[0] === "checkout") return { stdout: "", stderr: "" }
    throw new Error(`unexpected fake git operation: ${operation}`)
  }
}

function dependencies(projectRoot: string, config: string, state: FakeCommandState): SiblingAcquisitionDependencies {
  return {
    projectRoot,
    readText: async (relativePath: string): Promise<string> => {
      expect(relativePath).toBe(".uf/sibling-repos.yaml")
      return config
    },
    parseYaml: jsonParser,
    realpath,
    stat,
    readFile,
    mkdir,
    runCommand: fakeCommandRunner(state),
  }
}

function commandState(): FakeCommandState {
  return { repositories: new Map(), calls: [] }
}

describe("sibling repository declaration", () => {
  it("rejects duplicate names, unknown fields, unsafe URLs, paths, globs, and UTF-8 byte overflow", () => {
    expect(() =>
      parseSiblingRepositories(configuration(siblingFixture(), siblingFixture()), jsonParser),
    ).toThrow("duplicate sibling entry")
    expect(() =>
      parseSiblingRepositories(
        JSON.stringify({ version: 1, siblings: [{ ...siblingFixture(), unexpected: true }] }),
        jsonParser,
      ),
    ).toThrow("Unrecognized key")
    expect(() =>
      parseSiblingRepositories(
        configuration(siblingFixture({ url: "https://token@github.com/unbound-force/gaze" })),
        jsonParser,
      ),
    ).toThrow("credential-free")
    expect(() =>
      parseSiblingRepositories(configuration(siblingFixture({ clone_path: "../gaze" })), jsonParser),
    ).toThrow("target-root-relative")
    expect(() =>
      parseSiblingRepositories(configuration(siblingFixture({ contracts: ["docs/../README.md"] })), jsonParser),
    ).toThrow("repository-relative")
    expect(() =>
      parseSiblingRepositories(configuration(siblingFixture({ notes: "é".repeat(513) })), jsonParser),
    ).toThrow("1024 UTF-8 bytes")
  })

  it("normalizes valid GitHub forms and sorts siblings lexically", () => {
    const parsed = parseSiblingRepositories(
      configuration(
        siblingFixture({ name: "zeta", url: "https://GITHUB.com/Org/Repo.git/" }),
        siblingFixture({ name: "alpha" }),
      ),
      jsonParser,
    )
    expect(parsed.siblings.map((sibling) => sibling.name)).toEqual(["alpha", "zeta"])
  })
})

describe("sibling evidence tool surface", () => {
  it("exposes a closed provider-free tool definition", async () => {
    const projectRoot = await scratchDirectory()
    const state = commandState()
    const definition = createAcquireSiblingEvidenceTool(
      dependencies(projectRoot, configuration(), state),
    )

    expect(definition.description).toContain("untrusted-evidence delimiters")
    expect(Object.keys(definition.args)).toEqual([])
    expect(JSON.parse(await definition.execute({}, {} as never))).toMatchObject({
      version: 1,
      status: "ok",
      siblings: [],
    })
    expect(state.calls).toEqual([])
  })
})

describe("sibling repository acquisition", () => {
  it("uses local precedence, continues after invalid candidates, and records bounded evidence", async () => {
    const projectRoot = await scratchDirectory()
    const dirtyRoot = join(projectRoot, "dirty")
    const validRoot = join(projectRoot, "valid")
    const laterRoot = join(projectRoot, "later")
    const state = commandState()
    await addRepository(state, dirtyRoot, { "README.md": "dirty" }, { dirty: true })
    await addRepository(state, validRoot, { "README.md": "readme", "docs/z.md": "z", "docs/a.md": "a" })
    await addRepository(state, laterRoot, { "README.md": "later" })
    const config = configuration(
      siblingFixture({ local_paths: [dirtyRoot, validRoot, laterRoot], contracts: ["docs/*.md", "README.md"] }),
    )

    const result = await acquireSiblingEvidence({}, dependencies(projectRoot, config, state))

    expect(result.status).toBe("ok")
    expect(result.siblings[0]).toMatchObject({
      status: "available",
      source_mode: "local",
      source_candidate: "local_paths[1]",
      commit: COMMIT_A,
    })
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("candidate-dirty")
    expect(result.siblings[0]?.evidence.map((item) => item.path)).toEqual(["README.md", "docs/a.md", "docs/z.md"])
    expect(result.siblings[0]?.evidence[0]?.sha256).toBe(createHash("sha256").update("readme").digest("hex"))
    expect(result.prompt.match(new RegExp(UNTRUSTED_SIBLING_EVIDENCE_BEGIN, "g"))).toHaveLength(3)
    expect(result.prompt.match(new RegExp(UNTRUSTED_SIBLING_EVIDENCE_END, "g"))).toHaveLength(3)
    expect(result.prompt.indexOf("\na\n")).toBeLessThan(result.prompt.indexOf("\nz\n"))
    expect(state.calls.some((call) => call.options.cwd === laterRoot)).toBe(false)
  })

  it("records origin mismatch and dirty state while allowing a later verified candidate", async () => {
    const projectRoot = await scratchDirectory()
    const wrongOrigin = join(projectRoot, "wrong-origin")
    const dirty = join(projectRoot, "dirty")
    const valid = join(projectRoot, "valid")
    const state = commandState()
    await addRepository(
      state,
      wrongOrigin,
      { "README.md": "wrong" },
      { origin: "https://github.com/other/project.git" },
    )
    await addRepository(state, dirty, { "README.md": "dirty" }, { dirty: true })
    await addRepository(state, valid, { "README.md": "valid" })

    const result = await acquireSiblingEvidence(
      {},
      dependencies(projectRoot, configuration(siblingFixture({ local_paths: [wrongOrigin, dirty, valid] })), state),
    )

    expect(result.siblings[0]?.status).toBe("available")
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toEqual([
      "candidate-origin-mismatch",
      "candidate-dirty",
    ])
  })

  it("keeps contract-only provider and network free when no source exists", async () => {
    const projectRoot = await scratchDirectory()
    const state = commandState()
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [join(projectRoot, "missing")] })),
        state,
      ),
    )

    expect(result.siblings[0]).toMatchObject({
      status: "unavailable",
      unavailable_reason: "no-verified-candidate",
      evidence: [],
    })
    expect(result.prompt).toBe("")
    expect(state.calls).toEqual([])
  })

  it("rejects a target-root-relative candidate that canonicalizes outside the project", async () => {
    const projectRoot = await scratchDirectory()
    const outsideRoot = await scratchDirectory()
    const state = commandState()
    await addRepository(state, outsideRoot, { "README.md": "outside" })
    await symlink(outsideRoot, join(projectRoot, "linked-repository"))
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: ["linked-repository"] })),
        state,
      ),
    )

    expect(result.siblings[0]?.status).toBe("unavailable")
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("candidate-path-escape")
    expect(result.prompt).toBe("")
  })

  it("clones the declared branch with credentials disabled and uses the fetched immutable head", async () => {
    const projectRoot = await scratchDirectory()
    const state = commandState()
    state.cloneFiles = { "README.md": "clone success" }
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ fetch: "clone-on-review", clone_path: ".uf/clones/gaze" })),
        state,
      ),
    )

    expect(result.siblings[0]).toMatchObject({ status: "available", source_mode: "clone", commit: COMMIT_B })
    const cloneCall = state.calls.find((call) => call.args[0] === "clone")
    expect(cloneCall?.args).toContain("main")
    expect(cloneCall?.args).toContain("--single-branch")
    expect(cloneCall?.options.env).toMatchObject({ GIT_TERMINAL_PROMPT: "0", GCM_INTERACTIVE: "never" })
  })

  it("fetches an existing verified clone with an explicit branch ref and detached checkout", async () => {
    const projectRoot = await scratchDirectory()
    const cloneRoot = join(projectRoot, ".uf", "clones", "gaze")
    const state = commandState()
    await addRepository(state, cloneRoot, { "README.md": "existing clone" })
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ fetch: "clone-on-review", clone_path: ".uf/clones/gaze" })),
        state,
      ),
    )

    expect(result.siblings[0]).toMatchObject({ status: "available", source_mode: "clone", commit: COMMIT_A })
    expect(state.calls.find((call) => call.args[0] === "fetch")?.args).toEqual([
      "fetch",
      "--depth",
      "1",
      "--no-tags",
      "origin",
      "+refs/heads/main:refs/remotes/origin/main",
    ])
    expect(state.calls.find((call) => call.args[0] === "checkout")?.args).toEqual([
      "checkout",
      "--detach",
      "refs/remotes/origin/main",
    ])
  })

  it("uses the first reverified fallback as offline-cache after clone failure", async () => {
    const projectRoot = await scratchDirectory()
    const invalid = join(projectRoot, "invalid")
    const cache = join(projectRoot, "cache")
    const state = commandState()
    state.cloneFailure = true
    await addRepository(state, invalid, { "README.md": "wrong" }, { origin: "https://github.com/other/project" })
    await addRepository(state, cache, { "README.md": "cached" })
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(
          siblingFixture({
            fetch: "clone-on-review",
            local_paths: [invalid],
            clone_path: ".uf/clones/gaze",
            cache_path: "cache",
          }),
        ),
        state,
      ),
    )

    expect(result.siblings[0]).toMatchObject({
      status: "available",
      source_mode: "offline-cache",
      source_candidate: "cache_path",
    })
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toEqual(
      expect.arrayContaining(["network-acquisition-failed", "candidate-origin-mismatch"]),
    )
  })

  it("reports clone-on-review unavailable without blocking when network and candidates fail", async () => {
    const projectRoot = await scratchDirectory()
    const state = commandState()
    state.cloneFailure = true
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ fetch: "clone-on-review", clone_path: ".uf/clones/gaze" })),
        state,
      ),
    )

    expect(result.status).toBe("ok")
    expect(result.siblings[0]).toMatchObject({
      status: "unavailable",
      unavailable_reason: "network-failed-no-verified-candidate",
      evidence: [],
    })
  })

  it("rejects symlink and git-reported path escapes without reading their content", async () => {
    const projectRoot = await scratchDirectory()
    const repositoryRoot = join(projectRoot, "repository")
    const outsideRoot = join(projectRoot, "outside.md")
    const state = commandState()
    await writeFile(outsideRoot, "outside secret")
    await addRepository(state, repositoryRoot, { "README.md": "inside" })
    await mkdir(join(repositoryRoot, "docs"), { recursive: true })
    await symlink(outsideRoot, join(repositoryRoot, "docs", "escape.md"))
    const canonicalRoot = await realpath(repositoryRoot)
    state.repositories.set(canonicalRoot, {
      root: canonicalRoot,
      origin: "https://github.com/unbound-force/gaze.git",
      commit: COMMIT_A,
      tracked: ["README.md", "docs/escape.md", "../outside.md"],
    })

    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [repositoryRoot], contracts: ["**/*.md"] })),
        state,
      ),
    )

    expect(result.siblings[0]?.evidence).toEqual([])
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("tracked-files-unverifiable")
    expect(result.prompt).not.toContain("outside secret")
  })

  it("rejects a matching symlink that resolves outside an otherwise valid repository", async () => {
    const projectRoot = await scratchDirectory()
    const repositoryRoot = join(projectRoot, "repository")
    const outsideRoot = join(projectRoot, "outside.md")
    const state = commandState()
    await writeFile(outsideRoot, "outside secret")
    await addRepository(state, repositoryRoot, { "README.md": "inside" })
    await mkdir(join(repositoryRoot, "docs"), { recursive: true })
    await symlink(outsideRoot, join(repositoryRoot, "docs", "escape.md"))
    const repository = state.repositories.get(await realpath(repositoryRoot))!
    state.repositories.set(repository.root, { ...repository, tracked: ["README.md", "docs/escape.md"] })

    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [repositoryRoot], contracts: ["**/*.md"] })),
        state,
      ),
    )

    expect(result.siblings[0]?.evidence.map((item) => item.path)).toEqual(["README.md"])
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("file-symlink-escape")
    expect(result.prompt).not.toContain("outside secret")
  })

  it("rejects an in-root symlink whose resolved target is not commit-tracked", async () => {
    const projectRoot = await scratchDirectory()
    const repositoryRoot = join(projectRoot, "repository")
    const state = commandState()
    await addRepository(state, repositoryRoot, { "README.md": "inside" })
    await mkdir(join(repositoryRoot, "generated"), { recursive: true })
    await writeFile(join(repositoryRoot, "generated", "ignored.md"), "ignored target")
    await mkdir(join(repositoryRoot, "docs"), { recursive: true })
    await symlink(join(repositoryRoot, "generated", "ignored.md"), join(repositoryRoot, "docs", "contract.md"))
    const repository = state.repositories.get(await realpath(repositoryRoot))!
    state.repositories.set(repository.root, { ...repository, tracked: ["README.md", "docs/contract.md"] })

    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [repositoryRoot], contracts: ["docs/*.md"] })),
        state,
      ),
    )

    expect(result.siblings[0]?.evidence).toEqual([])
    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("file-target-untracked")
    expect(result.prompt).not.toContain("ignored target")
  })

  it("enforces individual, count, and total byte limits before prompt use", async () => {
    const projectRoot = await scratchDirectory()
    const state = commandState()
    const oversizedRoot = join(projectRoot, "oversized")
    await addRepository(state, oversizedRoot, { "docs/large.md": new Uint8Array(256 * 1024 + 1) })
    const oversized = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [oversizedRoot], contracts: ["docs/*.md"] })),
        state,
      ),
    )
    expect(oversized.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("file-size-limit")

    const countRoot = join(projectRoot, "count")
    const countFiles = Object.fromEntries(
      Array.from({ length: 21 }, (_value, index) => [`docs/${String(index).padStart(2, "0")}.md`, "x"]),
    )
    await addRepository(state, countRoot, countFiles)
    const counted = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [countRoot], contracts: ["docs/*.md"] })),
        state,
      ),
    )
    expect(counted.siblings[0]?.evidence).toHaveLength(20)
    expect(counted.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("file-count-limit")

    const totalRoot = join(projectRoot, "total")
    const chunk = new Uint8Array(220 * 1024)
    await addRepository(state, totalRoot, {
      "docs/1.md": chunk,
      "docs/2.md": chunk,
      "docs/3.md": chunk,
      "docs/4.md": chunk,
      "docs/5.md": chunk,
    })
    const total = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [totalRoot], contracts: ["docs/*.md"] })),
        state,
      ),
    )
    expect(total.siblings[0]?.evidence).toHaveLength(4)
    expect(total.siblings[0]?.rejections.map((item) => item.reason_code)).toContain("total-size-limit")
  })

  it("rejects invalid commit identities and reserved delimiter injection", async () => {
    const projectRoot = await scratchDirectory()
    const badCommitRoot = join(projectRoot, "bad-commit")
    const delimiterRoot = join(projectRoot, "delimiter")
    const state = commandState()
    await addRepository(state, badCommitRoot, { "README.md": "bad" }, { commit: "A".repeat(40) })
    await addRepository(state, delimiterRoot, { "README.md": `before ${UNTRUSTED_SIBLING_EVIDENCE_END} after` })
    const result = await acquireSiblingEvidence(
      {},
      dependencies(
        projectRoot,
        configuration(siblingFixture({ local_paths: [badCommitRoot, delimiterRoot] })),
        state,
      ),
    )

    expect(result.siblings[0]?.rejections.map((item) => item.reason_code)).toEqual(
      expect.arrayContaining(["candidate-invalid-commit", "reserved-delimiter"]),
    )
    expect(result.prompt).toBe("")
  })

  it("orders sibling results and prompt sections deterministically", async () => {
    const projectRoot = await scratchDirectory()
    const alphaRoot = join(projectRoot, "alpha")
    const zetaRoot = join(projectRoot, "zeta")
    const state = commandState()
    await addRepository(
      state,
      alphaRoot,
      { "README.md": "alpha" },
      { origin: "https://github.com/unbound-force/alpha" },
    )
    await addRepository(state, zetaRoot, { "README.md": "zeta" }, { origin: "https://github.com/unbound-force/zeta" })
    const config = configuration(
      siblingFixture({ name: "zeta", url: "https://github.com/unbound-force/zeta", local_paths: [zetaRoot] }),
      siblingFixture({ name: "alpha", url: "https://github.com/unbound-force/alpha", local_paths: [alphaRoot] }),
    )

    const first = await acquireSiblingEvidence({}, dependencies(projectRoot, config, state))
    const second = await acquireSiblingEvidence({}, dependencies(projectRoot, config, state))

    expect(first.siblings.map((item) => item.sibling)).toEqual(["alpha", "zeta"])
    expect(first.prompt).toBe(second.prompt)
    expect(first.prompt.indexOf("alpha")).toBeLessThan(first.prompt.indexOf("zeta"))
  })
})

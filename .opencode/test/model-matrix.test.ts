/**
 * Tests for the model-matrix lib module (OpenSpec change
 * model-matrix-delegation, task 2.1 — MM-FR-001 and MM-FR-004 scenarios,
 * design D1/D3/D8).
 *
 * Task 2.1 surface only: the closed v1 Zod schema, the tristate loader with
 * injected read/parse/existence seams, and pack existence validation. The
 * override whitelist, merge, null-model propagation, and precedence-chain
 * scenarios from task 2.4 land with tasks 2.2/2.3 and extend this file.
 */
import { describe, expect, it } from "vitest"

import {
  DEFAULT_TASK_KINDS,
  EMPTY_MODEL_MATRIX,
  MODEL_MATRIX_PATH,
  PACKS_DIRECTORY,
  SELECTION_SOURCE_VALUES,
  loadModelMatrix,
  parseModelMatrix,
  type ModelMatrixDependencies,
} from "../lib/model-matrix.js"

/** Starter tracked matrix per design D9: four empty capabilities, structure only. */
const STARTER_MATRIX = {
  version: 1,
  profiles: {
    "spec-authoring": {},
    "code-writing": {},
    "content-writing": {},
    mechanical: {},
  },
  packs: {
    "meta-prompt": "spec-authoring",
    go: "code-writing",
    typescript: "code-writing",
    python: "code-writing",
    content: "content-writing",
  },
  defaults: {
    plan: "spec-authoring",
    tasks: "spec-authoring",
    implement: "code-writing",
  },
  agents: {},
  classifier: {},
}

/** Pack names from the starter `packs` mapping. */
const STARTER_PACKS = ["meta-prompt", "go", "typescript", "python", "content"] as const

/**
 * Parser stub mirroring Bun.YAML for the shapes under test: comments-only and
 * empty documents parse to null; everything else is JSON (a superset of the
 * YAML the fixtures use, matching the review-dispatch test convention).
 */
function testParser(text: string): unknown {
  const content = text
    .split(/\r?\n/)
    .filter((line) => line.trim() !== "" && !line.trim().startsWith("#"))
    .join("\n")
  if (content === "") {
    return null
  }
  return JSON.parse(text) as unknown
}

/** Node-shaped ENOENT error so the injected reader reports an absent file. */
function enoent(what: string): NodeJS.ErrnoException {
  return Object.assign(new Error(`ENOENT: no such file or directory, ${what}`), { code: "ENOENT" })
}

type LoaderOverrides = {
  readonly matrixText?: string
  readonly absent?: boolean
  readonly readError?: unknown
  readonly packFiles?: readonly string[]
}

/**
 * Builds injectable loader dependencies plus the pack-probe log, so tests can
 * assert exactly which pack files were consulted.
 */
function loaderDependencies(overrides: LoaderOverrides = {}): {
  readonly dependencies: ModelMatrixDependencies
  readonly probed: readonly string[]
} {
  const packFiles = new Set(
    (overrides.packFiles ?? STARTER_PACKS).map((pack) => `${PACKS_DIRECTORY}/${pack}.md`),
  )
  const probed: string[] = []
  const dependencies: ModelMatrixDependencies = {
    readText: async (relativePath: string): Promise<string> => {
      if (relativePath !== MODEL_MATRIX_PATH) {
        throw new Error(`unexpected read: ${relativePath}`)
      }
      if (overrides.readError !== undefined) {
        throw overrides.readError
      }
      if (overrides.absent === true) {
        throw enoent(`open '${relativePath}'`)
      }
      return overrides.matrixText ?? JSON.stringify(STARTER_MATRIX)
    },
    parseYaml: testParser,
    fileExists: (relativePath: string): boolean => {
      probed.push(relativePath)
      return packFiles.has(relativePath)
    },
  }
  return { dependencies, probed }
}

/** Parses a matrix document through the throwing parse boundary. */
function parse(document: unknown): ReturnType<typeof parseModelMatrix> {
  return parseModelMatrix(JSON.stringify(document), testParser)
}

describe("model matrix closed v1 schema (MM-FR-001)", () => {
  it("should parse the starter tracked matrix with all sections present", () => {
    expect(parse(STARTER_MATRIX)).toEqual(STARTER_MATRIX)
  })

  it("should parse a minimal matrix with only version and one profile", () => {
    const minimal = parse({ version: 1, profiles: { mechanical: {} } })
    expect(minimal.version).toBe(1)
    expect(minimal.profiles).toEqual({ mechanical: {} })
    expect(minimal.packs).toBeUndefined()
    expect(minimal.defaults).toBeUndefined()
    expect(minimal.agents).toBeUndefined()
    expect(minimal.classifier).toBeUndefined()
  })

  it("should accept profiles that carry model and variant", () => {
    const matrix = parse({
      version: 1,
      profiles: { "code-writing": { model: "provider/model-id", variant: "high" } },
    })
    expect(matrix.profiles["code-writing"]).toEqual({ model: "provider/model-id", variant: "high" })
  })

  it("should reject unknown root keys", () => {
    expect(() => parse({ ...STARTER_MATRIX, unexpected: true })).toThrow('Unrecognized key: "unexpected"')
  })

  it("should reject unknown keys inside a profile", () => {
    const matrix = { ...STARTER_MATRIX, profiles: { "code-writing": { bogus: 1 } } }
    expect(() => parse(matrix)).toThrow('Unrecognized key: "bogus"')
  })

  it("should reject unknown keys inside classifier and classifier.llm", () => {
    expect(() => parse({ ...STARTER_MATRIX, classifier: { bogus: true } })).toThrow('Unrecognized key: "bogus"')
    expect(() => parse({ ...STARTER_MATRIX, classifier: { llm: { bogus: 1 } } })).toThrow('Unrecognized key: "bogus"')
  })

  it("should accept every closed classifier field and bounds-check the values", () => {
    expect(() =>
      parse({
        ...STARTER_MATRIX,
        classifier: {
          enabled: true,
          agent: "model-matrix-classifier",
          agent_timeout_ms: 45_000,
          llm: { endpoint: "https://example.invalid/v1", model: "classifier-model", apiKey: "{env:KEY}", request_timeout_ms: 30_000 },
          confidence_threshold: 0.6,
        },
      }),
    ).not.toThrow()
    expect(() => parse({ ...STARTER_MATRIX, classifier: { confidence_threshold: 1.5 } })).toThrow("Too big")
    expect(() => parse({ ...STARTER_MATRIX, classifier: { agent_timeout_ms: 0 } })).toThrow()
  })

  it("should reject an unknown defaults task kind naming the accepted key set", () => {
    expect(DEFAULT_TASK_KINDS).toEqual(["plan", "tasks", "implement"])
    expect(() => parse({ ...STARTER_MATRIX, defaults: { deploy: "mechanical" } })).toThrow(
      'Unrecognized key: "deploy" (accepted task kinds: plan, tasks, implement)',
    )
  })

  it("should reject a packs value naming an absent profile", () => {
    expect(() => parse({ ...STARTER_MATRIX, packs: { go: "no-such-capability" } })).toThrow(
      "packs.go: unknown profile no-such-capability",
    )
  })

  it("should reject a defaults value naming an absent profile", () => {
    expect(() => parse({ ...STARTER_MATRIX, defaults: { plan: "no-such-capability" } })).toThrow(
      "defaults.plan: unknown profile no-such-capability",
    )
  })

  it("should reject an agents value naming an absent profile", () => {
    expect(() => parse({ ...STARTER_MATRIX, agents: { "cobalt-crush-dev": "no-such-capability" } })).toThrow(
      "agents.cobalt-crush-dev: unknown profile no-such-capability",
    )
  })

  it("should reject reviewer agent pins while accepting non-reviewer pins", () => {
    expect(() => parse({ ...STARTER_MATRIX, agents: { "divisor-guard": "code-writing" } })).toThrow(
      "agents.divisor-guard: Invalid key in record",
    )
    expect(() => parse({ ...STARTER_MATRIX, agents: { "cobalt-crush-dev": "code-writing" } })).not.toThrow()
  })

  it("should reject an empty profiles map and a map above the 32-entry bound", () => {
    expect(() => parse({ version: 1, profiles: {} })).toThrow("at least one profile is required")
    const oversized: Record<string, unknown> = {}
    for (let index = 0; index < 33; index += 1) {
      oversized[`capability-${index}`] = {}
    }
    expect(() => parse({ version: 1, profiles: oversized })).toThrow("at most 32 profiles are allowed")
  })

  it("should reject version values other than 1", () => {
    expect(() => parse({ ...STARTER_MATRIX, version: 2 })).toThrow("Invalid input: expected 1")
  })

  it("should reject malformed names, model strings, and variants", () => {
    expect(() => parse({ version: 1, profiles: { "Bad_Name": {} } })).toThrow("Invalid key in record")
    expect(() => parse({ version: 1, profiles: { "code-writing": { model: "not-a-model" } } })).toThrow(
      "must match pattern",
    )
    expect(() => parse({ version: 1, profiles: { "code-writing": { variant: "!bad" } } })).toThrow(
      "must match pattern",
    )
    expect(() => parse({ ...STARTER_MATRIX, packs: { "Bad_Pack": "code-writing" } })).toThrow("Invalid key in record")
  })

  it("should reject YAML parse failures and merge references", () => {
    expect(() => parseModelMatrix("not-json", testParser)).toThrow("parse model matrix")
    expect(() => parseModelMatrix("profiles: &anchor {}\n", testParser)).toThrow(
      "model matrix contains forbidden YAML anchor",
    )
  })

  it("should treat comments-only and empty documents as the canonical no-op matrix (MM-FR-004)", () => {
    expect(parseModelMatrix("# tracked matrix is a no-op\n# second comment\n", testParser)).toBe(EMPTY_MODEL_MATRIX)
    expect(parseModelMatrix("", testParser)).toBe(EMPTY_MODEL_MATRIX)
  })

  it("should treat a parser that yields undefined for empty documents as a no-op", () => {
    // Parser variance: some YAML frontends normalize empty documents to
    // undefined instead of null; the parse boundary accepts both.
    expect(parseModelMatrix("", () => undefined)).toBe(EMPTY_MODEL_MATRIX)
  })
})

describe("loadModelMatrix tristate (MM-FR-004, design D8)", () => {
  it("should return absent when the matrix file is missing, probing no pack files", async () => {
    const { dependencies, probed } = loaderDependencies({ absent: true })
    await expect(loadModelMatrix(dependencies)).resolves.toEqual({ status: "absent" })
    expect(probed).toEqual([])
  })

  it("should return ok with the parsed matrix and probe every starter pack file", async () => {
    const { dependencies, probed } = loaderDependencies()
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.matrix).toEqual(STARTER_MATRIX)
    }
    expect(probed).toContain(`${PACKS_DIRECTORY}/go.md`)
    expect(probed).toHaveLength(STARTER_PACKS.length)
  })

  it("should return ok with the no-op matrix for a comments-only file without probing packs", async () => {
    const { dependencies, probed } = loaderDependencies({ matrixText: "# intentionally empty\n" })
    const result = await loadModelMatrix(dependencies)
    expect(result).toEqual({ status: "ok", matrix: EMPTY_MODEL_MATRIX })
    expect(probed).toEqual([])
  })

  it("should return ok for a minimal matrix with no packs section, probing no pack files", async () => {
    const { dependencies, probed } = loaderDependencies({
      matrixText: JSON.stringify({ version: 1, profiles: { mechanical: {} } }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.matrix.packs).toBeUndefined()
    }
    expect(probed).toEqual([])
  })

  it("should return invalid naming the path when validation fails", async () => {
    const { dependencies } = loaderDependencies({ matrixText: JSON.stringify({ ...STARTER_MATRIX, extra: 1 }) })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain(MODEL_MATRIX_PATH)
      expect(result.error).toContain('Unrecognized key: "extra"')
    }
  })

  it("should return invalid naming the offending pack key when its pack file is missing", async () => {
    const { dependencies } = loaderDependencies({ packFiles: STARTER_PACKS.filter((pack) => pack !== "go") })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain("packs key(s) go")
      expect(result.error).toContain(`${PACKS_DIRECTORY}/go.md`)
      expect(result.error).not.toContain("meta-prompt")
    }
  })

  it("should name every missing pack file, sorted, when several are missing", async () => {
    const { dependencies } = loaderDependencies({ packFiles: ["content"] })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain("packs key(s) go, meta-prompt, python, typescript")
      expect(result.error).toContain(
        `${PACKS_DIRECTORY}/go.md, ${PACKS_DIRECTORY}/meta-prompt.md, ${PACKS_DIRECTORY}/python.md, ${PACKS_DIRECTORY}/typescript.md`,
      )
    }
  })

  it("should return invalid when the read fails with a non-ENOENT error", async () => {
    const readError = Object.assign(new Error("EACCES: permission denied"), { code: "EACCES" })
    const { dependencies } = loaderDependencies({ readError })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain(`cannot read model matrix ${MODEL_MATRIX_PATH}`)
      expect(result.error).toContain("EACCES")
    }
  })

  it("should return invalid with a fallback message when the read rejects with a non-Error", async () => {
    const { dependencies } = loaderDependencies({ readError: "raw string failure" })
    const result = await loadModelMatrix(dependencies)
    expect(result).toEqual({
      status: "invalid",
      error: `cannot read model matrix ${MODEL_MATRIX_PATH}: unknown error`,
    })
  })
})

describe("selection audit surface (design D7, MM-FR-007)", () => {
  it("should expose the eight-value selection source enum", () => {
    expect(SELECTION_SOURCE_VALUES).toEqual([
      "explicit-model",
      "explicit-profile",
      "explicit-pack",
      "agent-pin",
      "default-kind",
      "classifier-agent",
      "classifier-llm",
      "host",
    ])
  })
})

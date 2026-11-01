/**
 * Tests for the model-matrix lib module (OpenSpec change
 * model-matrix-delegation — MM-FR-001 through MM-FR-004 scenarios,
 * design D1/D2/D3/D4/D8).
 *
 * Task 2.1 surface: the closed v1 Zod schema, the tristate loader with
 * injected read/parse/existence seams, and pack existence validation.
 * Task 2.2 surface: the override whitelist, nested-path merge, null-model
 * propagation, merged-result re-validation, and the loader override wiring.
 * Task 2.3 surface: the resolveSelection precedence chain — one proof test
 * per chain step and failure mode.
 * Task 2.4 rows: the tracked-role population rule — loader rejection of
 * populated tracked `classifier.llm` values naming the override file.
 * Task 2.5 rows: the plan-of-record reconciliation — profile-count and
 * agent-pin boundary rows, classifier timeout caps, cascade-bypass and
 * selection-level gating coverage, and the tracked-matrix extensibility
 * scenario (spec-review iteration-4 advisories 5, 7, 8, and 10). The
 * classifier cascade engine tests live in task 3.3's separate file
 * (model-matrix-classifier.test.ts), not here.
 */
import { describe, expect, it } from "vitest"

import {
  DEFAULT_TASK_KINDS,
  EMPTY_MODEL_MATRIX,
  MODEL_MATRIX_OVERRIDE_PATH,
  MODEL_MATRIX_PATH,
  PACKS_DIRECTORY,
  SELECTION_SOURCE_VALUES,
  loadModelMatrix,
  mergeOverride,
  parseModelMatrix,
  resolveSelection,
  validateModelMatrixOverride,
  type ClassifierOutcome,
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
  readonly overrideText?: string
  readonly overrideReadError?: unknown
}

/**
 * Builds injectable loader dependencies plus the pack-probe and file-read
 * logs, so tests can assert exactly which pack files were consulted and
 * which config paths were read. An undefined `overrideText` reports the
 * override as absent (ENOENT); an empty string reports it present but empty.
 */
function loaderDependencies(overrides: LoaderOverrides = {}): {
  readonly dependencies: ModelMatrixDependencies
  readonly probed: readonly string[]
  readonly reads: readonly string[]
} {
  const packFiles = new Set(
    (overrides.packFiles ?? STARTER_PACKS).map((pack) => `${PACKS_DIRECTORY}/${pack}.md`),
  )
  const probed: string[] = []
  const reads: string[] = []
  const dependencies: ModelMatrixDependencies = {
    readText: async (relativePath: string): Promise<string> => {
      reads.push(relativePath)
      if (relativePath === MODEL_MATRIX_PATH) {
        if (overrides.readError !== undefined) {
          throw overrides.readError
        }
        if (overrides.absent === true) {
          throw enoent(`open '${relativePath}'`)
        }
        return overrides.matrixText ?? JSON.stringify(STARTER_MATRIX)
      }
      if (relativePath === MODEL_MATRIX_OVERRIDE_PATH) {
        if (overrides.overrideReadError !== undefined) {
          throw overrides.overrideReadError
        }
        if (overrides.overrideText === undefined) {
          throw enoent(`open '${relativePath}'`)
        }
        return overrides.overrideText
      }
      throw new Error(`unexpected read: ${relativePath}`)
    },
    parseYaml: testParser,
    fileExists: (relativePath: string): boolean => {
      probed.push(relativePath)
      return packFiles.has(relativePath)
    },
  }
  return { dependencies, probed, reads }
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

  it("should accept matrices at the profile-count bounds (1 and 32 profiles)", () => {
    // Task 2.5 boundary row: the rejection test pins 0 and 33; the bounds
    // themselves must also accept their boundary values.
    const single = parse({ version: 1, profiles: { "code-writing": {} } })
    expect(Object.keys(single.profiles)).toHaveLength(1)

    const maximal: Record<string, unknown> = {}
    for (let index = 0; index < 32; index += 1) {
      maximal[`capability-${index}`] = {}
    }
    const bounded = parse({ version: 1, profiles: maximal })
    expect(Object.keys(bounded.profiles)).toHaveLength(32)
  })

  it("should bound the agent-pin count at 32 while accepting an empty agents map", () => {
    // Task 2.5 boundary row, reconciled with the landed schema per spec
    // review iteration-4 advisory 5: the bound is upper-only by design —
    // the tracked starter matrix ships `agents: {}` (design D9, task 4.1),
    // so zero pins MUST parse; the task text's "0 pins rejected" wording
    // was unimplementable as written and the landed semantics win.
    const pins = (count: number): Record<string, string> => {
      const map: Record<string, string> = {}
      for (let index = 0; index < count; index += 1) {
        map[`agent-${index}`] = "code-writing"
      }
      return map
    }
    const empty = parse({ ...STARTER_MATRIX, agents: {} })
    expect(empty.agents).toEqual({})

    const bounded = parse({ ...STARTER_MATRIX, agents: pins(32) })
    expect(Object.keys(bounded.agents ?? {})).toHaveLength(32)

    expect(() => parse({ ...STARTER_MATRIX, agents: pins(33) })).toThrow("at most 32 agent pins are allowed")
  })

  it("should bound classifier timeouts at their schema caps", () => {
    // Task 2.5 rows for the numeric caps the closed schema enforces
    // (spec-review iteration-4 advisory 6 records that the FR text states
    // only the defaults; these rows pin the enforced surface).
    expect(() => parse({ ...STARTER_MATRIX, classifier: { agent_timeout_ms: 1_800_000 } })).not.toThrow()
    expect(() => parse({ ...STARTER_MATRIX, classifier: { agent_timeout_ms: 1_800_001 } })).toThrow("Too big")
    expect(() => parse({ ...STARTER_MATRIX, classifier: { llm: { request_timeout_ms: 180_000 } } })).not.toThrow()
    expect(() => parse({ ...STARTER_MATRIX, classifier: { llm: { request_timeout_ms: 180_001 } } })).toThrow(
      "Too big",
    )
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

describe("override whitelist validation (MM-FR-002, design D4)", () => {
  it("should accept null, undefined, and empty-object documents as no-ops", () => {
    expect(() => validateModelMatrixOverride(null)).not.toThrow()
    expect(() => validateModelMatrixOverride(undefined)).not.toThrow()
    expect(() => validateModelMatrixOverride({})).not.toThrow()
  })

  it("should accept whitelisted profile model and variant leaves", () => {
    expect(() =>
      validateModelMatrixOverride({
        profiles: { "code-writing": { model: "provider/model-id", variant: "high" } },
      }),
    ).not.toThrow()
  })

  it("should accept any nested key under classifier.llm", () => {
    expect(() =>
      validateModelMatrixOverride({
        classifier: {
          llm: { endpoint: "https://example.invalid/v1", model: "classifier-model", apiKey: "{env:KEY}", request_timeout_ms: 30_000 },
        },
      }),
    ).not.toThrow()
  })

  it("should reject non-object documents naming the whitelist", () => {
    expect(() => validateModelMatrixOverride("just-a-string")).toThrow(
      "model matrix override must be an object — only profiles.*.model, profiles.*.variant, classifier.llm.* paths are allowed",
    )
    expect(() => validateModelMatrixOverride(["array"])).toThrow("model matrix override must be an object")
  })

  it("should reject unknown top-level keys naming the whitelist", () => {
    expect(() => validateModelMatrixOverride({ defaults: { implement: "mechanical" } })).toThrow(
      'model matrix override key "defaults" is not permitted — only profiles.*.model, profiles.*.variant, classifier.llm.* paths are allowed',
    )
  })

  it("should reject structural classifier keys outside llm naming the whitelist", () => {
    expect(() => validateModelMatrixOverride({ classifier: { enabled: false } })).toThrow(
      'model matrix override key "classifier.enabled" is not permitted — only profiles.*.model, profiles.*.variant, classifier.llm.* paths are allowed',
    )
    expect(() => validateModelMatrixOverride({ classifier: { agent: "other-agent" } })).toThrow(
      'model matrix override key "classifier.agent" is not permitted — only profiles.*.model, profiles.*.variant, classifier.llm.* paths are allowed',
    )
  })

  it("should reject non-whitelisted profile leaf keys naming the whitelist", () => {
    expect(() => validateModelMatrixOverride({ profiles: { "code-writing": { temperature: 0.2 } } })).toThrow(
      'model matrix override key "profiles.code-writing.temperature" is not permitted — only profiles.*.model, profiles.*.variant, classifier.llm.* paths are allowed',
    )
  })

  it("should reject non-object profiles, profile entries, classifier, and classifier.llm values", () => {
    expect(() => validateModelMatrixOverride({ profiles: "nope" })).toThrow("model matrix override profiles must be an object")
    expect(() => validateModelMatrixOverride({ profiles: { "code-writing": "nope" } })).toThrow(
      "model matrix override profiles.code-writing must be an object",
    )
    expect(() => validateModelMatrixOverride({ classifier: "nope" })).toThrow("model matrix override classifier must be an object")
    expect(() => validateModelMatrixOverride({ classifier: { llm: "nope" } })).toThrow(
      "model matrix override classifier.llm must be an object",
    )
  })

  it("should treat null profiles, profile entries, classifier, and classifier.llm as no-ops", () => {
    expect(() => validateModelMatrixOverride({ profiles: null })).not.toThrow()
    expect(() => validateModelMatrixOverride({ profiles: { "code-writing": null } })).not.toThrow()
    expect(() => validateModelMatrixOverride({ classifier: null })).not.toThrow()
    expect(() => validateModelMatrixOverride({ classifier: { llm: null } })).not.toThrow()
  })
})

describe("override nested-path merge (MM-FR-002, design D4)", () => {
  /** Base matrix with populated leaves so merges have something to replace. */
  const populatedBase = parse({
    ...STARTER_MATRIX,
    profiles: {
      "spec-authoring": {},
      "code-writing": { model: "provider/base-model", variant: "base" },
      "content-writing": {},
      mechanical: {},
    },
    classifier: { enabled: true, llm: { endpoint: "https://base.invalid/v1" } },
  })

  it("should replace the overridden model leaf and fall through untouched profiles", () => {
    const merged = mergeOverride(populatedBase, { profiles: { "spec-authoring": { model: "provider/override-model" } } })
    expect(merged.profiles["spec-authoring"]).toEqual({ model: "provider/override-model" })
    expect(merged.profiles["code-writing"]).toEqual({ model: "provider/base-model", variant: "base" })
    expect(merged.defaults).toEqual(STARTER_MATRIX.defaults)
    expect(merged.packs).toEqual(STARTER_MATRIX.packs)
  })

  it("should propagate an explicit null model as host fallback (absent key)", () => {
    // Design D4: the matching null model leaf replaces its base counterpart
    // (absent key → host fallback) while the variant leaf, absent from the
    // override, falls through to the base — the review-dispatch mergeOverride
    // precedent's exact semantics.
    const merged = mergeOverride(populatedBase, { profiles: { "code-writing": { model: null } } })
    expect(merged.profiles["code-writing"]).toEqual({ variant: "base" })
    expect(merged.profiles["code-writing"]).not.toHaveProperty("model")
  })

  it("should propagate an explicit null variant as an absent key", () => {
    const merged = mergeOverride(populatedBase, { profiles: { "code-writing": { variant: null } } })
    expect(merged.profiles["code-writing"]).toEqual({ model: "provider/base-model" })
  })

  it("should preserve base profile leaves the override does not mention", () => {
    const merged = mergeOverride(populatedBase, { profiles: { "code-writing": { variant: "override" } } })
    expect(merged.profiles["code-writing"]).toEqual({ model: "provider/base-model", variant: "override" })
  })

  it("should skip null profile entries while re-validating the equivalent merge", () => {
    const merged = mergeOverride(populatedBase, { profiles: { "code-writing": null } })
    expect(merged).toEqual(populatedBase)
  })

  it("should merge classifier.llm leaves while preserving base llm keys and sibling fields", () => {
    const merged = mergeOverride(populatedBase, {
      classifier: { llm: { model: "provider/override-model", apiKey: "{env:KEY}" } },
    })
    expect(merged.classifier).toEqual({
      enabled: true,
      llm: { endpoint: "https://base.invalid/v1", model: "provider/override-model", apiKey: "{env:KEY}" },
    })
  })

  it("should create classifier.llm when the base has no classifier section", () => {
    const base = parse({ version: 1, profiles: { "code-writing": {} } })
    const merged = mergeOverride(base, { classifier: { llm: { endpoint: "https://example.invalid/v1" } } })
    expect(merged.classifier).toEqual({ llm: { endpoint: "https://example.invalid/v1" } })
  })

  it("should propagate an explicit null classifier.llm leaf as an absent key", () => {
    const merged = mergeOverride(populatedBase, { classifier: { llm: { endpoint: null } } })
    expect(merged.classifier).toEqual({ enabled: true, llm: {} })
  })

  it("should fail closed naming the path when the override names an absent capability", () => {
    expect(() => mergeOverride(populatedBase, { profiles: { nonexistent: { model: "provider/model-id" } } })).toThrow(
      "model matrix override profiles.nonexistent is not a defined capability in the base matrix",
    )
  })

  it("should re-validate the merged result and fail closed on an invalid model string", () => {
    expect(() => mergeOverride(populatedBase, { profiles: { "code-writing": { model: "not-a-model" } } })).toThrow(
      "validate model matrix",
    )
  })

  it("should re-validate the merged result and fail closed on an unknown classifier.llm key", () => {
    expect(() => mergeOverride(populatedBase, { classifier: { llm: { bogus: true } } })).toThrow(
      'Unrecognized key: "bogus"',
    )
  })

  it("should fail closed when an override carries content against the no-op base matrix", () => {
    // A comments-only tracked matrix declares "no structure"; an override
    // carrying content contradicts that declaration. Fail closed rather
    // than silently ignoring per-developer configuration (design D4 policy).
    expect(() =>
      mergeOverride(EMPTY_MODEL_MATRIX, { profiles: { "spec-authoring": { model: "provider/model-id" } } }),
    ).toThrow("profiles.spec-authoring is not a defined capability")
    expect(() =>
      mergeOverride(EMPTY_MODEL_MATRIX, { classifier: { llm: { endpoint: "https://example.invalid/v1" } } }),
    ).toThrow("at least one profile is required")
  })

  it("should return the base unchanged for overrides carrying no whitelisted entries", () => {
    expect(mergeOverride(populatedBase, {})).toBe(populatedBase)
    expect(mergeOverride(populatedBase, { profiles: null })).toBe(populatedBase)
    expect(mergeOverride(populatedBase, { profiles: {} })).toBe(populatedBase)
    expect(mergeOverride(populatedBase, { classifier: { llm: {} } })).toBe(populatedBase)
  })

  it("should not mutate the base matrix during the merge", () => {
    const before = JSON.stringify(populatedBase)
    mergeOverride(populatedBase, {
      profiles: { "code-writing": { model: null } },
      classifier: { llm: { model: "provider/override-model" } },
    })
    expect(JSON.stringify(populatedBase)).toBe(before)
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

describe("loadModelMatrix override wiring (MM-FR-002, design D4/D8)", () => {
  it("should not read the override when the matrix is absent (design D8 zero-config)", async () => {
    const { dependencies, reads } = loaderDependencies({ absent: true })
    await expect(loadModelMatrix(dependencies)).resolves.toEqual({ status: "absent" })
    expect(reads).toEqual([MODEL_MATRIX_PATH])
  })

  it("should keep the base matrix when the override is absent", async () => {
    const { dependencies, reads } = loaderDependencies()
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.matrix).toEqual(STARTER_MATRIX)
    }
    expect(reads).toEqual([MODEL_MATRIX_PATH, MODEL_MATRIX_OVERRIDE_PATH])
  })

  it("should merge an override model into the loaded matrix (scenario: override supplies a profile model)", async () => {
    const { dependencies } = loaderDependencies({
      overrideText: JSON.stringify({ profiles: { "spec-authoring": { model: "some-provider/some-model" } } }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.matrix.profiles["spec-authoring"]).toEqual({ model: "some-provider/some-model" })
      expect(result.matrix.profiles["code-writing"]).toEqual({})
    }
  })

  it("should treat an empty or comments-only override as a no-op (scenario: empty override)", async () => {
    for (const overrideText of ["", "# comments only\n"]) {
      const { dependencies } = loaderDependencies({ overrideText })
      const result = await loadModelMatrix(dependencies)
      expect(result.status).toBe("ok")
      if (result.status === "ok") {
        expect(result.matrix).toEqual(STARTER_MATRIX)
      }
    }
  })

  it("should return invalid naming the whitelist for a non-whitelisted key, producing no merged configuration", async () => {
    const { dependencies } = loaderDependencies({
      overrideText: JSON.stringify({ defaults: { implement: "mechanical" } }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain(MODEL_MATRIX_OVERRIDE_PATH)
      expect(result.error).toContain('key "defaults" is not permitted')
      expect(result.error).toContain("profiles.*.model, profiles.*.variant, classifier.llm.*")
    }
  })

  it("should return invalid for a structural classifier override key (scenario: classifier.enabled)", async () => {
    const { dependencies } = loaderDependencies({
      overrideText: JSON.stringify({ classifier: { enabled: false } }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain('key "classifier.enabled" is not permitted')
    }
  })

  it("should return invalid naming the capability when the override references an absent profile", async () => {
    const { dependencies } = loaderDependencies({
      overrideText: JSON.stringify({ profiles: { nonexistent: { model: "provider/model-id" } } }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain("profiles.nonexistent is not a defined capability")
    }
  })

  it("should return invalid when the merged result fails re-validation", async () => {
    const { dependencies } = loaderDependencies({
      overrideText: JSON.stringify({ profiles: { "code-writing": { model: "not-a-model" } } }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain(MODEL_MATRIX_OVERRIDE_PATH)
      expect(result.error).toContain("must match pattern")
    }
  })

  it("should return invalid when the override read fails with a non-ENOENT error", async () => {
    const overrideReadError = Object.assign(new Error("EISDIR: illegal operation on a directory"), { code: "EISDIR" })
    const { dependencies } = loaderDependencies({ overrideReadError })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain(`cannot read model matrix override ${MODEL_MATRIX_OVERRIDE_PATH}`)
      expect(result.error).toContain("EISDIR")
    }
  })

  it("should return invalid when the override document fails to parse", async () => {
    const { dependencies } = loaderDependencies({ overrideText: "not-json" })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain(`model matrix override ${MODEL_MATRIX_OVERRIDE_PATH} is invalid`)
      expect(result.error).toContain("parse model matrix override")
    }
  })

  it("should return invalid when the override parses to a non-object document", async () => {
    const { dependencies } = loaderDependencies({ overrideText: JSON.stringify("just-a-string") })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain("model matrix override must be an object")
    }
  })
})

describe("tracked-role population rule (MM-FR-001, task 2.4)", () => {
  it("should fail load naming the override file for each populated tracked classifier.llm leaf", async () => {
    for (const [key, value] of [
      ["endpoint", "https://example.invalid/v1"],
      ["model", "provider/classifier-model"],
      ["apiKey", "{env:CLASSIFIER_KEY}"],
      ["request_timeout_ms", 30_000],
    ] as const) {
      const trackedLlm = parse({ ...STARTER_MATRIX, classifier: { llm: { [key]: value } } })
      const { dependencies, reads } = loaderDependencies({ matrixText: JSON.stringify(trackedLlm) })
      const result = await loadModelMatrix(dependencies)
      expect(result.status).toBe("invalid")
      if (result.status === "invalid") {
        expect(result.error).toContain("tracked classifier.llm carries populated values")
        expect(result.error).toContain(key)
        expect(result.error).toContain(MODEL_MATRIX_OVERRIDE_PATH)
      }
      // Fail-closed before the merge (MM-FR-001 scenario): no merged
      // configuration is produced, so the override is never read.
      expect(reads).toEqual([MODEL_MATRIX_PATH])
    }
  })

  it("should load the same classifier.llm values arriving via the override", async () => {
    const { dependencies } = loaderDependencies({
      overrideText: JSON.stringify({
        classifier: {
          llm: {
            endpoint: "https://example.invalid/v1",
            model: "provider/classifier-model",
            apiKey: "{env:CLASSIFIER_KEY}",
            request_timeout_ms: 30_000,
          },
        },
      }),
    })
    const result = await loadModelMatrix(dependencies)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.matrix.classifier?.llm).toEqual({
        endpoint: "https://example.invalid/v1",
        model: "provider/classifier-model",
        apiKey: "{env:CLASSIFIER_KEY}",
        request_timeout_ms: 30_000,
      })
    }
  })

  it("should load a structure-only tracked classifier.llm and the comments-only tracked file", async () => {
    const structureOnly = loaderDependencies({
      matrixText: JSON.stringify({ ...STARTER_MATRIX, classifier: { llm: {} } }),
    })
    const structureOnlyResult = await loadModelMatrix(structureOnly.dependencies)
    expect(structureOnlyResult.status).toBe("ok")

    // The comments-only tracked file resolves to the canonical no-op matrix
    // whose classifier carries no llm at all — the population rule leaves
    // the MM-FR-004 no-op valid (covered for parse in the schema block; the
    // loader row pins the full path).
    const noOp = loaderDependencies({ matrixText: "# intentionally empty\n" })
    const noOpResult = await loadModelMatrix(noOp.dependencies)
    expect(noOpResult).toEqual({ status: "ok", matrix: EMPTY_MODEL_MATRIX })
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

describe("resolveSelection precedence chain (MM-FR-003/MM-FR-004, design D2)", () => {
  /** Matrix with populated profiles, an agent pin, and the starter defaults and packs. */
  const selectionMatrix = parse({
    ...STARTER_MATRIX,
    profiles: {
      "spec-authoring": { model: "provider/spec-model" },
      "code-writing": { model: "provider/code-model" },
      "content-writing": { model: "provider/content-model" },
      mechanical: {},
    },
    agents: { "cobalt-crush-dev": "content-writing" },
  })

  /** Cascade outcome for calls where no classifier participated. */
  const notRun: ClassifierOutcome = { status: "not-run" }

  it("should let an explicit model win outright, bypassing pins, defaults, and the classifier", () => {
    const result = resolveSelection(
      selectionMatrix,
      { model: "provider/explicit-model", variant: "fast", checkpoint_step: "implement", agent: "cobalt-crush-dev" },
      { status: "selected", engine: "agent", pack: "go", confidence: 0.9 },
    )
    expect(result).toEqual({
      ok: true,
      selection: {
        selection_source: "explicit-model",
        // No explicit pack was supplied and the classifier cascade never
        // runs for a chain resolved at step 1 (MM-FR-003) — the synthetic
        // "selected" outcome above must not leak a pack directive.
        pack: null,
        capability: null,
        model: "provider/explicit-model",
        variant: "fast",
        confidence: null,
        fallback_reason: null,
      },
    })
  })

  it("should let an explicit profile beat an explicit pack for the model while retaining the pack", () => {
    const result = resolveSelection(selectionMatrix, { profile: "code-writing", pack: "content" }, notRun)
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("explicit-profile")
      expect(result.selection.capability).toBe("code-writing")
      expect(result.selection.model).toBe("provider/code-model")
      // The explicitly supplied pack is validated and retained for its
      // directive duty — never silently dropped (MM-FR-003, UF-FR-002).
      expect(result.selection.pack).toBe("content")
    }
  })

  it("should retain a valid explicit pack alongside an explicit model", () => {
    const result = resolveSelection(
      selectionMatrix,
      { model: "provider/explicit-model", pack: "go", checkpoint_step: "implement" },
      notRun,
    )
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("explicit-model")
      expect(result.selection.pack).toBe("go")
      expect(result.selection.capability).toBeNull()
      expect(result.selection.model).toBe("provider/explicit-model")
    }
  })

  it("should validate an explicit profile alongside an explicit model", () => {
    const result = resolveSelection(
      selectionMatrix,
      { model: "provider/explicit-model", profile: "content-writing" },
      notRun,
    )
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("explicit-model")
      expect(result.selection.capability).toBeNull()
      expect(result.selection.model).toBe("provider/explicit-model")
      expect(result.selection.pack).toBeNull()
    }
  })

  it("should fail closed on an invalid pack alongside an explicit model", () => {
    const result = resolveSelection(
      selectionMatrix,
      { model: "provider/explicit-model", pack: "no-such-pack" },
      notRun,
    )
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.code).toBe("INVALID_PACK")
      expect(result.retryable).toBe(false)
      expect(result.error).toContain('pack "no-such-pack" is not a packs key in the model matrix')
    }
  })

  it("should fail closed on an unknown profile alongside an explicit model", () => {
    const result = resolveSelection(
      selectionMatrix,
      { model: "provider/explicit-model", profile: "no-such-capability" },
      notRun,
    )
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.code).toBe("UNKNOWN_PROFILE")
      expect(result.retryable).toBe(false)
      expect(result.error).toContain('profile "no-such-capability" is not a capability declared in the model matrix')
    }
  })

  it("should fail closed on an invalid pack alongside a valid explicit profile", () => {
    const result = resolveSelection(selectionMatrix, { profile: "code-writing", pack: "no-such-pack" }, notRun)
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.code).toBe("INVALID_PACK")
      expect(result.retryable).toBe(false)
      expect(result.error).toContain('pack "no-such-pack" is not a packs key in the model matrix')
    }
  })

  it("should fail closed on combined explicit params without a matrix", () => {
    for (const matrix of [null, EMPTY_MODEL_MATRIX]) {
      for (const inputs of [
        { model: "provider/explicit-model", pack: "go" },
        { model: "provider/explicit-model", profile: "code-writing" },
      ]) {
        const result = resolveSelection(matrix, inputs, notRun)
        expect(result.ok).toBe(false)
        if (!result.ok) {
          expect(result.code).toBe("MATRIX_NOT_CONFIGURED")
          expect(result.retryable).toBe(false)
          expect(result.error).toContain("no model matrix is configured")
        }
      }
    }
  })

  it("should resolve an explicit pack through its mapped profile", () => {
    const result = resolveSelection(selectionMatrix, { pack: "go" }, notRun)
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("explicit-pack")
      expect(result.selection.pack).toBe("go")
      expect(result.selection.capability).toBe("code-writing")
      expect(result.selection.model).toBe("provider/code-model")
    }
  })

  it("should let an agent pin beat the defaults for implement", () => {
    const result = resolveSelection(
      selectionMatrix,
      { checkpoint_step: "implement", agent: "cobalt-crush-dev" },
      { status: "selected", engine: "llm", pack: "go", confidence: 0.9 },
    )
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("agent-pin")
      expect(result.selection.capability).toBe("content-writing")
      expect(result.selection.model).toBe("provider/content-model")
    }
  })

  it("should resolve plan and tasks statically, never consulting the classifier", () => {
    for (const checkpoint_step of ["plan", "tasks"] as const) {
      const result = resolveSelection(
        selectionMatrix,
        { checkpoint_step },
        { status: "selected", engine: "agent", pack: "go", confidence: 0.99 },
      )
      expect(result.ok).toBe(true)
      if (result.ok) {
        expect(result.selection.selection_source).toBe("default-kind")
        expect(result.selection.capability).toBe("spec-authoring")
        expect(result.selection.model).toBe("provider/spec-model")
        expect(result.selection.fallback_reason).toBeNull()
      }
    }
  })

  it("should resolve implement and unknown kinds through the classifier outcome", () => {
    const agentEngine = resolveSelection(
      selectionMatrix,
      { checkpoint_step: "implement" },
      { status: "selected", engine: "agent", pack: "go", confidence: 0.9 },
    )
    expect(agentEngine.ok).toBe(true)
    if (agentEngine.ok) {
      expect(agentEngine.selection.selection_source).toBe("classifier-agent")
      expect(agentEngine.selection.pack).toBe("go")
      expect(agentEngine.selection.capability).toBe("code-writing")
      expect(agentEngine.selection.model).toBe("provider/code-model")
      expect(agentEngine.selection.confidence).toBe(0.9)
    }
    const llmEngine = resolveSelection(
      selectionMatrix,
      { checkpoint_step: "analyze" },
      { status: "selected", engine: "llm", pack: "typescript", confidence: 0.8 },
    )
    expect(llmEngine.ok).toBe(true)
    if (llmEngine.ok) {
      expect(llmEngine.selection.selection_source).toBe("classifier-llm")
      expect(llmEngine.selection.pack).toBe("typescript")
      expect(llmEngine.selection.model).toBe("provider/code-model")
      expect(llmEngine.selection.confidence).toBe(0.8)
    }
  })

  it("should not leak a classifier outcome when an explicit profile or pack resolves the chain", () => {
    // Spec-review iteration-4 advisory 7 (lib level): an explicit model,
    // profile, or pack bypasses the cascade entirely, so a synthetic
    // selected outcome presented alongside explicit parameters must never
    // leak a pack directive. The model case is pinned above; these rows
    // pin profile and pack.
    const selected: ClassifierOutcome = { status: "selected", engine: "agent", pack: "go", confidence: 0.99 }

    const viaProfile = resolveSelection(
      selectionMatrix,
      { profile: "code-writing", checkpoint_step: "implement" },
      selected,
    )
    expect(viaProfile.ok).toBe(true)
    if (viaProfile.ok) {
      expect(viaProfile.selection.selection_source).toBe("explicit-profile")
      expect(viaProfile.selection.model).toBe("provider/code-model")
      expect(viaProfile.selection.pack).toBeNull()
    }

    const viaPack = resolveSelection(selectionMatrix, { pack: "content", checkpoint_step: "implement" }, selected)
    expect(viaPack.ok).toBe(true)
    if (viaPack.ok) {
      expect(viaPack.selection.selection_source).toBe("explicit-pack")
      expect(viaPack.selection.pack).toBe("content")
      expect(viaPack.selection.capability).toBe("content-writing")
      expect(viaPack.selection.model).toBe("provider/content-model")
    }
  })

  it("should discard a classifier outcome naming a matrix-invalid pack, falling back statically", () => {
    // Spec-review iteration-4 advisory 8: the selection-level
    // matrix-validity gate in resolveSelection (distinct from the cascade
    // engine's own check) — a pack outside the configured `packs` map is
    // discarded and the static default carries the auditable reason.
    const result = resolveSelection(
      selectionMatrix,
      { checkpoint_step: "implement" },
      { status: "selected", engine: "agent", pack: "no-such-pack", confidence: 0.99 },
    )
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("default-kind")
      expect(result.selection.model).toBe("provider/code-model")
      expect(result.selection.fallback_reason).toContain('classifier pack "no-such-pack" is not a packs key')
    }
  })

  it("should fall to host with the confidence recorded when a classifier-selected pack's profile carries no model", () => {
    // Advisory 8's second untested branch: an accepted classifier result
    // whose mapped profile is model-less resolves to host (step 7) with
    // the capability, pack, and confidence still recorded for the audit
    // trail (MM-FR-007).
    const matrix = parse({ ...STARTER_MATRIX, packs: { ops: "mechanical" } })
    const result = resolveSelection(
      matrix,
      { checkpoint_step: "implement" },
      { status: "selected", engine: "llm", pack: "ops", confidence: 0.8 },
    )
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("host")
      expect(result.selection.model).toBeNull()
      expect(result.selection.pack).toBe("ops")
      expect(result.selection.capability).toBe("mechanical")
      expect(result.selection.confidence).toBe(0.8)
      expect(result.selection.fallback_reason).toContain('profile "mechanical" carries no model')
    }
  })

  it("should resolve a non-starter capability added by matrix edit alone (MM-FR-001 extensibility scenario)", () => {
    // Spec-review iteration-4 advisory 10: the "new capability without
    // code changes" scenario — a fifth capability with its own pack
    // mapping and agent pin resolves through profile, pack, and pin
    // selections with no code changes. The override whitelist cannot add
    // capabilities (fail-closed elsewhere), so this row exercises
    // tracked-matrix extensibility specifically.
    const extended = parse({
      ...STARTER_MATRIX,
      profiles: {
        ...STARTER_MATRIX.profiles,
        "data-pipeline": { model: "provider/pipeline-model" },
      },
      packs: { ...STARTER_MATRIX.packs, airflow: "data-pipeline" },
      agents: { "data-worker-dev": "data-pipeline" },
    })

    const viaProfile = resolveSelection(extended, { profile: "data-pipeline" }, notRun)
    expect(viaProfile.ok).toBe(true)
    if (viaProfile.ok) {
      expect(viaProfile.selection.selection_source).toBe("explicit-profile")
      expect(viaProfile.selection.capability).toBe("data-pipeline")
      expect(viaProfile.selection.model).toBe("provider/pipeline-model")
    }

    const viaPack = resolveSelection(extended, { pack: "airflow" }, notRun)
    expect(viaPack.ok).toBe(true)
    if (viaPack.ok) {
      expect(viaPack.selection.selection_source).toBe("explicit-pack")
      expect(viaPack.selection.pack).toBe("airflow")
      expect(viaPack.selection.capability).toBe("data-pipeline")
      expect(viaPack.selection.model).toBe("provider/pipeline-model")
    }

    const viaPin = resolveSelection(extended, { agent: "data-worker-dev", checkpoint_step: "implement" }, notRun)
    expect(viaPin.ok).toBe(true)
    if (viaPin.ok) {
      expect(viaPin.selection.selection_source).toBe("agent-pin")
      expect(viaPin.selection.capability).toBe("data-pipeline")
      expect(viaPin.selection.model).toBe("provider/pipeline-model")
    }
  })

  it("should fall back to the static default when the classifier is below threshold or not run", () => {
    const belowThreshold = resolveSelection(
      selectionMatrix,
      { checkpoint_step: "implement" },
      { status: "selected", engine: "agent", pack: "go", confidence: 0.4 },
    )
    expect(belowThreshold.ok).toBe(true)
    if (belowThreshold.ok) {
      expect(belowThreshold.selection.selection_source).toBe("default-kind")
      expect(belowThreshold.selection.model).toBe("provider/code-model")
      expect(belowThreshold.selection.fallback_reason).toContain("below the threshold 0.6")
    }
    const skipped = resolveSelection(selectionMatrix, { checkpoint_step: "implement" }, notRun)
    expect(skipped.ok).toBe(true)
    if (skipped.ok) {
      expect(skipped.selection.selection_source).toBe("default-kind")
      expect(skipped.selection.fallback_reason).toContain("classifier not run")
    }
  })

  it("should fall back to host when the classifier fails with no static default for the kind", () => {
    const minimal = parse({ version: 1, profiles: { "code-writing": { model: "provider/code-model" } } })
    const result = resolveSelection(minimal, { checkpoint_step: "deploy" }, { status: "failed", reason: "agent timeout" })
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("host")
      expect(result.selection.model).toBeNull()
      expect(result.selection.fallback_reason).toContain("classifier cascade failed: agent timeout")
      expect(result.selection.fallback_reason).toContain('no static default for task kind "deploy"')
    }
  })

  it("should fall to host with a fallback reason when the resolved profile carries no model", () => {
    const result = resolveSelection(selectionMatrix, { profile: "mechanical" }, notRun)
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.selection.selection_source).toBe("host")
      expect(result.selection.capability).toBe("mechanical")
      expect(result.selection.model).toBeNull()
      expect(result.selection.fallback_reason).toContain('profile "mechanical" carries no model')
    }
  })

  it("should fail closed naming the profile when it is not declared", () => {
    const result = resolveSelection(selectionMatrix, { profile: "no-such-capability" }, notRun)
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.code).toBe("UNKNOWN_PROFILE")
      expect(result.retryable).toBe(false)
      expect(result.error).toContain('profile "no-such-capability" is not a capability declared in the model matrix')
      expect(result.error).toContain("Configured capabilities: code-writing, content-writing, mechanical, spec-authoring")
      expect(result.error).toContain("retrying will not fix this")
    }
  })

  it("should fail closed naming the pack when it is not a packs key", () => {
    const result = resolveSelection(selectionMatrix, { pack: "no-such-pack" }, notRun)
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.code).toBe("INVALID_PACK")
      expect(result.retryable).toBe(false)
      expect(result.error).toContain('pack "no-such-pack" is not a packs key in the model matrix')
      expect(result.error).toContain("Configured packs: content, go, meta-prompt, python, typescript")
    }
  })

  it("should fail closed stating no matrix is configured for explicit profile and pack", () => {
    for (const matrix of [null, EMPTY_MODEL_MATRIX]) {
      for (const inputs of [{ profile: "code-writing" }, { pack: "go" }]) {
        const result = resolveSelection(matrix, inputs, notRun)
        expect(result.ok).toBe(false)
        if (!result.ok) {
          expect(result.code).toBe("MATRIX_NOT_CONFIGURED")
          expect(result.retryable).toBe(false)
          expect(result.error).toContain("no model matrix is configured")
          expect(result.error).toContain(MODEL_MATRIX_PATH)
        }
      }
    }
  })

  it("should proceed on an explicit model without a matrix and host on zero params", () => {
    const explicit = resolveSelection(null, { model: "provider/explicit-model" }, notRun)
    expect(explicit.ok).toBe(true)
    if (explicit.ok) {
      expect(explicit.selection.selection_source).toBe("explicit-model")
      expect(explicit.selection.model).toBe("provider/explicit-model")
    }
    const zeroParams = resolveSelection(null, {}, notRun)
    expect(zeroParams.ok).toBe(true)
    if (zeroParams.ok) {
      expect(zeroParams.selection.selection_source).toBe("host")
      expect(zeroParams.selection.model).toBeNull()
      expect(zeroParams.selection.fallback_reason).toBeNull()
    }
    const noOpMatrix = resolveSelection(EMPTY_MODEL_MATRIX, {}, notRun)
    expect(noOpMatrix.ok).toBe(true)
    if (noOpMatrix.ok) {
      expect(noOpMatrix.selection.selection_source).toBe("host")
      expect(noOpMatrix.selection.fallback_reason).toBeNull()
    }
  })
})

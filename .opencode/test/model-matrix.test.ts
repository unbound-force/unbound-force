/**
 * Tests for the model-matrix lib module (OpenSpec change
 * model-matrix-delegation, tasks 2.1 + 2.2 — MM-FR-001, MM-FR-002, and
 * MM-FR-004 scenarios, design D1/D3/D4/D8).
 *
 * Task 2.1 surface: the closed v1 Zod schema, the tristate loader with
 * injected read/parse/existence seams, and pack existence validation.
 * Task 2.2 surface: the override whitelist, nested-path merge, null-model
 * propagation, merged-result re-validation, and the loader override wiring.
 * The precedence-chain and classifier scenarios from task 2.4 land with
 * tasks 2.3+ and extend this file.
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
  validateModelMatrixOverride,
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

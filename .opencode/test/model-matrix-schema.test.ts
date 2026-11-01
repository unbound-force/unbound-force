/**
 * JSON-Schema cross-validation harness for the model-matrix registry
 * (OpenSpec change model-matrix-delegation, task 2.6; design D3, MM-FR-008).
 *
 * Every `schemas/model-matrix/samples/` fixture is validated against the
 * canonical `v1.schema.json` AND against the runtime Zod surface
 * (parseModelMatrix), asserting identical verdicts so the two surfaces
 * cannot drift. The fixtures themselves land with task 4.2; this harness
 * landed first with the valid-starter-matrix fixture (the task 2.6
 * sanction) and discovers fixtures dynamically by name, so later fixtures
 * drop in without harness edits.
 *
 * Fixture families and their verdict contracts (naming convention mirrored
 * from the review-matrix samples registry, with disambiguating prefixes
 * for the surfaces that intentionally diverge):
 *
 * - `valid-*.json` — passes v1.schema.json AND the Zod surface.
 * - `invalid-*.json` (bare) — fails BOTH surfaces: closed-input, naming,
 *   and shared bound violations.
 * - `invalid-xref-*.json` — passes v1.schema.json, fails the Zod surface:
 *   cross-reference rules (mapped capabilities must be declared profiles)
 *   are runtime rules beyond plain JSON Schema expressiveness, so each
 *   surface's verdict is pinned separately.
 * - `invalid-loader-*.json` / `loader-rejection-*.json` — passes BOTH parse
 *   surfaces and fails only loadModelMatrix (pack-file existence, the
 *   tracked-role population rule).
 * - `invalid-override-*.json` — an override document, not a matrix: the
 *   Zod-side validateModelMatrixOverride must throw; JSON-Schema parity
 *   does not apply.
 *
 * The JSON-Schema side is checked by the shared compact validator in
 * test/helpers/json-schema-subset.ts (draft 2020-12 keyword subset: $ref,
 * type, const, enum, pattern, minLength, maxLength, minimum, maximum,
 * exclusiveMinimum, properties, patternProperties, additionalProperties,
 * propertyNames, required, minProperties, maxProperties) — extracted
 * from this file per default.md CS-004 when the tracked-matrix
 * conformance test (model-matrix-tracked.test.ts, task 4.1) needed to
 * validate the real tracked `.uf/model-matrix.yaml` against the same
 * surface. The suite's dependency
 * footprint stays pinned (zod + vitest only — no validator library), and
 * the self-verification probes below prove the validator rejects what the
 * Zod schema rejects, so the identical-verdict assertions have teeth even
 * while only the valid fixture has landed.
 */
import { readdirSync, readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

import { validateJsonSchema, type SchemaRoot, type SchemaVerdict } from "./helpers/json-schema-subset.js"

import {
  MODEL_MATRIX_OVERRIDE_PATH,
  MODEL_MATRIX_PATH,
  PACKS_DIRECTORY,
  loadModelMatrix,
  parseModelMatrix,
  validateModelMatrixOverride,
  type ModelMatrixDependencies,
} from "../lib/model-matrix.js"

/** Repository-root registry locations (test files live two levels deeper). */
const REGISTRY_ROOT = fileURLToPath(new URL("../../schemas/model-matrix/", import.meta.url))
const SCHEMA_PATH = `${REGISTRY_ROOT}v1.schema.json`
const SAMPLES_DIRECTORY = `${REGISTRY_ROOT}samples`

/**
 * Parses fixture text through the runtime Zod surface (parseModelMatrix
 * with a JSON parser), capturing the verdict without throwing.
 * @param text The fixture's JSON text.
 * @returns The verdict with the failure message, or valid.
 */
function zodVerdict(text: string): SchemaVerdict {
  try {
    parseModelMatrix(text, (source: string) => JSON.parse(source) as unknown)
    return { valid: true, error: null }
  } catch (error) {
    return { valid: false, error: error instanceof Error ? error.message : "unknown error" }
  }
}

/** Every sample fixture in the registry, sorted for deterministic runs. */
const FIXTURES = readdirSync(SAMPLES_DIRECTORY)
  .filter((name) => name.endsWith(".json"))
  .sort()

/** The fixture families the harness verdict contracts distinguish. */
type FixtureFamily = "valid" | "schema" | "xref" | "loader" | "override"

/**
 * Classifies a fixture by its documented name prefix. Specific divergent
 * families are matched before the bare `invalid-` prefix so a fixture
 * like `invalid-xref-absent-profile.json` never lands in the both-fail
 * family by accident; a name matching no documented family throws so the
 * discovery test fails loudly instead of silently skipping coverage.
 * @param name The fixture file name.
 * @returns The fixture's family.
 */
function fixtureFamily(name: string): FixtureFamily {
  if (name.startsWith("valid-")) {
    return "valid"
  }
  if (name.startsWith("invalid-xref-")) {
    return "xref"
  }
  if (name.startsWith("invalid-loader-") || name.startsWith("loader-rejection-")) {
    return "loader"
  }
  if (name.startsWith("invalid-override-")) {
    return "override"
  }
  if (name.startsWith("invalid-")) {
    return "schema"
  }
  throw new Error(`sample fixture "${name}" matches no documented family prefix`)
}

/** The fixture names of one family. */
function fixturesOf(family: FixtureFamily): readonly string[] {
  return FIXTURES.filter((name) => fixtureFamily(name) === family)
}

/** Reads a fixture's text and parses it as JSON. */
function readFixture(name: string): { readonly text: string; readonly document: unknown } {
  const text = readFileSync(`${SAMPLES_DIRECTORY}/${name}`, "utf8")
  return { text, document: JSON.parse(text) as unknown }
}

/** Starter pack files the loader family's fileExists seam reports. */
const STARTER_PACK_FILES = new Set(
  ["meta-prompt", "go", "typescript", "python", "content"].map((pack) => `${PACKS_DIRECTORY}/${pack}.md`),
)

/**
 * Builds loader dependencies serving one fixture as the tracked matrix
 * with no override and only the starter pack files present — the
 * loader-rejection family's environment.
 * @param text The fixture's JSON text served as the tracked matrix.
 * @returns Injectable loader dependencies.
 */
function loaderFor(text: string): ModelMatrixDependencies {
  return {
    readText: async (relativePath: string): Promise<string> => {
      if (relativePath === MODEL_MATRIX_PATH) {
        return text
      }
      if (relativePath === MODEL_MATRIX_OVERRIDE_PATH) {
        throw Object.assign(new Error(`ENOENT: no such file or directory, '${relativePath}'`), { code: "ENOENT" })
      }
      throw new Error(`unexpected read: ${relativePath}`)
    },
    parseYaml: (source: string) => JSON.parse(source) as unknown,
    fileExists: (relativePath: string): boolean => STARTER_PACK_FILES.has(relativePath),
  }
}

describe("model-matrix registry cross-validation (task 2.6, design D3, MM-FR-008)", () => {
  const schema = JSON.parse(readFileSync(SCHEMA_PATH, "utf8")) as SchemaRoot
  const starter = readFixture("valid-starter-matrix.json")

  it("should load v1.schema.json and discover sample fixtures, all classified", () => {
    expect(schema.$id).toBe("https://github.com/unbound-force/unbound-force/schemas/model-matrix/v1.schema.json")
    expect(FIXTURES.length).toBeGreaterThan(0)
    for (const name of FIXTURES) {
      expect(() => fixtureFamily(name)).not.toThrow()
    }
  })

  it("should carry the task 2.6 seed fixture (valid-starter-matrix.json)", () => {
    expect(FIXTURES).toContain("valid-starter-matrix.json")
  })

  it("valid fixtures pass v1.schema.json and the Zod surface (identical verdicts)", () => {
    const names = fixturesOf("valid")
    expect(names).toContain("valid-starter-matrix.json")
    for (const name of names) {
      const { text, document } = readFixture(name)
      const jsonVerdict = validateJsonSchema(schema, document)
      expect(jsonVerdict.valid, `${name}: ${jsonVerdict.error}`).toBe(true)
      const zod = zodVerdict(text)
      expect(zod.valid, `${name}: ${zod.error}`).toBe(true)
    }
  })

  it("schema-invalid fixtures fail v1.schema.json and the Zod surface (identical verdicts)", () => {
    for (const name of fixturesOf("schema")) {
      const { text, document } = readFixture(name)
      const jsonVerdict = validateJsonSchema(schema, document)
      expect(jsonVerdict.valid, `${name}: expected v1.schema.json rejection`).toBe(false)
      const zod = zodVerdict(text)
      expect(zod.valid, `${name}: expected Zod-surface rejection`).toBe(false)
    }
  })

  it("cross-reference fixtures pass v1.schema.json and fail the Zod surface", () => {
    // The divergence is the contract: cross-reference rules are runtime
    // rules the plain JSON Schema cannot express, so each side's verdict
    // is pinned (design D3's parity scope is the shared surface).
    for (const name of fixturesOf("xref")) {
      const { text, document } = readFixture(name)
      const jsonVerdict = validateJsonSchema(schema, document)
      expect(jsonVerdict.valid, `${name}: expected v1.schema.json pass`).toBe(true)
      const zod = zodVerdict(text)
      expect(zod.valid, `${name}: expected Zod-surface rejection`).toBe(false)
    }
  })

  it("loader-rejection fixtures pass both parse surfaces and fail loadModelMatrix", async () => {
    for (const name of fixturesOf("loader")) {
      const { text, document } = readFixture(name)
      const jsonVerdict = validateJsonSchema(schema, document)
      expect(jsonVerdict.valid, `${name}: expected v1.schema.json pass`).toBe(true)
      const zod = zodVerdict(text)
      expect(zod.valid, `${name}: expected Zod-surface pass`).toBe(true)
      const result = await loadModelMatrix(loaderFor(text))
      expect(result.status, `${name}: expected loader rejection`).toBe("invalid")
    }
  })

  it("override fixtures fail the override whitelist", () => {
    for (const name of fixturesOf("override")) {
      const { document } = readFixture(name)
      expect(() => validateModelMatrixOverride(document), `${name}: expected whitelist rejection`).toThrow()
    }
  })

  it("should reject closed-surface violations the Zod schema also rejects (harness self-verification)", () => {
    // Before task 4.2 lands the invalid fixtures, these probes prove the
    // harness's JSON-Schema validator actually rejects — a validator that
    // approved everything would otherwise pass the identical-verdict
    // assertions vacuously. Every probe must fail BOTH surfaces.
    const oversized: Record<string, unknown> = {}
    for (let index = 0; index < 33; index += 1) {
      oversized[`capability-${index}`] = {}
    }
    const probes: readonly (readonly [string, unknown])[] = [
      ["unknown root key", { ...starter.document, extra: 1 }],
      ["wrong version", { ...starter.document, version: 2 }],
      ["zero profiles", { version: 1, profiles: {} }],
      ["33 profiles", { version: 1, profiles: oversized }],
      ["unknown defaults kind", { ...starter.document, defaults: { deploy: "mechanical" } }],
      ["unknown profile leaf key", { version: 1, profiles: { "code-writing": { bogus: 1 } } }],
      ["malformed capability name", { version: 1, profiles: { Bad_Name: {} } }],
      ["malformed model string", { version: 1, profiles: { "code-writing": { model: "not-a-model" } } }],
      ["divisor agent pin", { ...starter.document, agents: { "divisor-guard": "code-writing" } }],
      ["unknown classifier.llm key", { ...starter.document, classifier: { llm: { bogus: 1 } } }],
      ["agent timeout over cap", { ...starter.document, classifier: { agent_timeout_ms: 1_800_001 } }],
      ["llm request timeout over cap", { ...starter.document, classifier: { llm: { request_timeout_ms: 180_001 } } }],
    ]
    for (const [label, document] of probes) {
      const jsonVerdict = validateJsonSchema(schema, document)
      expect(jsonVerdict.valid, `${label}: expected v1.schema.json rejection`).toBe(false)
      const zod = zodVerdict(JSON.stringify(document))
      expect(zod.valid, `${label}: expected Zod-surface rejection`).toBe(false)
    }
  })

  it("should exercise the loader's tracked-role rejection through the starter fixture", async () => {
    // Self-verification for the loader family: the seed fixture plus a
    // populated tracked classifier.llm must fail loadModelMatrix while
    // both parse surfaces accept it — the exact contract the
    // loader-rejection fixtures will pin (MM-FR-001 tracked-role rule).
    const populated = {
      ...starter.document,
      classifier: { llm: { endpoint: "https://example.invalid/v1" } },
    }
    const text = JSON.stringify(populated)
    expect(validateJsonSchema(schema, populated).valid).toBe(true)
    expect(zodVerdict(text).valid).toBe(true)
    const result = await loadModelMatrix(loaderFor(text))
    expect(result.status).toBe("invalid")
    if (result.status === "invalid") {
      expect(result.error).toContain("tracked classifier.llm carries populated values")
    }
  })
})

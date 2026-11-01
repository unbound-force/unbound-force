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
 * The JSON-Schema side is checked by a compact validator for the exact
 * keyword subset v1.schema.json uses (draft 2020-12: $ref, type, const,
 * enum, pattern, minLength, maxLength, minimum, maximum, exclusiveMinimum,
 * properties, patternProperties, additionalProperties, propertyNames,
 * required, minProperties, maxProperties). The suite's dependency
 * footprint stays pinned (zod + vitest only — no validator library), and
 * the self-verification probes below prove the validator rejects what the
 * Zod schema rejects, so the identical-verdict assertions have teeth even
 * while only the valid fixture has landed.
 */
import { readdirSync, readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

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
 * The JSON-Schema keyword subset v1.schema.json uses (draft 2020-12).
 * Unknown keywords are ignored, matching lenient validator behavior for
 * annotation keywords; every keyword the schema relies on is implemented.
 */
interface SchemaSubset {
  readonly $ref?: string
  readonly type?: "object" | "string" | "number" | "integer" | "boolean"
  readonly const?: unknown
  readonly enum?: readonly unknown[]
  readonly pattern?: string
  readonly minLength?: number
  readonly maxLength?: number
  readonly minimum?: number
  readonly maximum?: number
  readonly exclusiveMinimum?: number
  readonly properties?: Readonly<Record<string, SchemaSubset>>
  readonly patternProperties?: Readonly<Record<string, SchemaSubset>>
  readonly additionalProperties?: boolean | SchemaSubset
  readonly propertyNames?: SchemaSubset
  readonly required?: readonly string[]
  readonly minProperties?: number
  readonly maxProperties?: number
}

/** A schema document with its (optional) local definition table. */
interface SchemaRoot extends SchemaSubset {
  readonly $defs?: Readonly<Record<string, SchemaSubset>>
}

/** First-failure validation outcome: null error means valid. */
type SchemaVerdict = { readonly valid: boolean; readonly error: string | null }

/**
 * Resolves a local `#/$defs/<name>` reference against the root document.
 * @param root The full schema document carrying `$defs`.
 * @param ref The reference string used by v1.schema.json.
 * @returns The referenced sub-schema.
 * @throws When the reference is not a supported local $defs form.
 */
function resolveRef(root: SchemaRoot, ref: string): SchemaSubset {
  if (!ref.startsWith("#/$defs/")) {
    throw new Error(`harness does not support $ref form: ${ref}`)
  }
  const target = root.$defs?.[ref.slice("#/$defs/".length)]
  if (target === undefined) {
    throw new Error(`unknown $ref: ${ref}`)
  }
  return target
}

/**
 * Validates one value against one sub-schema, returning the first failure
 * as a path-qualified message or null when valid. Short-circuits on the
 * first failure — the harness asserts verdicts, not error inventories.
 * @param schema The sub-schema to apply.
 * @param root The root document for `$ref` resolution.
 * @param value The value under validation.
 * @param path The JSON-pointer-ish location for diagnostics.
 * @returns The first failure message, or null when the value is valid.
 */
function validateNode(schema: SchemaSubset, root: SchemaRoot, value: unknown, path: string): string | null {
  if (schema.$ref !== undefined) {
    return validateNode(resolveRef(root, schema.$ref), root, value, path)
  }

  if (schema.type !== undefined) {
    const matchesType =
      schema.type === "object"
        ? typeof value === "object" && value !== null && !Array.isArray(value)
        : schema.type === "string"
          ? typeof value === "string"
          : schema.type === "boolean"
            ? typeof value === "boolean"
            : schema.type === "number"
              ? typeof value === "number"
              : schema.type === "integer"
                ? typeof value === "number" && Number.isInteger(value)
                : false
    if (!matchesType) {
      return `${path}: expected type ${schema.type}`
    }
  }

  if ("const" in schema && value !== schema.const) {
    return `${path}: expected const ${JSON.stringify(schema.const)}`
  }
  if (schema.enum !== undefined && !schema.enum.some((option) => value === option)) {
    return `${path}: value not in enum`
  }

  if (typeof value === "string") {
    if (schema.pattern !== undefined && !new RegExp(schema.pattern).test(value)) {
      return `${path}: does not match pattern ${schema.pattern}`
    }
    if (schema.minLength !== undefined && value.length < schema.minLength) {
      return `${path}: shorter than minLength ${schema.minLength}`
    }
    if (schema.maxLength !== undefined && value.length > schema.maxLength) {
      return `${path}: longer than maxLength ${schema.maxLength}`
    }
  }

  if (typeof value === "number") {
    if (schema.minimum !== undefined && value < schema.minimum) {
      return `${path}: less than minimum ${schema.minimum}`
    }
    if (schema.maximum !== undefined && value > schema.maximum) {
      return `${path}: greater than maximum ${schema.maximum}`
    }
    if (schema.exclusiveMinimum !== undefined && value <= schema.exclusiveMinimum) {
      return `${path}: not greater than exclusiveMinimum ${schema.exclusiveMinimum}`
    }
  }

  if (typeof value === "object" && value !== null && !Array.isArray(value)) {
    const record = value as Readonly<Record<string, unknown>>
    const keys = Object.keys(record)
    for (const key of schema.required ?? []) {
      if (!Object.hasOwn(record, key)) {
        return `${path}: missing required property "${key}"`
      }
    }
    if (schema.minProperties !== undefined && keys.length < schema.minProperties) {
      return `${path}: fewer than minProperties ${schema.minProperties}`
    }
    if (schema.maxProperties !== undefined && keys.length > schema.maxProperties) {
      return `${path}: more than maxProperties ${schema.maxProperties}`
    }
    if (schema.propertyNames !== undefined) {
      for (const key of keys) {
        const nameError = validateNode(schema.propertyNames, root, key, `${path} (property name "${key}")`)
        if (nameError !== null) {
          return nameError
        }
      }
    }
    for (const key of keys) {
      let valueSchema: SchemaSubset | undefined = schema.properties?.[key]
      if (valueSchema === undefined && schema.patternProperties !== undefined) {
        for (const [pattern, patternSchema] of Object.entries(schema.patternProperties)) {
          if (new RegExp(pattern).test(key)) {
            valueSchema = patternSchema
            break
          }
        }
      }
      if (valueSchema === undefined) {
        // Draft 2020-12: `additionalProperties` governs keys not covered
        // by `properties`/`patternProperties` — `false` rejects them, a
        // schema validates them, and absence permits them.
        if (schema.additionalProperties === false) {
          return `${path}: additional property "${key}" is not allowed`
        }
        if (schema.additionalProperties !== undefined && schema.additionalProperties !== true) {
          const error = validateNode(schema.additionalProperties, root, record[key], `${path}.${key}`)
          if (error !== null) {
            return error
          }
        }
        continue
      }
      const error = validateNode(valueSchema, root, record[key], `${path}.${key}`)
      if (error !== null) {
        return error
      }
    }
  }

  return null
}

/**
 * Validates a parsed document against the full schema document.
 * @param root The schema document (with `$defs`).
 * @param document The parsed value under validation.
 * @returns The verdict with the first failure message, or valid.
 */
function validateJsonSchema(root: SchemaRoot, document: unknown): SchemaVerdict {
  const error = validateNode(root, root, document, "$")
  return { valid: error === null, error }
}

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

/**
 * Tracked-file conformance test for the repo-root `.uf/model-matrix.yaml`
 * (OpenSpec change model-matrix-delegation, task 4.1; design D9,
 * MM-FR-001).
 *
 * Loads the REAL tracked file through `loadModelMatrix`: the injected
 * reader serves the file's actual on-disk text, the parse seam runs the
 * fail-closed yaml-subset helper, and the existence seam probes the
 * REAL pack files at their repository locations — so every starter
 * `packs` key must exist on disk, including task 1.1's meta-prompt.md.
 * The per-developer override is reported absent: the conformance
 * contract is the tracked file's own load, never a machine-local
 * merged view.
 *
 * MM-FR-001's starter-mapping rule lives HERE by design — in the
 * tracked-file verification test, not in the schema or the loader:
 * arbitrary conforming matrices MAY map `plan` and `tasks` to any
 * present profile and the loader MUST NOT reject such a mapping, so
 * the obligation is pinned as a test-level conformance check (spec
 * scenario "Starter plan/tasks mapping is pinned"). The drifted-mapping
 * scenarios prove the split: a drifted document still passes
 * `schemas/model-matrix/v1.schema.json` and still loads cleanly, and
 * only the conformance check fails, naming the drifted mapping and the
 * rule it violates.
 */
import { existsSync, readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

import { validateJsonSchema, type SchemaRoot } from "./helpers/json-schema-subset.js"
import { parseYamlSubset } from "./helpers/yaml-subset.js"
import {
  MODEL_MATRIX_OVERRIDE_PATH,
  MODEL_MATRIX_PATH,
  loadModelMatrix,
  type ModelMatrix,
  type ModelMatrixDependencies,
} from "../lib/model-matrix.js"

/** Repository root (test files live two levels beneath it). */
const REPOSITORY_ROOT = fileURLToPath(new URL("../../", import.meta.url))

/** The real tracked matrix on disk (design D9's starter shape). */
const TRACKED_MATRIX_FILE = `${REPOSITORY_ROOT}${MODEL_MATRIX_PATH}`

/** The canonical JSON Schema the closed v1 surface mirrors (design D3). */
const SCHEMA_FILE = fileURLToPath(new URL("../../schemas/model-matrix/v1.schema.json", import.meta.url))

/** The tracked file's text, read once per run from disk. */
const TRACKED_TEXT = readFileSync(TRACKED_MATRIX_FILE, "utf8")

/** The canonical v1 schema document the drifted scenarios validate against. */
const SCHEMA = JSON.parse(readFileSync(SCHEMA_FILE, "utf8")) as SchemaRoot

/** The pinned starter capability set (design D9). */
const STARTER_CAPABILITIES = ["spec-authoring", "code-writing", "content-writing", "mechanical"] as const

/**
 * The pinned starter `packs` mapping (design D9). Exact equality is the
 * drift guard, and it carries the `mechanical` explicit-profile-only
 * fact: no pack's capability is `mechanical`, so only an explicit
 * `profile: mechanical` parameter can select it.
 */
const STARTER_PACKS: Readonly<Record<string, string>> = {
  "meta-prompt": "spec-authoring",
  go: "code-writing",
  typescript: "code-writing",
  python: "code-writing",
  content: "content-writing",
}

/**
 * The pinned starter `defaults` mapping — the MM-FR-001 conformance
 * rule: `plan` and `tasks` MUST map to `spec-authoring` so meta-prompt
 * authoring work runs on the meta-prompt model, and `implement` maps to
 * `code-writing` as the classifier cascade's static fallback (design D9).
 */
const STARTER_DEFAULTS: Readonly<Record<"plan" | "tasks" | "implement", string>> = {
  plan: "spec-authoring",
  tasks: "spec-authoring",
  implement: "code-writing",
}

/** Node-shaped ENOENT error so the injected reader reports an absent file. */
function enoent(relativePath: string): NodeJS.ErrnoException {
  return Object.assign(new Error(`ENOENT: no such file or directory, '${relativePath}'`), { code: "ENOENT" })
}

/**
 * Builds loader dependencies for the tracked-file conformance run: the
 * given text serves as the tracked matrix, the parse seam runs the
 * fail-closed yaml-subset helper, the existence seam probes the REAL
 * pack files at their repository locations, and the per-developer
 * override is reported absent (the merged view is machine-local state
 * the conformance contract must not depend on).
 * @param matrixText The tracked matrix text under load.
 * @returns Injectable loader dependencies.
 */
function trackedLoader(matrixText: string): ModelMatrixDependencies {
  return {
    readText: async (relativePath: string): Promise<string> => {
      if (relativePath === MODEL_MATRIX_PATH) {
        return matrixText
      }
      if (relativePath === MODEL_MATRIX_OVERRIDE_PATH) {
        throw enoent(relativePath)
      }
      throw new Error(`unexpected read: ${relativePath}`)
    },
    parseYaml: parseYamlSubset,
    fileExists: (relativePath: string): boolean => existsSync(`${REPOSITORY_ROOT}${relativePath}`),
  }
}

/**
 * Runs the tracked-file conformance check and asserts the failure names
 * every expected fragment, so a drifted mapping is reported with its
 * path and the rule it violates rather than an anonymous mismatch.
 * @param matrix The drifted matrix under check.
 * @param fragments Message fragments the failure MUST name.
 */
function expectConformanceFailure(matrix: ModelMatrix, fragments: readonly string[]): void {
  let failure: unknown
  try {
    checkTrackedStarterConformance(matrix)
  } catch (error: unknown) {
    failure = error
  }
  if (!(failure instanceof Error)) {
    throw new Error("the tracked-file conformance check must reject the drifted starter matrix")
  }
  for (const fragment of fragments) {
    expect(failure.message, `the conformance failure must name "${fragment}"`).toContain(fragment)
  }
}

/**
 * The tracked-file conformance check (MM-FR-001, design D9): asserts a
 * matrix loaded from the tracked file's text carries the pinned starter
 * shape — the four starter capabilities, every tracked profile empty
 * (no model strings; per-developer identities arrive via the override),
 * the exact starter `packs` mapping, the `plan`/`tasks` →
 * `spec-authoring` and `implement` → `code-writing` `defaults` mapping,
 * and no agent pins. Field-level failures name the drifted mapping and
 * the rule it violates.
 * @param matrix A matrix loaded from the tracked file's text.
 * @throws When any pinned starter obligation is violated.
 */
function checkTrackedStarterConformance(matrix: ModelMatrix): void {
  for (const capability of STARTER_CAPABILITIES) {
    if (!Object.hasOwn(matrix.profiles, capability)) {
      throw new Error(`tracked starter matrix must declare the starter capability "${capability}" (design D9)`)
    }
  }
  for (const [capability, profile] of Object.entries(matrix.profiles)) {
    const populated = Object.keys(profile)
    if (populated.length > 0) {
      throw new Error(
        `tracked profile "${capability}" carries ${populated.join(", ")} — the tracked file is structure only; ` +
          `per-developer identities arrive via .uf/model-matrix.override.yaml (MM-FR-001)`,
      )
    }
  }
  const packs = matrix.packs ?? {}
  const actualPackKeys = Object.keys(packs).sort()
  const expectedPackKeys = Object.keys(STARTER_PACKS).sort()
  if (actualPackKeys.join(",") !== expectedPackKeys.join(",")) {
    throw new Error(
      `tracked starter packs keys are [${actualPackKeys.join(", ")}] but the starter set pins ` +
        `[${expectedPackKeys.join(", ")}] (design D9)`,
    )
  }
  for (const [pack, capability] of Object.entries(STARTER_PACKS)) {
    if (packs[pack] !== capability) {
      throw new Error(
        `tracked starter packs.${pack} maps to "${packs[pack]}" but the starter set pins "${capability}" (design D9)`,
      )
    }
  }
  const defaults = matrix.defaults ?? {}
  for (const kind of ["plan", "tasks", "implement"] as const) {
    const required = STARTER_DEFAULTS[kind]
    const actual = defaults[kind]
    if (actual !== required) {
      throw new Error(
        `tracked starter matrix conformance rule (MM-FR-001, design D9): defaults.${kind} maps to ` +
          `${JSON.stringify(actual)} but the starter matrix pins "${required}" — plan and tasks must run ` +
          `meta-prompt authoring work on the spec-authoring model, and implement must fall back to code-writing`,
      )
    }
  }
  const pins = Object.keys(matrix.agents ?? {})
  if (pins.length > 0) {
    throw new Error(`tracked starter matrix pins agents [${pins.join(", ")}] but the starter set carries none (design D9)`)
  }
}

describe("tracked starter matrix conformance (task 4.1, design D9, MM-FR-001)", () => {
  it("loads the real tracked .uf/model-matrix.yaml cleanly through loadModelMatrix", async () => {
    const result = await loadModelMatrix(trackedLoader(TRACKED_TEXT))
    if (result.status !== "ok") {
      throw new Error(`the real tracked matrix must load cleanly: ${result.error}`)
    }
    expect(result.matrix.version).toBe(1)
  })

  it("validates the real tracked document against schemas/model-matrix/v1.schema.json", () => {
    const verdict = validateJsonSchema(SCHEMA, parseYamlSubset(TRACKED_TEXT))
    expect(verdict.valid, verdict.error ?? "v1.schema.json rejected the tracked document").toBe(true)
  })

  it("conforms to the pinned starter shape and the plan/tasks conformance rule", async () => {
    const result = await loadModelMatrix(trackedLoader(TRACKED_TEXT))
    if (result.status !== "ok") {
      throw new Error(`the real tracked matrix must load cleanly: ${result.error}`)
    }
    expect(() => checkTrackedStarterConformance(result.matrix)).not.toThrow()
  })

  it("rejects a drifted defaults.plan mapping while the document stays schema-valid (MM-FR-001 starter-mapping scenario)", async () => {
    const driftedText = TRACKED_TEXT.replace("  plan: spec-authoring", "  plan: code-writing")
    expect(driftedText, "the plan drift must apply to the tracked text").not.toBe(TRACKED_TEXT)
    const verdict = validateJsonSchema(SCHEMA, parseYamlSubset(driftedText))
    expect(
      verdict.valid,
      "the drifted mapping must still pass v1.schema.json — the rule is a tracked-file obligation, not a schema constraint",
    ).toBe(true)
    const result = await loadModelMatrix(trackedLoader(driftedText))
    if (result.status !== "ok") {
      throw new Error(`the loader must accept the drifted but conforming mapping: ${result.error}`)
    }
    expectConformanceFailure(result.matrix, ["defaults.plan", "spec-authoring"])
  })

  it("rejects a drifted defaults.implement mapping while the document stays schema-valid", async () => {
    const driftedText = TRACKED_TEXT.replace("  implement: code-writing", "  implement: content-writing")
    expect(driftedText, "the implement drift must apply to the tracked text").not.toBe(TRACKED_TEXT)
    const verdict = validateJsonSchema(SCHEMA, parseYamlSubset(driftedText))
    expect(
      verdict.valid,
      "the drifted mapping must still pass v1.schema.json — the rule is a tracked-file obligation, not a schema constraint",
    ).toBe(true)
    const result = await loadModelMatrix(trackedLoader(driftedText))
    if (result.status !== "ok") {
      throw new Error(`the loader must accept the drifted but conforming mapping: ${result.error}`)
    }
    expectConformanceFailure(result.matrix, ["defaults.implement", "code-writing"])
  })
})

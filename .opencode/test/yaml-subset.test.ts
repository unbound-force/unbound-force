/**
 * Self-verification probes for the fail-closed YAML-subset helper
 * (OpenSpec change model-matrix-delegation, task 4.1).
 *
 * The tracked-file conformance test (model-matrix-tracked.test.ts)
 * parses the real `.uf/model-matrix.yaml` through this helper, and its
 * result is double-validated by the closed Zod surface and
 * v1.schema.json — but those checks cannot prove the helper THROWS on
 * syntax outside its subset, the property that keeps a tracked-file
 * edit from silently misparsing. These probes pin both directions:
 * the documented subset parses to the exact expected document, and
 * every out-of-subset form fails loudly (the model-matrix-schema
 * harness self-verification precedent).
 */
import { describe, expect, it } from "vitest"

import { parseYamlSubset } from "./helpers/yaml-subset.js"

/** A representative sample of the tracked file's syntax, in the documented subset. */
const SAMPLE = [
  "# leading comment",
  "version: 1",
  "",
  "profiles:",
  "  spec-authoring: {}",
  "  mechanical: {}",
  "packs:",
  "  meta-prompt: spec-authoring",
  "  go: code-writing",
  "defaults:",
  "  plan: spec-authoring",
  "  implement: code-writing",
  "agents: {}",
  "classifier: {}",
  "# trailing comment",
].join("\n")

describe("yaml-subset helper (task 4.1 parse seam)", () => {
  it("parses the documented subset to the exact expected document", () => {
    expect(parseYamlSubset(SAMPLE)).toEqual({
      version: 1,
      profiles: { "spec-authoring": {}, mechanical: {} },
      packs: { "meta-prompt": "spec-authoring", go: "code-writing" },
      defaults: { plan: "spec-authoring", implement: "code-writing" },
      agents: {},
      classifier: {},
    })
  })

  it("parses empty and comments-only documents to null (MM-FR-004 no-op shape)", () => {
    expect(parseYamlSubset("")).toBeNull()
    expect(parseYamlSubset("# one comment\n\n# another\n")).toBeNull()
  })

  it("throws on syntax outside the subset instead of guessing", () => {
    const probes: readonly (readonly [string, string])[] = [
      ["list item", "packs:\n  - go\n"],
      ["tab indentation", "packs:\n\tgo: code-writing\n"],
      ["quoted scalar", "defaults:\n  plan: \"spec-authoring\"\n"],
      ["trailing comment", "defaults:\n  plan: spec-authoring # note\n"],
      ["value with spaces", "defaults:\n  plan: spec authoring\n"],
      ["value with a colon", "classifier:\n  llm:\n    endpoint: https://example.invalid\n"],
      ["inconsistent indent step", "packs:\n  go: code-writing\n    extra: deep\n"],
      ["stray deeper line", "version: 1\n  nested: surprise\n"],
      ["block with no entries", "packs:\nversion: 1\n"],
      ["duplicate key", "version: 1\nversion: 2\n"],
      ["line without a key", "just text\n"],
      ["bare numeric line", "-1\n"],
    ]
    for (const [label, text] of probes) {
      expect(() => parseYamlSubset(text), `${label}: expected a loud failure`).toThrow(/yaml subset:/)
    }
  })
})

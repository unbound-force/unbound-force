import { describe, expect, it } from "vitest"

import {
  parseReviewMatrix,
  parseReviewerManifest,
  planReviewDispatch,
  type PlannerDependencies,
} from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

function dependencies(matrix: unknown = matrixFixture, manifest: unknown = manifestFixture): PlannerDependencies {
  const documents: Readonly<Record<string, string>> = {
    ".uf/review-matrix.yaml": JSON.stringify(matrix),
    ".uf/reviewer-capabilities.yaml": JSON.stringify(manifest),
  }
  return {
    readText: async (relativePath: string): Promise<string> => {
      const document = documents[relativePath]
      if (document === undefined) throw new Error(`unexpected policy read: ${relativePath}`)
      return document
    },
    parseYaml: jsonParser,
  }
}

function changedFile(path: string): { path: string; additions: number; deletions: number } {
  return { path, additions: 5, deletions: 0 }
}

describe("plan_review_dispatch input validation boundaries", () => {
  it("rejects duplicate discovered agents", async () => {
    await expect(
      planReviewDispatch(
        { mode: "code", discovered_agents: ["divisor-guard", "divisor-guard"], changed_files: [changedFile("a.go")] },
        dependencies(),
      ),
    ).rejects.toThrow("items must be unique")
  })

  it("requires changed files for non-triage modes", async () => {
    await expect(
      planReviewDispatch({ mode: "code", discovered_agents: ["divisor-guard"] }, dependencies()),
    ).rejects.toThrow("requires changed files")
  })

  it("requires issue content for triage mode", async () => {
    await expect(
      planReviewDispatch({ mode: "triage", discovered_agents: ["divisor-guard"] }, dependencies()),
    ).rejects.toThrow("requires issue content")
  })

  it("rejects a triage mode that synthesizes a diff", async () => {
    await expect(
      planReviewDispatch(
        {
          mode: "triage",
          discovered_agents: ["divisor-guard"],
          issue: { title: "t", body: null, comments: [] },
          changed_files: [changedFile("a.go")],
        },
        dependencies(),
      ),
    ).rejects.toThrow("must not synthesize a diff")
  })

  it("maps test mode to the code matrix mode", async () => {
    const plan = await planReviewDispatch(
      { mode: "test", discovered_agents: ["divisor-testing"], changed_files: [changedFile("a_test.go")] },
      dependencies(),
    )
    expect(plan.mode).toBe("test")
    expect(plan.matrix_mode).toBe("code")
  })
})

describe("review matrix and manifest schema superRefine boundaries", () => {
  it("rejects an unknown profile referenced by a default", () => {
    const matrix = structuredClone(matrixFixture)
    ;(matrix as { defaults: Record<string, string> }).defaults.code = "missing"
    expect(() => parseReviewMatrix(JSON.stringify(matrix), jsonParser)).toThrow("unknown profile")
  })

  it("rejects an unknown profile referenced by always", () => {
    const matrix = structuredClone(matrixFixture)
    ;(matrix as { always: Record<string, string> }).always["divisor-adversary"] = "missing"
    expect(() => parseReviewMatrix(JSON.stringify(matrix), jsonParser)).toThrow("unknown profile")
  })

  it("rejects a review-capable reviewer with empty scopes", () => {
    const manifest = structuredClone(manifestFixture)
    ;(manifest as { reviewers: unknown[] }).reviewers.push({ agent: "divisor-empty", capability: "review", scopes: [] })
    expect(() => parseReviewerManifest(JSON.stringify(manifest), jsonParser)).toThrow("review scopes must not be empty")
  })

  it("rejects a content-capable reviewer with scopes", () => {
    const manifest = structuredClone(manifestFixture)
    ;(manifest as { reviewers: unknown[] }).reviewers.push({
      agent: "divisor-content",
      capability: "content",
      scopes: ["standard"],
    })
    expect(() => parseReviewerManifest(JSON.stringify(manifest), jsonParser)).toThrow("content scopes must be empty")
  })
})

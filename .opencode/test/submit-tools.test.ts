import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it } from "vitest"

import { ReviewDispatchPlugin, _getSubmissionStore } from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

type BunRuntime = { readonly YAML: { readonly parse: (text: string) => unknown } }

function withBun<T>(run: () => T | Promise<T>): Promise<T> {
  const runtime = globalThis as typeof globalThis & { Bun?: BunRuntime }
  const original = runtime.Bun
  runtime.Bun = { YAML: { parse: (text: string): unknown => JSON.parse(text) as unknown } }
  const finish = (): void => {
    runtime.Bun = original
  }
  try {
    return Promise.resolve(run()).finally(finish)
  } catch (error) {
    finish()
    throw error
  }
}

const scratchDirectories: string[] = []

async function scratchProject(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-submit-tools-"))
  scratchDirectories.push(directory)
  await mkdir(join(directory, ".opencode"), { recursive: true })
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(join(directory, ".uf", "sibling-repos.yaml"), JSON.stringify({ version: 1, siblings: [] }))
  return directory
}

function childContext(sessionID: string): ToolContext {
  return {
    sessionID,
    messageID: "msg-child",
    agent: "divisor-guard",
    directory: "/workspace",
    worktree: "/workspace",
    abort: new AbortController().signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("not used")
    }) as ToolContext["ask"],
  }
}

afterEach(async () => {
  _getSubmissionStore().clear()
  await Promise.all(scratchDirectories.splice(0).map((d) => rm(d, { recursive: true, force: true })))
})

beforeEach(() => {
  _getSubmissionStore().clear()
})

describe("submit_review_findings", () => {
  it("stores findings in the submission store keyed by session ID", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const ctx = childContext("child-session-1")
      const result = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            {
              severity: "HIGH",
              category: "security",
              description: "SQL injection vulnerability",
              root_cause: "Unsanitized input in query",
              file: "src/db.ts",
              line: 42,
            },
          ],
        },
        ctx,
      )

      expect(result.output).toBe("1 finding(s) submitted")
      const store = _getSubmissionStore()
      expect(store.has("child-session-1")).toBe(true)
      const data = store.get("child-session-1")!
      expect(data.findings).toHaveLength(1)
      expect(data.findings[0].severity).toBe("HIGH")
      expect(data.findings[0].file).toBe("src/db.ts")
      expect(data.proposals).toHaveLength(0)
    })
  })

  it("accumulates findings across multiple calls in the same session", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const ctx = childContext("child-session-2")
      await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { severity: "LOW", category: "style", description: "Naming", root_cause: "Inconsistent", file: "a.ts", line: 1 },
          ],
        },
        ctx,
      )
      await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { severity: "MEDIUM", category: "logic", description: "Missing null check", root_cause: "No guard", file: "b.ts", line: 10 },
            { severity: "HIGH", category: "security", description: "XSS risk", root_cause: "Unescaped", file: "c.ts", line: 20 },
          ],
        },
        ctx,
      )

      const data = _getSubmissionStore().get("child-session-2")!
      expect(data.findings).toHaveLength(3)
      expect(data.findings[0].severity).toBe("LOW")
      expect(data.findings[1].severity).toBe("MEDIUM")
      expect(data.findings[2].severity).toBe("HIGH")
    })
  })

  it("keeps separate sessions isolated", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      await hooks.tool.submit_review_findings.execute(
        {
          findings: [{ severity: "LOW", category: "style", description: "d1", root_cause: "c1", file: "a.ts", line: 1 }],
        },
        childContext("session-A"),
      )
      await hooks.tool.submit_review_findings.execute(
        {
          findings: [{ severity: "HIGH", category: "security", description: "d2", root_cause: "c2", file: "b.ts", line: 2 }],
        },
        childContext("session-B"),
      )

      const store = _getSubmissionStore()
      expect(store.get("session-A")!.findings).toHaveLength(1)
      expect(store.get("session-B")!.findings).toHaveLength(1)
      expect(store.get("session-A")!.findings[0].severity).toBe("LOW")
      expect(store.get("session-B")!.findings[0].severity).toBe("HIGH")
    })
  })

  it("accepts findings with null file and line", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { severity: "MEDIUM", category: "arch", description: "Missing abstraction", root_cause: "Coupling", file: null, line: null },
          ],
        },
        childContext("child-null-fields"),
      )

      expect(result.output).toBe("1 finding(s) submitted")
      const data = _getSubmissionStore().get("child-null-fields")!
      expect(data.findings[0].file).toBeNull()
      expect(data.findings[0].line).toBeNull()
    })
  })

  it("rejects findings missing severity without storing anything", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { category: "security", description: "Missing severity key", root_cause: "No severity", file: "a.ts", line: 1 },
          ],
        } as never,
        childContext("child-invalid-missing-severity"),
      )

      expect(result.output).toContain("submit_review_findings validation failed")
      expect(result.output).toContain("findings.0.severity")
      expect(result.output).toContain("Invalid option: expected one of")
      expect(_getSubmissionStore().has("child-invalid-missing-severity")).toBe(false)
    })
  })

  it("rejects an invalid severity enum value with the expected-values message", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { severity: "medium", category: "security", description: "Lowercase severity", root_cause: "Case mismatch", file: "a.ts", line: 1 },
          ],
        } as never,
        childContext("child-invalid-lowercase"),
      )

      expect(result.output).toContain("submit_review_findings validation failed")
      expect(result.output).toContain("findings.0.severity")
      expect(result.output).toContain("Invalid option: expected one of")
      expect(_getSubmissionStore().has("child-invalid-lowercase")).toBe(false)
    })
  })

  it("rejects findings carrying an unrecognized key (strict violation)", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            {
              severity: "HIGH",
              category: "security",
              description: "Extra key",
              root_cause: "Strict violation",
              file: "a.ts",
              line: 1,
              recommendation: "fix it",
            },
          ],
        } as never,
        childContext("child-invalid-extra-key"),
      )

      expect(result.output).toContain("submit_review_findings validation failed")
      expect(result.output).toContain("findings.0")
      expect(result.output).toContain("Unrecognized key")
      expect(_getSubmissionStore().has("child-invalid-extra-key")).toBe(false)
    })
  })

  it("rejects an empty findings array (min 1)", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_review_findings.execute(
        { findings: [] },
        childContext("child-invalid-empty"),
      )

      expect(result.output).toContain("submit_review_findings validation failed")
      expect(result.output).toContain("findings")
      expect(result.output).toContain("Too small")
      expect(_getSubmissionStore().has("child-invalid-empty")).toBe(false)
    })
  })

  it("stores nothing when rejected and accepts a later valid submission in the same session", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const ctx = childContext("child-recover")
      const rejected = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { severity: "medium", category: "security", description: "Invalid", root_cause: "Lowercase", file: "a.ts", line: 1 },
          ],
        } as never,
        ctx,
      )
      expect(rejected.output).toContain("submit_review_findings validation failed")

      const accepted = await hooks.tool.submit_review_findings.execute(
        {
          findings: [
            { severity: "MEDIUM", category: "security", description: "Fixed", root_cause: "Lowercase", file: "a.ts", line: 1 },
          ],
        },
        ctx,
      )
      expect(accepted.output).toBe("1 finding(s) submitted")

      const data = _getSubmissionStore().get("child-recover")!
      expect(data.findings).toHaveLength(1)
      expect(data.findings[0].severity).toBe("MEDIUM")
      expect(data.proposals).toHaveLength(0)
    })
  })
})

describe("submit_lesson_proposal", () => {
  it("stores a lesson proposal in the submission store", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const ctx = childContext("child-lesson-1")
      const result = await hooks.tool.submit_lesson_proposal.execute(
        {
          information: "Always validate @suffix is preserved in model slugs",
          tag: "model-resolution",
          category: "gotcha",
        },
        ctx,
      )

      expect(result.output).toBe("lesson proposal submitted (tag: model-resolution)")
      const data = _getSubmissionStore().get("child-lesson-1")!
      expect(data.proposals).toHaveLength(1)
      expect(data.proposals[0].information).toBe("Always validate @suffix is preserved in model slugs")
      expect(data.proposals[0].tag).toBe("model-resolution")
      expect(data.proposals[0].category).toBe("gotcha")
      expect(data.findings).toHaveLength(0)
    })
  })

  it("stores a proposal without optional category", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_lesson_proposal.execute(
        {
          information: "Review matrix files are passthrough for model slugs",
          tag: "review-dispatch",
        },
        childContext("child-lesson-no-cat"),
      )

      expect(result.output).toBe("lesson proposal submitted (tag: review-dispatch)")
      const data = _getSubmissionStore().get("child-lesson-no-cat")!
      expect(data.proposals[0].category).toBeUndefined()
    })
  })

  it("accumulates proposals and findings in the same session", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const ctx = childContext("child-mixed")
      await hooks.tool.submit_review_findings.execute(
        {
          findings: [{ severity: "LOW", category: "style", description: "Minor", root_cause: "Convention", file: "x.ts", line: 5 }],
        },
        ctx,
      )
      await hooks.tool.submit_lesson_proposal.execute(
        { information: "Style checks should be automated", tag: "automation", category: "pattern" },
        ctx,
      )

      const data = _getSubmissionStore().get("child-mixed")!
      expect(data.findings).toHaveLength(1)
      expect(data.proposals).toHaveLength(1)
      expect(data.findings[0].severity).toBe("LOW")
      expect(data.proposals[0].tag).toBe("automation")
    })
  })

  it("rejects a proposal missing the required tag without storing anything", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_lesson_proposal.execute(
        { information: "Learning without a tag" } as never,
        childContext("child-lesson-no-tag"),
      )

      expect(result.output).toContain("submit_lesson_proposal validation failed")
      expect(result.output).toContain("tag: Invalid input")
      expect(_getSubmissionStore().has("child-lesson-no-tag")).toBe(false)
    })
  })

  it("rejects a proposal with empty information", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_lesson_proposal.execute(
        { information: "", tag: "empty-information" },
        childContext("child-lesson-empty-info"),
      )

      expect(result.output).toContain("submit_lesson_proposal validation failed")
      expect(result.output).toContain("information: Too small")
      expect(_getSubmissionStore().has("child-lesson-empty-info")).toBe(false)
    })
  })

  it("rejects a proposal with an invalid category enum value", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_lesson_proposal.execute(
        { information: "Learning", tag: "bad-category", category: "meta" } as never,
        childContext("child-lesson-bad-category"),
      )

      expect(result.output).toContain("submit_lesson_proposal validation failed")
      expect(result.output).toContain("category: Invalid option: expected one of")
      expect(_getSubmissionStore().has("child-lesson-bad-category")).toBe(false)
    })
  })

  it("rejects a proposal carrying an unrecognized key (strict violation)", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.submit_lesson_proposal.execute(
        { information: "Learning", tag: "extra-key", recommendation: "store it" } as never,
        childContext("child-lesson-extra-key"),
      )

      expect(result.output).toContain("submit_lesson_proposal validation failed")
      expect(result.output).toContain("Unrecognized key")
      expect(_getSubmissionStore().has("child-lesson-extra-key")).toBe(false)
    })
  })
})

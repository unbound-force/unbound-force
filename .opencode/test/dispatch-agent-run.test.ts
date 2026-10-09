import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { rm } from "node:fs/promises"
import { join } from "node:path"
import { tmpdir } from "node:os"
import { describe, expect, it, vi, afterEach } from "vitest"

import {
  createDispatchAgentRunTool,
  dispatchAgentRun,
  _getSessionCorrelationMap,
  type DispatchAgentRunDependencies,
  type PlannerDependencies,
} from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/

// Track dispatch session dirs created by dispatchAgentRun for cleanup.
const cleanupDirs: string[] = []

afterEach(async () => {
  _getSessionCorrelationMap().clear()
  for (const dir of cleanupDirs) {
    try {
      await rm(dir, { recursive: true, force: true })
    } catch {
      // best-effort
    }
  }
  cleanupDirs.length = 0
})

type Client = PluginInput["client"]

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

/**
 * Minimal fake client that succeeds for dispatch_agent_run calls.
 * Prompt response echoes the requested model identity so
 * model-mismatch checks pass.
 */
function fakeClient(): Client {
  const session = {
    create: async (): Promise<unknown> => ({
      data: { id: "child-session" },
      error: undefined,
    }),
    message: async (): Promise<unknown> => ({
      data: {
        info: {
          id: "m",
          sessionID: "s",
          role: "assistant",
          providerID: "p",
          modelID: "m",
          variant: "active",
        },
        parts: [],
      },
      error: undefined,
    }),
    prompt: async (request: unknown): Promise<unknown> => {
      // Echo the requested model so model-mismatch checks pass.
      const req = request as { body?: { model?: { providerID: string; modelID: string } } }
      const providerID = req.body?.model?.providerID ?? "provider"
      const modelID = req.body?.model?.modelID ?? "standard"
      return {
        data: {
          info: { role: "assistant", providerID, modelID },
          parts: [
            { id: "t", sessionID: "child-session", messageID: "m", type: "text", text: "review output", ignored: false },
          ],
        },
        error: undefined,
      }
    },
    abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
  }
  return { session } as unknown as Client
}

function toolContext(): ToolContext {
  return {
    sessionID: "parent-session",
    messageID: "message-parent",
    agent: "build",
    directory: "/workspace",
    worktree: "/workspace",
    abort: new AbortController().signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("permission prompts are not used")
    }) as ToolContext["ask"],
  }
}

/** Track a dispatch result for temp dir cleanup. */
function trackForCleanup(correlationId: string): void {
  cleanupDirs.push(join(tmpdir(), "opencode", `dispatch-${correlationId}`))
}

function makeDeps(overrides?: {
  readText?: (path: string) => Promise<string>
  client?: Client
  directory?: string
}): DispatchAgentRunDependencies {
  const readText =
    overrides?.readText ??
    vi.fn(async (path: string): Promise<string> => {
      if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
      if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
      if (path === ".uf/review-matrix.override.yaml") {
        const err = new Error("ENOENT") as NodeJS.ErrnoException
        err.code = "ENOENT"
        throw err
      }
      throw new Error(`unexpected read: ${path}`)
    })

  return {
    plannerDependencies: {
      readText,
      parseYaml: jsonParser,
    },
    client: overrides?.client ?? fakeClient(),
    directory: overrides?.directory ?? "/workspace",
  }
}

describe("dispatch_agent_run", () => {
  describe("tier-based model resolution", () => {
    it("resolves model from review matrix for 'standard' tier", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.text).toBe("review output")
      // Auto-generated correlation_id and run_id are valid UUIDs.
      expect(result.correlation_id).toMatch(UUID_RE)
      expect(result.run_id).toMatch(UUID_RE)
      // The matrix fixture has standard profile with model "provider/standard".
      expect(result.provenance.requested_model).toBe("provider/standard")
      expect(result.provenance.requested_variant).toBe("high")
      expect(result.provenance.agent).toBe("divisor-guard")
    })

    it("resolves model from review matrix for 'lightweight' tier", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "lightweight" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      // The matrix fixture has lightweight profile with model "provider/light".
      expect(result.provenance.requested_model).toBe("provider/light")
      expect(result.provenance.requested_variant).toBe("fast")
    })
  })

  describe("explicit model bypass", () => {
    it("uses model parameter directly without loading matrix for resolution", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", model: "custom-provider/custom-model@v2" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.requested_model).toBe("custom-provider/custom-model@v2")
      expect(result.provenance.agent).toBe("divisor-guard")
    })

    it("preserves variant when model is provided directly", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        {
          agent: "divisor-guard",
          prompt: "Review this change.",
          model: "custom-provider/custom-model",
          variant: "deep",
        },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.requested_model).toBe("custom-provider/custom-model")
      expect(result.provenance.requested_variant).toBe("deep")
    })
  })

  describe("default tier fallback", () => {
    it("defaults to 'standard' tier when neither tier nor model is provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.correlation_id).toMatch(UUID_RE)
      expect(result.run_id).toMatch(UUID_RE)
      // Should resolve via the "standard" profile: model "provider/standard".
      expect(result.provenance.requested_model).toBe("provider/standard")
      expect(result.provenance.requested_variant).toBe("high")
    })
  })

  describe("mutual exclusivity validation", () => {
    it("rejects when both tier and model are provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", model: "provider/model" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("tier and model are mutually exclusive")
    })

    it("rejects when both prompt and promptFile are provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "inline text", promptFile: "/tmp/prompt.txt" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("prompt and promptFile are mutually exclusive")
    })

    it("rejects when neither prompt nor promptFile is provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", tier: "standard" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("one of prompt or promptFile is required")
    })
  })

  describe("unknown agent rejection", () => {
    it("returns failed result for an agent not in the manifest", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-nonexistent", prompt: "Review this change." },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("unknown_agent")
      expect(result.error?.message).toContain("divisor-nonexistent")
      expect(result.error?.message).toContain("absent from the reviewer manifest")
    })
  })

  describe("missing review matrix", () => {
    it("returns a failed result when loadPolicies throws", async () => {
      const readText = vi.fn(async (_path: string): Promise<string> => {
        throw new Error("ENOENT: no such file /Users/alice/.uf/review-matrix.yaml")
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("policy_load_failed")
      expect(result.error?.retryable).toBe(true)
      // Verify path sanitization — /Users/alice should be redacted.
      expect(result.error?.message).not.toContain("/Users/alice")
      expect(result.error?.message).toContain("[REDACTED]")
    })
  })

  describe("tier with no model configured", () => {
    it("returns tier_model_unavailable when the profile has no model", async () => {
      const emptyProfileMatrix = {
        ...matrixFixture,
        profiles: { ...matrixFixture.profiles, standard: {} },
      }
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(emptyProfileMatrix)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("tier_model_unavailable")
      expect(result.error?.message).toContain("standard")
      expect(result.error?.message).toContain("no model configured")
      expect(result.error?.retryable).toBe(false)
    })
  })

  describe("invalid timeout", () => {
    it("rejects a timeout exceeding the schema maximum", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", timeout: 2_000_000 },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.retryable).toBe(false)
    })
  })

  describe("provenance metadata", () => {
    it("reflects tier-resolved model in requested_model", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "heavy" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance).toMatchObject({
        agent: "divisor-guard",
        requested_model: "provider/heavy",
        resolved_parent_model: null,
        resolved_parent_variant: null,
        read_only: true,
        model_mismatch: false,
      })
      // Heavy profile has no variant configured, so requested_variant should be null.
      expect(result.provenance.requested_variant).toBeNull()
    })

    it("reflects direct model in requested_model", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        {
          agent: "divisor-adversary",
          prompt: "Review this change.",
          model: "google-vertex/claude-opus-4@default",
          variant: "fast",
          read_only: false,
        },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance).toMatchObject({
        agent: "divisor-adversary",
        requested_model: "google-vertex/claude-opus-4@default",
        requested_variant: "fast",
        resolved_parent_model: null,
        resolved_parent_variant: null,
        read_only: false,
      })
    })

    it("sets read_only to true by default", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.read_only).toBe(true)
    })
  })

  describe("promptFile support", () => {
    it("reads prompt text from promptFile and executes successfully", async () => {
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/prompt.txt") return "File-based prompt content."
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/prompt.txt" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.text).toBe("review output")
      expect(readText).toHaveBeenCalledWith("/tmp/prompt.txt")
    })

    it("returns a failed result when promptFile read fails", async () => {
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/missing.txt") throw new Error("ENOENT: no such file /Users/alice/secret/prompt.txt")
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/missing.txt" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("prompt_file_read_failed")
      expect(result.error?.retryable).toBe(false)
      // Verify path sanitization.
      expect(result.error?.message).not.toContain("/Users/alice")
      expect(result.error?.message).toContain("[REDACTED]")
    })

    it("rejects promptFile content exceeding 4 MiB", async () => {
      const oversizedContent = "x".repeat(4 * 1024 * 1024 + 1)
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/big-prompt.txt") return oversizedContent
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/big-prompt.txt" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("promptFile content exceeds 4 MiB")
    })
  })

  describe("session-based correlation", () => {
    it("reuses the same correlation_id for multiple calls in the same session", async () => {
      const deps = makeDeps()
      const ctx = toolContext()
      const result1 = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "First review." },
        ctx,
        deps,
      )
      trackForCleanup(result1.correlation_id)
      const result2 = await dispatchAgentRun(
        { agent: "divisor-adversary", prompt: "Second review." },
        ctx,
        deps,
      )

      expect(result1.correlation_id).toMatch(UUID_RE)
      expect(result2.correlation_id).toBe(result1.correlation_id)
      expect(result1.run_id).not.toBe(result2.run_id)
    })
  })

  describe("createDispatchAgentRunTool factory", () => {
    it("creates a tool with description and expected args", () => {
      const deps = makeDeps()
      const dispatchTool = createDispatchAgentRunTool(deps)

      expect(dispatchTool.description).toContain("Divisor review-agent run")
      expect(Object.keys(dispatchTool.args).sort()).toEqual([
        "agent",
        "model",
        "prompt",
        "promptFile",
        "read_only",
        "sequence",
        "session_metadata",
        "source",
        "tier",
        "timeout",
        "variant",
      ])
    })
  })
})

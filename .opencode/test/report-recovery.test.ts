import type { PluginInput } from "@opencode-ai/plugin"
import { describe, expect, it } from "vitest"

import {
  MAX_DELEGATE_FINAL_TEXT_LENGTH,
  MAX_RECOVERY_MESSAGES,
  boundFinalText,
  readChildTranscript,
  summarizeTranscript,
} from "../lib/report-recovery.js"

type Client = PluginInput["client"]

function clientWith(messages: (request: unknown) => Promise<unknown>): Client {
  return { session: { messages } } as unknown as Client
}

describe("boundFinalText", () => {
  it("returns null for empty or blank text", () => {
    expect(boundFinalText("")).toBeNull()
    expect(boundFinalText("  \n ")).toBeNull()
  })

  it("keeps short text trimmed and unchanged", () => {
    expect(boundFinalText("  ## Demo\nall green  ")).toBe("## Demo\nall green")
  })

  it("keeps the tail behind a truncation marker when over the bound", () => {
    // One long line: the head (x) must be dropped, the tail (END) kept.
    const text = `${"x".repeat(200)}${"b".repeat(MAX_DELEGATE_FINAL_TEXT_LENGTH - 3)}END`
    const bounded = boundFinalText(text) as string
    expect(bounded.startsWith("…[truncated 200 chars]\n")).toBe(true)
    expect(bounded.endsWith("END")).toBe(true)
    expect(bounded).not.toContain("x")
    expect(bounded.length).toBe("…[truncated 200 chars]\n".length + MAX_DELEGATE_FINAL_TEXT_LENGTH)
  })

  it("redacts secrets and home paths on every line", () => {
    const bounded = boundFinalText("token=abc123\nsee /Users/someone/project/file.ts") as string
    expect(bounded).toContain("token=[REDACTED]")
    expect(bounded).not.toContain("abc123")
    expect(bounded).not.toContain("/Users/someone")
  })
})

describe("summarizeTranscript", () => {
  it("summarizes an empty transcript", () => {
    expect(summarizeTranscript([])).toEqual({
      messages_examined: 0,
      last_assistant_text: null,
      last_tool: null,
      last_input_tokens: null,
      final_turn: null,
    })
  })

  it("takes the latest non-empty assistant text and the latest tool", () => {
    const summary = summarizeTranscript([
      { info: { role: "assistant", providerID: "p", modelID: "m" }, parts: [{ type: "text", text: "first" }] },
      {
        info: { role: "assistant", providerID: "p", modelID: "m", tokens: { input: 12 } },
        parts: [
          { type: "text", text: "second" },
          { type: "text", text: "hidden", ignored: true },
          { type: "tool", tool: "read", state: { status: "completed" } },
        ],
      },
      { info: { role: "assistant", providerID: "p", modelID: "m" }, parts: [{ type: "text", text: "  " }] },
    ])
    expect(summary.last_assistant_text).toBe("second")
    expect(summary.last_tool).toEqual({ tool: "read", status: "completed" })
    expect(summary.last_input_tokens).toBe(12)
    expect(summary.final_turn).toBeNull()
  })

  it("flags a final user turn whose text parts are all synthetic as injected", () => {
    const summary = summarizeTranscript([
      { info: { role: "user", agent: "dev", model: { providerID: "p", modelID: "m" } }, parts: [{ type: "text", text: "go" }] },
      {
        info: { role: "user", agent: "build", model: { providerID: "p", modelID: "x" } },
        parts: [{ type: "text", text: "[supervisor] fix", synthetic: true }],
      },
      { info: { role: "assistant", providerID: "p", modelID: "x" }, parts: [] },
    ])
    expect(summary.final_turn).toEqual({ agent: "build", model: "p/x", assistant_model: "p/x", injected: true })
  })

  it("does not flag a mixed or text-less user turn as injected", () => {
    const mixed = summarizeTranscript([
      {
        info: { role: "user" },
        parts: [
          { type: "text", text: "a", synthetic: true },
          { type: "text", text: "b" },
        ],
      },
    ])
    expect(mixed.final_turn).toEqual({ agent: null, model: null, assistant_model: null, injected: false })
    const empty = summarizeTranscript([{ info: { role: "user" }, parts: [{ type: "file" }] }])
    expect(empty.final_turn?.injected).toBe(false)
  })

  it("skips malformed entries without failing", () => {
    const summary = summarizeTranscript([null, 7, { parts: [] }, { info: { role: "system" }, parts: "x" }])
    expect(summary.messages_examined).toBe(2)
    expect(summary.final_turn).toBeNull()
    expect(summary.last_assistant_text).toBeNull()
  })
})

describe("readChildTranscript", () => {
  it("reads the most recent messages of the child session", async () => {
    const requests: unknown[] = []
    const summary = await readChildTranscript(
      clientWith(async (request) => {
        requests.push(request)
        return { data: [{ info: { role: "assistant" }, parts: [{ type: "text", text: "done" }] }] }
      }),
      "child",
      "/dir",
    )
    expect(summary?.last_assistant_text).toBe("done")
    expect(requests).toEqual([{ path: { id: "child" }, query: { directory: "/dir", limit: MAX_RECOVERY_MESSAGES } }])
  })

  it("fails open on a missing data array or a rejected read", async () => {
    expect(await readChildTranscript(clientWith(async () => ({ error: { name: "NotFound" } })), "c", "/d")).toBeNull()
    expect(
      await readChildTranscript(
        clientWith(async () => {
          throw new Error("boom")
        }),
        "c",
        "/d",
      ),
    ).toBeNull()
  })
})

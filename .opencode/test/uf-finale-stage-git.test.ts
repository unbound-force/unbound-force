// Layer 3 round trip for uf_finale_stage against a real git repository
// (feedback-loop-scope LPS-FR-001). uf-finale.test.ts mocks
// node:child_process, so the real executor runs in this file instead.
import { execFileSync } from "node:child_process"
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import { createUfFinaleStageTool, defaultGitExec } from "../plugins/uf-finale/index.js"

const repos: string[] = []

afterEach(() => {
  for (const dir of repos.splice(0)) rmSync(dir, { recursive: true, force: true })
})

function git(dir: string, ...args: string[]): string {
  return execFileSync("git", args, { cwd: dir, encoding: "utf-8" }).trim()
}

function repo(withCommit: boolean): string {
  const dir = mkdtempSync(join(tmpdir(), "uf-finale-stage-"))
  repos.push(dir)
  git(dir, "init", "-q")
  git(dir, "config", "user.email", "test@example.com")
  git(dir, "config", "user.name", "Test")
  git(dir, "config", "commit.gpgsign", "false")
  if (withCommit) git(dir, "commit", "-q", "--allow-empty", "-m", "init")
  writeFileSync(join(dir, "src.ts"), "export {}\n")
  return dir
}

function learning(dir: string): void {
  mkdirSync(join(dir, ".uf", "dewey", "learnings"), { recursive: true })
  writeFileSync(join(dir, ".uf", "dewey", "learnings", "x.md"), "lesson\n")
}

async function stage(dir: string): Promise<{ status: string; data?: { fileCount: number; files: string[] } }> {
  const tool = createUfFinaleStageTool(defaultGitExec, dir) as unknown as {
    execute: (args: Record<string, unknown>, ctx: { sessionID: string }) => Promise<{ output: string }>
  }
  const result = await tool.execute({}, { sessionID: "test" })
  return JSON.parse(result.output) as { status: string; data?: { fileCount: number; files: string[] } }
}

describe("uf_finale_stage with real git", () => {
  it("succeeds when no learnings directory exists", async () => {
    const dir = repo(true)
    const result = await stage(dir)
    expect(result.status).toBe("ok")
    expect(result.data?.files).toEqual(["src.ts"])
  })

  it("succeeds in a repository with no commits yet", async () => {
    const dir = repo(false)
    learning(dir)
    const result = await stage(dir)
    expect(result.status).toBe("ok")
    expect(result.data?.files).toEqual(["src.ts"])
  })

  it("leaves untracked learnings unstaged", async () => {
    const dir = repo(true)
    learning(dir)
    const result = await stage(dir)
    expect(result.data?.files).toEqual(["src.ts"])
    expect(git(dir, "status", "--porcelain", "--", ".uf/dewey/learnings")).toBe("?? .uf/dewey/learnings/")
  })

  it("unstages learnings that were already staged", async () => {
    const dir = repo(true)
    learning(dir)
    git(dir, "add", ".uf/dewey/learnings")
    const result = await stage(dir)
    expect(result.data?.files).toEqual(["src.ts"])
  })

  it("still stages compiled articles", async () => {
    const dir = repo(true)
    learning(dir)
    mkdirSync(join(dir, ".uf", "dewey", "compiled"), { recursive: true })
    writeFileSync(join(dir, ".uf", "dewey", "compiled", "topic.md"), "article\n")
    const result = await stage(dir)
    expect(result.data?.files).toEqual([".uf/dewey/compiled/topic.md", "src.ts"])
  })
})

import { createHash } from "node:crypto"
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { afterEach, describe, expect, it, vi } from "vitest"

import { escapeControlChars, findRoot, git, isMainModule, runMain, spawnDetail } from "../lib/cli-common.js"

const roots: string[] = []
afterEach(() => {
  for (const r of roots.splice(0)) rmSync(r, { recursive: true, force: true })
})

describe("cli-common", () => {
  it("findRoot locates the nearest ancestor holding a .opencode directory", () => {
    const root = mkdtempSync(join(tmpdir(), "cli-common-"))
    roots.push(root)
    mkdirSync(join(root, ".opencode"))
    mkdirSync(join(root, "a", "b"), { recursive: true })
    expect(findRoot(join(root, "a", "b"))).toBe(root)
    expect(findRoot(root)).toBe(root)
  })

  it("findRoot skips a file named .opencode and keeps walking up", () => {
    const root = mkdtempSync(join(tmpdir(), "cli-common-"))
    roots.push(root)
    mkdirSync(join(root, "a"), { recursive: true })
    writeFileSync(join(root, "a", ".opencode"), "file, not a directory")
    expect(findRoot(join(root, "a"))).toBeNull()
    mkdirSync(join(root, ".opencode"))
    expect(findRoot(join(root, "a"))).toBe(root)
  })

  it("git captures stdout and feeds stdin", () => {
    const expected = createHash("sha1").update("blob 5\0hello").digest("hex")
    expect(git(tmpdir(), ["hash-object", "--stdin"], "hello").toString("utf8").trim()).toBe(expected)
  })

  it("spawnDetail prefers trimmed stderr and falls back to the message", () => {
    const e = Object.assign(new Error("boom"), { stderr: Buffer.from("  git said no  \n") })
    expect(spawnDetail(e)).toBe("git said no")
    expect(spawnDetail(new Error("only message"))).toBe("only message")
    // Null/undefined errors and plain-string stderr must not throw.
    expect(spawnDetail(null)).toBe("null")
    expect(spawnDetail(undefined)).toBe("undefined")
    expect(spawnDetail(Object.assign(new Error("x"), { stderr: "  plain  " }))).toBe("plain")
  })

  it("runMain wires process writers and records the exit code", () => {
    const prior = process.exitCode
    const out = vi.spyOn(process.stdout, "write").mockImplementation(() => true)
    const err = vi.spyOn(process.stderr, "write").mockImplementation(() => true)
    let cwd = ""
    try {
      runMain((_argv, io) => {
        cwd = io.cwd
        io.out("wrote out")
        io.err("wrote err")
        return 3
      }, "test-cli")
      expect(cwd).toBe(process.cwd())
      expect(out).toHaveBeenCalledWith("wrote out\n")
      expect(err).toHaveBeenCalledWith("wrote err\n")
      expect(process.exitCode).toBe(3)
    } finally {
      out.mockRestore()
      err.mockRestore()
      process.exitCode = prior
    }
  })

  it("runMain converts a crash to exit 2 with a stderr message", () => {
    const prior = process.exitCode
    const err = vi.spyOn(process.stderr, "write").mockImplementation(() => true)
    try {
      runMain(() => {
        throw new Error("kaboom")
      }, "test-cli")
      expect(process.exitCode).toBe(2)
      expect(err).toHaveBeenCalledWith("test-cli: could not run: kaboom\n")
    } finally {
      err.mockRestore()
      process.exitCode = prior
    }
  })

  it("escapeControlChars escapes C0 and C1 controls, keeping tab and newline", () => {
    expect(escapeControlChars("a\u001b[31mb\u009b\rc\td\ne")).toBe("a\\x1b[31mb\\x9b\\x0dc\td\ne")
  })

  it("isMainModule is false outside the real entry path", () => {
    const prior = process.argv[1]
    try {
      // Under vitest the entry is the runner binary, never this module.
      expect(isMainModule(import.meta.url)).toBe(false)
      process.argv[1] = undefined as unknown as string
      expect(isMainModule(import.meta.url)).toBe(false)
      process.argv[1] = join(tmpdir(), "definitely-not-a-real-entry.js")
      expect(isMainModule(import.meta.url)).toBe(false)
    } finally {
      process.argv[1] = prior
    }
  })
})

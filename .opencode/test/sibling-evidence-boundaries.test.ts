import { mkdtemp, rm } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import {
  createSiblingAcquisitionDependencies,
  parseSiblingRepositories,
} from "../plugins/review-dispatch/index.js"

const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

function config(overrides: Record<string, unknown> = {}): string {
  return JSON.stringify({
    version: 1,
    siblings: [
      {
        name: "gaze",
        url: "https://github.com/unbound-force/gaze",
        default_branch: "main",
        role: "sibling",
        fetch: "contract-only",
        contracts: ["README.md"],
        ...overrides,
      },
    ],
  })
}

describe("sibling evidence declaration boundaries", () => {
  it("rejects a non-HTTPS or malformed repository URL", () => {
    expect(() => parseSiblingRepositories(config({ url: "git@github.com:unbound-force/gaze.git" }), jsonParser)).toThrow(
      "absolute HTTPS",
    )
    expect(() => parseSiblingRepositories(config({ url: "https://github.com/only-owner" }), jsonParser)).toThrow(
      "one GitHub owner and repository",
    )
  })

  it("rejects YAML merge and alias references and quoted comment stripping", () => {
    expect(() => parseSiblingRepositories("siblings:\n  - &alias\n", jsonParser)).toThrow("forbidden YAML")
    expect(() => parseSiblingRepositories("version: 1\nsiblings: []\n  <<: *alias\n", jsonParser)).toThrow(
      "forbidden YAML",
    )
  })

  it("rejects unsafe clone and cache target paths", () => {
    expect(() => parseSiblingRepositories(config({ clone_path: "/absolute/path" }), jsonParser)).toThrow(
      "target-root-relative",
    )
  })
})

describe("production sibling acquisition dependencies", () => {
  it("runs a local command without shell interpretation", async () => {
    const directory = await mkdtemp(join(tmpdir(), "uf-sibling-boundaries-"))
    scratchDirectories.push(directory)
    const dependencies = createSiblingAcquisitionDependencies(directory, jsonParser)

    const result = await dependencies.runCommand("node", ["--version"])
    expect(result.stdout).toMatch(/^v\d+\./)
    expect(dependencies.projectRoot).toBe(directory)
    expect(typeof dependencies.readText).toBe("function")
    expect(typeof dependencies.realpath).toBe("function")
    expect(typeof dependencies.stat).toBe("function")
    expect(typeof dependencies.readFile).toBe("function")
    expect(typeof dependencies.mkdir).toBe("function")
  })
})

import { describe, expect, it } from "vitest"

import {
  createBunYamlParser,
  parseReviewMatrix,
  parseReviewerManifest,
  planReviewDispatch,
  serializeDispatchPlan,
  type DispatchPlan,
  type PlannerDependencies,
} from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"
import { derivePerRunTimeoutMs, resolveBaseWindowMs } from "../lib/run-timeout.js"

type JsonObject = Record<string, unknown>
type RunFixture = { profile?: string; model?: string; variant?: string }
type MutableMatrixFixture = Omit<typeof matrixFixture, "runs"> & {
  runs: { code: Record<string, RunFixture[]> }
}

function cloneFixture<T>(fixture: T): T {
  return structuredClone(fixture)
}

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

function dependencies(
  matrix: unknown = matrixFixture,
  manifest: unknown = manifestFixture,
  override?: unknown,
): PlannerDependencies {
  const documents: Record<string, string> = {
    ".uf/review-matrix.yaml": JSON.stringify(matrix),
    ".uf/reviewer-capabilities.yaml": JSON.stringify(manifest),
  }
  if (override !== undefined) {
    documents[".uf/review-matrix.override.yaml"] = JSON.stringify(override)
  }
  return {
    readText: async (relativePath: string): Promise<string> => {
      const document = documents[relativePath]
      if (document === undefined) {
        // Simulate ENOENT for the override file so loadPolicies can distinguish
        // "file not found" from other read errors.
        if (relativePath === ".uf/review-matrix.override.yaml") {
          const err = new Error(`ENOENT: no such file or directory, open '${relativePath}'`)
          ;(err as NodeJS.ErrnoException).code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected policy read: ${relativePath}`)
      }
      return document
    },
    parseYaml: jsonParser,
  }
}

function changedFile(path: string, lines = 1): {
  readonly path: string
  readonly additions: number
  readonly deletions: number
} {
  return { path, additions: lines, deletions: 0 }
}

function included(plan: DispatchPlan, agent: string): DispatchPlan["entries"] {
  return plan.entries.filter((entry) => entry.agent === agent && entry.decision === "include")
}

describe("review dispatch policy schemas", () => {
  it("closed-validates matrix and canonical manifest fixtures", () => {
    const matrix = parseReviewMatrix(JSON.stringify(matrixFixture), jsonParser)
    const manifest = parseReviewerManifest(JSON.stringify(manifestFixture), jsonParser)

    expect(matrix.version).toBe(3)
    expect(matrix.profiles.standard).toEqual({ model: "provider/standard", variant: "high" })
    expect(manifest.reviewers).toHaveLength(9)
    expect(manifest.reviewers[0]?.scopes).toEqual(["security", "dependencies", "standard"])
  })

  it("rejects unknown fields, malformed input, aliases, and canonical scope drift", () => {
    const unknownMatrix = { ...cloneFixture(matrixFixture), unexpected: true }
    expect(() => parseReviewMatrix(JSON.stringify(unknownMatrix), jsonParser)).toThrow(
      "Unrecognized key",
    )
    expect(() => parseReviewMatrix("not-json", jsonParser)).toThrow("parse review matrix")
    expect(() => parseReviewMatrix("version: 2\nprofiles: &profiles {}\n", () => ({}))).toThrow(
      "forbidden YAML",
    )

    const invalidLimits = cloneFixture(matrixFixture)
    invalidLimits.limits.max_total_runs = 65
    expect(() => parseReviewMatrix(JSON.stringify(invalidLimits), jsonParser)).toThrow(
      "Too big",
    )

    const manifest = cloneFixture(manifestFixture)
    manifest.reviewers[0]!.scopes = ["dependencies", "security", "standard"]
    expect(() => parseReviewerManifest(JSON.stringify(manifest), jsonParser)).toThrow(
      "conflicts with canonical policy",
    )
  })

  it("fails clearly when Bun YAML support is unavailable", () => {
    const runtime = globalThis as typeof globalThis & { Bun?: unknown }
    const original = runtime.Bun
    Reflect.deleteProperty(runtime, "Bun")
    try {
      expect(() => createBunYamlParser()).toThrow("Bun.YAML.parse is unavailable")
    } finally {
      if (original !== undefined) {
        Object.defineProperty(runtime, "Bun", { configurable: true, value: original })
      }
    }
  })
})

describe("diff profiling and reviewer selection", () => {
  it("normalizes specs to spec while retaining command mode", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "specs",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("specs/042-feature/spec.md", 20)],
      },
      dependencies(),
    )

    expect(plan.mode).toBe("specs")
    expect(plan.matrix_mode).toBe("spec")
    expect(included(plan, "divisor-guard")[0]?.source).toBe("advisor")
  })

  it.each([
    [50, 1, "lightweight"],
    [51, 1, "standard"],
    [300, 1, "standard"],
    [301, 1, "heavy"],
  ] as const)("classifies %i changed lines as %s", async (lines, _files, tier) => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", lines)],
      },
      dependencies(),
    )
    expect(plan.change_profile.tier).toBe(tier)
  })

  it("uses component span and security precedence for total tiering", async () => {
    const componentPlan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("a/file.ts", 20), changedFile("b/file.ts", 20)],
      },
      dependencies(),
    )
    const securityPlan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        changed_files: [changedFile("api/auth.ts", 1)],
      },
      dependencies(),
    )

    expect(componentPlan.change_profile.tier).toBe("standard")
    expect(securityPlan.change_profile.tier).toBe("heavy")
    expect(securityPlan.change_profile.security_sensitive).toBe(true)
  })

  it.each([
    [3, "lightweight"],
    [4, "standard"],
    [10, "standard"],
    [11, "heavy"],
  ] as const)("classifies %i changed files as %s", async (count, tier) => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: Array.from({ length: count }, (_value, index) =>
          changedFile(`component/file-${index}.ts`),
        ),
      },
      dependencies(),
    )
    expect(plan.change_profile.tier).toBe(tier)
  })

  it("selects by exact scopes while excluding content and pruning Curator", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: [
          "divisor-sre",
          "divisor-scribe",
          "divisor-guard",
          "divisor-curator",
          "divisor-architect",
          "divisor-adversary",
        ],
        changed_files: [changedFile(".github/workflows/ci.yml", 4)],
      },
      dependencies(),
    )

    expect(included(plan, "divisor-adversary")).toHaveLength(2)
    expect(included(plan, "divisor-architect")).toHaveLength(1)
    expect(included(plan, "divisor-guard")).toHaveLength(1)
    expect(included(plan, "divisor-sre")).toHaveLength(1)
    expect(plan.entries.find((entry) => entry.agent === "divisor-curator")?.reason_code).toBe(
      "curator-pruned",
    )
    expect(plan.entries.find((entry) => entry.agent === "divisor-scribe")?.reason_code).toBe(
      "content-capability",
    )
  })
})

describe("issue-content profiling", () => {
  it("matches the fixed framing hash and test-quality reviewer selection", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-testing", "divisor-guard", "divisor-adversary"],
        issue: { title: "Coverage regression", body: null, comments: [] },
      },
      dependencies(),
    )

    expect(plan.change_profile).toMatchObject({
      kind: "issue",
      text_bytes: 19,
      content_sha256: "60a7be4394be0186439a588e93c57dc9095c1b96f42f27464dbfeb2ce5cfdbf5",
      categories: ["test-quality"],
      matched_rules: ["test-quality:coverage", "test-quality:regression"],
      tier: "lightweight",
    })
    expect(included(plan, "divisor-testing")).toHaveLength(1)
  })

  describe("tier-gated agent caps", () => {
  it("lightweight change caps included agents at default tier cap of 2", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(),
    )
    expect(plan.status).toBe("ready")

    // adversary and guard are always-required — kept under cap of 2
    // (adversary has 2 explicit runs in code mode, so count unique agents)
    const uniqueIncluded = [...new Set(
      plan.entries
        .filter((e) => e.decision === "include")
        .map((e) => e.agent),
    )]
    expect(uniqueIncluded).toHaveLength(2)
    expect(uniqueIncluded).toContain("divisor-adversary")
    expect(uniqueIncluded).toContain("divisor-guard")

    // architect (scope-match) exceeds the cap — tier-capped
    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(1)
    expect(tierCapSkips[0]!.agent).toBe("divisor-architect")
    expect(tierCapSkips[0]!.decision).toBe("skip")
    expect(tierCapSkips[0]!.reason).toContain("tier cap")
    expect(tierCapSkips[0]!.reason).toContain("lightweight")
  })

  it("standard change does not apply tier cap", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        changed_files: [changedFile("src/feature.ts", 60)], // >50 lines → standard
      },
      dependencies(),
    )
    expect(plan.status).toBe("ready")

    // All three relevant agents included — standard tier has null cap
    const includedAgents = plan.entries
      .filter((e) => e.decision === "include")
      .map((e) => e.agent)
    expect(includedAgents).toContain("divisor-adversary")
    expect(includedAgents).toContain("divisor-architect")
    expect(includedAgents).toContain("divisor-guard")

    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(0)
  })

  it("full flag bypasses tier caps for lightweight changes", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        full: true,
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(),
    )
    expect(plan.status).toBe("ready")

    // All agents included — full bypasses tier cap
    const includedAgents = plan.entries
      .filter((e) => e.decision === "include")
      .map((e) => e.agent)
    expect(includedAgents).toContain("divisor-adversary")
    expect(includedAgents).toContain("divisor-architect")
    expect(includedAgents).toContain("divisor-guard")

    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(0)
  })

  it("triage mode does not apply tier caps", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        issue: { title: "bug: login fails", body: "Login page error", comments: [] },
      },
      dependencies(),
    )

    // No tier-cap skip entries in triage mode
    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(0)
  })

  it("feedback mode does not apply tier caps", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "feedback",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(),
    )

    // No tier-cap skip entries in feedback mode
    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(0)
  })

  it("custom tier_caps override via matrix limits", async () => {
    const customMatrix = cloneFixture(matrixFixture) as Record<string, unknown>
    ;(customMatrix as Record<string, unknown>).limits = {
      tier_caps: { lightweight: 3 },
    }

    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(customMatrix),
    )
    expect(plan.status).toBe("ready")

    // All 3 agents fit within custom cap of 3
    const includedAgents = plan.entries
      .filter((e) => e.decision === "include")
      .map((e) => e.agent)
    expect(includedAgents).toContain("divisor-adversary")
    expect(includedAgents).toContain("divisor-architect")
    expect(includedAgents).toContain("divisor-guard")

    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(0)
  })

  it("floor enforcement clamps cap below always-required count", async () => {
    const customMatrix = cloneFixture(matrixFixture) as Record<string, unknown>
    ;(customMatrix as Record<string, unknown>).limits = {
      tier_caps: { lightweight: 1 },
    }

    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-guard",
        ],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(customMatrix),
    )
    expect(plan.status).toBe("ready")

    // cap=1 but 2 always-required agents → effectiveCap = max(1, 2) = 2
    // Both always-required agents kept, scope-match agent skipped
    const uniqueIncluded = [...new Set(
      plan.entries
        .filter((e) => e.decision === "include")
        .map((e) => e.agent),
    )]
    expect(uniqueIncluded).toHaveLength(2)
    expect(uniqueIncluded).toContain("divisor-adversary")
    expect(uniqueIncluded).toContain("divisor-guard")

    // architect (scope-match) exceeds the effective cap — tier-capped
    const tierCapSkips = plan.entries.filter((e) => e.reason_code === "tier-cap")
    expect(tierCapSkips).toHaveLength(1)
    expect(tierCapSkips[0]!.agent).toBe("divisor-architect")
    // reason should mention the effective cap (2), not the configured cap (1)
    expect(tierCapSkips[0]!.reason).toContain("tier cap of 2")
  })

  it("emits advisory when heavy tier cap is configured for security-sensitive change", async () => {
    const customMatrix = cloneFixture(matrixFixture) as Record<string, unknown>
    ;(customMatrix as Record<string, unknown>).limits = {
      tier_caps: { heavy: 3 },
    }

    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        // .github/workflows/ path makes it security-sensitive, >300 lines → heavy
        changed_files: [changedFile(".github/workflows/ci.yml", 310)],
      },
      dependencies(customMatrix),
    )

    const tierCapAdvisories = plan.advisories.filter((a) => a.severity === "HIGH")
    expect(tierCapAdvisories).toHaveLength(1)
    expect(tierCapAdvisories[0]!.description).toContain("tier_caps.heavy")
    expect(tierCapAdvisories[0]!.description).toContain("security-sensitive")
  })

  it("does not emit advisory when heavy tier cap is null (default)", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        changed_files: [changedFile(".github/workflows/ci.yml", 310)],
      },
      dependencies(),
    )

    // Default config has tier_caps.heavy = null — no tier-cap advisory
    expect(plan.advisories.filter((a) => a.description.includes("tier_caps.heavy"))).toHaveLength(0)
  })
})


  it.each([
    [4_096, "4f30e0423cec84abfc13ae44c9fe73dde044a0852c5492884e52ba020f7b8275", "lightweight"],
    [4_097, "19c10c78070cae1fb718628a7e276ed1b9d05d7d3f017bfb32d09bc309aa7207", "standard"],
    [32_768, "5e06dbe70b16a7d00fb864d511e8542bfdbc1473275d33076f33d5a917be0168", "standard"],
    [32_769, "15b20fab5e843a6526a81d7809f0c847f1aa237408cd6351b8c2c997d71ae3fe", "heavy"],
  ] as const)("uses exact %i-byte hash and tier boundary", async (length, hash, tier) => {
    const plan = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: { title: "x".repeat(length), body: null, comments: [] },
      },
      dependencies(),
    )
    expect(plan.change_profile).toMatchObject({ text_bytes: length, content_sha256: hash, tier })
  })

  it("normalizes line endings and orders comments by time then numeric id", async () => {
    const first = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: {
          title: "Docs\r\nupdate",
          body: "guide\rtext",
          comments: [
            { id: 10, created_at: "2026-01-01T00:00:00Z", body: "second" },
            { id: 2, created_at: "2026-01-01T00:00:00.000Z", body: "first" },
          ],
        },
      },
      dependencies(),
    )
    const second = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: {
          title: "Docs\nupdate",
          body: "guide\ntext",
          comments: [
            { id: 2, created_at: "2026-01-01T00:00:00Z", body: "first" },
            { id: 10, created_at: "2025-12-31T19:00:00-05:00", body: "second" },
          ],
        },
      },
      dependencies(),
    )
    expect(first.change_profile).toEqual(second.change_profile)
  })

  it("distinguishes null from empty bodies and applies comment-count boundaries", async () => {
    const issue = (body: string | null, count: number) => ({
      title: "question",
      body,
      comments: Array.from({ length: count }, (_value, index) => ({
        id: index + 1,
        created_at: `2026-01-01T00:00:${String(index).padStart(2, "0")}Z`,
        body: "note",
      })),
    })
    const nullBody = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: issue(null, 0),
      },
      dependencies(),
    )
    const emptyBody = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: issue("", 0),
      },
      dependencies(),
    )
    const fiveComments = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: issue(null, 5),
      },
      dependencies(),
    )
    const sixComments = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: issue(null, 6),
      },
      dependencies(),
    )
    const twentyOneComments = await planReviewDispatch(
      {
        mode: "triage",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        issue: issue(null, 21),
      },
      dependencies(),
    )

    expect(nullBody.change_profile).not.toMatchObject(emptyBody.change_profile)
    expect(fiveComments.change_profile.tier).toBe("lightweight")
    expect(sixComments.change_profile.tier).toBe("standard")
    expect(twentyOneComments.change_profile.tier).toBe("heavy")
  })
})

describe("run resolution and fail-closed limits", () => {
  it("emits explicit, advisor, and host sources with variant precedence", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard", "divisor-testing"],
        changed_files: [changedFile("feature_test.go", 60)],
      },
      dependencies(),
    )

    expect(included(plan, "divisor-adversary").map((entry) => entry.source)).toEqual([
      "explicit",
      "explicit",
    ])
    expect(included(plan, "divisor-adversary")[0]).toMatchObject({
      model: "provider/standard",
      variant: "low",
    })
    expect(included(plan, "divisor-guard")[0]?.source).toBe("advisor")
    expect(included(plan, "divisor-testing")[0]?.source).toBe("host")
  })

  it("applies opt-in augmentation and records first-occurrence dedupe", async () => {
    const matrix = cloneFixture(matrixFixture) as MutableMatrixFixture
    matrix.runs.code["divisor-adversary"] = [{ profile: "standard" }]
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        augment: true,
        changed_files: [changedFile("api/auth.ts", 10)],
      },
      dependencies(matrix),
    )

    expect(included(plan, "divisor-adversary")).toHaveLength(1)
    expect(plan.omissions).toEqual([
      {
        agent: "divisor-adversary",
        reason_code: "duplicate-model-variant",
        model: "provider/standard",
        variant: "high",
      },
    ])
  })

  it("fails when Curator has more than one assessment run", async () => {
    const matrix = cloneFixture(matrixFixture) as MutableMatrixFixture
    matrix.runs.code["divisor-curator"] = [
      { profile: "lightweight" },
      { profile: "standard" },
    ]
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-curator"],
        changed_files: [changedFile("docs/guide.md", 10)],
      },
      dependencies(matrix),
    )

    expect(plan.status).toBe("inconclusive")
    expect(plan.errors).toContain("divisor-curator may have at most one assessment run")
  })

  it("fails malformed policy before producing runnable entries", async () => {
    const malformed = { ...cloneFixture(matrixFixture), forbidden: true }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts")],
      },
      dependencies(malformed),
    )

    expect(plan.status).toBe("inconclusive")
    expect(plan.entries).toEqual([])
    expect(plan.errors[0]).toContain("validate review matrix")
  })

  it("fails a full panel without truncation when configured limits are too small", async () => {
    const matrix = cloneFixture(matrixFixture)
    matrix.limits.max_personas = 1
    matrix.limits.max_total_runs = 1
    const discovered = manifestFixture.reviewers.map((reviewer) => reviewer.agent)
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: discovered,
        full: true,
        augment: true,
        changed_files: [changedFile("feature.ts")],
      },
      dependencies(matrix),
    )

    expect(plan.status).toBe("inconclusive")
    expect(plan.limit_state).toEqual({
      required_personas: 6,
      configured_personas: 1,
      required_runs: 6,
      configured_runs: 1,
    })
    expect(plan.entries.filter((entry) => entry.decision === "include")).toHaveLength(6)
    expect(plan.entries.filter((entry) => entry.reason_code === "content-capability")).toHaveLength(3)
    expect(
      plan.entries
        .filter((entry) => entry.decision === "include")
        .every((entry) => entry.tier === "standard"),
    ).toBe(true)
  })

  it("emits byte-stable plan serialization", async () => {
    const input = {
      mode: "code" as const,
      discovered_agents: ["divisor-testing", "divisor-guard", "divisor-adversary"],
      changed_files: [changedFile("feature_test.go", 10)],
    }
    const first = serializeDispatchPlan(await planReviewDispatch(input, dependencies()))
    const second = serializeDispatchPlan(await planReviewDispatch(input, dependencies()))

    expect(first).toBe(second)
    expect(first.endsWith("\n")).toBe(true)
    expect(JSON.parse(first) as JsonObject).toHaveProperty("plan_version", 1)
  })
})

describe("optional profile models", () => {
  it("parses version-3 matrix with omitted model field", () => {
    const v3 = cloneFixture(matrixFixture)
    delete (v3.profiles as Record<string, unknown>).lightweight.model
    const matrix = parseReviewMatrix(JSON.stringify(v3), jsonParser)
    expect(matrix.version).toBe(3)
    expect(matrix.profiles.lightweight.model).toBeUndefined()
    expect(matrix.profiles.standard.model).toBe("provider/standard")
  })

  it("parses version-3 matrix with explicit model field", () => {
    const matrix = parseReviewMatrix(JSON.stringify(matrixFixture), jsonParser)
    expect(matrix.version).toBe(3)
    expect(matrix.profiles.standard.model).toBe("provider/standard")
  })

  it("accepts version-2 matrix with all models present", () => {
    // RD-FR-023: "Version 2 matrix loads under v2 schema"
    const v2 = { ...cloneFixture(matrixFixture), version: 2 }
    const matrix = parseReviewMatrix(JSON.stringify(v2), jsonParser)
    expect(matrix.version).toBe(2)
    expect(matrix.profiles.standard.model).toBe("provider/standard")
  })

  it("rejects version-2 matrix with missing model", () => {
    // RD-FR-023: "Version 2 matrix missing model field fails"
    const v2 = { ...cloneFixture(matrixFixture), version: 2 }
    delete (v2.profiles as Record<string, unknown>).heavy.model
    expect(() => parseReviewMatrix(JSON.stringify(v2), jsonParser)).toThrow(
      /version 2 requires model/,
    )
  })

  it("rejects version beyond 3", () => {
    const v4 = { ...cloneFixture(matrixFixture), version: 4 }
    expect(() => parseReviewMatrix(JSON.stringify(v4), jsonParser)).toThrow()
  })

  it("rejects profile with invalid model", () => {
    const bad = cloneFixture(matrixFixture)
    ;(bad.profiles as Record<string, unknown>).standard = { model: "not-a-valid-model!!!" }
    expect(() => parseReviewMatrix(JSON.stringify(bad), jsonParser)).toThrow()
  })

  it("returns model: null from profilePair when profile has no model", async () => {
    const v3 = cloneFixture(matrixFixture)
    delete (v3.profiles as Record<string, unknown>).lightweight.model
    // Use a CI file to match divisor-sre's ci-cd scope
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-sre"],
        changed_files: [changedFile(".github/workflows/ci.yml", 5)],
      },
      dependencies(v3),
    )
    const sreEntry = plan.entries.find((e) => e.agent === "divisor-sre" && e.decision === "include")
    expect(sreEntry).toBeDefined()
    expect(sreEntry!.model).toBeNull()
  })

  it("returns explicit model from profilePair when profile has model", async () => {
    // Use a matrix without explicit runs so the advisor path is taken
    const noRunsMatrix = cloneFixture(matrixFixture)
    delete (noRunsMatrix as Record<string, unknown>).runs
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(noRunsMatrix),
    )
    const advEntry = plan.entries.find((e) => e.agent === "divisor-adversary" && e.decision === "include")
    expect(advEntry).toBeDefined()
    // adversary is in always→heavy, so model should be provider/heavy
    expect(advEntry!.model).toBe("provider/heavy")
  })
})

describe("planner tier resolvability (RD-FR-063/064)", () => {
  // D1 (RD-FR-063) advisories carry a signature phrase no other advisory
  // shares; stale-manifest and tier-cap advisories share the array.
  const inheritanceAdvisories = (plan: DispatchPlan): DispatchPlan["advisories"] =>
    plan.advisories.filter((advisory) => advisory.description.includes("has no model configured"))

  // Matrix version 3 allows profiles with no model; the planner treats
  // them as host-model inheritance destinations for advisor runs.
  const matrixWithoutModel = (...profiles: string[]): MutableMatrixFixture => {
    const matrix = cloneFixture(matrixFixture) as MutableMatrixFixture
    const profileRecords = matrix.profiles as Record<string, { model?: string }>
    for (const profile of profiles) {
      delete profileRecords[profile]!.model
    }
    return matrix
  }

  it("signals host-model inheritance for an advisor run on a model-less profile", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(matrixWithoutModel("standard")),
    )

    expect(plan.status).toBe("ready")
    const entry = included(plan, "divisor-guard")[0]
    expect(entry).toBeDefined()
    expect(entry).toMatchObject({ source: "advisor", model: null })

    const advisories = inheritanceAdvisories(plan)
    expect(advisories).toHaveLength(1)
    expect(advisories[0]!.severity).toBe("LOW")
    expect(advisories[0]!.description).toContain('profile "standard"')
    expect(advisories[0]!.description).toContain("divisor-guard")
    expect(advisories[0]!.description).toContain("advisor host-model fallback")
    expect(advisories[0]!.description).toContain("configure profiles.standard.model")
  })

  it("emits one advisory for a full panel on a model-less standard profile", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        full: true,
        discovered_agents: manifestFixture.reviewers.map((reviewer) => reviewer.agent),
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(matrixWithoutModel("standard")),
    )

    expect(plan.status).toBe("ready")
    const includedEntries = plan.entries.filter((entry) => entry.decision === "include")
    expect(includedEntries).toHaveLength(6)
    expect(
      includedEntries.every((entry) => entry.source === "advisor" && entry.model === null),
    ).toBe(true)

    // One advisory for the profile names the review-capable agents in plan
    // order — not one advisory per entry.
    const advisories = inheritanceAdvisories(plan)
    expect(advisories).toHaveLength(1)
    expect(advisories[0]!.description).toContain('profile "standard"')
    expect(advisories[0]!.description).toContain(
      "advisor-destination runs (divisor-adversary, divisor-architect, divisor-curator, " +
        "divisor-guard, divisor-sre, divisor-testing)",
    )
  })

  it("emits an advisory when opt-in augmentation selects a model-less adjacent profile", async () => {
    const matrix = matrixWithoutModel("standard")
    // Drop adversary's explicit runs so its base run resolves through the
    // advisor branch (always-pinned heavy) and only the augmentation run
    // lands on the model-less adjacent profile.
    delete matrix.runs.code["divisor-adversary"]
    const plan = await planReviewDispatch(
      {
        mode: "code",
        augment: true,
        discovered_agents: ["divisor-adversary"],
        changed_files: [changedFile("api/auth.ts", 10)],
      },
      dependencies(matrix),
    )

    expect(plan.status).toBe("ready")
    const entries = included(plan, "divisor-adversary")
    expect(entries).toHaveLength(2)
    expect(entries[0]).toMatchObject({ source: "advisor", model: "provider/heavy" })
    expect(entries[1]).toMatchObject({
      source: "advisor",
      model: null,
      tier: "standard",
      reason_code: "risk-augmentation",
    })

    const advisories = inheritanceAdvisories(plan)
    expect(advisories).toHaveLength(1)
    expect(advisories[0]!.description).toContain('profile "standard"')
    expect(advisories[0]!.description).toContain("advisor host-model fallback")
  })

  it("emits no advisory when every profile referenced by the plan has a model", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-architect", "divisor-guard"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(),
    )

    expect(plan.status).toBe("ready")
    const architectEntry = included(plan, "divisor-architect")[0]
    const guardEntry = included(plan, "divisor-guard")[0]
    expect(architectEntry).toBeDefined()
    expect(guardEntry).toBeDefined()
    expect(architectEntry!.model).toBe("provider/standard")
    expect(guardEntry!.model).toBe("provider/standard")
    expect(inheritanceAdvisories(plan)).toHaveLength(0)
  })

  it("emits no advisory for host-fallback entries on a model-less matrix", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard", "divisor-sre"],
        changed_files: [changedFile(".github/workflows/ci.yml", 5)],
      },
      dependencies(matrixWithoutModel("standard")),
    )

    expect(plan.status).toBe("ready")
    // divisor-sre is absent from runs and advisor.code for code mode, so it
    // takes the host fallback — model null with its existing self-describing
    // reason, and no advisory of its own.
    const sreEntry = included(plan, "divisor-sre")[0]
    expect(sreEntry).toBeDefined()
    expect(sreEntry).toMatchObject({ source: "host", model: null })
    expect(sreEntry!.reason).toContain("active host model")

    const advisories = inheritanceAdvisories(plan)
    expect(advisories).toHaveLength(1)
    expect(advisories[0]!.description).toContain("divisor-guard")
    expect(advisories[0]!.description).not.toContain("divisor-sre")
  })

  it("produces identical advisory arrays across identical plans, sorted by profile name", async () => {
    const matrix = matrixWithoutModel("standard", "heavy")
    delete matrix.runs.code["divisor-adversary"]
    const input = {
      mode: "code" as const,
      discovered_agents: ["divisor-adversary", "divisor-guard"],
      changed_files: [changedFile("src/feature.ts", 5)],
    }
    const first = await planReviewDispatch(input, dependencies(matrix))
    const second = await planReviewDispatch(input, dependencies(matrix))

    expect(first.status).toBe("ready")
    expect(first.advisories).toEqual(second.advisories)

    const advisories = inheritanceAdvisories(first)
    expect(advisories).toHaveLength(2)
    // one advisory per distinct model-less profile, sorted by profile name
    expect(advisories[0]!.description).toContain('profile "heavy"')
    expect(advisories[0]!.description).toContain("divisor-adversary")
    expect(advisories[1]!.description).toContain('profile "standard"')
    expect(advisories[1]!.description).toContain("divisor-guard")
  })

  it("fails closed on an explicit profile run against a model-less profile", async () => {
    const matrix = matrixWithoutModel("heavy")
    matrix.runs.code["divisor-architect"] = [{ profile: "heavy" }]
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-architect"],
        changed_files: [changedFile("src/feature.ts", 60)],
      },
      dependencies(matrix),
    )

    expect(plan.status).toBe("inconclusive")
    expect(plan.workflow_result).toBe("INCONCLUSIVE")

    const entry = plan.entries.find((e) => e.agent === "divisor-architect")
    expect(entry).toBeDefined()
    expect(entry!.decision).toBe("skip")
    expect(entry!.reason_code).toBe("invalid-explicit-run")
    expect(entry!.validation_errors).toHaveLength(1)
    expect(entry!.validation_errors[0]).toContain('profile "heavy" has no model configured')
    expect(entry!.validation_errors[0]).toContain(
      "explicit-source runs cannot inherit the host model",
    )
    expect(entry!.validation_errors[0]).toContain("configure profiles.heavy.model")

    const error = plan.errors.find((e) => e.includes("divisor-architect"))
    expect(error).toBeDefined()
    expect(error).toContain("divisor-architect run 1")
    expect(error).toContain('profile "heavy"')
    expect(error).toContain("configure profiles.heavy.model")

    // The fail-closed rejection rides the error channel, not an advisory.
    expect(inheritanceAdvisories(plan)).toHaveLength(0)
  })

  it("keeps direct-model explicit runs unaffected", async () => {
    const matrix = cloneFixture(matrixFixture) as MutableMatrixFixture
    matrix.runs.code["divisor-architect"] = [{ model: "provider/direct-model" }]
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-architect"],
        changed_files: [changedFile("src/feature.ts", 60)],
      },
      dependencies(matrix),
    )

    expect(plan.status).toBe("ready")
    const entry = included(plan, "divisor-architect")[0]
    expect(entry).toBeDefined()
    expect(entry).toMatchObject({
      source: "explicit",
      model: "provider/direct-model",
      reason_code: "explicit-run",
    })
    expect(entry!.validation_errors).toHaveLength(0)
    expect(inheritanceAdvisories(plan)).toHaveLength(0)
  })

  it("keeps advisor runs on the same model-less profile includable per RD-FR-063", async () => {
    const matrix = matrixWithoutModel("heavy")
    delete matrix.runs.code["divisor-adversary"]
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(matrix),
    )

    expect(plan.status).toBe("ready")
    expect(plan.workflow_result).toBeNull()
    const entry = included(plan, "divisor-adversary")[0]
    expect(entry).toBeDefined()
    expect(entry).toMatchObject({
      source: "advisor",
      model: null,
      reason_code: "advisor-run",
    })

    const advisories = inheritanceAdvisories(plan)
    expect(advisories).toHaveLength(1)
    expect(advisories[0]!.severity).toBe("LOW")
    expect(advisories[0]!.description).toContain('profile "heavy"')
    expect(advisories[0]!.description).toContain("divisor-adversary")
    expect(advisories[0]!.description).toContain("advisor host-model fallback")
  })
})

describe("override file loading and merging", () => {
  it("uses base matrix when no override file exists", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(),
    )
    expect(plan.status).toBe("ready")
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.model).toBe("provider/standard")
  })

  it("merges override model into base matrix", async () => {
    const override = {
      profiles: {
        standard: { model: "provider/override-model" },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    expect(plan.status).toBe("ready")
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.model).toBe("provider/override-model")
  })

  it("merges override variant into base matrix", async () => {
    const override = {
      profiles: {
        standard: { variant: "override-variant" },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.variant).toBe("override-variant")
  })

  it("absent override keys fall through to base", async () => {
    const override = {
      profiles: {
        heavy: { model: "provider/override-heavy" },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    // divisor-guard uses standard profile, not heavy — should be unchanged
    expect(guardEntry!.model).toBe("provider/standard")
  })

  it("empty override returns base unchanged", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, {}),
    )
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.model).toBe("provider/standard")
  })

  it("override with model: null explicitly sets null", async () => {
    const override = {
      profiles: {
        standard: { model: null },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.model).toBeNull()
  })

  it("override with variant only (no model) works", async () => {
    const override = {
      profiles: {
        standard: { variant: "variant-only" },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.variant).toBe("variant-only")
    expect(guardEntry!.model).toBe("provider/standard") // model unchanged
  })

  it("override value replaces base value (last wins semantics)", async () => {
    const override = {
      profiles: {
        standard: { model: "provider/last" },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.model).toBe("provider/last")
  })
})

describe("override validation", () => {
  it("rejects non-whitelisted top-level key", async () => {
    const override = { defaults: { code: "lightweight" } }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    expect(plan.status).toBe("inconclusive")
    expect(plan.workflow_result).toBe("INCONCLUSIVE")
    expect(plan.errors.some((e) => e.includes("defaults"))).toBe(true)
  })

  it("rejects non-whitelisted nested key in profile", async () => {
    const override = {
      profiles: {
        standard: { unexpected_key: "value" },
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    expect(plan.status).toBe("inconclusive")
    expect(plan.workflow_result).toBe("INCONCLUSIVE")
    expect(plan.errors.some((e) => e.includes("unexpected_key"))).toBe(true)
  })

  it("rejects non-whitelisted key like 'always'", async () => {
    const override = { always: { "divisor-guard": "lightweight" } }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(matrixFixture, manifestFixture, override),
    )
    expect(plan.status).toBe("inconclusive")
    expect(plan.workflow_result).toBe("INCONCLUSIVE")
    expect(plan.errors.some((e) => e.includes("always"))).toBe(true)
  })

  it("override with no model in any profile loads successfully (soft gate only)", async () => {
    const noModelMatrix = cloneFixture(matrixFixture)
    delete (noModelMatrix.profiles as Record<string, unknown>).lightweight.model
    delete (noModelMatrix.profiles as Record<string, unknown>).standard.model
    delete (noModelMatrix.profiles as Record<string, unknown>).heavy.model
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(noModelMatrix),
    )
    // Should load successfully — soft gate only warns, doesn't fail
    expect(plan.status).toBe("ready")
    const guardEntry = plan.entries.find((e) => e.agent === "divisor-guard" && e.decision === "include")
    expect(guardEntry!.model).toBeNull()
  })
})

describe("override file read failure", () => {
  it("produces INCONCLUSIVE when override file is unparseable", async () => {
    let parseCallCount = 0
    const deps: PlannerDependencies = {
      readText: async (path: string) => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") return "not valid yaml {{{"
        throw new Error(`unexpected read: ${path}`)
      },
      parseYaml: (text: string) => {
        parseCallCount++
        // First two calls are for base matrix and manifest — parse normally.
        // Third call is for the override — simulate parse failure.
        if (parseCallCount <= 2) return JSON.parse(text) as unknown
        throw new Error("YAML parse failure")
      },
    }
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      deps,
    )
    expect(plan.status).toBe("inconclusive")
    expect(plan.workflow_result).toBe("INCONCLUSIVE")
    expect(plan.errors.some((e) => e.includes("INCONCLUSIVE") || e.includes("override"))).toBe(true)
  })
})

describe("planner error remediation (unleash-pipeline-hardening)", () => {
  it("enriches the unregistered discovered-agent error with remediation and both lists", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "specs",
        discovered_agents: ["divisor-guard", "divisor-ghost"],
        changed_files: [changedFile("specs/042-feature/spec.md", 20)],
      },
      dependencies(),
    )

    expect(plan.status).toBe("inconclusive")
    expect(plan.workflow_result).toBe("INCONCLUSIVE")
    const ghostError = plan.errors.find((e) => e.includes("divisor-ghost"))
    expect(ghostError).toBeDefined()
    expect(ghostError).toContain("discovered agent divisor-ghost has no reviewer-manifest entry")
    expect(ghostError).toContain("Remediation: add an entry for divisor-ghost to .uf/reviewer-capabilities.yaml")
    expect(ghostError).toContain("Discovered agents: divisor-ghost, divisor-guard")
    expect(ghostError).toContain(
      "Manifest-eligible agents: divisor-adversary, divisor-architect, divisor-curator, " +
        "divisor-envoy, divisor-guard, divisor-herald, divisor-scribe, divisor-sre, divisor-testing",
    )
  })
})

describe("planner drift check (unleash-pipeline-hardening)", () => {
  it("warns as a LOW advisory when a manifest reviewer was not discovered", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-guard"],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(),
    )

    expect(plan.status).toBe("ready")
    const staleAdvisories = plan.advisories.filter((a) => a.description.includes("stale"))
    expect(staleAdvisories.length).toBeGreaterThan(0)
    const adversaryAdvisory = staleAdvisories.find((a) => a.description.includes("divisor-adversary"))
    expect(adversaryAdvisory).toBeDefined()
    expect(adversaryAdvisory!.severity).toBe("LOW")
    expect(adversaryAdvisory!.description).toContain("stale .uf/reviewer-capabilities.yaml entry")
    expect(adversaryAdvisory!.description).toContain("it will not be scheduled")
  })

  it("emits no stale advisory when every manifest reviewer was discovered", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: [
          "divisor-adversary",
          "divisor-architect",
          "divisor-curator",
          "divisor-envoy",
          "divisor-guard",
          "divisor-herald",
          "divisor-scribe",
          "divisor-sre",
          "divisor-testing",
        ],
        changed_files: [changedFile("feature.ts", 5)],
      },
      dependencies(),
    )

    expect(plan.status).toBe("ready")
    expect(plan.advisories.filter((a) => a.description.includes("stale"))).toHaveLength(0)
  })
})

describe("per-run timeout derivation matrix fields (RD-FR-009)", () => {
  it("parses optional ceiling and per-tier override fields", () => {
    const matrix = cloneFixture(matrixFixture) as Record<string, unknown>
    matrix.limits = {
      per_run_timeout_seconds: 1200,
      per_run_timeout_ceiling_seconds: 1500,
      per_run_timeout_tiers: { lightweight: 900, heavy: 1500 },
    }

    const parsed = parseReviewMatrix(JSON.stringify(matrix), jsonParser)

    expect(parsed.limits?.per_run_timeout_seconds).toBe(1200)
    expect(parsed.limits?.per_run_timeout_ceiling_seconds).toBe(1500)
    expect(parsed.limits?.per_run_timeout_tiers?.lightweight).toBe(900)
    expect(parsed.limits?.per_run_timeout_tiers?.heavy).toBe(1500)
    expect(parsed.limits?.per_run_timeout_tiers?.standard).toBeUndefined()
  })

  it("keeps matrix-defaults fallback when derivation fields are absent", async () => {
    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(),
    )
    expect(plan.status).toBe("ready")

    // Fixture limits carry no derivation fields: the base stays 600 s,
    // the ceiling defaults to the 1800 s schema bound, and every
    // per-tier override stays unset (design D6 / RD-FR-009).
    expect(plan.limits.per_run_timeout_seconds).toBe(600)
    expect(plan.limits.per_run_timeout_ceiling_seconds).toBe(1800)
    expect(plan.limits.per_run_timeout_tiers).toEqual({
      lightweight: undefined,
      standard: undefined,
      heavy: undefined,
    })
  })

  it("falls back when the matrix carries no limits block at all", async () => {
    const noLimits = cloneFixture(matrixFixture) as Record<string, unknown>
    delete noLimits.limits

    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(noLimits),
    )
    expect(plan.status).toBe("ready")
    expect(plan.limits.per_run_timeout_seconds).toBe(600)
    expect(plan.limits.per_run_timeout_ceiling_seconds).toBe(1800)
  })

  it("rejects out-of-range ceiling and tier override values", () => {
    const tooSmall = cloneFixture(matrixFixture) as Record<string, unknown>
    tooSmall.limits = { per_run_timeout_ceiling_seconds: 29 }
    expect(() => parseReviewMatrix(JSON.stringify(tooSmall), jsonParser)).toThrow("Too small")

    const tooBig = cloneFixture(matrixFixture) as Record<string, unknown>
    tooBig.limits = { per_run_timeout_ceiling_seconds: 1801 }
    expect(() => parseReviewMatrix(JSON.stringify(tooBig), jsonParser)).toThrow("Too big")

    const tierTooSmall = cloneFixture(matrixFixture) as Record<string, unknown>
    tierTooSmall.limits = { per_run_timeout_tiers: { heavy: 29 } }
    expect(() => parseReviewMatrix(JSON.stringify(tierTooSmall), jsonParser)).toThrow("Too small")

    const unknownTier = cloneFixture(matrixFixture) as Record<string, unknown>
    unknownTier.limits = { per_run_timeout_tiers: { medium: 1200 } }
    expect(() => parseReviewMatrix(JSON.stringify(unknownTier), jsonParser)).toThrow(
      "Unrecognized key",
    )
  })

  it("resolves per-tier overrides alongside the scalar base window (RD-FR-029)", async () => {
    const customMatrix = cloneFixture(matrixFixture) as Record<string, unknown>
    customMatrix.limits = {
      per_run_timeout_seconds: 1200,
      per_run_timeout_tiers: { heavy: 1500 },
    }

    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(customMatrix),
    )
    expect(plan.status).toBe("ready")
    expect(plan.limits.per_run_timeout_seconds).toBe(1200)
    expect(plan.limits.per_run_timeout_tiers.heavy).toBe(1500)

    const config = {
      per_run_timeout_seconds: plan.limits.per_run_timeout_seconds,
      per_run_timeout_tiers: plan.limits.per_run_timeout_tiers,
    }
    // The heavy override wins for heavy; standard and tier-less runs
    // fall back to the scalar limit.
    expect(resolveBaseWindowMs("heavy", config)).toBe(1_500_000)
    expect(resolveBaseWindowMs("standard", config)).toBe(1_200_000)
    expect(resolveBaseWindowMs(null, config)).toBe(1_200_000)
  })

  it("resolves a ceiling below the base window to the base window (floor precedence)", async () => {
    const customMatrix = cloneFixture(matrixFixture) as Record<string, unknown>
    customMatrix.limits = {
      per_run_timeout_seconds: 1200,
      per_run_timeout_ceiling_seconds: 600,
    }

    const plan = await planReviewDispatch(
      {
        mode: "code",
        discovered_agents: ["divisor-adversary", "divisor-guard"],
        changed_files: [changedFile("src/feature.ts", 5)],
      },
      dependencies(customMatrix),
    )
    expect(plan.status).toBe("ready")
    expect(plan.limits.per_run_timeout_seconds).toBe(1200)
    expect(plan.limits.per_run_timeout_ceiling_seconds).toBe(600)

    // The derivation raises a configured ceiling below the resolved
    // base window back to the base (RD-FR-028 floor precedence).
    const baseMs = resolveBaseWindowMs("standard", {
      per_run_timeout_seconds: plan.limits.per_run_timeout_seconds,
    })
    const result = derivePerRunTimeoutMs({
      baseMs,
      promptBytes: 65_536,
      resolvedTier: "standard",
      sessionProfileTier: "standard",
      maxParallelRuns: 1,
      configuredCeilingMs: plan.limits.per_run_timeout_ceiling_seconds * 1000,
    })
    expect(result.effectiveMs).toBe(baseMs)
    expect(result.provenance.ceiling_ms).toBe(baseMs)
  })
})

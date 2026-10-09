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

    expect(plan.advisories).toHaveLength(1)
    expect(plan.advisories[0]!.severity).toBe("HIGH")
    expect(plan.advisories[0]!.description).toContain("tier_caps.heavy")
    expect(plan.advisories[0]!.description).toContain("security-sensitive")
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

    // Default config has tier_caps.heavy = null — no advisory
    expect(plan.advisories).toHaveLength(0)
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

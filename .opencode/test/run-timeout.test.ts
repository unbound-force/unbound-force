/**
 * Tests for the shared per-run and per-task timeout derivations
 * (OpenSpec change derive-per-run-timeout, task 1.2 — RD-FR-028
 * scenarios; OpenSpec change per-task-delegation-budgets, task 1.2 —
 * PTB-FR-001 scenarios).
 *
 * Both derivations are pure functions of scalar signals (constitution
 * Principles II/IV): formula arithmetic, monotonicity in prompt bytes
 * and tier, contention gating and cap, floor-at-base, ceiling clamp
 * including the 1_800_000 executor bound, ceiling-below-base floor
 * precedence, tier-less zero profile term, determinism, and the
 * explicit-source passthrough (design D1/D2/D5). The per-task variant
 * adds the phase-shape signals: monotonicity in phase task and file
 * count, the flat-default replacement, and the full per-task
 * provenance record (per-task-delegation-budgets design D1).
 */
import { describe, expect, it } from "vitest"

import {
  CONTENTION_FACTOR_PER_PARALLEL_RUN,
  MANIFEST_RETRIEVAL_ALLOWANCE_MS,
  MAX_CONTENTION_MULTIPLIER,
  MAX_TIMEOUT_MS,
  MS_PER_KIB_ABOVE_FREE,
  MS_PER_PHASE_FILE_ABOVE_FREE,
  MS_PER_PHASE_TASK_ABOVE_FREE,
  PHASE_FREE_FILES,
  PHASE_FREE_TASKS,
  PROMPT_FREE_BYTES,
  TIER_TERM_MS,
  contentionMultiplier,
  derivePerRunTimeoutMs,
  derivePerTaskTimeoutMs,
  phaseFileTermMs,
  phaseTaskTermMs,
  profileTermMs,
  promptTermMs,
  resolveBaseWindowMs,
  type PerTaskTimeoutSignals,
  type TimeoutDerivationSignals,
} from "../lib/run-timeout.js"
import { TimeoutSchema } from "../lib/agent-executor.js"

/** Baseline signals matching the 2026-10-27 incident panel shape. */
function signals(overrides: Partial<TimeoutDerivationSignals> = {}): TimeoutDerivationSignals {
  return {
    baseMs: 1_200_000,
    promptBytes: PROMPT_FREE_BYTES,
    resolvedTier: "standard",
    sessionProfileTier: "standard",
    maxParallelRuns: 1,
    configuredCeilingMs: MAX_TIMEOUT_MS,
    ...overrides,
  }
}

/**
 * Baseline per-task signals matching the issue #1080 incident panel
 * shape: a single-task phase (the delegated task is the free window),
 * the free core file set, a prompt at the free window, no tier
 * signal, and a sequential fan-out. The 20-minute base matches the
 * panel's successful-delegation upper bound.
 */
function perTaskSignals(overrides: Partial<PerTaskTimeoutSignals> = {}): PerTaskTimeoutSignals {
  return {
    baseMs: 1_200_000,
    phaseTaskCount: PHASE_FREE_TASKS,
    phaseFileCount: PHASE_FREE_FILES,
    promptBytes: PROMPT_FREE_BYTES,
    tier: null,
    maxParallelRuns: 1,
    configuredCeilingMs: MAX_TIMEOUT_MS,
    ...overrides,
  }
}

describe("promptTermMs", () => {
  it("adds nothing at or below the free window", () => {
    expect(promptTermMs(0)).toBe(0)
    expect(promptTermMs(PROMPT_FREE_BYTES)).toBe(0)
  })

  it("charges whole KiB above the free window (D5 example)", () => {
    // 179_200 - 65_536 = 113_664 bytes = 111 KiB; 111 x 1_500 = 166_500.
    expect(promptTermMs(179_200)).toBe(166_500)
  })

  it("truncates to whole KiB at the free-window boundary (D5)", () => {
    // 179_200 above is exactly 111 KiB above the free window, where
    // floor, round, and ceil agree — these boundary pairs pin the
    // whole-KiB truncation contract itself: a partial KiB charges
    // nothing, the byte that completes it charges the full term.
    expect(promptTermMs(PROMPT_FREE_BYTES + 1023)).toBe(0)
    expect(promptTermMs(PROMPT_FREE_BYTES + 1024)).toBe(MS_PER_KIB_ABOVE_FREE)
  })

  it("is monotonically non-decreasing in prompt bytes", () => {
    let previous = -1
    for (let bytes = 0; bytes <= 400_000; bytes += 517) {
      const term = promptTermMs(bytes)
      expect(term).toBeGreaterThanOrEqual(previous)
      previous = term
    }
  })
})

describe("profileTermMs", () => {
  it("applies the tier additive terms", () => {
    expect(profileTermMs("lightweight")).toBe(TIER_TERM_MS.lightweight)
    expect(profileTermMs("standard")).toBe(TIER_TERM_MS.standard)
    expect(profileTermMs("heavy")).toBe(TIER_TERM_MS.heavy)
    expect(TIER_TERM_MS.lightweight).toBe(0)
    expect(TIER_TERM_MS.standard).toBe(120_000)
    expect(TIER_TERM_MS.heavy).toBe(300_000)
  })

  it("derives zero for a tier-less (direct-model) entry", () => {
    expect(profileTermMs(null)).toBe(0)
  })
})

describe("contentionMultiplier", () => {
  it("is 1 at or below one parallel run", () => {
    expect(contentionMultiplier(1)).toBe(1)
    expect(contentionMultiplier(0)).toBe(1)
  })

  it("applies the configured factor above one run", () => {
    expect(contentionMultiplier(2)).toBe(1 + CONTENTION_FACTOR_PER_PARALLEL_RUN)
    expect(contentionMultiplier(4)).toBe(1.45)
  })

  it("caps at the maximum multiplier", () => {
    expect(contentionMultiplier(5)).toBe(MAX_CONTENTION_MULTIPLIER)
    expect(contentionMultiplier(8)).toBe(MAX_CONTENTION_MULTIPLIER)
    expect(MAX_CONTENTION_MULTIPLIER).toBe(1.5)
  })
})

describe("resolveBaseWindowMs", () => {
  it("uses the scalar limit for tier-less runs", () => {
    expect(resolveBaseWindowMs(null, { per_run_timeout_seconds: 1200 })).toBe(1_200_000)
  })

  it("uses the per-tier override when configured for the resolved tier", () => {
    expect(
      resolveBaseWindowMs("heavy", {
        per_run_timeout_seconds: 1200,
        per_run_timeout_tiers: { heavy: 1500 },
      }),
    ).toBe(1_500_000)
  })

  it("falls back to the scalar limit when the tier has no override", () => {
    expect(
      resolveBaseWindowMs("standard", {
        per_run_timeout_seconds: 1200,
        per_run_timeout_tiers: { heavy: 1500 },
      }),
    ).toBe(1_200_000)
  })
})

describe("derivePerRunTimeoutMs", () => {
  it("composes the formula: (base + prompt + profile) x multiplier", () => {
    const result = derivePerRunTimeoutMs(
      signals({ promptBytes: 179_200, resolvedTier: "heavy", maxParallelRuns: 1 }),
    )
    // 1_200_000 + 166_500 + 300_000 = 1_666_500, no multiplier.
    expect(result.effectiveMs).toBe(1_666_500)
    expect(result.provenance.source).toBe("derived")
    expect(result.provenance.clamped).toBe(false)
  })

  it("matches the incident-panel calibration (4-parallel, clamped at ceiling)", () => {
    const result = derivePerRunTimeoutMs(
      signals({ promptBytes: 179_200, resolvedTier: "heavy", maxParallelRuns: 4 }),
    )
    // 1_666_500 x 1.45 = 2_416_425 -> clamped to the 1_800_000 ceiling.
    expect(result.effectiveMs).toBe(MAX_TIMEOUT_MS)
    expect(result.provenance.clamped).toBe(true)
    expect(result.provenance.contention_multiplier).toBe(1.45)
    expect(result.provenance.ceiling_ms).toBe(MAX_TIMEOUT_MS)
  })

  it("is monotonically non-decreasing in prompt bytes", () => {
    let previous = -1
    for (let bytes = 0; bytes <= 400_000; bytes += 4_031) {
      const { effectiveMs } = derivePerRunTimeoutMs(signals({ promptBytes: bytes }))
      expect(effectiveMs).toBeGreaterThanOrEqual(previous)
      previous = effectiveMs
    }
  })

  it("is monotonically non-decreasing in tier", () => {
    const tiers = ["lightweight", "standard", "heavy"] as const
    let previous = -1
    for (const tier of tiers) {
      const { effectiveMs } = derivePerRunTimeoutMs(signals({ resolvedTier: tier }))
      expect(effectiveMs).toBeGreaterThanOrEqual(previous)
      previous = effectiveMs
    }
  })

  it("never reduces below the base window (floor)", () => {
    // Zero prompt and zero profile terms leave exactly the base.
    const result = derivePerRunTimeoutMs(
      signals({ promptBytes: 0, resolvedTier: "lightweight", maxParallelRuns: 1 }),
    )
    expect(result.effectiveMs).toBe(1_200_000)
  })

  it("clamps at the executor bound even with a higher configured ceiling", () => {
    const result = derivePerRunTimeoutMs(
      signals({
        promptBytes: 4_000_000,
        resolvedTier: "heavy",
        maxParallelRuns: 8,
        configuredCeilingMs: 3_600_000,
      }),
    )
    expect(result.effectiveMs).toBe(MAX_TIMEOUT_MS)
    expect(result.provenance.ceiling_ms).toBe(MAX_TIMEOUT_MS)
  })

  it("raises a configured ceiling below the base to the base (floor precedence)", () => {
    const result = derivePerRunTimeoutMs(signals({ configuredCeilingMs: 300_000 }))
    expect(result.effectiveMs).toBe(1_200_000)
    expect(result.provenance.ceiling_ms).toBe(1_200_000)
  })

  it("derives a zero profile term for tier-less entries", () => {
    const result = derivePerRunTimeoutMs(
      signals({ resolvedTier: null, sessionProfileTier: "heavy" }),
    )
    expect(result.effectiveMs).toBe(1_200_000)
    expect(result.provenance.profile_term_ms).toBe(0)
    expect(result.provenance.resolved_tier).toBeNull()
    // The session profile tier is provenance only and never alters the term.
    expect(result.provenance.session_profile_tier).toBe("heavy")
  })

  it("records every derivation input in the provenance record (D5)", () => {
    const result = derivePerRunTimeoutMs(
      signals({ promptBytes: 179_200, resolvedTier: "heavy", maxParallelRuns: 4 }),
    )
    expect(result.provenance).toEqual({
      effective_ms: MAX_TIMEOUT_MS,
      source: "derived",
      base_ms: 1_200_000,
      prompt_bytes: 179_200,
      prompt_term_ms: 166_500,
      profile_term_ms: 300_000,
      resolved_tier: "heavy",
      session_profile_tier: "standard",
      contention_multiplier: 1.45,
      max_parallel_runs: 4,
      ceiling_ms: MAX_TIMEOUT_MS,
      clamped: true,
    })
  })

  it("is deterministic: identical inputs produce identical output", () => {
    const input = signals({ promptBytes: 200_000, resolvedTier: "standard", maxParallelRuns: 4 })
    expect(derivePerRunTimeoutMs(input)).toEqual(derivePerRunTimeoutMs(input))
  })

  it("passes an explicit timeout through unchanged with source explicit", () => {
    const result = derivePerRunTimeoutMs(signals({ explicitTimeoutMs: 900_000 }))
    expect(result.effectiveMs).toBe(900_000)
    expect(result.provenance).toEqual({ effective_ms: 900_000, source: "explicit" })
  })
})

describe("manifest-mode retrieval allowance (RD-FR-036)", () => {
  it("manifest-mode run gains exactly the allowance (absent clamping)", () => {
    // Baseline shape: base 1.2M + standard 120k term, no prompt term,
    // no contention. Inline: 1_320_000. Manifest: +300_000.
    const inline = derivePerRunTimeoutMs(signals({ manifestMode: false }))
    const manifest = derivePerRunTimeoutMs(signals({ manifestMode: true }))
    expect(inline.effectiveMs).toBe(1_320_000)
    expect(manifest.effectiveMs).toBe(inline.effectiveMs + MANIFEST_RETRIEVAL_ALLOWANCE_MS)
    expect(manifest.effectiveMs).toBe(1_620_000)
    expect(manifest.provenance.retrieval_allowance_ms).toBe(MANIFEST_RETRIEVAL_ALLOWANCE_MS)
  })

  it("allowance applies after the contention multiplier, not inside it", () => {
    // A smaller base keeps both variants under the ceiling so the
    // difference is observable unclamped: with multiplier 1.45 the
    // manifest-vs-inline gap must remain exactly the allowance.
    const base = { baseMs: 600_000, resolvedTier: "standard" as const, maxParallelRuns: 4 }
    const inline = derivePerRunTimeoutMs(signals({ ...base, manifestMode: false }))
    const manifest = derivePerRunTimeoutMs(signals({ ...base, manifestMode: true }))
    expect(inline.provenance.contention_multiplier).toBe(1.45)
    // round((600_000 + 0 + 120_000) x 1.45) = 1_044_000.
    expect(inline.effectiveMs).toBe(1_044_000)
    expect(manifest.effectiveMs).toBe(1_044_000 + MANIFEST_RETRIEVAL_ALLOWANCE_MS)
    expect(manifest.provenance.retrieval_allowance_ms).toBe(MANIFEST_RETRIEVAL_ALLOWANCE_MS)
    expect(manifest.provenance.clamped).toBe(false)
  })

  it("inline-mode derivation identical to pre-change for the same inputs", () => {
    // Absent signal, explicit false, and explicit true-with-no-allowance
    // (undefined) are the pre-change shapes; the derivation result and
    // the provenance record must match the pre-change byte shape — no
    // retrieval_allowance_ms key at all.
    const pre = derivePerRunTimeoutMs(signals({}))
    const off = derivePerRunTimeoutMs(signals({ manifestMode: false }))
    expect(off).toEqual(pre)
    expect(pre.provenance).not.toHaveProperty("retrieval_allowance_ms")
    expect(off.provenance).not.toHaveProperty("retrieval_allowance_ms")
  })

  it("explicit argument still wins; the allowance does not modify it", () => {
    const result = derivePerRunTimeoutMs(signals({ manifestMode: true, explicitTimeoutMs: 900_000 }))
    expect(result.effectiveMs).toBe(900_000)
    expect(result.provenance).toEqual({ effective_ms: 900_000, source: "explicit" })
    expect(result.provenance).not.toHaveProperty("retrieval_allowance_ms")
  })

  it("allowance participates in ceiling clamping like any derived term", () => {
    // Inline derivation 1_320_000 fits under a 1_500_000 ceiling; the
    // manifest-mode 1_620_000 does not and clamps to the ceiling.
    const inline = derivePerRunTimeoutMs(signals({ configuredCeilingMs: 1_500_000 }))
    const manifest = derivePerRunTimeoutMs(signals({ configuredCeilingMs: 1_500_000, manifestMode: true }))
    expect(inline.effectiveMs).toBe(1_320_000)
    expect(inline.provenance.clamped).toBe(false)
    expect(manifest.effectiveMs).toBe(1_500_000)
    expect(manifest.provenance.clamped).toBe(true)
    expect(manifest.provenance.retrieval_allowance_ms).toBe(MANIFEST_RETRIEVAL_ALLOWANCE_MS)
  })

  it("derivation stays deterministic with the allowance applied", () => {
    const input = signals({ promptBytes: 200_000, manifestMode: true, maxParallelRuns: 4 })
    expect(derivePerRunTimeoutMs(input)).toEqual(derivePerRunTimeoutMs(input))
  })
})

describe("phaseTaskTermMs", () => {
  it("adds nothing at or below the free window", () => {
    expect(phaseTaskTermMs(0)).toBe(0)
    expect(phaseTaskTermMs(PHASE_FREE_TASKS)).toBe(0)
  })

  it("charges per sibling task above the free window (D1 calibration)", () => {
    expect(PHASE_FREE_TASKS).toBe(1)
    expect(MS_PER_PHASE_TASK_ABOVE_FREE).toBe(60_000)
    // The delegated task itself is the free window; the #1080
    // incident phase held 11 tasks — 10 siblings at 60 s each.
    expect(phaseTaskTermMs(11)).toBe(600_000)
  })

  it("charges the full term for the first task above the free window", () => {
    expect(phaseTaskTermMs(PHASE_FREE_TASKS + 1)).toBe(MS_PER_PHASE_TASK_ABOVE_FREE)
  })

  it("is monotonically non-decreasing in the phase task count", () => {
    let previous = -1
    for (let tasks = 0; tasks <= 40; tasks += 3) {
      const term = phaseTaskTermMs(tasks)
      expect(term).toBeGreaterThanOrEqual(previous)
      previous = term
    }
  })
})

describe("phaseFileTermMs", () => {
  it("adds nothing at or below the free window", () => {
    expect(phaseFileTermMs(0)).toBe(0)
    expect(phaseFileTermMs(PHASE_FREE_FILES)).toBe(0)
  })

  it("charges per file above the free window (D1 calibration)", () => {
    expect(PHASE_FREE_FILES).toBe(4)
    expect(MS_PER_PHASE_FILE_ABOVE_FREE).toBe(30_000)
    expect(phaseFileTermMs(9)).toBe(150_000)
  })

  it("charges the full term for the first file above the free window", () => {
    expect(phaseFileTermMs(PHASE_FREE_FILES + 1)).toBe(MS_PER_PHASE_FILE_ABOVE_FREE)
  })

  it("is monotonically non-decreasing in the phase file count", () => {
    let previous = -1
    for (let files = 0; files <= 40; files += 3) {
      const term = phaseFileTermMs(files)
      expect(term).toBeGreaterThanOrEqual(previous)
      previous = term
    }
  })
})

describe("derivePerTaskTimeoutMs", () => {
  it("composes the formula: (base + tier + task + file + prompt) x multiplier", () => {
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({
        baseMs: 600_000,
        phaseTaskCount: 6,
        phaseFileCount: 6,
        promptBytes: 179_200,
        tier: "heavy",
      }),
    )
    // 600_000 + 300_000 (heavy) + 300_000 (5 tasks above free) +
    // 60_000 (2 files above free) + 166_500 (111 KiB above free) =
    // 1_426_500, no multiplier.
    expect(result.effectiveMs).toBe(1_426_500)
    expect(result.provenance.source).toBe("derived")
    expect(result.provenance.clamped).toBe(false)
  })

  it("replaces the flat 1_800_000 per-task default with the derived window", () => {
    // PTB-FR-001 scenario "Flat default no longer applies for
    // per-task delegations": with no explicit timeout the effective
    // wall clock is the derived value from work signals — not the
    // 1_800_000 ms flat default the #1080 incident delegated under —
    // and the derivation provenance is present on the result.
    const result = derivePerTaskTimeoutMs(perTaskSignals({ phaseTaskCount: 3 }))
    // 1_200_000 + 2 x 60_000 = 1_320_000, under the flat default.
    expect(result.effectiveMs).toBe(1_320_000)
    expect(result.effectiveMs).not.toBe(MAX_TIMEOUT_MS)
    expect(result.provenance.source).toBe("derived")
    expect(result.provenance.phase_task_count).toBe(3)
  })

  it("scales the window with larger phases (more tasks or files)", () => {
    // PTB-FR-001 scenario "Larger phases scale the window": two
    // identical delegations except phase size — the larger phase's
    // effective timeout is greater than or equal to the smaller's.
    const small = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 2, phaseFileCount: 5 }),
    )
    const moreTasks = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 12, phaseFileCount: 5 }),
    )
    const moreFiles = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 2, phaseFileCount: 15 }),
    )
    // 1_290_000 vs 1_890_000 (clamped to 1_800_000) vs 1_590_000.
    expect(moreTasks.effectiveMs).toBeGreaterThan(small.effectiveMs)
    expect(moreFiles.effectiveMs).toBeGreaterThan(small.effectiveMs)
  })

  it("is monotonically non-decreasing in the phase task count", () => {
    let previous = -1
    for (let tasks = 0; tasks <= 30; tasks += 3) {
      const { effectiveMs } = derivePerTaskTimeoutMs(
        perTaskSignals({ phaseTaskCount: tasks }),
      )
      expect(effectiveMs).toBeGreaterThanOrEqual(previous)
      previous = effectiveMs
    }
  })

  it("is monotonically non-decreasing in the phase file count", () => {
    let previous = -1
    for (let files = 0; files <= 30; files += 3) {
      const { effectiveMs } = derivePerTaskTimeoutMs(
        perTaskSignals({ phaseFileCount: files }),
      )
      expect(effectiveMs).toBeGreaterThanOrEqual(previous)
      previous = effectiveMs
    }
  })

  it("is monotonically non-decreasing in tier", () => {
    const tiers = ["lightweight", "standard", "heavy"] as const
    let previous = -1
    for (const tier of tiers) {
      const { effectiveMs } = derivePerTaskTimeoutMs(perTaskSignals({ tier }))
      expect(effectiveMs).toBeGreaterThanOrEqual(previous)
      previous = effectiveMs
    }
  })

  it("reuses the shared per-run tier terms", () => {
    const lightweight = derivePerTaskTimeoutMs(perTaskSignals({ tier: "lightweight" }))
    const standard = derivePerTaskTimeoutMs(perTaskSignals({ tier: "standard" }))
    const heavy = derivePerTaskTimeoutMs(perTaskSignals({ tier: "heavy" }))
    // TIER_TERM_MS: lightweight 0, standard 120_000, heavy 300_000.
    expect(lightweight.effectiveMs).toBe(1_200_000)
    expect(standard.effectiveMs).toBe(1_320_000)
    expect(heavy.effectiveMs).toBe(1_500_000)
  })

  it("applies no contention at or below one parallel run", () => {
    // PTB-FR-001 item 4: the multiplier applies only when the
    // configured fan-out exceeds one concurrent delegation.
    const single = derivePerTaskTimeoutMs(perTaskSignals({ maxParallelRuns: 1 }))
    const zero = derivePerTaskTimeoutMs(perTaskSignals({ maxParallelRuns: 0 }))
    expect(single.provenance.contention_multiplier).toBe(1)
    expect(zero.provenance.contention_multiplier).toBe(1)
    expect(zero.effectiveMs).toBe(single.effectiveMs)
  })

  it("multiplies the derived sum by the contention multiplier", () => {
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({ baseMs: 600_000, phaseTaskCount: 6, maxParallelRuns: 2 }),
    )
    // (600_000 + 300_000) x 1.15 = 1_035_000.
    expect(result.effectiveMs).toBe(1_035_000)
    expect(result.provenance.contention_multiplier).toBe(1.15)
  })

  it("multiplies the window under parallel contention", () => {
    // PTB-FR-001 scenario "Parallel contention multiplies the
    // window": effective >= the single-run derivation, and the
    // unclamped derivation strictly greater.
    const single = derivePerTaskTimeoutMs(perTaskSignals({ phaseTaskCount: 6 }))
    const parallel = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 6, maxParallelRuns: 4 }),
    )
    expect(parallel.effectiveMs).toBeGreaterThanOrEqual(single.effectiveMs)
    expect(parallel.provenance.derived_ms).toBeGreaterThan(
      single.provenance.derived_ms,
    )
    // 1_500_000 x 1.45 = 2_175_000 -> clamped to the ceiling.
    expect(parallel.effectiveMs).toBe(MAX_TIMEOUT_MS)
    expect(parallel.provenance.contention_multiplier).toBe(1.45)
  })

  it("caps the contention multiplier at 1.5", () => {
    const five = derivePerTaskTimeoutMs(perTaskSignals({ maxParallelRuns: 5 }))
    const eight = derivePerTaskTimeoutMs(perTaskSignals({ maxParallelRuns: 8 }))
    expect(five.provenance.contention_multiplier).toBe(MAX_CONTENTION_MULTIPLIER)
    expect(eight.provenance.contention_multiplier).toBe(MAX_CONTENTION_MULTIPLIER)
    expect(MAX_CONTENTION_MULTIPLIER).toBe(1.5)
    expect(eight.effectiveMs).toBe(five.effectiveMs)
  })

  it("never reduces below the base window (floor)", () => {
    // Zero tier, task, file, and prompt terms leave exactly the base.
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 0, phaseFileCount: 0, promptBytes: 0 }),
    )
    expect(result.effectiveMs).toBe(1_200_000)
    expect(result.provenance.clamped).toBe(false)
  })

  it("clamps at a configured ceiling below the derived value", () => {
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 11, configuredCeilingMs: 1_500_000 }),
    )
    // 1_200_000 + 600_000 = 1_800_000 unclamped; the ceiling wins.
    expect(result.effectiveMs).toBe(1_500_000)
    expect(result.provenance.ceiling_ms).toBe(1_500_000)
    expect(result.provenance.clamped).toBe(true)
  })

  it("clamps at the ceiling with clamped provenance", () => {
    // PTB-FR-001 scenario "Derived value clamped at the ceiling":
    // work signals large enough that the unclamped derivation exceeds
    // the ceiling; the ceiling never exceeds the 1_800_000 bound.
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({
        phaseTaskCount: 11,
        phaseFileCount: 20,
        tier: "heavy",
        maxParallelRuns: 4,
      }),
    )
    // (1_200_000 + 300_000 + 600_000 + 480_000) x 1.45 = 3_741_000.
    expect(result.provenance.derived_ms).toBe(3_741_000)
    expect(result.effectiveMs).toBe(MAX_TIMEOUT_MS)
    expect(result.provenance.ceiling_ms).toBe(MAX_TIMEOUT_MS)
    expect(result.provenance.clamped).toBe(true)
  })

  it("clamps at the executor bound even with a higher configured ceiling", () => {
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({
        phaseTaskCount: 30,
        phaseFileCount: 40,
        tier: "heavy",
        maxParallelRuns: 8,
        configuredCeilingMs: 3_600_000,
      }),
    )
    expect(result.effectiveMs).toBe(MAX_TIMEOUT_MS)
    expect(result.provenance.ceiling_ms).toBe(MAX_TIMEOUT_MS)
  })

  it("raises a configured ceiling below the base to the base (floor precedence)", () => {
    // PTB-FR-001: when the configured ceiling is below the resolved
    // base window the base floor takes precedence — the effective
    // ceiling is the base window.
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({ phaseTaskCount: 6, configuredCeilingMs: 300_000 }),
    )
    // Derived 1_500_000; the ceiling is raised to the 1_200_000 base.
    expect(result.effectiveMs).toBe(1_200_000)
    expect(result.provenance.ceiling_ms).toBe(1_200_000)
    expect(result.provenance.clamped).toBe(true)
  })

  it("derives a zero tier term for tier-less entries", () => {
    const tierless = derivePerTaskTimeoutMs(perTaskSignals({ tier: null }))
    const lightweight = derivePerTaskTimeoutMs(perTaskSignals({ tier: "lightweight" }))
    expect(tierless.provenance.tier_term_ms).toBe(0)
    expect(tierless.provenance.tier).toBeNull()
    // Tier-less and lightweight both contribute a zero tier term.
    expect(tierless.effectiveMs).toBe(lightweight.effectiveMs)
  })

  it("records every derivation input in the provenance record", () => {
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({
        phaseTaskCount: 11,
        phaseFileCount: 9,
        promptBytes: 179_200,
        tier: "heavy",
        maxParallelRuns: 4,
      }),
    )
    // (1_200_000 + 300_000 + 600_000 + 150_000 + 166_500) x 1.45 =
    // 3_503_925 -> clamped to the 1_800_000 executor bound.
    expect(result.provenance).toEqual({
      effective_ms: MAX_TIMEOUT_MS,
      source: "derived",
      base_ms: 1_200_000,
      phase_task_count: 11,
      phase_task_term_ms: 600_000,
      phase_file_count: 9,
      phase_file_term_ms: 150_000,
      prompt_bytes: 179_200,
      prompt_term_ms: 166_500,
      tier: "heavy",
      tier_term_ms: 300_000,
      contention_multiplier: 1.45,
      max_parallel_runs: 4,
      derived_ms: 3_503_925,
      ceiling_ms: MAX_TIMEOUT_MS,
      clamped: true,
    })
  })

  it("is deterministic: identical inputs produce identical output", () => {
    const input = perTaskSignals({
      phaseTaskCount: 11,
      promptBytes: 200_000,
      tier: "standard",
      maxParallelRuns: 4,
    })
    expect(derivePerTaskTimeoutMs(input)).toEqual(derivePerTaskTimeoutMs(input))
  })

  it("passes an explicit timeout through unchanged with source explicit", () => {
    // PTB-FR-001 scenario "Explicit timeout bypasses derivation":
    // any combination of work signals — the explicit argument wins
    // outright, no derivation term modifies it, and the provenance
    // carries no derivation fields at all.
    const result = derivePerTaskTimeoutMs(
      perTaskSignals({
        phaseTaskCount: 11,
        phaseFileCount: 20,
        tier: "heavy",
        maxParallelRuns: 4,
        explicitTimeoutMs: 900_000,
      }),
    )
    expect(result.effectiveMs).toBe(900_000)
    expect(result.provenance).toEqual({ effective_ms: 900_000, source: "explicit" })
  })
})

describe("executor bound drift pin (RD-FR-028)", () => {
  it("keeps MAX_TIMEOUT_MS aligned with the executor TimeoutSchema bound", () => {
    // RD-FR-028: the derivation ceiling re-declares the executor's
    // 1_800_000 ms TimeoutSchema bound as a zod-free constant — the two
    // are coupled by comment only at authoring time. This test-only
    // safeParse pin (run-timeout.ts stays zod-free; the schema is
    // imported here) fails the suite the moment either side drifts: a
    // schema bound lowered below MAX_TIMEOUT_MS would reject the ceiling
    // the derivation already emits, and one raised above it would let the
    // executor accept windows the derivation can never produce.
    expect(TimeoutSchema.safeParse(MAX_TIMEOUT_MS).success).toBe(true)
    expect(TimeoutSchema.safeParse(MAX_TIMEOUT_MS + 1).success).toBe(false)
  })
})

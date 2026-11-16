import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"

import { describe, expect, it } from "vitest"

import {
  isDeclaredUnleashCheckpointStep,
  UNLEASH_DECLARED_STEP_ALLOWLIST,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Scope (task 4.2 synthetic-compaction-continue-prompt-rec, SYNTH-003)
// ---------------------------------------------------------------------------

// Two writer surfaces record checkpoint steps under command uf.unleash: the
// command markdown prescribes step:/checkpoint_step: literals for the
// pipeline's user-facing steps, and the uf-workflow plugin's
// recordAutoCheckpoint machinery records its own literal steps plus the
// delegation-supplied step. Task 4.1 gates the checkpoint tool on the
// declared step set — UNLEASH_DECLARED_STEP_ALLOWLIST plus the
// implement-phase-N family — rejecting any other name with undeclared_step.
// This test holds both writer surfaces in lockstep with that single
// definition: every step name either surface records under uf.unleash must
// be declared, or the test fails naming the file and the step. The allowlist
// and the declaration predicate are imported from the plugin, never
// re-hardcoded here, so the plugin's own definition is what the surfaces
// are measured against.
//
// The check is one-directional by design: declared names that neither
// enumerated surface records (clarify and retrospective arrive through
// other writer paths) do not fail this test — only an undeclared recorded
// step does. The reply-comments step belongs to uf.address-feedback, so
// that command's recordAutoCheckpoint site contributes no assertions.

// ---------------------------------------------------------------------------
// Surfaces under test
// ---------------------------------------------------------------------------

/** Display names for failure messages. */
const COMMAND_FILE = ".opencode/commands/uf.unleash.md"
const PLUGIN_FILE = ".opencode/plugins/uf-workflow/index.ts"

/**
 * Real command markdown and plugin source — both writer surfaces, resolved
 * from this test file's own URL so the suite never depends on the working
 * directory vitest was invoked from.
 */
const COMMAND_FILE_PATH = fileURLToPath(new URL("../commands/uf.unleash.md", import.meta.url))
const PLUGIN_SOURCE_PATH = fileURLToPath(new URL("../plugins/uf-workflow/index.ts", import.meta.url))

const commandMarkdown = readFileSync(COMMAND_FILE_PATH, "utf8")
const pluginSource = readFileSync(PLUGIN_SOURCE_PATH, "utf8")

// ---------------------------------------------------------------------------
// Command-surface extraction
// ---------------------------------------------------------------------------

/**
 * Step literals the command markdown prescribes: `step:` and
 * `checkpoint_step:` arguments carrying a double-quoted value. The leading
 * [^\w] guard keeps longer identifiers ending in `_step` from matching, and
 * requiring a quoted value keeps unquoted prose out. Blockquotes that
 * restate the prescribed arguments (the documented checkpoint contracts)
 * are prescriptions too, so they capture deliberately.
 */
const COMMAND_STEP_PATTERN = /(?:^|[^\w])(?:checkpoint_)?step:\s*"([^"]+)"/gm

/**
 * The implement-phase family prefix. The command prescribes the family as a
 * placeholder (implement-phase-[N]), never a concrete literal, so a capture
 * starting with this prefix maps to the pattern-matched family instead of a
 * literal allowlist name.
 */
const IMPLEMENT_PHASE_FAMILY_PREFIX = "implement-phase-"

/** Every step-literal capture the command surface prescribes, in document order. */
function captureCommandStepLiterals(markdown: string): string[] {
  return [...markdown.matchAll(COMMAND_STEP_PATTERN)].map((match) => match[1])
}

/**
 * Split command-surface captures into literal step names and
 * implement-phase family placeholders.
 */
function partitionCommandSteps(captures: readonly string[]): {
  literals: string[]
  familyPlaceholders: string[]
} {
  const literals: string[] = []
  const familyPlaceholders: string[] = []
  for (const capture of captures) {
    if (capture.startsWith(IMPLEMENT_PHASE_FAMILY_PREFIX)) {
      familyPlaceholders.push(capture)
    } else {
      literals.push(capture)
    }
  }
  return { literals, familyPlaceholders }
}

// ---------------------------------------------------------------------------
// Plugin-surface extraction
// ---------------------------------------------------------------------------

/**
 * One recordAutoCheckpoint call site: the command literal it records under,
 * plus either its literal step name or the variable expression supplying
 * the step at runtime — exactly one of the two.
 */
interface AutoCheckpointCallSite {
  command: string
  stepLiteral: string | undefined
  stepVariable: string | undefined
}

/**
 * recordAutoCheckpoint call sites take the shape
 * (cwd, key, <command literal>, <step>, ...). The (?<!function\s) guard
 * skips the function declaration; [^"]*? stops at the first double quote —
 * the command literal, because the leading cwd/key arguments never carry
 * string literals; the alternation then captures the step as either its
 * quoted literal or its variable reference. The grammar is deliberately
 * strict: call-shape drift fails the extraction sanity assertions below
 * instead of passing vacuously.
 */
const AUTO_CHECKPOINT_CALL_PATTERN =
  /(?<!function\s)recordAutoCheckpoint\([^"]*?"(?<command>uf\.[a-z-]+)",\s*(?:"(?<stepLiteral>[^"]*)"|(?<stepVariable>[A-Za-z_$][\w$.]*))/g

/** Every recordAutoCheckpoint call site in the plugin source, in order. */
function extractAutoCheckpointCallSites(source: string): AutoCheckpointCallSite[] {
  return [...source.matchAll(AUTO_CHECKPOINT_CALL_PATTERN)].map((match) => ({
    command: match.groups?.command ?? "",
    stepLiteral: match.groups?.stepLiteral,
    stepVariable: match.groups?.stepVariable,
  }))
}

// ---------------------------------------------------------------------------
// Enumerated surfaces
// ---------------------------------------------------------------------------

const commandStepPartition = partitionCommandSteps(captureCommandStepLiterals(commandMarkdown))

/** Literal step names the command prescribes, deduplicated, in document order. */
const commandLiteralSteps = [...new Set(commandStepPartition.literals)]

/** implement-phase family placeholders the command prescribes (e.g. implement-phase-[N]). */
const commandFamilyPlaceholders = commandStepPartition.familyPlaceholders

/**
 * recordAutoCheckpoint sites recording under command uf.unleash. The
 * uf.address-feedback site (its reply-comments path) is out of scope by
 * task design and contributes no assertions.
 */
const unleashCallSites = extractAutoCheckpointCallSites(pluginSource).filter(
  (site) => site.command === "uf.unleash",
)

/** Literal machinery step names the plugin records under uf.unleash, deduplicated. */
const pluginLiteralSteps = [
  ...new Set(
    unleashCallSites
      .map((site) => site.stepLiteral)
      .filter((step): step is string => step !== undefined),
  ),
]

/**
 * uf.unleash sites whose step arrives as a runtime variable rather than a
 * literal — the delegation site whose input.checkpoint_step values the
 * command surface enumerates.
 */
const pluginVariableStepSites = unleashCallSites.filter((site) => site.stepVariable !== undefined)

// ---------------------------------------------------------------------------
// The lockstep assertion
// ---------------------------------------------------------------------------

/**
 * Assert one enumerated step name is declared for uf.unleash checkpoints: a
 * UNLEASH_DECLARED_STEP_ALLOWLIST member or an implement-phase-N family
 * member. IMPLEMENT_PHASE_STEP_PATTERN is private to the plugin, but
 * isDeclaredUnleashCheckpointStep is the exported production disjunction
 * over exactly those two (task 4.1), so this imports the predicate and the
 * allowlist rather than re-hardcoding either — the lockstep runs against
 * the plugin's single definition. The message names the offending file and
 * step so a failure points at the drifting surface.
 */
function expectDeclaredStep(file: string, step: string): void {
  const declaredSteps = UNLEASH_DECLARED_STEP_ALLOWLIST.join(", ")
  expect(
    isDeclaredUnleashCheckpointStep(step),
    `${file} records step "${step}" under command uf.unleash, but it is absent from UNLEASH_DECLARED_STEP_ALLOWLIST [${declaredSteps}] and does not match the implement-phase-N family`,
  ).toBe(true)
}

/**
 * Extraction sanity pin for the plugin surface: the exact literal machinery
 * steps the plugin records under uf.unleash. This pins the extraction, not
 * a second copy of the allowlist — each extracted name still runs through
 * expectDeclaredStep below.
 */
const EXPECTED_PLUGIN_LITERAL_STEPS: readonly string[] = [
  "branch-safety-gate",
  "demo",
  "resumability-detection",
  "tier-routing",
]

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("uf.unleash checkpoint step lockstep (SYNTH-003 task 4.2)", () => {
  describe(`command surface (${COMMAND_FILE})`, () => {
    it("extracts at least 8 distinct prescribed steps including the review, gate, and budget steps", () => {
      // Extraction sanity: a regex drift that captures nothing fails here
      // instead of passing vacuously.
      expect(commandLiteralSteps.length).toBeGreaterThanOrEqual(8)
      for (const required of ["code-review", "quick-fix-budget", "quick-pre-flight", "spec-review", "tier-routing"]) {
        expect(
          commandLiteralSteps,
          `${COMMAND_FILE} no longer prescribes the "${required}" checkpoint step`,
        ).toContain(required)
      }
    })

    it("prescribes the implement-phase family as a placeholder and the family stays declared", () => {
      expect(
        commandFamilyPlaceholders.length,
        `${COMMAND_FILE} no longer prescribes an implement-phase-* family placeholder`,
      ).toBeGreaterThan(0)
      // The placeholder (implement-phase-[N]) denotes the pattern-matched
      // family, not a literal allowlist name: the phase machinery records
      // implement-phase-<N> for every positive N. The family pattern is
      // private to the plugin, so probe the exported declaration predicate
      // with a concrete member to prove the family accepts real instances.
      expect(
        isDeclaredUnleashCheckpointStep("implement-phase-1"),
        `${PLUGIN_FILE} no longer declares the implement-phase-N family that ${COMMAND_FILE} prescribes`,
      ).toBe(true)
    })

    it("declares every step name the command prescribes", () => {
      for (const step of commandLiteralSteps) {
        expectDeclaredStep(COMMAND_FILE, step)
      }
    })
  })

  describe(`plugin surface (${PLUGIN_FILE})`, () => {
    it("extracts exactly the four literal machinery steps recorded under uf.unleash", () => {
      // Extraction sanity: the exact set of literal machinery steps, so a
      // call-shape or regex drift fails here instead of passing vacuously.
      expect([...pluginLiteralSteps].sort()).toEqual(EXPECTED_PLUGIN_LITERAL_STEPS)
    })

    it("recognizes the single delegation site whose step arrives as a variable", () => {
      // The delegation site passes input.checkpoint_step (from
      // uf_delegate_command callers) rather than a literal, so its values —
      // plan, tasks, implement — are covered by the command-surface
      // enumeration above. Pinning the count keeps a removed, added, or
      // reshaped variable-step site from silently dropping out of coverage.
      expect(
        pluginVariableStepSites,
        `${PLUGIN_FILE} records ${pluginVariableStepSites.length} uf.unleash recordAutoCheckpoint sites with a variable step; expected exactly the one delegation site`,
      ).toHaveLength(1)
    })

    it("declares every literal machinery step the plugin records under uf.unleash", () => {
      for (const step of pluginLiteralSteps) {
        expectDeclaredStep(PLUGIN_FILE, step)
      }
    })
  })
})

/**
 * Minimal fail-closed YAML-subset parser for the tracked model-matrix
 * config file (OpenSpec change model-matrix-delegation, task 4.1).
 *
 * The plugin suite's dependency footprint is pinned to zod + vitest —
 * no YAML library — and Bun.YAML, the parser behind the production
 * `createBunYamlParser` seam, is unavailable under the Node-side
 * vitest run (the `withBun` shim is JSON-backed). The tracked-file
 * conformance test (model-matrix-tracked.test.ts) must nonetheless
 * parse the REAL `.uf/model-matrix.yaml` text, so this helper
 * implements exactly the subset that file uses and throws on any
 * other syntax:
 *
 * - full-line comments and blank lines (a comments-only document
 *   parses to null, matching Bun.YAML and the MM-FR-004 no-op
 *   contract);
 * - `key: scalar` leaves (bare tokens and integers);
 * - `key:` block mappings nested by consistent indentation;
 * - the `{}` empty flow mapping.
 *
 * A tracked-file edit that introduces syntax outside the subset fails
 * the conformance test loudly instead of misparsing, and every parsed
 * result is validated downstream by both the closed Zod surface
 * (`loadModelMatrix`) and `schemas/model-matrix/v1.schema.json` — two
 * independent checks on the parser's correctness for this file.
 */

/** One significant (non-comment, non-blank) document line. */
interface SignificantLine {
  readonly indent: number
  readonly key: string
  /** The raw scalar text, or null when the line opens a child block. */
  readonly value: string | null
}

/** Keys and bare-token scalars the subset admits (no spaces, quotes, or colons). */
const TOKEN_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._-]*$/

/**
 * Parses the tracked-config YAML subset.
 * @param text YAML text restricted to the documented subset.
 * @returns The parsed document (nested plain mappings), or null for an
 *   empty or comments-only document.
 * @throws When any line, key, scalar, or indentation step falls outside
 *   the subset — the parser never guesses at syntax it does not cover.
 */
export function parseYamlSubset(text: string): unknown {
  const lines: SignificantLine[] = []
  for (const raw of text.split(/\r?\n/)) {
    if (raw.includes("\t")) {
      throw new Error("yaml subset: tab indentation is not supported")
    }
    const trimmed = raw.trim()
    if (trimmed === "" || trimmed.startsWith("#")) {
      continue
    }
    const indent = raw.length - raw.trimStart().length
    const separator = trimmed.indexOf(":")
    if (separator <= 0) {
      throw new Error(`yaml subset: unsupported line syntax: ${trimmed}`)
    }
    const key = trimmed.slice(0, separator)
    if (!TOKEN_PATTERN.test(key)) {
      throw new Error(`yaml subset: unsupported key syntax: ${key}`)
    }
    const value = trimmed.slice(separator + 1).trim()
    lines.push({ indent, key, value: value === "" ? null : value })
  }
  if (lines.length === 0) {
    return null
  }
  let cursor = 0
  const parseBlock = (indent: number): Record<string, unknown> => {
    const mapping: Record<string, unknown> = {}
    while (cursor < lines.length) {
      const line = lines[cursor]
      if (line.indent < indent) {
        return mapping
      }
      if (line.indent > indent) {
        throw new Error(
          `yaml subset: inconsistent indentation at "${line.key}" (indent ${line.indent} inside indent ${indent})`,
        )
      }
      if (Object.hasOwn(mapping, line.key)) {
        throw new Error(`yaml subset: duplicate key "${line.key}"`)
      }
      cursor += 1
      if (line.value === null) {
        if (cursor >= lines.length || lines[cursor].indent <= indent) {
          throw new Error(`yaml subset: key "${line.key}" opens a block with no nested entries`)
        }
        mapping[line.key] = parseBlock(lines[cursor].indent)
        continue
      }
      mapping[line.key] = parseScalar(line.key, line.value)
    }
    return mapping
  }
  return parseBlock(0)
}

/**
 * Parses one scalar leaf value in the subset: `{}`, an integer, or a
 * bare token. Anything else — quoted strings, flow collections, values
 * with internal spaces or trailing comments — throws.
 * @param key The leaf's key, for diagnostics.
 * @param value The raw scalar text.
 * @returns The parsed scalar value.
 * @throws When the value falls outside the subset.
 */
function parseScalar(key: string, value: string): unknown {
  if (value === "{}") {
    return {}
  }
  if (/^-?\d+$/.test(value)) {
    return Number(value)
  }
  if (TOKEN_PATTERN.test(value)) {
    return value
  }
  throw new Error(`yaml subset: unsupported value syntax for "${key}": ${value}`)
}

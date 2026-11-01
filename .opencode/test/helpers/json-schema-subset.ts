/**
 * Compact JSON-Schema validator for the exact keyword subset
 * `schemas/model-matrix/v1.schema.json` uses (draft 2020-12: $ref,
 * type, const, enum, pattern, minLength, maxLength, minimum, maximum,
 * exclusiveMinimum, properties, patternProperties, additionalProperties,
 * propertyNames, required, minProperties, maxProperties).
 *
 * Extracted from model-matrix-schema.test.ts per default.md CS-004
 * [MUST] when the tracked-matrix conformance test
 * (model-matrix-tracked.test.ts, task 4.1) needed to validate the real
 * tracked `.uf/model-matrix.yaml` and its drifted variants against the
 * same schema: identical logic in two locations lives in one shared
 * module. The suite's dependency footprint stays pinned (zod + vitest
 * only — no validator library), and the schema test's
 * self-verification probes prove the validator rejects what the Zod
 * schema rejects, so identical-verdict and schema-validity assertions
 * keep their teeth.
 */

/**
 * The JSON-Schema keyword subset v1.schema.json uses (draft 2020-12).
 * Unknown keywords are ignored, matching lenient validator behavior for
 * annotation keywords; every keyword the schema relies on is implemented.
 */
export interface SchemaSubset {
  readonly $ref?: string
  readonly type?: "object" | "string" | "number" | "integer" | "boolean"
  readonly const?: unknown
  readonly enum?: readonly unknown[]
  readonly pattern?: string
  readonly minLength?: number
  readonly maxLength?: number
  readonly minimum?: number
  readonly maximum?: number
  readonly exclusiveMinimum?: number
  readonly properties?: Readonly<Record<string, SchemaSubset>>
  readonly patternProperties?: Readonly<Record<string, SchemaSubset>>
  readonly additionalProperties?: boolean | SchemaSubset
  readonly propertyNames?: SchemaSubset
  readonly required?: readonly string[]
  readonly minProperties?: number
  readonly maxProperties?: number
}

/** A schema document with its (optional) local definition table. */
export interface SchemaRoot extends SchemaSubset {
  readonly $defs?: Readonly<Record<string, SchemaSubset>>
}

/** First-failure validation outcome: null error means valid. */
export type SchemaVerdict = { readonly valid: boolean; readonly error: string | null }

/**
 * Resolves a local `#/$defs/<name>` reference against the root document.
 * @param root The full schema document carrying `$defs`.
 * @param ref The reference string used by v1.schema.json.
 * @returns The referenced sub-schema.
 * @throws When the reference is not a supported local $defs form.
 */
function resolveRef(root: SchemaRoot, ref: string): SchemaSubset {
  if (!ref.startsWith("#/$defs/")) {
    throw new Error(`harness does not support $ref form: ${ref}`)
  }
  const target = root.$defs?.[ref.slice("#/$defs/".length)]
  if (target === undefined) {
    throw new Error(`unknown $ref: ${ref}`)
  }
  return target
}

/**
 * Validates one value against one sub-schema, returning the first failure
 * as a path-qualified message or null when valid. Short-circuits on the
 * first failure — the harness asserts verdicts, not error inventories.
 * @param schema The sub-schema to apply.
 * @param root The root document for `$ref` resolution.
 * @param value The value under validation.
 * @param path The JSON-pointer-ish location for diagnostics.
 * @returns The first failure message, or null when the value is valid.
 */
function validateNode(schema: SchemaSubset, root: SchemaRoot, value: unknown, path: string): string | null {
  if (schema.$ref !== undefined) {
    return validateNode(resolveRef(root, schema.$ref), root, value, path)
  }

  if (schema.type !== undefined) {
    const matchesType =
      schema.type === "object"
        ? typeof value === "object" && value !== null && !Array.isArray(value)
        : schema.type === "string"
          ? typeof value === "string"
          : schema.type === "boolean"
            ? typeof value === "boolean"
            : schema.type === "number"
              ? typeof value === "number"
              : schema.type === "integer"
                ? typeof value === "number" && Number.isInteger(value)
                : false
    if (!matchesType) {
      return `${path}: expected type ${schema.type}`
    }
  }

  if ("const" in schema && value !== schema.const) {
    return `${path}: expected const ${JSON.stringify(schema.const)}`
  }
  if (schema.enum !== undefined && !schema.enum.some((option) => value === option)) {
    return `${path}: value not in enum`
  }

  if (typeof value === "string") {
    if (schema.pattern !== undefined && !new RegExp(schema.pattern).test(value)) {
      return `${path}: does not match pattern ${schema.pattern}`
    }
    if (schema.minLength !== undefined && value.length < schema.minLength) {
      return `${path}: shorter than minLength ${schema.minLength}`
    }
    if (schema.maxLength !== undefined && value.length > schema.maxLength) {
      return `${path}: longer than maxLength ${schema.maxLength}`
    }
  }

  if (typeof value === "number") {
    if (schema.minimum !== undefined && value < schema.minimum) {
      return `${path}: less than minimum ${schema.minimum}`
    }
    if (schema.maximum !== undefined && value > schema.maximum) {
      return `${path}: greater than maximum ${schema.maximum}`
    }
    if (schema.exclusiveMinimum !== undefined && value <= schema.exclusiveMinimum) {
      return `${path}: not greater than exclusiveMinimum ${schema.exclusiveMinimum}`
    }
  }

  if (typeof value === "object" && value !== null && !Array.isArray(value)) {
    const record = value as Readonly<Record<string, unknown>>
    const keys = Object.keys(record)
    for (const key of schema.required ?? []) {
      if (!Object.hasOwn(record, key)) {
        return `${path}: missing required property "${key}"`
      }
    }
    if (schema.minProperties !== undefined && keys.length < schema.minProperties) {
      return `${path}: fewer than minProperties ${schema.minProperties}`
    }
    if (schema.maxProperties !== undefined && keys.length > schema.maxProperties) {
      return `${path}: more than maxProperties ${schema.maxProperties}`
    }
    if (schema.propertyNames !== undefined) {
      for (const key of keys) {
        const nameError = validateNode(schema.propertyNames, root, key, `${path} (property name "${key}")`)
        if (nameError !== null) {
          return nameError
        }
      }
    }
    for (const key of keys) {
      let valueSchema: SchemaSubset | undefined = schema.properties?.[key]
      if (valueSchema === undefined && schema.patternProperties !== undefined) {
        for (const [pattern, patternSchema] of Object.entries(schema.patternProperties)) {
          if (new RegExp(pattern).test(key)) {
            valueSchema = patternSchema
            break
          }
        }
      }
      if (valueSchema === undefined) {
        // Draft 2020-12: `additionalProperties` governs keys not covered
        // by `properties`/`patternProperties` — `false` rejects them, a
        // schema validates them, and absence permits them.
        if (schema.additionalProperties === false) {
          return `${path}: additional property "${key}" is not allowed`
        }
        if (schema.additionalProperties !== undefined && schema.additionalProperties !== true) {
          const error = validateNode(schema.additionalProperties, root, record[key], `${path}.${key}`)
          if (error !== null) {
            return error
          }
        }
        continue
      }
      const error = validateNode(valueSchema, root, record[key], `${path}.${key}`)
      if (error !== null) {
        return error
      }
    }
  }

  return null
}

/**
 * Validates a parsed document against the full schema document.
 * @param root The schema document (with `$defs`).
 * @param document The parsed value under validation.
 * @returns The verdict with the first failure message, or valid.
 */
export function validateJsonSchema(root: SchemaRoot, document: unknown): SchemaVerdict {
  const error = validateNode(root, root, document, "$")
  return { valid: error === null, error }
}

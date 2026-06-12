import { evaluateComputed, getAffectedComputedFields } from './expression-eval'
import type { Field } from '@/types/kora'

/**
 * Re-evaluates all computed fields on a document and returns the updated form data.
 * Call this whenever any field value changes.
 */
export function applyComputedFields(
  fields: Field[],
  formData: Record<string, any>,
): Record<string, any> {
  const computedFields = fields.filter((f) => f.computed)
  if (computedFields.length === 0) return formData

  const updated = { ...formData }

  // Build evaluation context.
  const ctx = {
    fields: updated,
    tables: extractChildTables(fields, updated),
  }

  // Evaluate each computed field.
  for (const f of computedFields) {
    const result = evaluateComputed(f.computed!, ctx)
    if (result !== null) {
      updated[f.fieldname] = result
    }
  }

  return updated
}

/**
 * Returns which computed fields should be re-evaluated when fieldname changes.
 */
export function getDependentComputedFields(
  fields: Field[],
  changedFieldname: string,
): { fieldname: string; computed: string }[] {
  return getAffectedComputedFields(fields, changedFieldname)
}

function extractChildTables(
  fields: Field[],
  formData: Record<string, any>,
): Record<string, Record<string, any>[]> {
  const tables: Record<string, Record<string, any>[]> = {}
  for (const f of fields) {
    if (f.fieldtype === 'Table' && formData[f.fieldname]) {
      tables[f.fieldname] = formData[f.fieldname]
    }
  }
  return tables
}

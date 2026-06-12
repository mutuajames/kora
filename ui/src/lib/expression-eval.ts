/**
 * Evaluates a computed expression against document field values.
 * Supports:
 *   - Arithmetic: + - * /
 *   - Field references: fieldname → value
 *   - SUM(table.field) → sum of field across child table rows
 *   - ROUND(expr, N) → round to N decimal places
 *
 * No external dependencies — uses Function constructor for safe evaluation.
 */

export interface EvalContext {
  /** Current field values (scalar fields on this document). */
  fields: Record<string, any>
  /** Child table rows, keyed by table fieldname. */
  tables: Record<string, Record<string, any>[]>
}

/**
 * Finds field names referenced in a computed expression.
 * Returns an array of field names (including child table references like "items").
 */
export function findDependencies(expr: string): string[] {
  if (!expr) return []
  const deps = new Set<string>()
  // Match fieldname patterns: word characters and underscores.
  const fieldPattern = /\b([a-z_][a-z0-9_]*)\b/gi
  let match
  while ((match = fieldPattern.exec(expr)) !== null) {
    const name = match[1].toLowerCase()
    // Skip keywords.
    if (['sum', 'round'].includes(name)) continue
    deps.add(name)
  }
  return Array.from(deps)
}

/**
 * Evaluates a computed expression string against the given context.
 */
export function evaluateComputed(expr: string, ctx: EvalContext): number | null {
  if (!expr || !expr.trim()) return null

  try {
    let processed = expr.trim()

    // Expand SUM(table.field) → numeric sum.
    processed = processed.replace(/SUM\s*\(\s*(\w+)\s*\.\s*(\w+)\s*\)/gi, (_, table, field) => {
      const rows = ctx.tables[table] || ctx.tables[table.toLowerCase()] || []
      const sum = rows.reduce((acc: number, row: any) => {
        return acc + (parseFloat(row[field]) || 0)
      }, 0)
      return String(sum)
    })

    // Expand ROUND(expr, N)
    processed = processed.replace(/ROUND\s*\(\s*([^,]+)\s*,\s*(\d+)\s*\)/gi, (_, inner, places) => {
      const val = evalExpr(inner, ctx)
      const factor = Math.pow(10, parseInt(places))
      return String(Math.round(val * factor) / factor)
    })

    return evalExpr(processed, ctx)
  } catch {
    return null
  }
}

function evalExpr(expr: string, ctx: EvalContext): number {
  // Replace field references with their numeric values.
  const safe = expr.replace(/\b([a-z_][a-z0-9_]*)\b/gi, (match: string) => {
    const key = match.toLowerCase()
    if (key in ctx.fields) {
      const val = parseFloat(ctx.fields[key])
      return isNaN(val) ? '0' : String(val)
    }
    return '0'
  })

  // Use Function constructor — safe because we control the input (config values).
  // eslint-disable-next-line no-new-func
  const fn = new Function(`return (${safe})`)
  const result = fn()
  return isNaN(result) || !isFinite(result) ? 0 : result
}

/**
 * Returns all computed fields on a doctype that should be recalculated
 * when a given field changes.
 */
export function getAffectedComputedFields(
  fields: { fieldname: string; computed?: string }[],
  changedFieldname: string,
): { fieldname: string; computed: string }[] {
  const result: { fieldname: string; computed: string }[] = []
  for (const f of fields) {
    if (!f.computed) continue
    const deps = findDependencies(f.computed)
    if (deps.includes(changedFieldname.toLowerCase())) {
      result.push({ fieldname: f.fieldname, computed: f.computed })
    }
  }
  return result
}

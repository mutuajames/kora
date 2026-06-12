import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { fetchDoctypeSchema } from '@/lib/api/system'
import { FieldRenderer } from './FieldRenderer'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Plus, Trash2, GripVertical } from 'lucide-react'
import { applyComputedFields } from '@/lib/computed-fields'
import { cn } from '@/lib/utils'
import type { Field } from '@/types/kora'

interface ChildTableEditorProps {
  field: Field
  value: Record<string, any>[]
  onChange: (fieldname: string, value: Record<string, any>[]) => void
  onRowsChange?: (rows: Record<string, any>[]) => void
  disabled: boolean
  errors?: Record<number, Record<string, string>>
}

export function ChildTableEditor({ field, value, onChange, onRowsChange, disabled, errors }: ChildTableEditorProps) {
  const childDoctype = field.options // "Order Item"
  const rows: Record<string, any>[] = value ?? []
  const [localRows, setLocalRows] = useState<Record<string, any>[]>(rows)

  // Fetch child doctype schema for its fields.
  const schemaQuery = useQuery({
    queryKey: ['doctype', childDoctype],
    queryFn: () => fetchDoctypeSchema(childDoctype),
    enabled: !!childDoctype,
    staleTime: 5 * 60_000,
  })

  const childFields: Field[] =
    schemaQuery.data?.doctype?.fields?.filter(
      (f: Field) =>
        !['Section Break', 'Column Break', 'Heading', 'Table'].includes(f.fieldtype),
    ) ?? []

  // Sync local rows with parent value.
  const updateRows = (newRows: Record<string, any>[]) => {
    setLocalRows(newRows)
    onChange(field.fieldname, newRows)
    onRowsChange?.(newRows)
  }

  const addRow = () => {
    const newRow: Record<string, any> = {}
    childFields.forEach((f) => {
      if (f.default) newRow[f.fieldname] = f.default
    })
    updateRows([...localRows, newRow])
  }

  const removeRow = (idx: number) => {
    updateRows(localRows.filter((_, i) => i !== idx))
  }

  const updateRow = (idx: number, fieldname: string, val: any) => {
    const updated = localRows.map((row, i) =>
      i === idx ? { ...row, [fieldname]: val } : row,
    )

    // Generic linked_field auto-population: when a Link field changes,
    // fetch the linked document and populate fields that reference it.
    const changedField = childFields.find((f) => f.fieldname === fieldname)
    if (changedField && (changedField.fieldtype === 'Link' || changedField.fieldtype === 'Dynamic Link') && val) {
      const targetDoctype = changedField.options
      const linkedFields = childFields.filter(
        (f) => f.linked_field?.startsWith(fieldname + '.'),
      )
      if (linkedFields.length > 0 && targetDoctype) {
        import('@/lib/api/resources').then(({ fetchDocument }) => {
          fetchDocument(targetDoctype, val).then((doc) => {
            setLocalRows((prev) => {
              const newRows = prev.map((row, i) => {
                if (i !== idx) return row
                const newRow = { ...row }
                for (const lf of linkedFields) {
                  const sourceField = lf.linked_field!.split('.')[1]
                  if (doc[sourceField] !== undefined) {
                    newRow[lf.fieldname] = doc[sourceField]
                  }
                }
                // Apply computed fields after linked fields are populated.
                return applyComputedFields(childFields as Field[], newRow)
              })
              // Notify parent of changes (triggers computed fields on parent).
              onChange(field.fieldname, newRows)
              onRowsChange?.(newRows)
              return newRows
            })
          }).catch(() => {})
        })
      }
    }

    // Apply computed field expressions from config.
    const row = applyComputedFields(childFields as Field[], updated[idx])
    updated[idx] = row

    updateRows(updated)
  }

  if (schemaQuery.isLoading) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-24 w-full" />
      </div>
    )
  }

  const rowErrors = (idx: number) => errors?.[idx] ?? {}

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold">{field.label}</h4>
        <Button
          variant="outline"
          size="sm"
          onClick={addRow}
          disabled={disabled}
          type="button"
        >
          <Plus className="mr-1 h-3.5 w-3.5" />
          Add Row
        </Button>
      </div>

      {localRows.length === 0 ? (
        <div className="rounded-lg border border-dashed p-8 text-center">
          <p className="text-sm text-muted-foreground">
            No {childDoctype || 'items'} added yet. Click "Add Row" to add one.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {localRows.map((row, idx) => (
            <div
              key={idx}
              className={cn(
                'group relative rounded-lg border bg-muted/30 px-3 py-2',
                Object.keys(rowErrors(idx)).length > 0 && 'border-destructive',
              )}
            >
              <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4 items-start">
                {childFields.map((cf) => (
                  <FieldRenderer
                    key={cf.fieldname}
                    field={cf}
                    value={row[cf.fieldname] ?? null}
                    onChange={(fn, v) => updateRow(idx, fn, v)}
                    disabled={disabled}
                    error={rowErrors(idx)[cf.fieldname]}
                    compact
                  />
                ))}
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="absolute -right-1 -top-1 h-6 w-6 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={() => removeRow(idx)}
                disabled={disabled}
                type="button"
              >
                <Trash2 className="h-3 w-3 text-muted-foreground hover:text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

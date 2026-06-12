import { useQuery } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { fetchList } from '@/lib/api/resources'
import { fetchDoctypeSchema } from '@/lib/api/system'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronRight, LinkIcon } from 'lucide-react'
import type { ReferenceInfo } from '@/types/kora'

interface RelatedDocsProps {
  doctype: string
  name: string
}

function RelatedList({ ref, docName }: { ref: ReferenceInfo; docName: string }) {
  const navigate = useNavigate()

  const schemaQuery = useQuery({
    queryKey: ['doctype', ref.doctype],
    queryFn: () => fetchDoctypeSchema(ref.doctype),
    staleTime: 5 * 60_000,
  })

  const titleField = schemaQuery.data?.doctype?.title_field || 'name'

  const listQuery = useQuery({
    queryKey: ['resource', ref.doctype, 'related', docName],
    queryFn: () =>
      fetchList(ref.doctype, {
        limit: 5,
        fields: ['name', titleField],
        filters: JSON.stringify([[ref.fieldname, '=', docName]]),
      }),
    staleTime: 15_000,
  })

  const docs = listQuery.data?.data ?? []
  const total = listQuery.data?.meta?.total ?? 0

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 text-sm font-medium">
          <LinkIcon className="h-4 w-4 text-muted-foreground" />
          {ref.doctype}s
          {listQuery.isLoading ? (
            <Skeleton className="h-4 w-8" />
          ) : (
            <span className="text-muted-foreground">({total})</span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-1 pt-0">
        {listQuery.isLoading ? (
          <div className="space-y-2">
            {[1, 2].map((i) => (
              <Skeleton key={i} className="h-6 w-full" />
            ))}
          </div>
        ) : docs.length === 0 ? (
          <p className="text-xs text-muted-foreground">No {ref.doctype.toLowerCase()}s yet.</p>
        ) : (
          docs.map((doc) => (
            <button
              key={doc.name}
              type="button"
              className="flex w-full items-center justify-between rounded-md px-2 py-1 text-left text-sm transition-colors hover:bg-muted"
              onClick={() =>
                navigate({
                  to: '/workspace/$doctype/$name',
                  params: { doctype: ref.doctype, name: doc.name },
                })
              }
            >
              <span>
                <span className="font-mono text-xs">{doc.name}</span>
                <span className="ml-2 text-muted-foreground">
                  {String(doc[titleField] ?? '')}
                </span>
              </span>
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            </button>
          ))
        )}
      </CardContent>
    </Card>
  )
}

export function RelatedDocs({ doctype, name }: RelatedDocsProps) {
  const schemaQuery = useQuery({
    queryKey: ['doctype', doctype],
    queryFn: () => fetchDoctypeSchema(doctype),
    staleTime: 5 * 60_000,
  })

  const refs: ReferenceInfo[] = schemaQuery.data?.referenced_by ?? []

  if (refs.length === 0) return null

  return (
    <div className="mt-8 space-y-4">
      <h3 className="text-lg font-semibold">Related Documents</h3>
      <div className="grid gap-4 sm:grid-cols-2">
        {refs.map((ref) => (
          <RelatedList key={ref.doctype + ref.fieldname} ref={ref} docName={name} />
        ))}
      </div>
    </div>
  )
}

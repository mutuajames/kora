import { api } from './client'
import type { NavigationConfig } from '@/types/api'
import type { DocTypeSchema } from '@/types/kora'

export async function fetchNavigation(): Promise<NavigationConfig> {
  return api.get<NavigationConfig>('/api/system/navigation')
}

export async function fetchDoctypeSchema(
  doctype: string,
  state?: string,
): Promise<DocTypeSchema> {
  const params: Record<string, string | undefined> = {}
  if (state) params.state = state
  return api.get<DocTypeSchema>(`/api/system/doctype/${encodeURIComponent(doctype)}`, params)
}

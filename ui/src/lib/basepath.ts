// Returns the site prefix for path-based access (/s/:site), or empty string for Host-based.
export function getSitePrefix(): string {
  const m = window.location.pathname.match(/^(\/s\/[^/]+)/)
  return m ? m[1] : ''
}

// Prepends the site prefix to a path.
// sitePath('/workspace') → '/workspace' or '/s/fieldwork/workspace'
export function sitePath(path: string): string {
  return getSitePrefix() + path
}

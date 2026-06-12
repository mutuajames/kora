import { create } from 'zustand'

type Theme = 'light' | 'dark'

interface UIState {
  theme: Theme
  sidebarOpen: boolean
  sidebarCollapsed: boolean
  configVersion: number

  toggleTheme: () => void
  setTheme: (theme: Theme) => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setConfigVersion: (version: number) => void
}

function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'light'
  const stored = localStorage.getItem('kora-theme')
  if (stored === 'dark' || stored === 'light') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyTheme(theme: Theme) {
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', theme === 'dark')
  localStorage.setItem('kora-theme', theme)
}

export const useUIStore = create<UIState>((set, get) => ({
  theme: getStoredTheme(),
  sidebarOpen: false,
  sidebarCollapsed: false,
  configVersion: 0,

  toggleTheme: () => {
    const next = get().theme === 'light' ? 'dark' : 'light'
    applyTheme(next)
    set({ theme: next })
  },

  setTheme: (theme) => {
    applyTheme(theme)
    set({ theme })
  },

  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
  setConfigVersion: (version) => set({ configVersion: version }),
}))

// Apply initial theme.
if (typeof document !== 'undefined') {
  applyTheme(getStoredTheme())
}

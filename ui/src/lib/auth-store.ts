import { create } from 'zustand'
import type { CurrentUser } from '@/types/api'
import { sitePath } from './basepath'
import * as authApi from './api/auth'

interface AuthState {
  user: CurrentUser | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  checkAuth: () => Promise<boolean>
  clearError: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,

  login: async (email, password) => {
    set({ isLoading: true, error: null })
    try {
      const user = await authApi.login({ email, password })
      set({ user, isAuthenticated: true, isLoading: false })
    } catch (err: any) {
      set({ isLoading: false, error: err.message || 'Login failed' })
      throw err
    }
  },

  logout: async () => {
    try {
      await authApi.logout()
    } catch {
      // Ignore logout errors.
    }
    set({ user: null, isAuthenticated: false, isLoading: false })
    // Redirect to login, preserving site prefix.
    window.location.href = sitePath('/workspace/auth/login')
  },

  checkAuth: async () => {
    if (get().isAuthenticated) return true
    set({ isLoading: true })
    try {
      const user = await authApi.fetchMe()
      set({ user, isAuthenticated: true, isLoading: false })
      return true
    } catch {
      set({ user: null, isAuthenticated: false, isLoading: false })
      return false
    }
  },

  clearError: () => set({ error: null }),
}))

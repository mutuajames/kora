import { createRootRoute, Outlet, useNavigate, useRouter } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useAuthStore } from '@/lib/auth-store'
import { RootLayout } from '@/components/layout/RootLayout'

function AuthGuard() {
  const { isAuthenticated, isLoading, checkAuth } = useAuthStore()
  const navigate = useNavigate()
  const router = useRouter()

  useEffect(() => {
    checkAuth()
  }, [])

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      const currentPath = router.state.location.pathname
      if (!currentPath.includes('/auth/login')) {
        navigate({ to: '/workspace/auth/login', search: { redirect: currentPath } })
      }
    }
  }, [isAuthenticated, isLoading])

  if (isLoading) return null
  if (!isAuthenticated) return null

  return <RootLayout />
}

export const Route = createRootRoute({
  component: AuthGuard,
})

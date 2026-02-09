import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from './use-auth'

type GuardProps = {
  children: ReactNode
}

export function ProtectedRoute({ children }: GuardProps) {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) {
    return <Navigate to="/auth" replace />
  }
  return <>{children}</>
}

export function AuthOnlyRoute({ children }: GuardProps) {
  const { isAuthenticated } = useAuth()
  if (isAuthenticated) {
    return <Navigate to="/boards" replace />
  }
  return <>{children}</>
}

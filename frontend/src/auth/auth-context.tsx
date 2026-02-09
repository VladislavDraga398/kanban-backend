import { useMemo, useState, type ReactNode } from 'react'
import { clearAuthToken, getAuthToken, setAuthToken } from '../shared/auth/token'
import { AuthContext, type AuthContextValue } from './auth-context-store'

type AuthProviderProps = {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [token, setToken] = useState<string | null>(() => getAuthToken())

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      isAuthenticated: Boolean(token),
      login: (nextToken: string) => {
        setAuthToken(nextToken)
        setToken(nextToken)
      },
      logout: () => {
        clearAuthToken()
        setToken(null)
      },
    }),
    [token],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

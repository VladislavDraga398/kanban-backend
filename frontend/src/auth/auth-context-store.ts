import { createContext } from 'react'

export type AuthContextValue = {
  token: string | null
  isAuthenticated: boolean
  login: (token: string) => void
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined)

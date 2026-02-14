import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthContext, type AuthContextValue } from './auth-context-store'
import { AuthOnlyRoute, ProtectedRoute } from './route-guards'

function renderGuards(initialPath: string, isAuthenticated: boolean) {
  const value: AuthContextValue = {
    token: isAuthenticated ? 'token' : null,
    isAuthenticated,
    login: () => undefined,
    logout: () => undefined,
  }

  return render(
    <AuthContext.Provider value={value}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/auth" element={<div>Auth Screen</div>} />
          <Route path="/boards" element={<div>Boards Screen</div>} />
          <Route
            path="/private"
            element={
              <ProtectedRoute>
                <div>Private Screen</div>
              </ProtectedRoute>
            }
          />
          <Route
            path="/guest"
            element={
              <AuthOnlyRoute>
                <div>Guest Screen</div>
              </AuthOnlyRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    </AuthContext.Provider>,
  )
}

describe('route guards', () => {
  it('redirects unauthenticated users from protected routes', () => {
    renderGuards('/private', false)
    expect(screen.getByText('Auth Screen')).toBeInTheDocument()
  })

  it('allows authenticated users to protected routes', () => {
    renderGuards('/private', true)
    expect(screen.getByText('Private Screen')).toBeInTheDocument()
  })

  it('redirects authenticated users from auth-only routes', () => {
    renderGuards('/guest', true)
    expect(screen.getByText('Boards Screen')).toBeInTheDocument()
  })

  it('allows guests to auth-only routes', () => {
    renderGuards('/guest', false)
    expect(screen.getByText('Guest Screen')).toBeInTheDocument()
  })
})

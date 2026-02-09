import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import { getAuthToken, setAuthToken } from '../shared/auth/token'
import { AuthProvider } from './auth-context'
import { useAuth } from './use-auth'

function AuthHarness() {
  const { token, isAuthenticated, login, logout } = useAuth()

  return (
    <div>
      <p data-testid="token">{token ?? 'none'}</p>
      <p data-testid="authenticated">{String(isAuthenticated)}</p>
      <button type="button" onClick={() => login('token-123')}>
        login
      </button>
      <button type="button" onClick={logout}>
        logout
      </button>
    </div>
  )
}

function InvalidHookUsage() {
  useAuth()
  return null
}

describe('AuthProvider', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('loads token from localStorage on startup', () => {
    setAuthToken('stored-token')

    render(
      <AuthProvider>
        <AuthHarness />
      </AuthProvider>,
    )

    expect(screen.getByTestId('token')).toHaveTextContent('stored-token')
    expect(screen.getByTestId('authenticated')).toHaveTextContent('true')
  })

  it('persists token on login', async () => {
    const user = userEvent.setup()

    render(
      <AuthProvider>
        <AuthHarness />
      </AuthProvider>,
    )

    await user.click(screen.getByRole('button', { name: 'login' }))

    expect(screen.getByTestId('token')).toHaveTextContent('token-123')
    expect(screen.getByTestId('authenticated')).toHaveTextContent('true')
    expect(getAuthToken()).toBe('token-123')
  })

  it('clears token on logout', async () => {
    const user = userEvent.setup()

    render(
      <AuthProvider>
        <AuthHarness />
      </AuthProvider>,
    )

    await user.click(screen.getByRole('button', { name: 'login' }))
    await user.click(screen.getByRole('button', { name: 'logout' }))

    expect(screen.getByTestId('token')).toHaveTextContent('none')
    expect(screen.getByTestId('authenticated')).toHaveTextContent('false')
    expect(getAuthToken()).toBeNull()
  })
})

describe('useAuth', () => {
  it('throws when used outside provider', () => {
    expect(() => render(<InvalidHookUsage />)).toThrowError(
      'useAuth must be used inside AuthProvider',
    )
  })
})

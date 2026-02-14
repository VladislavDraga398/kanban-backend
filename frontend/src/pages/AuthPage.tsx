import { useMutation } from '@tanstack/react-query'
import { useMemo, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/use-auth'
import { login, register } from '../features/auth/api'
import { getErrorMessage } from '../shared/api/errors'

type AuthMode = 'login' | 'register'

export function AuthPage() {
  const [mode, setMode] = useState<AuthMode>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const { login: storeLogin } = useAuth()
  const navigate = useNavigate()

  const authMutation = useMutation({
    mutationFn: async () => {
      const payload = {
        email: email.trim(),
        password: password.trim(),
      }
      return mode === 'login' ? login(payload) : register(payload)
    },
    onSuccess: (data) => {
      storeLogin(data.token)
      navigate('/boards', { replace: true })
    },
  })

  const pageTitle = useMemo(
    () => (mode === 'login' ? 'Войти в Kanban' : 'Создать аккаунт'),
    [mode],
  )

  const pageSubtitle = useMemo(
    () =>
      mode === 'login'
        ? 'Используй существующую учетную запись.'
        : 'Регистрация сразу возвращает JWT токен.',
    [mode],
  )

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!email.trim() || !password.trim()) {
      return
    }
    authMutation.mutate()
  }

  return (
    <main className="auth-page">
      <section className="auth-card">
        <div className="auth-card__heading">
          <p className="badge">KANBAN CONTROL</p>
          <h1>{pageTitle}</h1>
          <p>{pageSubtitle}</p>
        </div>

        <div className="auth-switch">
          <button
            type="button"
            className={mode === 'login' ? 'is-active' : ''}
            onClick={() => setMode('login')}
          >
            Login
          </button>
          <button
            type="button"
            className={mode === 'register' ? 'is-active' : ''}
            onClick={() => setMode('register')}
          >
            Register
          </button>
        </div>

        <form className="auth-form" onSubmit={onSubmit}>
          <label>
            <span>Email</span>
            <input
              type="email"
              value={email}
              autoComplete="email"
              onChange={(event) => setEmail(event.target.value)}
              placeholder="user@example.com"
              required
            />
          </label>

          <label>
            <span>Password</span>
            <input
              type="password"
              value={password}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Минимум 6 символов"
              required
            />
          </label>

          {authMutation.isError && (
            <p className="error-text">{getErrorMessage(authMutation.error)}</p>
          )}

          <button type="submit" className="primary-button" disabled={authMutation.isPending}>
            {authMutation.isPending
              ? 'Подключаем...'
              : mode === 'login'
                ? 'Войти'
                : 'Создать аккаунт'}
          </button>
        </form>
      </section>
    </main>
  )
}

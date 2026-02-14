import { Component, type ErrorInfo, type ReactNode } from 'react'

type AppErrorBoundaryProps = {
  children: ReactNode
}

type AppErrorBoundaryState = {
  hasError: boolean
}

export class AppErrorBoundary extends Component<
  AppErrorBoundaryProps,
  AppErrorBoundaryState
> {
  state: AppErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError(): AppErrorBoundaryState {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Неперехваченная ошибка приложения', error, info)
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="page-shell">
          <section className="panel" role="alert" aria-live="assertive">
            <p className="badge">KANBAN CONTROL</p>
            <h1>Приложение временно недоступно</h1>
            <p>Произошла ошибка на клиенте. Обнови страницу и попробуй снова.</p>
          </section>
        </main>
      )
    }

    return this.props.children
  }
}

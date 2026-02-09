import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from './auth/use-auth'
import { AuthOnlyRoute, ProtectedRoute } from './auth/route-guards'
import { AuthPage } from './pages/AuthPage'
import { BoardPage } from './pages/BoardPage'
import { BoardsPage } from './pages/BoardsPage'

function RootRedirect() {
  const { isAuthenticated } = useAuth()
  return <Navigate to={isAuthenticated ? '/boards' : '/auth'} replace />
}

export default function App() {
  return (
    <Routes>
      <Route
        path="/auth"
        element={
          <AuthOnlyRoute>
            <AuthPage />
          </AuthOnlyRoute>
        }
      />
      <Route
        path="/boards"
        element={
          <ProtectedRoute>
            <BoardsPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/boards/:boardId"
        element={
          <ProtectedRoute>
            <BoardPage />
          </ProtectedRoute>
        }
      />
      <Route path="/" element={<RootRedirect />} />
      <Route path="*" element={<RootRedirect />} />
    </Routes>
  )
}

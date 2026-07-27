import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from './AuthContext'

export function ProtectedRoute() {
  const { token, loading } = useAuth()

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-brand-bg text-sm text-brand-text-muted">
        Carregando…
      </div>
    )
  }

  if (!token) return <Navigate to="/login" replace />

  return <Outlet />
}

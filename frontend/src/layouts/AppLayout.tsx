import { Outlet } from 'react-router-dom'
import { Sidebar } from '../components/layout/Sidebar'
import { useAuth } from '../lib/auth/AuthContext'

export function AppLayout() {
  const { tenant, user, logout } = useAuth()

  return (
    <div className="flex h-screen bg-brand-bg">
      <Sidebar clinicName={tenant?.name ?? '—'} isAdmin={user?.role === 'TENANT_ADMIN'} onLogout={logout} />
      <div className="flex flex-1 flex-col overflow-y-auto">
        <Outlet />
      </div>
    </div>
  )
}

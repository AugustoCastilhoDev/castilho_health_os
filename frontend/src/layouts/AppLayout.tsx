import { Outlet } from 'react-router-dom'
import { Sidebar } from '../components/layout/Sidebar'
import { useAuth } from '../lib/auth/AuthContext'

export function AppLayout() {
  const { tenant, logout } = useAuth()

  return (
    <div className="flex h-screen bg-brand-bg">
      <Sidebar clinicName={tenant?.name ?? '—'} onLogout={logout} />
      <div className="flex flex-1 flex-col overflow-y-auto">
        <Outlet />
      </div>
    </div>
  )
}

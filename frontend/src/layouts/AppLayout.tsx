import { Outlet } from 'react-router-dom'
import { Sidebar } from '../components/layout/Sidebar'
import { mockSession } from '../lib/mockSession'

export function AppLayout() {
  return (
    <div className="flex h-screen bg-brand-bg">
      <Sidebar clinicName={mockSession.clinicName} />
      <div className="flex flex-1 flex-col overflow-y-auto">
        <Outlet />
      </div>
    </div>
  )
}

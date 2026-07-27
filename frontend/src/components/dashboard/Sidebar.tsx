import { Activity, Calendar, Users, Wallet, Package } from 'lucide-react'
import type { ComponentType } from 'react'

interface NavItem {
  label: string
  icon: ComponentType<{ size?: number; className?: string }>
  active?: boolean
}

const navItems: NavItem[] = [
  { label: 'Agenda', icon: Calendar, active: true },
  { label: 'Pacientes', icon: Users },
  { label: 'Financeiro', icon: Wallet },
  { label: 'Estoque', icon: Package },
]

interface SidebarProps {
  clinicName: string
}

export function Sidebar({ clinicName }: SidebarProps) {
  return (
    <aside className="flex h-screen w-64 shrink-0 flex-col bg-brand-sidebar text-slate-100">
      <div className="flex items-center gap-2 px-6 py-6">
        <Activity className="text-brand-action" size={28} />
        <span className="text-lg font-semibold tracking-tight text-white">
          Castilho Health OS
        </span>
      </div>

      <div className="mx-4 mb-6 rounded-lg bg-white/5 px-4 py-3">
        <p className="text-xs uppercase tracking-wide text-slate-400">Clínica</p>
        <div className="mt-1 flex items-center gap-2">
          <span className="h-2 w-2 shrink-0 rounded-full bg-brand-success" />
          <p className="truncate text-sm font-medium text-white">{clinicName}</p>
        </div>
      </div>

      <nav className="flex-1 space-y-1 px-4">
        {navItems.map(({ label, icon: Icon, active }) => (
          <a
            key={label}
            href="#"
            className={`flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors ${
              active
                ? 'bg-brand-action text-white'
                : 'text-slate-300 hover:bg-white/5 hover:text-white'
            }`}
          >
            <Icon size={18} />
            {label}
          </a>
        ))}
      </nav>
    </aside>
  )
}

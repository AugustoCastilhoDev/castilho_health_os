import { Link, NavLink } from 'react-router-dom'
import { Activity, Calendar, LogOut, Users, Wallet, Package } from 'lucide-react'

interface SidebarProps {
  clinicName: string
  onLogout: () => void
}

function navLinkClass(isActive: boolean): string {
  return `flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors ${
    isActive ? 'bg-brand-action text-white' : 'text-slate-300 hover:bg-white/5 hover:text-white'
  }`
}

// Financeiro/Estoque don't have screens yet — rendered inert rather than
// linking somewhere that 404s, so the shell still reads as complete.
const comingSoon = [
  { label: 'Financeiro', icon: Wallet },
  { label: 'Estoque', icon: Package },
]

export function Sidebar({ clinicName, onLogout }: SidebarProps) {
  return (
    <aside className="flex h-screen w-64 shrink-0 flex-col bg-brand-sidebar text-slate-100">
      <Link to="/" className="flex items-center gap-2 px-6 py-6">
        <Activity className="text-brand-action" size={28} />
        <span className="text-lg font-semibold tracking-tight text-white">
          Castilho Health OS
        </span>
      </Link>

      <div className="mx-4 mb-6 rounded-lg bg-white/5 px-4 py-3">
        <p className="text-xs uppercase tracking-wide text-slate-400">Clínica</p>
        <div className="mt-1 flex items-center gap-2">
          <span className="h-2 w-2 shrink-0 rounded-full bg-brand-success" />
          <p className="truncate text-sm font-medium text-white">{clinicName}</p>
        </div>
      </div>

      <nav className="flex-1 space-y-1 px-4">
        <NavLink to="/agenda" className={({ isActive }) => navLinkClass(isActive)}>
          <Calendar size={18} />
          Agenda
        </NavLink>
        <NavLink to="/pacientes" className={({ isActive }) => navLinkClass(isActive)}>
          <Users size={18} />
          Pacientes
        </NavLink>
        {comingSoon.map(({ label, icon: Icon }) => (
          <div
            key={label}
            className="flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-500"
          >
            <Icon size={18} />
            {label}
            <span className="ml-auto rounded-full bg-white/5 px-2 py-0.5 text-[10px] uppercase tracking-wide text-slate-500">
              em breve
            </span>
          </div>
        ))}
      </nav>

      <button
        type="button"
        onClick={onLogout}
        className="mx-4 mb-6 flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-300 transition-colors hover:bg-white/5 hover:text-white"
      >
        <LogOut size={18} />
        Sair
      </button>
    </aside>
  )
}

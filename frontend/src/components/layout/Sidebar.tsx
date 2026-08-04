import { Link, NavLink } from 'react-router-dom'
import { Activity, Calendar, LogOut, Users, Wallet, Package, FileText, UserCog, Settings } from 'lucide-react'

interface SidebarProps {
  clinicName: string
  isAdmin: boolean
  onLogout: () => void
}

function navLinkClass(isActive: boolean): string {
  return `flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors ${
    isActive ? 'bg-brand-action text-white' : 'text-slate-300 hover:bg-white/5 hover:text-white'
  }`
}

export function Sidebar({ clinicName, isAdmin, onLogout }: SidebarProps) {
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
        <NavLink to="/financeiro" className={({ isActive }) => navLinkClass(isActive)}>
          <Wallet size={18} />
          Financeiro
        </NavLink>
        <NavLink to="/documentos" className={({ isActive }) => navLinkClass(isActive)}>
          <FileText size={18} />
          Documentos
        </NavLink>
        <NavLink to="/estoque" className={({ isActive }) => navLinkClass(isActive)}>
          <Package size={18} />
          Estoque
        </NavLink>
        {isAdmin && (
          <NavLink to="/usuarios" className={({ isActive }) => navLinkClass(isActive)}>
            <UserCog size={18} />
            Usuários
          </NavLink>
        )}
        {isAdmin && (
          <NavLink to="/configuracoes" className={({ isActive }) => navLinkClass(isActive)}>
            <Settings size={18} />
            Configurações
          </NavLink>
        )}
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

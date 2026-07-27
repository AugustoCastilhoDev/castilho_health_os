import { Plus } from 'lucide-react'
import { formatTodayLong, initials } from '../../lib/format'

interface DashboardHeaderProps {
  professionalName: string
  professionalRole: string
  onNewAppointment?: () => void
}

export function DashboardHeader({
  professionalName,
  professionalRole,
  onNewAppointment,
}: DashboardHeaderProps) {
  return (
    <header className="flex items-center justify-between border-b border-slate-200 bg-brand-surface px-8 py-5">
      <div>
        <p className="text-sm text-brand-text-muted">{formatTodayLong()}</p>
        <h1 className="text-xl font-semibold text-brand-text">Bem-vindo(a) de volta</h1>
      </div>

      <div className="flex items-center gap-5">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-slate-200 text-sm font-semibold text-brand-text">
            {initials(professionalName)}
          </div>
          <div className="text-right">
            <p className="text-sm font-medium text-brand-text">{professionalName}</p>
            <p className="text-xs text-brand-text-muted">{professionalRole}</p>
          </div>
        </div>

        <button
          type="button"
          onClick={onNewAppointment}
          className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
        >
          <Plus size={18} />
          Novo Agendamento
        </button>
      </div>
    </header>
  )
}

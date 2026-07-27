import { useMemo, useState } from 'react'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { WeeklyAgendaGrid, type AgendaAppointment } from '../components/agenda/WeeklyAgendaGrid'
import { addDays, formatWeekRange, startOfWeek } from '../lib/format'
import { APPOINTMENT_STATUS_LABEL, APPOINTMENT_STATUS_STYLE, type AppointmentStatus } from '../lib/appointmentStatus'

// Placeholder until wired to GET /api/appointments?professional_id=&from=&to=
// — deterministic per week so navigating weeks still looks plausible.
function generateMockWeek(weekStart: Date): AgendaAppointment[] {
  const at = (dayOffset: number, hour: number, minute: number) => {
    const d = addDays(weekStart, dayOffset)
    d.setHours(hour, minute, 0, 0)
    return d
  }

  return [
    { id: '1', patientName: 'Mariana Costa', scheduledAt: at(0, 9, 0), durationMin: 30, status: 'CONFIRMED' },
    { id: '2', patientName: 'João Pereira', scheduledAt: at(0, 10, 30), durationMin: 60, status: 'COMPLETED' },
    { id: '3', patientName: 'Beatriz Lima', scheduledAt: at(1, 8, 30), durationMin: 30, status: 'SCHEDULED' },
    { id: '4', patientName: 'Carlos Nunes', scheduledAt: at(1, 14, 0), durationMin: 30, status: 'CANCELLED' },
    { id: '5', patientName: 'Fernanda Alves', scheduledAt: at(2, 9, 30), durationMin: 30, status: 'IN_PROGRESS' },
    { id: '6', patientName: 'Rafael Souza', scheduledAt: at(2, 11, 0), durationMin: 30, status: 'WAITING' },
    { id: '7', patientName: 'Patrícia Gomes', scheduledAt: at(3, 15, 0), durationMin: 45, status: 'COMPLETED' },
    { id: '8', patientName: 'Lucas Martins', scheduledAt: at(4, 10, 0), durationMin: 30, status: 'NO_SHOW' },
    { id: '9', patientName: 'Camila Rocha', scheduledAt: at(4, 16, 30), durationMin: 30, status: 'CONFIRMED' },
  ]
}

const LEGEND_STATUSES: AppointmentStatus[] = [
  'SCHEDULED',
  'CONFIRMED',
  'IN_PROGRESS',
  'COMPLETED',
  'CANCELLED',
  'NO_SHOW',
]

export function AgendaPage() {
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()))
  const appointments = useMemo(() => generateMockWeek(weekStart), [weekStart])

  return (
    <>
      <header className="flex items-center justify-between border-b border-slate-200 bg-brand-surface px-8 py-5">
        <div>
          <p className="text-sm text-brand-text-muted">Agenda</p>
          <h1 className="text-xl font-semibold text-brand-text">
            {formatWeekRange(weekStart, addDays(weekStart, 6))}
          </h1>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex items-center rounded-lg ring-1 ring-slate-200">
            <button
              type="button"
              onClick={() => setWeekStart((w) => addDays(w, -7))}
              className="rounded-l-lg p-2 text-brand-text-muted hover:bg-slate-50"
              aria-label="Semana anterior"
            >
              <ChevronLeft size={18} />
            </button>
            <button
              type="button"
              onClick={() => setWeekStart(startOfWeek(new Date()))}
              className="border-x border-slate-200 px-3 py-2 text-sm font-medium text-brand-text hover:bg-slate-50"
            >
              Hoje
            </button>
            <button
              type="button"
              onClick={() => setWeekStart((w) => addDays(w, 7))}
              className="rounded-r-lg p-2 text-brand-text-muted hover:bg-slate-50"
              aria-label="Próxima semana"
            >
              <ChevronRight size={18} />
            </button>
          </div>

          <button
            type="button"
            className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
          >
            <Plus size={18} />
            Novo Agendamento
          </button>
        </div>
      </header>

      <main className="p-6">
        <WeeklyAgendaGrid weekStart={weekStart} appointments={appointments} />

        <div className="mt-4 flex flex-wrap gap-x-6 gap-y-2">
          {LEGEND_STATUSES.map((status) => (
            <div key={status} className="flex items-center gap-2 text-xs text-brand-text-muted">
              <span className={`h-2.5 w-2.5 rounded-full border ${APPOINTMENT_STATUS_STYLE[status]}`} />
              {APPOINTMENT_STATUS_LABEL[status]}
            </div>
          ))}
        </div>
      </main>
    </>
  )
}

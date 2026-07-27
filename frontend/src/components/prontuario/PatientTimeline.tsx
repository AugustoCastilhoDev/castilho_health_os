import { StatusBadge } from '../common/StatusBadge'
import { formatDateLong, formatTime } from '../../lib/format'
import { APPOINTMENT_STATUS_DOT, type AppointmentStatus } from '../../lib/appointmentStatus'

// One entry per Appointment. The backend doesn't model clinical notes yet
// (see models.Patient's doc comment — PEP/prontuário content is a separate,
// not-yet-built module), so `notes` here is only ever the
// CancellationReason when present; there is no free-text encounter note.
export interface TimelineEntry {
  id: string
  date: Date
  professionalName: string
  status: AppointmentStatus
  notes?: string
}

interface PatientTimelineProps {
  entries: TimelineEntry[]
}

export function PatientTimeline({ entries }: PatientTimelineProps) {
  if (entries.length === 0) {
    return (
      <p className="rounded-xl bg-brand-surface p-6 text-sm text-brand-text-muted ring-1 ring-slate-200">
        Nenhum registro no histórico ainda.
      </p>
    )
  }

  return (
    <ol className="relative border-l-2 border-slate-200 pl-8">
      {entries.map((entry) => (
        <li key={entry.id} className="relative mb-8 last:mb-0">
          <span
            className={`absolute -left-[37px] top-1.5 h-4 w-4 rounded-full ring-4 ring-brand-bg ${APPOINTMENT_STATUS_DOT[entry.status]}`}
          />
          <div className="rounded-xl bg-brand-surface p-5 shadow-sm ring-1 ring-slate-200">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <span className="text-xs font-medium uppercase tracking-wide text-brand-text-muted">
                {formatDateLong(entry.date)} · {formatTime(entry.date)}
              </span>
              <StatusBadge status={entry.status} />
            </div>
            <h3 className="mt-1 text-base font-semibold text-brand-text">
              Consulta com {entry.professionalName}
            </h3>
            <p className="mt-3 text-sm text-brand-text">
              {entry.notes || 'Nenhuma observação registrada para este atendimento.'}
            </p>
          </div>
        </li>
      ))}
    </ol>
  )
}

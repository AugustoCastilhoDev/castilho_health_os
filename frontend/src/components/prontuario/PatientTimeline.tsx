import { StatusBadge } from '../common/StatusBadge'
import { formatDateLong } from '../../lib/format'
import { APPOINTMENT_STATUS_DOT, type AppointmentStatus } from '../../lib/appointmentStatus'

export type EncounterType = 'CONSULTA' | 'RETORNO' | 'PROCEDIMENTO' | 'EXAME'

const ENCOUNTER_TYPE_LABEL: Record<EncounterType, string> = {
  CONSULTA: 'Consulta',
  RETORNO: 'Retorno',
  PROCEDIMENTO: 'Procedimento',
  EXAME: 'Exame',
}

export interface TimelineEntry {
  id: string
  date: Date
  type: EncounterType
  title: string
  professionalName: string
  status: AppointmentStatus
  notes: string
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
                {formatDateLong(entry.date)}
              </span>
              <StatusBadge status={entry.status} />
            </div>
            <h3 className="mt-1 text-base font-semibold text-brand-text">{entry.title}</h3>
            <p className="text-sm text-brand-text-muted">
              {entry.professionalName} · {ENCOUNTER_TYPE_LABEL[entry.type]}
            </p>
            <p className="mt-3 text-sm text-brand-text">{entry.notes}</p>
          </div>
        </li>
      ))}
    </ol>
  )
}

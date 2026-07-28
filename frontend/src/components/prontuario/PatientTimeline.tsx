import { FileText, Lock } from 'lucide-react'
import { StatusBadge } from '../common/StatusBadge'
import { formatDateLong, formatTime } from '../../lib/format'
import { APPOINTMENT_STATUS_DOT, type AppointmentStatus } from '../../lib/appointmentStatus'
import { RECORD_TYPE_LABEL, type MedicalRecordType } from '../../lib/medicalRecord'

export interface AppointmentTimelineEntry {
  kind: 'appointment'
  id: string
  date: Date
  professionalName: string
  status: AppointmentStatus
  notes?: string
}

export interface RecordTimelineEntry {
  kind: 'record'
  id: string
  date: Date
  professionalName: string
  type: MedicalRecordType
  content: string
  isLocked: boolean
  onEdit?: () => void
  onLock?: () => void
}

export type TimelineEntry = AppointmentTimelineEntry | RecordTimelineEntry

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
      {entries.map((entry) =>
        entry.kind === 'appointment' ? (
          <li key={`appt-${entry.id}`} className="relative mb-8 last:mb-0">
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
        ) : (
          <li key={`rec-${entry.id}`} className="relative mb-8 last:mb-0">
            <span className="absolute -left-[37px] top-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-brand-action ring-4 ring-brand-bg" />
            <div className="rounded-xl bg-brand-surface p-5 shadow-sm ring-1 ring-slate-200">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="text-xs font-medium uppercase tracking-wide text-brand-text-muted">
                  {formatDateLong(entry.date)} · {formatTime(entry.date)}
                </span>
                <span className="inline-flex items-center gap-1 rounded-full border border-sky-200 bg-sky-50 px-2.5 py-0.5 text-xs font-medium text-brand-action">
                  <FileText size={12} />
                  {RECORD_TYPE_LABEL[entry.type] ?? entry.type}
                </span>
              </div>
              <h3 className="mt-1 flex items-center gap-2 text-base font-semibold text-brand-text">
                Registro de {entry.professionalName}
                {entry.isLocked && (
                  <span className="inline-flex items-center gap-1 text-xs font-medium text-brand-text-muted">
                    <Lock size={12} />
                    Finalizado
                  </span>
                )}
              </h3>
              <p className="mt-3 whitespace-pre-wrap text-sm text-brand-text">{entry.content}</p>
              {!entry.isLocked && (entry.onEdit || entry.onLock) && (
                <div className="mt-4 flex gap-2">
                  {entry.onEdit && (
                    <button
                      type="button"
                      onClick={entry.onEdit}
                      className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                    >
                      Editar
                    </button>
                  )}
                  {entry.onLock && (
                    <button
                      type="button"
                      onClick={entry.onLock}
                      className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                    >
                      Finalizar registro
                    </button>
                  )}
                </div>
              )}
            </div>
          </li>
        ),
      )}
    </ol>
  )
}

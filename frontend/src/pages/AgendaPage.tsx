import { useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { WeeklyAgendaGrid, type AgendaAppointment } from '../components/agenda/WeeklyAgendaGrid'
import { NewAppointmentModal } from '../components/agenda/NewAppointmentModal'
import { AppointmentDetailModal } from '../components/agenda/AppointmentDetailModal'
import { addDays, formatWeekRange, startOfWeek } from '../lib/format'
import { APPOINTMENT_STATUS_LABEL, APPOINTMENT_STATUS_STYLE, type AppointmentStatus } from '../lib/appointmentStatus'
import { useAuth } from '../lib/auth/AuthContext'
import { useProfessionalScope } from '../hooks/useProfessionalScope'
import { ApiError } from '../lib/api/client'
import type { AppointmentDTO, PatientDTO } from '../lib/api/types'

const LEGEND_STATUSES: AppointmentStatus[] = [
  'SCHEDULED',
  'CONFIRMED',
  'IN_PROGRESS',
  'COMPLETED',
  'CANCELLED',
  'NO_SHOW',
]

export function AgendaPage() {
  const { apiFetch } = useAuth()
  const { professionals, professionalId, setProfessionalId } = useProfessionalScope()
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()))
  const [appointments, setAppointments] = useState<AgendaAppointment[]>([])
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [showNewAppointment, setShowNewAppointment] = useState(false)
  const [selectedAppointment, setSelectedAppointment] = useState<AgendaAppointment | null>(null)

  useEffect(() => {
    if (!professionalId) return
    let cancelled = false

    async function load() {
      setError(null)
      try {
        const from = weekStart.toISOString()
        const to = addDays(weekStart, 7).toISOString()
        const data = await apiFetch<AppointmentDTO[]>(
          `/api/appointments?professional_id=${professionalId}&from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
        )

        // Appointment doesn't embed the patient (only patient_id) — resolve
        // display names with one lookup per unique patient in the week.
        const uniquePatientIds = [...new Set(data.map((a) => a.patient_id))]
        const patients = await Promise.all(
          uniquePatientIds.map((id) => apiFetch<PatientDTO>(`/api/patients/${id}`)),
        )
        const nameByPatientId = new Map(patients.map((p) => [p.id, p.name]))

        if (!cancelled) {
          setAppointments(
            data.map((a) => ({
              id: a.id,
              patientName: nameByPatientId.get(a.patient_id) ?? 'Paciente',
              scheduledAt: new Date(a.scheduled_at),
              durationMin: a.duration_min,
              status: a.status as AppointmentStatus,
              cid: a.cid,
            })),
          )
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar a agenda.')
        }
      }
    }

    load()
    return () => {
      cancelled = true
    }
  }, [professionalId, weekStart, reloadTick, apiFetch])

  return (
    <>
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 bg-brand-surface px-8 py-5">
        <div>
          <p className="text-sm text-brand-text-muted">Agenda</p>
          <h1 className="text-xl font-semibold text-brand-text">
            {formatWeekRange(weekStart, addDays(weekStart, 6))}
          </h1>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          {professionals.length > 1 && (
            <select
              value={professionalId ?? ''}
              onChange={(e) => setProfessionalId(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm text-brand-text focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              {professionals.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          )}

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
            onClick={() => setShowNewAppointment(true)}
            disabled={!professionalId}
            className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
          >
            <Plus size={18} />
            Novo Agendamento
          </button>
        </div>
      </header>

      <main className="p-6">
        {error && (
          <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
            {error}
          </p>
        )}

        <WeeklyAgendaGrid
          weekStart={weekStart}
          appointments={appointments}
          onAppointmentClick={setSelectedAppointment}
        />

        <div className="mt-4 flex flex-wrap gap-x-6 gap-y-2">
          {LEGEND_STATUSES.map((status) => (
            <div key={status} className="flex items-center gap-2 text-xs text-brand-text-muted">
              <span className={`h-2.5 w-2.5 rounded-full border ${APPOINTMENT_STATUS_STYLE[status]}`} />
              {APPOINTMENT_STATUS_LABEL[status]}
            </div>
          ))}
        </div>
      </main>

      {showNewAppointment && (
        <NewAppointmentModal
          professionals={professionals}
          defaultProfessionalId={professionalId}
          onClose={() => setShowNewAppointment(false)}
          onCreated={() => {
            setShowNewAppointment(false)
            setReloadTick((t) => t + 1)
          }}
        />
      )}

      {selectedAppointment && (
        <AppointmentDetailModal
          appointment={selectedAppointment}
          onClose={() => setSelectedAppointment(null)}
          onUpdated={() => {
            setSelectedAppointment(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}
    </>
  )
}

import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Cake, IdCard, Mail, Pencil, Phone, Plus } from 'lucide-react'
import { PatientTimeline, type TimelineEntry } from '../components/prontuario/PatientTimeline'
import { calculateAge, initials } from '../lib/format'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import type { AppointmentDTO, PatientDTO, UserDTO } from '../lib/api/types'
import type { AppointmentStatus } from '../lib/appointmentStatus'

export function PatientRecordPage() {
  const { patientId } = useParams<{ patientId: string }>()
  const { apiFetch } = useAuth()
  const [patient, setPatient] = useState<PatientDTO | null>(null)
  const [timeline, setTimeline] = useState<TimelineEntry[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!patientId) return
    let cancelled = false

    async function load() {
      setError(null)
      try {
        const [patientData, appointments] = await Promise.all([
          apiFetch<PatientDTO>(`/api/patients/${patientId}`),
          apiFetch<AppointmentDTO[]>(`/api/patients/${patientId}/appointments`),
        ])

        // Appointment doesn't embed the professional (only professional_id)
        // — resolve display names with one lookup per unique professional.
        const uniqueProfessionalIds = [...new Set(appointments.map((a) => a.professional_id))]
        const professionals = await Promise.all(
          uniqueProfessionalIds.map((id) => apiFetch<UserDTO>(`/api/users/${id}`)),
        )
        const nameByProfessionalId = new Map(professionals.map((p) => [p.id, p.name]))

        if (cancelled) return
        setPatient(patientData)
        setTimeline(
          appointments
            .slice()
            .sort((a, b) => new Date(b.scheduled_at).getTime() - new Date(a.scheduled_at).getTime())
            .map((a) => ({
              id: a.id,
              date: new Date(a.scheduled_at),
              professionalName: nameByProfessionalId.get(a.professional_id) ?? 'Profissional',
              status: a.status as AppointmentStatus,
              notes: a.cancellation_reason,
            })),
        )
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar o prontuário.')
        }
      }
    }

    load()
    return () => {
      cancelled = true
    }
  }, [patientId, apiFetch])

  if (error) {
    return (
      <main className="p-6">
        <p className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
          {error}
        </p>
      </main>
    )
  }

  if (!patient) {
    return <main className="p-6 text-sm text-brand-text-muted">Carregando…</main>
  }

  const age = patient.birth_date ? calculateAge(new Date(patient.birth_date)) : null

  return (
    <>
      <header className="border-b border-slate-200 bg-brand-surface px-8 py-5">
        <p className="text-sm text-brand-text-muted">Prontuário Eletrônico</p>
        <h1 className="text-xl font-semibold text-brand-text">{patient.name}</h1>
      </header>

      <main className="space-y-6 p-6">
        <div className="flex flex-wrap items-center justify-between gap-4 rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200">
          <div className="flex items-center gap-4">
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-slate-200 text-lg font-semibold text-brand-text">
              {initials(patient.name)}
            </div>
            <div>
              <p className="text-base font-semibold text-brand-text">{patient.name}</p>
              {age !== null && <p className="text-sm text-brand-text-muted">{age} anos</p>}
            </div>
          </div>

          <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-brand-text-muted">
            {age !== null && (
              <span className="flex items-center gap-1.5">
                <Cake size={16} />
                {age} anos
              </span>
            )}
            {patient.document && (
              <span className="flex items-center gap-1.5">
                <IdCard size={16} />
                {patient.document}
              </span>
            )}
            {patient.phone && (
              <span className="flex items-center gap-1.5">
                <Phone size={16} />
                {patient.phone}
              </span>
            )}
            {patient.email && (
              <span className="flex items-center gap-1.5">
                <Mail size={16} />
                {patient.email}
              </span>
            )}
          </div>

          <div className="flex gap-2">
            <Link
              to={`/pacientes/${patient.id}/editar`}
              className="flex items-center gap-2 rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium text-brand-text hover:bg-slate-50"
            >
              <Pencil size={16} />
              Editar
            </Link>
            <button
              type="button"
              className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
            >
              <Plus size={18} />
              Novo Registro
            </button>
          </div>
        </div>

        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-brand-text-muted">
            Histórico
          </h2>
          <PatientTimeline entries={timeline} />
        </div>
      </main>
    </>
  )
}

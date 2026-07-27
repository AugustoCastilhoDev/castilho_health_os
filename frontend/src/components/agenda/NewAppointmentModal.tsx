import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import type { AppointmentDTO, PatientDTO, UserDTO } from '../../lib/api/types'

interface NewAppointmentModalProps {
  professionals: UserDTO[]
  defaultProfessionalId: string | null
  onClose: () => void
  onCreated: (appt: AppointmentDTO) => void
}

export function NewAppointmentModal({
  professionals,
  defaultProfessionalId,
  onClose,
  onCreated,
}: NewAppointmentModalProps) {
  const { apiFetch } = useAuth()
  const [professionalId, setProfessionalId] = useState(defaultProfessionalId ?? '')
  const [patientQuery, setPatientQuery] = useState('')
  const [patientResults, setPatientResults] = useState<PatientDTO[]>([])
  const [selectedPatient, setSelectedPatient] = useState<PatientDTO | null>(null)
  const [scheduledAt, setScheduledAt] = useState('')
  const [durationMin, setDurationMin] = useState(30)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (selectedPatient || patientQuery.trim().length < 2) {
      setPatientResults([])
      return
    }
    let cancelled = false
    const handle = setTimeout(() => {
      apiFetch<PatientDTO[]>(`/api/patients?q=${encodeURIComponent(patientQuery)}&limit=6`)
        .then((data) => {
          if (!cancelled) setPatientResults(data ?? [])
        })
        .catch(() => {
          if (!cancelled) setPatientResults([])
        })
    }, 250)
    return () => {
      cancelled = true
      clearTimeout(handle)
    }
  }, [patientQuery, selectedPatient, apiFetch])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!selectedPatient) {
      setError('Selecione um paciente.')
      return
    }
    if (!professionalId) {
      setError('Selecione um profissional.')
      return
    }
    if (!scheduledAt) {
      setError('Informe a data e o horário.')
      return
    }
    setSubmitting(true)
    try {
      const appt = await apiFetch<AppointmentDTO>('/api/appointments/', {
        method: 'POST',
        body: JSON.stringify({
          patient_id: selectedPatient.id,
          professional_id: professionalId,
          scheduled_at: new Date(scheduledAt).toISOString(),
          duration_min: durationMin,
        }),
      })
      onCreated(appt)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível criar o agendamento.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="Novo Agendamento" onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Paciente</label>
          {selectedPatient ? (
            <div className="flex items-center justify-between rounded-lg border border-slate-300 px-3 py-2 text-sm">
              <span className="text-brand-text">{selectedPatient.name}</span>
              <button
                type="button"
                onClick={() => {
                  setSelectedPatient(null)
                  setPatientQuery('')
                }}
                className="text-xs text-brand-action hover:underline"
              >
                trocar
              </button>
            </div>
          ) : (
            <div className="relative">
              <input
                value={patientQuery}
                onChange={(e) => setPatientQuery(e.target.value)}
                placeholder="Buscar paciente por nome…"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
              />
              {patientResults.length > 0 && (
                <ul className="absolute z-10 mt-1 w-full rounded-lg border border-slate-200 bg-brand-surface shadow-sm">
                  {patientResults.map((p) => (
                    <li key={p.id}>
                      <button
                        type="button"
                        onClick={() => {
                          setSelectedPatient(p)
                          setPatientResults([])
                        }}
                        className="block w-full px-3 py-2 text-left text-sm hover:bg-slate-50"
                      >
                        {p.name}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </div>

        {professionals.length > 1 && (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Profissional</label>
            <select
              value={professionalId}
              onChange={(e) => setProfessionalId(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              <option value="" disabled>
                Selecione…
              </option>
              {professionals.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Data e horário</label>
            <input
              type="datetime-local"
              required
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Duração (min)</label>
            <input
              type="number"
              min={5}
              step={5}
              required
              value={durationMin}
              onChange={(e) => setDurationMin(Number(e.target.value))}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Agendando…' : 'Agendar'}
        </button>
      </form>
    </Modal>
  )
}

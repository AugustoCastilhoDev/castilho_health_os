import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Cake, FileOutput, IdCard, Mail, Pencil, Phone, Plus } from 'lucide-react'
import { PatientTimeline, type TimelineEntry } from '../components/prontuario/PatientTimeline'
import { MedicalRecordFormModal } from '../components/prontuario/MedicalRecordFormModal'
import { GenerateDocumentModal } from '../components/prontuario/GenerateDocumentModal'
import { PatientDocumentsPanel } from '../components/prontuario/PatientDocumentsPanel'
import { MemedPrescriptionsPanel } from '../components/prontuario/MemedPrescriptionsPanel'
import { OdontogramaPanel } from '../components/prontuario/OdontogramaPanel'
import { calculateAge, initials } from '../lib/format'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import type { AppointmentDTO, MedicalRecordDTO, PatientDTO, UserDTO } from '../lib/api/types'
import type { AppointmentStatus } from '../lib/appointmentStatus'
import type { MedicalRecordType } from '../lib/medicalRecord'
import { HEALTH_PROFESSIONAL_ROLES } from '../lib/roles'

const CAN_WRITE_RECORDS_ROLES = HEALTH_PROFESSIONAL_ROLES

export function PatientRecordPage() {
  const { patientId } = useParams<{ patientId: string }>()
  const { apiFetch, user } = useAuth()
  const [patient, setPatient] = useState<PatientDTO | null>(null)
  const [timeline, setTimeline] = useState<TimelineEntry[]>([])
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [editingRecord, setEditingRecord] = useState<MedicalRecordDTO | 'new' | null>(null)
  const [generatingDocument, setGeneratingDocument] = useState(false)

  const canWriteRecords = user ? CAN_WRITE_RECORDS_ROLES.has(user.role) : false

  useEffect(() => {
    if (!patientId) return
    let cancelled = false

    async function load() {
      setError(null)
      try {
        const [patientData, appointments, records] = await Promise.all([
          apiFetch<PatientDTO>(`/api/patients/${patientId}`),
          apiFetch<AppointmentDTO[]>(`/api/patients/${patientId}/appointments`),
          apiFetch<MedicalRecordDTO[]>(`/api/patients/${patientId}/medical-records`),
        ])

        // Neither Appointment nor MedicalRecord embeds the professional
        // (only professional_id) — resolve display names with one lookup
        // per unique professional across both lists.
        const uniqueProfessionalIds = [
          ...new Set([...appointments.map((a) => a.professional_id), ...records.map((r) => r.professional_id)]),
        ]
        const professionals = await Promise.all(
          uniqueProfessionalIds.map((id) => apiFetch<UserDTO>(`/api/users/${id}`)),
        )
        const nameByProfessionalId = new Map(professionals.map((p) => [p.id, p.name]))

        if (cancelled) return
        setPatient(patientData)

        const appointmentEntries: TimelineEntry[] = appointments.map((a) => ({
          kind: 'appointment',
          id: a.id,
          date: new Date(a.scheduled_at),
          professionalName: nameByProfessionalId.get(a.professional_id) ?? 'Profissional',
          status: a.status as AppointmentStatus,
          notes: a.cancellation_reason,
        }))
        const recordEntries: TimelineEntry[] = records.map((r) => ({
          kind: 'record',
          id: r.id,
          date: new Date(r.created_at),
          professionalName: nameByProfessionalId.get(r.professional_id) ?? 'Profissional',
          type: r.type as MedicalRecordType,
          content: r.content,
          cid: r.cid,
          isLocked: r.is_locked,
          onEdit: canWriteRecords ? () => setEditingRecord(r) : undefined,
          onLock: canWriteRecords ? () => handleLock(r.id) : undefined,
        }))

        setTimeline(
          [...appointmentEntries, ...recordEntries].sort((a, b) => b.date.getTime() - a.date.getTime()),
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [patientId, apiFetch, reloadTick])

  async function handleLock(recordId: string) {
    setError(null)
    try {
      await apiFetch(`/api/medical-records/${recordId}/lock`, { method: 'POST' })
      setReloadTick((t) => t + 1)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível finalizar o registro.')
    }
  }

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
              onClick={() => setGeneratingDocument(true)}
              className="flex items-center gap-2 rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium text-brand-text hover:bg-slate-50"
            >
              <FileOutput size={16} />
              Gerar Documento
            </button>
            {canWriteRecords && (
              <button
                type="button"
                onClick={() => setEditingRecord('new')}
                className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
              >
                <Plus size={18} />
                Novo Registro
              </button>
            )}
          </div>
        </div>

        <MemedPrescriptionsPanel patient={patient} />

        <OdontogramaPanel patientId={patient.id} />

        <PatientDocumentsPanel patientId={patient.id} />

        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-brand-text-muted">
            Histórico
          </h2>
          <PatientTimeline entries={timeline} />
        </div>
      </main>

      {editingRecord && user && (
        <MedicalRecordFormModal
          patientId={patient.id}
          professionalId={user.id}
          professionalRole={user.role}
          existingRecord={editingRecord === 'new' ? undefined : editingRecord}
          onClose={() => setEditingRecord(null)}
          onSaved={() => {
            setEditingRecord(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}

      {generatingDocument && (
        <GenerateDocumentModal
          patientId={patient.id}
          professionalId={canWriteRecords ? user?.id : undefined}
          onClose={() => setGeneratingDocument(false)}
        />
      )}
    </>
  )
}

import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { RichTextEditor } from '../common/RichTextEditor'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { DEFAULT_RECORD_TYPE_BY_ROLE, RECORD_TYPE_LABEL, type MedicalRecordType } from '../../lib/medicalRecord'
import type { MedicalRecordDTO } from '../../lib/api/types'

// An empty Tiptap doc serializes to "<p></p>", not "" — strip tags before
// checking for blank content so that doesn't slip past validation as
// "non-empty".
function isHtmlContentEmpty(html: string): boolean {
  return !html.replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').trim()
}

interface MedicalRecordFormModalProps {
  patientId: string
  professionalId: string
  professionalRole: string
  existingRecord?: MedicalRecordDTO
  onClose: () => void
  onSaved: (record: MedicalRecordDTO) => void
}

export function MedicalRecordFormModal({
  patientId,
  professionalId,
  professionalRole,
  existingRecord,
  onClose,
  onSaved,
}: MedicalRecordFormModalProps) {
  const { apiFetch } = useAuth()
  const isEditing = !!existingRecord
  const [type, setType] = useState<MedicalRecordType>(
    (existingRecord?.type as MedicalRecordType) ?? DEFAULT_RECORD_TYPE_BY_ROLE[professionalRole] ?? 'MEDICA',
  )
  const [content, setContent] = useState(existingRecord?.content ?? '')
  const [cid, setCid] = useState(existingRecord?.cid ?? '')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (isHtmlContentEmpty(content)) {
      setError('Escreva o conteúdo do registro.')
      return
    }
    setSubmitting(true)
    try {
      const record = isEditing
        ? await apiFetch<MedicalRecordDTO>(`/api/medical-records/${existingRecord.id}`, {
            method: 'PUT',
            body: JSON.stringify({ type, content, cid: cid.trim() || null }),
          })
        : await apiFetch<MedicalRecordDTO>('/api/medical-records/', {
            method: 'POST',
            body: JSON.stringify({
              patient_id: patientId,
              professional_id: professionalId,
              type,
              content,
              cid: cid.trim() || null,
            }),
          })
      onSaved(record)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível salvar o registro.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={isEditing ? 'Editar Registro' : 'Novo Registro'} onClose={onClose} maxWidthClassName="max-w-xl">
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Tipo</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as MedicalRecordType)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            {Object.entries(RECORD_TYPE_LABEL).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">CID (opcional)</label>
          <input
            value={cid}
            onChange={(e) => setCid(e.target.value)}
            placeholder="Ex: F41.1 - Transtorno de ansiedade generalizada"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Evolução</label>
          <RichTextEditor
            value={content}
            onChange={setContent}
            placeholder="Descreva a evolução clínica do paciente…"
          />
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : isEditing ? 'Salvar Alterações' : 'Adicionar Registro'}
        </button>
      </form>
    </Modal>
  )
}

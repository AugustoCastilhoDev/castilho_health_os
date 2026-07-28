import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import type { PatientDocumentDTO } from '../../lib/api/types'

interface UploadDocumentModalProps {
  patientId: string
  onClose: () => void
  onUploaded: (doc: PatientDocumentDTO) => void
}

export function UploadDocumentModal({ patientId, onClose, onUploaded }: UploadDocumentModalProps) {
  const { apiFetch } = useAuth()
  const [file, setFile] = useState<File | null>(null)
  const [description, setDescription] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!file) {
      setError('Selecione um arquivo.')
      return
    }
    setSubmitting(true)
    try {
      // Step 1: get a presigned PUT URL scoped to this patient/file.
      const { upload_url, file_key } = await apiFetch<{ upload_url: string; file_key: string }>(
        `/api/patients/${patientId}/documents/upload-url`,
        { method: 'POST', body: JSON.stringify({ file_name: file.name, mime_type: file.type || 'application/octet-stream' }) },
      )

      // Step 2: the browser PUTs the bytes straight to R2 — never through our API.
      const uploadRes = await fetch(upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': file.type || 'application/octet-stream' },
        body: file,
      })
      if (!uploadRes.ok) {
        throw new ApiError(uploadRes.status, 'Falha ao enviar o arquivo para o armazenamento.')
      }

      // Step 3: persist the metadata row now that the upload succeeded.
      const doc = await apiFetch<PatientDocumentDTO>(`/api/patients/${patientId}/documents`, {
        method: 'POST',
        body: JSON.stringify({
          file_key,
          file_name: file.name,
          file_size: file.size,
          mime_type: file.type || 'application/octet-stream',
          description,
        }),
      })
      onUploaded(doc)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível enviar o documento.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="Enviar Documento" onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Arquivo</label>
          <input
            type="file"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm file:mr-3 file:rounded-md file:border-0 file:bg-slate-100 file:px-3 file:py-1.5 file:text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Descrição (opcional)</label>
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Ex: Exame de sangue, laudo de raio-x…"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Enviando…' : 'Enviar Documento'}
        </button>
      </form>
    </Modal>
  )
}

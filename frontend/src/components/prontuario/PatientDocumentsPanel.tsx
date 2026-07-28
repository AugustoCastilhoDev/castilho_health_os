import { useEffect, useState } from 'react'
import { Download, Paperclip, Plus, Trash2 } from 'lucide-react'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { formatFileSize } from '../../lib/format'
import { UploadDocumentModal } from './UploadDocumentModal'
import type { PatientDocumentDTO } from '../../lib/api/types'

interface PatientDocumentsPanelProps {
  patientId: string
}

const CAN_DELETE_ROLES = new Set(['TENANT_ADMIN'])

export function PatientDocumentsPanel({ patientId }: PatientDocumentsPanelProps) {
  const { apiFetch, user } = useAuth()
  const [documents, setDocuments] = useState<PatientDocumentDTO[]>([])
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [uploading, setUploading] = useState(false)

  const canDelete = user ? CAN_DELETE_ROLES.has(user.role) : false

  useEffect(() => {
    let cancelled = false
    apiFetch<PatientDocumentDTO[]>(`/api/patients/${patientId}/documents`)
      .then((data) => {
        if (!cancelled) setDocuments(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) {
          // 503 here means R2 isn't configured yet — read the error message
          // straight from the API instead of a generic fallback, since it's
          // the useful one ("document storage is not configured").
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar os documentos.')
        }
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, patientId, reloadTick])

  async function handleDownload(doc: PatientDocumentDTO) {
    setError(null)
    try {
      const { url } = await apiFetch<{ url: string }>(`/api/patient-documents/${doc.id}/download-url`)
      window.open(url, '_blank', 'noopener')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível baixar o documento.')
    }
  }

  async function handleDelete(doc: PatientDocumentDTO) {
    setError(null)
    try {
      await apiFetch(`/api/patient-documents/${doc.id}`, { method: 'DELETE' })
      setReloadTick((t) => t + 1)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível excluir o documento.')
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-brand-text-muted">
          Documentos Anexados
        </h2>
        <button
          type="button"
          onClick={() => setUploading(true)}
          className="flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
        >
          <Plus size={14} />
          Enviar Documento
        </button>
      </div>

      {error && (
        <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
          {error}
        </p>
      )}

      <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
        {documents.length === 0 && !error && (
          <p className="p-6 text-sm text-brand-text-muted">Nenhum documento anexado ainda.</p>
        )}
        {documents.length > 0 && (
          <ul className="divide-y divide-slate-100">
            {documents.map((doc) => (
              <li key={doc.id} className="flex items-center justify-between gap-4 px-6 py-4">
                <div className="flex min-w-0 items-center gap-3">
                  <Paperclip size={18} className="shrink-0 text-brand-text-muted" />
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-brand-text">{doc.file_name}</p>
                    <p className="text-xs text-brand-text-muted">
                      {formatFileSize(doc.file_size)}
                      {doc.description ? ` · ${doc.description}` : ''}
                    </p>
                  </div>
                </div>
                <div className="flex shrink-0 gap-2">
                  <button
                    type="button"
                    onClick={() => handleDownload(doc)}
                    className="flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                  >
                    <Download size={14} />
                    Baixar
                  </button>
                  {canDelete && (
                    <button
                      type="button"
                      onClick={() => handleDelete(doc)}
                      className="flex items-center gap-1.5 rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-alert-text hover:bg-rose-50"
                    >
                      <Trash2 size={14} />
                      Excluir
                    </button>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {uploading && (
        <UploadDocumentModal
          patientId={patientId}
          onClose={() => setUploading(false)}
          onUploaded={() => {
            setUploading(false)
            setReloadTick((t) => t + 1)
          }}
        />
      )}
    </div>
  )
}

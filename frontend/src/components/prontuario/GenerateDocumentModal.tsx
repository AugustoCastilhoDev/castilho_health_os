import { useEffect, useState } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { API_BASE_URL } from '../../lib/api/client'
import { DOCUMENT_TEMPLATE_TYPE_LABEL, type DocumentTemplateType } from '../../lib/medicalRecord'
import type { DocumentTemplateDTO } from '../../lib/api/types'

interface GenerateDocumentModalProps {
  patientId: string
  professionalId?: string
  onClose: () => void
}

// patient_name/professional_name/date are always resolved server-side (see
// DocumentTemplateHandler.Generate) — only tags beyond these three need a
// value from this form.
const AUTO_VARS = new Set(['patient_name', 'professional_name', 'date'])

function extractTags(content: string): string[] {
  const matches = [...content.matchAll(/\{\{\s*(\w+)\s*\}\}/g)]
  return [...new Set(matches.map((m) => m[1]))].filter((t) => !AUTO_VARS.has(t))
}

function prettifyTag(tag: string): string {
  const spaced = tag.replace(/_/g, ' ')
  return spaced.charAt(0).toUpperCase() + spaced.slice(1)
}

async function downloadGeneratedPdf(
  token: string | null,
  templateId: string,
  body: { patient_id: string; professional_id?: string; vars: Record<string, string> },
) {
  const res = await fetch(`${API_BASE_URL}/api/document-templates/${templateId}/generate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const text = await res.text()
    let message = `Erro ${res.status}`
    try {
      const data = JSON.parse(text)
      if (data?.error) message = data.error
    } catch {
      /* body wasn't JSON, keep the generic message */
    }
    throw new ApiError(res.status, message)
  }
  const blob = await res.blob()
  const disposition = res.headers.get('Content-Disposition') ?? ''
  const filename = disposition.match(/filename="?([^"]+)"?/)?.[1] ?? 'documento.pdf'

  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function GenerateDocumentModal({ patientId, professionalId, onClose }: GenerateDocumentModalProps) {
  const { apiFetch, token } = useAuth()
  const [templates, setTemplates] = useState<DocumentTemplateDTO[]>([])
  const [templateId, setTemplateId] = useState('')
  const [extraVars, setExtraVars] = useState<Record<string, string>>({})
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    apiFetch<DocumentTemplateDTO[]>('/api/document-templates')
      .then((data) => {
        const active = (data ?? []).filter((t) => t.is_active)
        setTemplates(active)
        setTemplateId((current) => current || active[0]?.id || '')
      })
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Não foi possível carregar os modelos.'))
  }, [apiFetch])

  const selectedTemplate = templates.find((t) => t.id === templateId)
  const tags = selectedTemplate ? extractTags(selectedTemplate.content) : []

  async function handleSubmit() {
    if (!selectedTemplate) return
    setError(null)
    setSubmitting(true)
    try {
      await downloadGeneratedPdf(token, selectedTemplate.id, {
        patient_id: patientId,
        professional_id: professionalId,
        vars: extraVars,
      })
      onClose()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível gerar o documento.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="Gerar Documento" onClose={onClose}>
      <div className="space-y-4">
        {templates.length === 0 ? (
          <p className="text-sm text-brand-text-muted">
            Nenhum modelo de documento ativo. Cadastre um em Documentos antes de gerar.
          </p>
        ) : (
          <>
            <div>
              <label className="mb-1 block text-sm font-medium text-brand-text">Modelo</label>
              <select
                value={templateId}
                onChange={(e) => {
                  setTemplateId(e.target.value)
                  setExtraVars({})
                }}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
              >
                {templates.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name} ({DOCUMENT_TEMPLATE_TYPE_LABEL[t.type as DocumentTemplateType] ?? t.type})
                  </option>
                ))}
              </select>
            </div>

            {tags.map((tag) => (
              <div key={tag}>
                <label className="mb-1 block text-sm font-medium text-brand-text">{prettifyTag(tag)}</label>
                <input
                  value={extraVars[tag] ?? ''}
                  onChange={(e) => setExtraVars((prev) => ({ ...prev, [tag]: e.target.value }))}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                />
              </div>
            ))}

            {error && <p className="text-sm text-brand-alert-text">{error}</p>}

            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting}
              className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
            >
              {submitting ? 'Gerando…' : 'Gerar PDF'}
            </button>
          </>
        )}
      </div>
    </Modal>
  )
}

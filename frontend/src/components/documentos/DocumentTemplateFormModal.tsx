import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { DOCUMENT_TEMPLATE_TYPE_LABEL, type DocumentTemplateType } from '../../lib/medicalRecord'
import type { DocumentTemplateDTO } from '../../lib/api/types'

interface DocumentTemplateFormModalProps {
  existingTemplate?: DocumentTemplateDTO
  onClose: () => void
  onSaved: (tmpl: DocumentTemplateDTO) => void
}

export function DocumentTemplateFormModal({ existingTemplate, onClose, onSaved }: DocumentTemplateFormModalProps) {
  const { apiFetch } = useAuth()
  const isEditing = !!existingTemplate
  const [name, setName] = useState(existingTemplate?.name ?? '')
  const [type, setType] = useState<DocumentTemplateType>((existingTemplate?.type as DocumentTemplateType) ?? 'ATESTADO')
  const [content, setContent] = useState(existingTemplate?.content ?? '')
  const [isActive, setIsActive] = useState(existingTemplate?.is_active ?? true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!name.trim() || !content.trim()) {
      setError('Preencha nome e conteúdo do modelo.')
      return
    }
    setSubmitting(true)
    try {
      const body = { name, type, content, is_active: isActive }
      const tmpl = isEditing
        ? await apiFetch<DocumentTemplateDTO>(`/api/document-templates/${existingTemplate.id}`, {
            method: 'PUT',
            body: JSON.stringify(body),
          })
        : await apiFetch<DocumentTemplateDTO>('/api/document-templates/', {
            method: 'POST',
            body: JSON.stringify(body),
          })
      onSaved(tmpl)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível salvar o modelo.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={isEditing ? 'Editar Modelo de Documento' : 'Novo Modelo de Documento'} onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Nome</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Atestado de comparecimento"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Tipo</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as DocumentTemplateType)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            {Object.entries(DOCUMENT_TEMPLATE_TYPE_LABEL).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Conteúdo</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={8}
            placeholder={'Atesto que {{patient_name}} esteve em consulta em {{date}} com {{professional_name}}.'}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 font-mono text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
          <p className="mt-1 text-xs text-brand-text-muted">
            Use {'{{patient_name}}'}, {'{{professional_name}}'} e {'{{date}}'} — preenchidos automaticamente — ou
            crie suas próprias tags (ex: {'{{cid}}'}), preenchidas na hora de gerar o documento.
          </p>
        </div>

        {isEditing && (
          <label className="flex items-center gap-2 text-sm text-brand-text">
            <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
            Modelo ativo
          </label>
        )}

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : isEditing ? 'Salvar Alterações' : 'Criar Modelo'}
        </button>
      </form>
    </Modal>
  )
}

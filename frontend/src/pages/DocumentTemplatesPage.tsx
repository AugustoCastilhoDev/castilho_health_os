import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import { DOCUMENT_TEMPLATE_TYPE_LABEL, type DocumentTemplateType } from '../lib/medicalRecord'
import { DocumentTemplateFormModal } from '../components/documentos/DocumentTemplateFormModal'
import type { DocumentTemplateDTO } from '../lib/api/types'

const CAN_MANAGE_TEMPLATES_ROLES = new Set(['TENANT_ADMIN'])

export function DocumentTemplatesPage() {
  const { apiFetch, user } = useAuth()
  const [templates, setTemplates] = useState<DocumentTemplateDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [editingTemplate, setEditingTemplate] = useState<DocumentTemplateDTO | 'new' | null>(null)

  const canManage = user ? CAN_MANAGE_TEMPLATES_ROLES.has(user.role) : false

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    apiFetch<DocumentTemplateDTO[]>('/api/document-templates')
      .then((data) => {
        if (!cancelled) setTemplates(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar os modelos.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, reloadTick])

  async function toggleActive(tmpl: DocumentTemplateDTO) {
    setError(null)
    try {
      await apiFetch<DocumentTemplateDTO>(`/api/document-templates/${tmpl.id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...tmpl, is_active: !tmpl.is_active }),
      })
      setReloadTick((t) => t + 1)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível atualizar o modelo.')
    }
  }

  return (
    <>
      <header className="flex items-center justify-between border-b border-slate-200 bg-brand-surface px-8 py-5">
        <div>
          <p className="text-sm text-brand-text-muted">Documentos</p>
          <h1 className="text-xl font-semibold text-brand-text">Modelos de Documentos</h1>
        </div>
        {canManage && (
          <button
            type="button"
            onClick={() => setEditingTemplate('new')}
            className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
          >
            <Plus size={18} />
            Novo Modelo
          </button>
        )}
      </header>

      <main className="p-6">
        {error && (
          <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
            {error}
          </p>
        )}

        <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
          {!loading && templates.length === 0 && (
            <p className="p-6 text-sm text-brand-text-muted">Nenhum modelo de documento cadastrado.</p>
          )}
          {templates.length > 0 && (
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
                <tr>
                  <th className="px-6 py-3 font-medium">Nome</th>
                  <th className="px-6 py-3 font-medium">Tipo</th>
                  <th className="px-6 py-3 font-medium">Situação</th>
                  {canManage && <th className="px-6 py-3 font-medium" />}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {templates.map((tmpl) => (
                  <tr key={tmpl.id}>
                    <td className="px-6 py-4 font-medium text-brand-text">{tmpl.name}</td>
                    <td className="px-6 py-4 text-brand-text-muted">
                      {DOCUMENT_TEMPLATE_TYPE_LABEL[tmpl.type as DocumentTemplateType] ?? tmpl.type}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${
                          tmpl.is_active
                            ? 'border-emerald-200 bg-emerald-50 text-brand-success-text'
                            : 'border-slate-300 bg-slate-100 text-slate-600'
                        }`}
                      >
                        {tmpl.is_active ? 'Ativo' : 'Inativo'}
                      </span>
                    </td>
                    {canManage && (
                      <td className="px-6 py-4 text-right">
                        <div className="flex justify-end gap-2">
                          <button
                            type="button"
                            onClick={() => setEditingTemplate(tmpl)}
                            className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                          >
                            Editar
                          </button>
                          <button
                            type="button"
                            onClick={() => toggleActive(tmpl)}
                            className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                          >
                            {tmpl.is_active ? 'Desativar' : 'Ativar'}
                          </button>
                        </div>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>

      {editingTemplate && (
        <DocumentTemplateFormModal
          existingTemplate={editingTemplate === 'new' ? undefined : editingTemplate}
          onClose={() => setEditingTemplate(null)}
          onSaved={() => {
            setEditingTemplate(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}
    </>
  )
}

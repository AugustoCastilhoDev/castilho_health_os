import { useRef, useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import type { StockImportResultDTO } from '../../lib/api/types'

interface StockImportModalProps {
  onClose: () => void
  onImported: () => void
}

const CSV_TEMPLATE = 'nome,unidade,quantidade_minima,quantidade_inicial\nLuva Cirúrgica,par,20,100\nGaze,cx,,\n'

function downloadCSVTemplate() {
  const blob = new Blob([CSV_TEMPLATE], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'modelo-estoque.csv'
  link.click()
  URL.revokeObjectURL(url)
}

export function StockImportModal({ onClose, onImported }: StockImportModalProps) {
  const { apiFetch } = useAuth()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [result, setResult] = useState<StockImportResultDTO | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    const file = fileInputRef.current?.files?.[0]
    if (!file) {
      setError('Selecione um arquivo CSV.')
      return
    }
    setSubmitting(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const imported = await apiFetch<StockImportResultDTO>('/api/stock-items/import', {
        method: 'POST',
        body: formData,
      })
      setResult(imported)
      if (imported.created.length > 0) onImported()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível importar o arquivo.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="Importar Estoque via CSV" onClose={onClose} maxWidthClassName="max-w-lg">
      {!result ? (
        <form className="space-y-4" onSubmit={handleSubmit}>
          <p className="text-sm text-brand-text-muted">
            O arquivo precisa ter as colunas <strong>nome</strong> e <strong>unidade</strong> (obrigatórias) e,
            opcionalmente, <strong>quantidade_minima</strong> e <strong>quantidade_inicial</strong>. Um item com
            nome já existente é pulado, nunca sobrescrito.
          </p>

          <button
            type="button"
            onClick={downloadCSVTemplate}
            className="text-sm font-medium text-brand-action hover:underline"
          >
            Baixar modelo CSV
          </button>

          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Arquivo CSV</label>
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv,text/csv"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>

          {error && <p className="text-sm text-brand-alert-text">{error}</p>}

          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
          >
            {submitting ? 'Importando…' : 'Importar'}
          </button>
        </form>
      ) : (
        <div className="space-y-4">
          <div className="grid grid-cols-3 gap-3 text-center">
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 p-3">
              <p className="text-2xl font-semibold text-brand-success-text">{result.created.length}</p>
              <p className="text-xs text-brand-text-muted">Criados</p>
            </div>
            <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
              <p className="text-2xl font-semibold text-brand-text">{result.skipped.length}</p>
              <p className="text-xs text-brand-text-muted">Pulados</p>
            </div>
            <div className="rounded-lg border border-rose-200 bg-rose-50 p-3">
              <p className="text-2xl font-semibold text-brand-alert-text">{result.failed.length}</p>
              <p className="text-xs text-brand-text-muted">Com erro</p>
            </div>
          </div>

          {(result.skipped.length > 0 || result.failed.length > 0) && (
            <div className="max-h-60 overflow-y-auto rounded-lg border border-slate-200">
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
                  <tr>
                    <th className="px-3 py-2 font-medium">Linha</th>
                    <th className="px-3 py-2 font-medium">Nome</th>
                    <th className="px-3 py-2 font-medium">Motivo</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {[...result.skipped, ...result.failed]
                    .sort((a, b) => a.row - b.row)
                    .map((issue, i) => (
                      <tr key={i}>
                        <td className="px-3 py-2 text-brand-text-muted">{issue.row}</td>
                        <td className="px-3 py-2 text-brand-text">{issue.name || '—'}</td>
                        <td className="px-3 py-2 text-brand-text-muted">{issue.reason}</td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          )}

          <button
            type="button"
            onClick={onClose}
            className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
          >
            Concluir
          </button>
        </div>
      )}
    </Modal>
  )
}

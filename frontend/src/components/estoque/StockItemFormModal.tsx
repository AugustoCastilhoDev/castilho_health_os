import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import type { StockItemDTO } from '../../lib/api/types'

interface StockItemFormModalProps {
  existingItem?: StockItemDTO
  onClose: () => void
  onSaved: (item: StockItemDTO) => void
}

export function StockItemFormModal({ existingItem, onClose, onSaved }: StockItemFormModalProps) {
  const { apiFetch } = useAuth()
  const isEditing = !!existingItem
  const [name, setName] = useState(existingItem?.name ?? '')
  const [unit, setUnit] = useState(existingItem?.unit ?? '')
  const [minQuantity, setMinQuantity] = useState(
    existingItem?.min_quantity != null ? String(existingItem.min_quantity) : '',
  )
  const [initialQuantity, setInitialQuantity] = useState('')
  const [isActive, setIsActive] = useState(existingItem?.is_active ?? true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!name.trim() || !unit.trim()) {
      setError('Preencha nome e unidade.')
      return
    }
    const minQty = minQuantity.trim() === '' ? null : Number(minQuantity)
    if (minQty !== null && (!Number.isInteger(minQty) || minQty < 0)) {
      setError('Quantidade mínima deve ser um número inteiro não-negativo.')
      return
    }
    const initialQty = initialQuantity.trim() === '' ? 0 : Number(initialQuantity)
    if (!Number.isInteger(initialQty) || initialQty < 0) {
      setError('Quantidade inicial deve ser um número inteiro não-negativo.')
      return
    }

    setSubmitting(true)
    try {
      const item = isEditing
        ? await apiFetch<StockItemDTO>(`/api/stock-items/${existingItem.id}`, {
            method: 'PUT',
            body: JSON.stringify({ name, unit, min_quantity: minQty, is_active: isActive }),
          })
        : await apiFetch<StockItemDTO>('/api/stock-items', {
            method: 'POST',
            body: JSON.stringify({ name, unit, min_quantity: minQty, initial_quantity: initialQty }),
          })
      onSaved(item)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível salvar o item.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={isEditing ? 'Editar Item' : 'Novo Item'} onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Nome</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Luva cirúrgica"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Unidade</label>
            <input
              value={unit}
              onChange={(e) => setUnit(e.target.value)}
              placeholder="un, cx, par, frasco…"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Estoque mínimo (opcional)</label>
            <input
              type="number"
              min={0}
              value={minQuantity}
              onChange={(e) => setMinQuantity(e.target.value)}
              placeholder="Sem alerta se vazio"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        </div>

        {!isEditing && (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Quantidade inicial (opcional)</label>
            <input
              type="number"
              min={0}
              value={initialQuantity}
              onChange={(e) => setInitialQuantity(e.target.value)}
              placeholder="0"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
            <p className="mt-1 text-xs text-brand-text-muted">
              Se você já tem esse item em estoque, informe a quantidade atual aqui — isso registra uma entrada
              inicial no histórico do item.
            </p>
          </div>
        )}

        {isEditing && (
          <label className="flex items-center gap-2 text-sm text-brand-text">
            <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
            Item ativo
          </label>
        )}

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : isEditing ? 'Salvar Alterações' : 'Criar Item'}
        </button>
      </form>
    </Modal>
  )
}

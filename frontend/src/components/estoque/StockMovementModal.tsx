import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { STOCK_MOVEMENT_TYPE_LABEL, type StockMovementType } from '../../lib/stock'
import type { StockItemDTO } from '../../lib/api/types'

interface StockMovementModalProps {
  item: StockItemDTO
  onClose: () => void
  onSaved: (item: StockItemDTO) => void
}

export function StockMovementModal({ item, onClose, onSaved }: StockMovementModalProps) {
  const { apiFetch } = useAuth()
  const [type, setType] = useState<StockMovementType>('IN')
  const [quantity, setQuantity] = useState('')
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    const qty = Number(quantity)
    if (!Number.isInteger(qty) || qty <= 0) {
      setError('Informe uma quantidade inteira maior que zero.')
      return
    }
    setSubmitting(true)
    try {
      const { item: updated } = await apiFetch<{ item: StockItemDTO }>(`/api/stock-items/${item.id}/movements`, {
        method: 'POST',
        body: JSON.stringify({ type, quantity: qty, note: note.trim() || null }),
      })
      onSaved(updated)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível registrar a movimentação.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={`Movimentar: ${item.name}`} onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <p className="text-sm text-brand-text-muted">
          Quantidade atual: <span className="font-semibold text-brand-text">{item.quantity_on_hand}</span> {item.unit}
        </p>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Tipo</label>
            <select
              value={type}
              onChange={(e) => setType(e.target.value as StockMovementType)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              {Object.entries(STOCK_MOVEMENT_TYPE_LABEL).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Quantidade</label>
            <input
              type="number"
              min={1}
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Nota (opcional)</label>
          <input
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Ex: compra do fornecedor X, uso no atendimento…"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : 'Registrar Movimentação'}
        </button>
      </form>
    </Modal>
  )
}

import { useEffect, useState } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { STOCK_MOVEMENT_TYPE_LABEL } from '../../lib/stock'
import { formatDateLong, formatTime } from '../../lib/format'
import type { StockItemDTO, StockMovementDTO, UserDTO } from '../../lib/api/types'

interface StockMovementHistoryModalProps {
  item: StockItemDTO
  onClose: () => void
}

export function StockMovementHistoryModal({ item, onClose }: StockMovementHistoryModalProps) {
  const { apiFetch } = useAuth()
  const [movements, setMovements] = useState<StockMovementDTO[]>([])
  const [userNames, setUserNames] = useState<Map<string, string>>(new Map())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    apiFetch<StockMovementDTO[]>(`/api/stock-items/${item.id}/movements`)
      .then(async (data) => {
        if (cancelled) return
        setMovements(data ?? [])
        const uniqueUserIds = [...new Set((data ?? []).map((m) => m.created_by_id))]
        const users = await Promise.all(
          uniqueUserIds.map((id) => apiFetch<UserDTO>(`/api/users/${id}`).catch(() => null)),
        )
        if (cancelled) return
        setUserNames(new Map(users.filter((u): u is UserDTO => u !== null).map((u) => [u.id, u.name])))
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar o histórico.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, item.id])

  return (
    <Modal title={`Histórico: ${item.name}`} onClose={onClose} maxWidthClassName="max-w-lg">
      {error && <p className="mb-3 text-sm text-brand-alert-text">{error}</p>}
      {loading && <p className="text-sm text-brand-text-muted">Carregando…</p>}
      {!loading && movements.length === 0 && (
        <p className="text-sm text-brand-text-muted">Nenhuma movimentação registrada ainda.</p>
      )}
      {movements.length > 0 && (
        <div className="max-h-96 overflow-y-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
              <tr>
                <th className="px-3 py-2 font-medium">Data</th>
                <th className="px-3 py-2 font-medium">Tipo</th>
                <th className="px-3 py-2 font-medium">Qtd.</th>
                <th className="px-3 py-2 font-medium">Por</th>
                <th className="px-3 py-2 font-medium">Nota</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {movements.map((m) => (
                <tr key={m.id}>
                  <td className="px-3 py-2 whitespace-nowrap text-brand-text-muted">
                    {formatDateLong(new Date(m.created_at))} · {formatTime(new Date(m.created_at))}
                  </td>
                  <td className="px-3 py-2">
                    <span
                      className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium ${
                        m.type === 'IN'
                          ? 'border-emerald-200 bg-emerald-50 text-brand-success-text'
                          : 'border-rose-200 bg-rose-50 text-brand-alert-text'
                      }`}
                    >
                      {STOCK_MOVEMENT_TYPE_LABEL[m.type]}
                    </span>
                  </td>
                  <td className="px-3 py-2 font-medium text-brand-text">{m.quantity}</td>
                  <td className="px-3 py-2 text-brand-text-muted">{userNames.get(m.created_by_id) ?? '—'}</td>
                  <td className="px-3 py-2 text-brand-text-muted">{m.note ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Modal>
  )
}

import { useEffect, useState } from 'react'
import { Plus, Upload } from 'lucide-react'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import { StockItemFormModal } from '../components/estoque/StockItemFormModal'
import { StockMovementModal } from '../components/estoque/StockMovementModal'
import { StockMovementHistoryModal } from '../components/estoque/StockMovementHistoryModal'
import { StockImportModal } from '../components/estoque/StockImportModal'
import type { StockItemDTO } from '../lib/api/types'

const CAN_MANAGE_STOCK_ROLES = new Set(['TENANT_ADMIN', 'RECEPTIONIST'])

function isLowStock(item: StockItemDTO): boolean {
  return item.min_quantity != null && item.quantity_on_hand <= item.min_quantity
}

export function EstoquePage() {
  const { apiFetch, user } = useAuth()
  const [items, setItems] = useState<StockItemDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [editingItem, setEditingItem] = useState<StockItemDTO | 'new' | null>(null)
  const [movingItem, setMovingItem] = useState<StockItemDTO | null>(null)
  const [historyItem, setHistoryItem] = useState<StockItemDTO | null>(null)
  const [importing, setImporting] = useState(false)

  const canManage = user ? CAN_MANAGE_STOCK_ROLES.has(user.role) : false

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    apiFetch<StockItemDTO[]>('/api/stock-items')
      .then((data) => {
        if (!cancelled) setItems(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar o estoque.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, reloadTick])

  return (
    <>
      <header className="flex items-center justify-between border-b border-slate-200 bg-brand-surface px-8 py-5">
        <div>
          <p className="text-sm text-brand-text-muted">Clínica</p>
          <h1 className="text-xl font-semibold text-brand-text">Estoque</h1>
        </div>
        {canManage && (
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setImporting(true)}
              className="flex items-center gap-2 rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium text-brand-text hover:bg-slate-50"
            >
              <Upload size={18} />
              Importar CSV
            </button>
            <button
              type="button"
              onClick={() => setEditingItem('new')}
              className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
            >
              <Plus size={18} />
              Novo Item
            </button>
          </div>
        )}
      </header>

      <main className="p-6">
        {error && (
          <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
            {error}
          </p>
        )}

        <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
          {!loading && items.length === 0 && (
            <p className="p-6 text-sm text-brand-text-muted">Nenhum item de estoque cadastrado.</p>
          )}
          {items.length > 0 && (
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
                <tr>
                  <th className="px-6 py-3 font-medium">Nome</th>
                  <th className="px-6 py-3 font-medium">Unidade</th>
                  <th className="px-6 py-3 font-medium">Quantidade</th>
                  <th className="px-6 py-3 font-medium">Situação</th>
                  <th className="px-6 py-3 font-medium" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {items.map((item) => (
                  <tr key={item.id}>
                    <td className="px-6 py-4 font-medium text-brand-text">{item.name}</td>
                    <td className="px-6 py-4 text-brand-text-muted">{item.unit}</td>
                    <td className="px-6 py-4 text-brand-text-muted">
                      {item.quantity_on_hand}
                      {item.min_quantity != null && (
                        <span className="ml-1 text-xs text-brand-text-muted">(mín. {item.min_quantity})</span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1.5">
                        {isLowStock(item) && (
                          <span className="inline-flex items-center rounded-full border border-rose-200 bg-rose-50 px-2.5 py-0.5 text-xs font-medium text-brand-alert-text">
                            Estoque baixo
                          </span>
                        )}
                        <span
                          className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${
                            item.is_active
                              ? 'border-emerald-200 bg-emerald-50 text-brand-success-text'
                              : 'border-slate-300 bg-slate-100 text-slate-600'
                          }`}
                        >
                          {item.is_active ? 'Ativo' : 'Inativo'}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => setHistoryItem(item)}
                          className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                        >
                          Histórico
                        </button>
                        {canManage && (
                          <>
                            <button
                              type="button"
                              onClick={() => setMovingItem(item)}
                              className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                            >
                              Movimentar
                            </button>
                            <button
                              type="button"
                              onClick={() => setEditingItem(item)}
                              className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                            >
                              Editar
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>

      {editingItem && (
        <StockItemFormModal
          existingItem={editingItem === 'new' ? undefined : editingItem}
          onClose={() => setEditingItem(null)}
          onSaved={() => {
            setEditingItem(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}

      {movingItem && (
        <StockMovementModal
          item={movingItem}
          onClose={() => setMovingItem(null)}
          onSaved={() => {
            setMovingItem(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}

      {historyItem && <StockMovementHistoryModal item={historyItem} onClose={() => setHistoryItem(null)} />}

      {importing && (
        <StockImportModal
          onClose={() => setImporting(false)}
          onImported={() => setReloadTick((t) => t + 1)}
        />
      )}
    </>
  )
}

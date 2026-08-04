import { useEffect, useState } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { TOOTH_CONDITION_LABEL, TOOTH_CONDITION_STYLE, type ToothCondition } from '../../lib/odontograma'
import { formatDateLong, formatTime } from '../../lib/format'
import type { OdontogramaEntryDTO, UserDTO } from '../../lib/api/types'

interface OdontogramaHistoryModalProps {
  patientId: string
  onClose: () => void
}

export function OdontogramaHistoryModal({ patientId, onClose }: OdontogramaHistoryModalProps) {
  const { apiFetch } = useAuth()
  const [entries, setEntries] = useState<OdontogramaEntryDTO[]>([])
  const [userNames, setUserNames] = useState<Map<string, string>>(new Map())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    apiFetch<OdontogramaEntryDTO[]>(`/api/patients/${patientId}/odontograma-entries`)
      .then(async (data) => {
        if (cancelled) return
        setEntries(data ?? [])
        const uniqueUserIds = [...new Set((data ?? []).map((e) => e.recorded_by_id))]
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
  }, [apiFetch, patientId])

  return (
    <Modal title="Histórico do Odontograma" onClose={onClose} maxWidthClassName="max-w-lg">
      {error && <p className="mb-3 text-sm text-brand-alert-text">{error}</p>}
      {loading && <p className="text-sm text-brand-text-muted">Carregando…</p>}
      {!loading && entries.length === 0 && (
        <p className="text-sm text-brand-text-muted">Nenhum achado registrado ainda.</p>
      )}
      {entries.length > 0 && (
        <div className="max-h-96 overflow-y-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
              <tr>
                <th className="px-3 py-2 font-medium">Data</th>
                <th className="px-3 py-2 font-medium">Dente</th>
                <th className="px-3 py-2 font-medium">Condição</th>
                <th className="px-3 py-2 font-medium">Por</th>
                <th className="px-3 py-2 font-medium">Nota</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {entries.map((e) => (
                <tr key={e.id}>
                  <td className="px-3 py-2 whitespace-nowrap text-brand-text-muted">
                    {formatDateLong(new Date(e.created_at))} · {formatTime(new Date(e.created_at))}
                  </td>
                  <td className="px-3 py-2 font-medium text-brand-text">{e.tooth_number}</td>
                  <td className="px-3 py-2">
                    <span
                      className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium ${
                        TOOTH_CONDITION_STYLE[e.condition as ToothCondition]
                      }`}
                    >
                      {TOOTH_CONDITION_LABEL[e.condition as ToothCondition] ?? e.condition}
                    </span>
                  </td>
                  <td className="px-3 py-2 text-brand-text-muted">{userNames.get(e.recorded_by_id) ?? '—'}</td>
                  <td className="px-3 py-2 text-brand-text-muted">{e.note ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Modal>
  )
}

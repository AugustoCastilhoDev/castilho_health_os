import { useEffect, useState } from 'react'
import { FilePlus2, Pill } from 'lucide-react'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { formatDateLong } from '../../lib/format'
import { IssuePrescriptionModal } from './IssuePrescriptionModal'
import type { MemedPrescriptionLogDTO, PatientDTO } from '../../lib/api/types'

interface MemedPrescriptionsPanelProps {
  patient: PatientDTO
}

const CAN_ISSUE_ROLES = new Set(['DOCTOR', 'DENTIST'])

// This panel only ever shows the audit trail (who/when) — the prescription
// content itself lives entirely in Memed, never here (see
// ROADMAP.md's "Backend nunca é dono do conteúdo da receita médica").
export function MemedPrescriptionsPanel({ patient }: MemedPrescriptionsPanelProps) {
  const { apiFetch, user } = useAuth()
  const [logs, setLogs] = useState<MemedPrescriptionLogDTO[]>([])
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [issuing, setIssuing] = useState(false)

  const canIssue = user ? CAN_ISSUE_ROLES.has(user.role) : false

  useEffect(() => {
    let cancelled = false
    apiFetch<MemedPrescriptionLogDTO[]>(`/api/patients/${patient.id}/memed-prescriptions`)
      .then((data) => {
        if (!cancelled) setLogs(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar as receitas.')
        }
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, patient.id, reloadTick])

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-brand-text-muted">Receitas (Memed)</h2>
        {canIssue && (
          <button
            type="button"
            onClick={() => setIssuing(true)}
            className="flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
          >
            <FilePlus2 size={14} />
            Emitir Receita
          </button>
        )}
      </div>

      {error && (
        <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
          {error}
        </p>
      )}

      <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
        {logs.length === 0 && !error && (
          <p className="p-6 text-sm text-brand-text-muted">Nenhuma receita emitida ainda.</p>
        )}
        {logs.length > 0 && (
          <ul className="divide-y divide-slate-100">
            {logs.map((log) => (
              <li key={log.id} className="flex items-center justify-between gap-4 px-6 py-4">
                <div className="flex items-center gap-3">
                  <Pill size={18} className="shrink-0 text-brand-text-muted" />
                  <p className="text-sm text-brand-text">{formatDateLong(new Date(log.issued_at))}</p>
                </div>
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    log.status === 'CANCELLED' ? 'bg-slate-100 text-brand-text-muted' : 'bg-emerald-50 text-emerald-700'
                  }`}
                >
                  {log.status === 'CANCELLED' ? 'Cancelada' : 'Emitida'}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>

      {issuing && (
        <IssuePrescriptionModal
          patient={patient}
          onClose={() => setIssuing(false)}
          onIssued={() => {
            setIssuing(false)
            setReloadTick((t) => t + 1)
          }}
        />
      )}
    </div>
  )
}

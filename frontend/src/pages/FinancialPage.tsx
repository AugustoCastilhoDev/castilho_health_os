import { useEffect, useState } from 'react'
import { CheckCircle2, Plus } from 'lucide-react'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import { formatCurrencyBRL } from '../lib/format'
import {
  TRANSACTION_STATUS_LABEL,
  TRANSACTION_STATUS_STYLE,
  TRANSACTION_TYPE_LABEL,
  type TransactionStatus,
  type TransactionType,
} from '../lib/financial'
import { StatCard } from '../components/dashboard/StatCard'
import { RegisterPaymentModal } from '../components/financial/RegisterPaymentModal'
import { FinancialRulesPanel } from '../components/financial/FinancialRulesPanel'
import { useProfessionalScope } from '../hooks/useProfessionalScope'
import type { FinancialTransactionDTO, FinancialTransactionListDTO, PatientDTO, UserDTO } from '../lib/api/types'

const PAGE_SIZE = 20
const CAN_MARK_PAID_ROLES = new Set(['TENANT_ADMIN', 'FINANCE'])

type Tab = 'ledger' | 'rules'

function LedgerPanel() {
  const { apiFetch, user } = useAuth()
  const { professionals } = useProfessionalScope()
  const [statusFilter, setStatusFilter] = useState<TransactionStatus | ''>('')
  const [typeFilter, setTypeFilter] = useState<TransactionType | ''>('')
  const [page, setPage] = useState(1)
  const [transactions, setTransactions] = useState<FinancialTransactionDTO[]>([])
  const [total, setTotal] = useState(0)
  const [patientNames, setPatientNames] = useState<Map<string, string>>(new Map())
  const [professionalNames, setProfessionalNames] = useState<Map<string, string>>(new Map())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showRegisterModal, setShowRegisterModal] = useState(false)
  const [reloadTick, setReloadTick] = useState(0)
  const [pendingReceivableCents, setPendingReceivableCents] = useState(0)
  const [pendingPayoutCents, setPendingPayoutCents] = useState(0)

  const canMarkPaid = user ? CAN_MARK_PAID_ROLES.has(user.role) : false

  useEffect(() => {
    let cancelled = false
    const params = new URLSearchParams({ page: String(page), page_size: String(PAGE_SIZE) })
    if (statusFilter) params.set('status', statusFilter)
    if (typeFilter) params.set('type', typeFilter)

    setLoading(true)
    apiFetch<FinancialTransactionListDTO>(`/api/financial-transactions?${params}`)
      .then(async (res) => {
        if (cancelled) return
        setTransactions(res.items)
        setTotal(res.total)

        const uniquePatientIds = [...new Set(res.items.map((t) => t.patient_id).filter((id): id is string => !!id))]
        const uniqueProfessionalIds = [
          ...new Set(res.items.map((t) => t.professional_id).filter((id): id is string => !!id)),
        ]
        const [patients, profs] = await Promise.all([
          Promise.all(uniquePatientIds.map((id) => apiFetch<PatientDTO>(`/api/patients/${id}`))),
          Promise.all(uniqueProfessionalIds.map((id) => apiFetch<UserDTO>(`/api/users/${id}`))),
        ])
        if (cancelled) return
        setPatientNames(new Map(patients.map((p) => [p.id, p.name])))
        setProfessionalNames(new Map(profs.map((p) => [p.id, p.name])))
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar os lançamentos.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [apiFetch, statusFilter, typeFilter, page, reloadTick])

  // No tenant-wide "sum of pending amounts" endpoint exists (same accepted
  // gap as the Dashboard's revenue total) — capped at 100 rows per side,
  // which covers the current scale without a new aggregate query.
  useEffect(() => {
    let cancelled = false
    Promise.all([
      apiFetch<FinancialTransactionListDTO>(
        '/api/financial-transactions?type=PATIENT_PAYMENT&status=PENDING&page_size=100',
      ),
      apiFetch<FinancialTransactionListDTO>(
        '/api/financial-transactions?type=PROFESSIONAL_PAYOUT&status=PENDING&page_size=100',
      ),
    ])
      .then(([receivable, payout]) => {
        if (cancelled) return
        setPendingReceivableCents(receivable.items.reduce((sum, t) => sum + t.gross_amount_cents, 0))
        setPendingPayoutCents(payout.items.reduce((sum, t) => sum + t.net_amount_cents, 0))
      })
      .catch(() => {
        /* best-effort summary cards; the table below is the source of truth */
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, reloadTick])

  async function handleMarkPaid(id: string) {
    setError(null)
    try {
      await apiFetch(`/api/financial-transactions/${id}/mark-paid`, { method: 'POST' })
      setReloadTick((t) => t + 1)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível marcar como pago.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <>
      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <StatCard label="A Receber (pendente)" value={formatCurrencyBRL(pendingReceivableCents)} icon={Plus} accent="action" />
        <StatCard
          label="A Repassar (pendente)"
          value={formatCurrencyBRL(pendingPayoutCents)}
          icon={CheckCircle2}
          accent="success"
        />
      </div>

      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap gap-2">
          <select
            value={statusFilter}
            onChange={(e) => {
              setPage(1)
              setStatusFilter(e.target.value as TransactionStatus | '')
            }}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            <option value="">Todas as situações</option>
            {Object.entries(TRANSACTION_STATUS_LABEL).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
          <select
            value={typeFilter}
            onChange={(e) => {
              setPage(1)
              setTypeFilter(e.target.value as TransactionType | '')
            }}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            <option value="">Recebimentos e repasses</option>
            {Object.entries(TRANSACTION_TYPE_LABEL).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>
        <button
          type="button"
          onClick={() => setShowRegisterModal(true)}
          className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
        >
          <Plus size={18} />
          Registrar Pagamento
        </button>
      </div>

      {error && (
        <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
          {error}
        </p>
      )}

      <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
        {!loading && transactions.length === 0 && (
          <p className="p-6 text-sm text-brand-text-muted">Nenhum lançamento encontrado.</p>
        )}
        {transactions.length > 0 && (
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
              <tr>
                <th className="px-6 py-3 font-medium">Data</th>
                <th className="px-6 py-3 font-medium">Quem</th>
                <th className="px-6 py-3 font-medium">Tipo</th>
                <th className="px-6 py-3 font-medium">Valor</th>
                <th className="px-6 py-3 font-medium">Situação</th>
                <th className="px-6 py-3 font-medium" />
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {transactions.map((tx) => {
                // A payout's PatientID is only carried over from its source
                // PATIENT_PAYMENT for traceability — the party actually
                // being paid is the professional, so that takes priority
                // here even though both fields may be set on the row.
                const who =
                  tx.type === 'PROFESSIONAL_PAYOUT'
                    ? (tx.professional_id && professionalNames.get(tx.professional_id)) ||
                      (tx.patient_id && patientNames.get(tx.patient_id)) ||
                      '—'
                    : (tx.patient_id && patientNames.get(tx.patient_id)) ||
                      (tx.professional_id && professionalNames.get(tx.professional_id)) ||
                      '—'
                return (
                  <tr key={tx.id}>
                    <td className="px-6 py-4 text-brand-text-muted">
                      {new Date(tx.created_at).toLocaleDateString('pt-BR')}
                    </td>
                    <td className="px-6 py-4 text-brand-text">{who}</td>
                    <td className="px-6 py-4 text-brand-text-muted">{TRANSACTION_TYPE_LABEL[tx.type as TransactionType] ?? tx.type}</td>
                    <td className="px-6 py-4 font-medium text-brand-text">
                      {formatCurrencyBRL(tx.type === 'PROFESSIONAL_PAYOUT' ? tx.net_amount_cents : tx.gross_amount_cents)}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${
                          TRANSACTION_STATUS_STYLE[tx.status as TransactionStatus]
                        }`}
                      >
                        {TRANSACTION_STATUS_LABEL[tx.status as TransactionStatus] ?? tx.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      {tx.status === 'PENDING' && canMarkPaid && (
                        <button
                          type="button"
                          onClick={() => handleMarkPaid(tx.id)}
                          className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                        >
                          Marcar como pago
                        </button>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </div>

      {total > PAGE_SIZE && (
        <div className="mt-4 flex items-center justify-between text-sm text-brand-text-muted">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
            className="rounded-lg border border-slate-300 px-3 py-1.5 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Anterior
          </button>
          <span>
            Página {page} de {totalPages}
          </span>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
            className="rounded-lg border border-slate-300 px-3 py-1.5 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Próxima
          </button>
        </div>
      )}

      {showRegisterModal && (
        <RegisterPaymentModal
          professionals={professionals}
          onClose={() => setShowRegisterModal(false)}
          onCreated={() => {
            setShowRegisterModal(false)
            setPage(1)
            setReloadTick((t) => t + 1)
          }}
        />
      )}
    </>
  )
}

export function FinancialPage() {
  const [tab, setTab] = useState<Tab>('ledger')

  return (
    <>
      <header className="border-b border-slate-200 bg-brand-surface px-8 py-5">
        <p className="text-sm text-brand-text-muted">Financeiro</p>
        <h1 className="text-xl font-semibold text-brand-text">Lançamentos e Repasses</h1>
        <div className="mt-4 flex gap-1">
          <button
            type="button"
            onClick={() => setTab('ledger')}
            className={`rounded-lg px-3 py-1.5 text-sm font-medium ${
              tab === 'ledger' ? 'bg-brand-action text-white' : 'text-brand-text-muted hover:bg-slate-100'
            }`}
          >
            Lançamentos
          </button>
          <button
            type="button"
            onClick={() => setTab('rules')}
            className={`rounded-lg px-3 py-1.5 text-sm font-medium ${
              tab === 'rules' ? 'bg-brand-action text-white' : 'text-brand-text-muted hover:bg-slate-100'
            }`}
          >
            Regras de Repasse
          </button>
        </div>
      </header>

      <main className="p-6">{tab === 'ledger' ? <LedgerPanel /> : <FinancialRulesPanel />}</main>
    </>
  )
}

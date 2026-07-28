import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { parseCurrencyToCents } from '../../lib/format'
import { PAYMENT_METHOD_LABEL, type PaymentMethod } from '../../lib/financial'
import type { FinancialTransactionDTO, PatientDTO, UserDTO } from '../../lib/api/types'

interface RegisterPaymentModalProps {
  professionals: UserDTO[]
  onClose: () => void
  onCreated: (tx: FinancialTransactionDTO) => void
}

export function RegisterPaymentModal({ professionals, onClose, onCreated }: RegisterPaymentModalProps) {
  const { apiFetch } = useAuth()
  const [patientQuery, setPatientQuery] = useState('')
  const [patientResults, setPatientResults] = useState<PatientDTO[]>([])
  const [selectedPatient, setSelectedPatient] = useState<PatientDTO | null>(null)
  const [professionalId, setProfessionalId] = useState('')
  const [amount, setAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('CASH')
  const [insurancePlan, setInsurancePlan] = useState('')
  const [notes, setNotes] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (selectedPatient || patientQuery.trim().length < 2) {
      setPatientResults([])
      return
    }
    let cancelled = false
    const handle = setTimeout(() => {
      apiFetch<PatientDTO[]>(`/api/patients?q=${encodeURIComponent(patientQuery)}&limit=6`)
        .then((data) => {
          if (!cancelled) setPatientResults(data ?? [])
        })
        .catch(() => {
          if (!cancelled) setPatientResults([])
        })
    }, 250)
    return () => {
      cancelled = true
      clearTimeout(handle)
    }
  }, [patientQuery, selectedPatient, apiFetch])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!selectedPatient) {
      setError('Selecione um paciente.')
      return
    }
    const grossAmountCents = parseCurrencyToCents(amount)
    if (grossAmountCents === null) {
      setError('Informe um valor válido.')
      return
    }
    setSubmitting(true)
    try {
      const tx = await apiFetch<FinancialTransactionDTO>('/api/financial-transactions/', {
        method: 'POST',
        body: JSON.stringify({
          type: 'PATIENT_PAYMENT',
          patient_id: selectedPatient.id,
          professional_id: professionalId || undefined,
          gross_amount_cents: grossAmountCents,
          fee_amount_cents: 0,
          payment_method: paymentMethod,
          insurance_plan: insurancePlan || undefined,
          notes,
        }),
      })
      onCreated(tx)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível registrar o pagamento.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="Registrar Pagamento" onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Paciente</label>
          {selectedPatient ? (
            <div className="flex items-center justify-between rounded-lg border border-slate-300 px-3 py-2 text-sm">
              <span className="text-brand-text">{selectedPatient.name}</span>
              <button
                type="button"
                onClick={() => {
                  setSelectedPatient(null)
                  setPatientQuery('')
                }}
                className="text-xs text-brand-action hover:underline"
              >
                trocar
              </button>
            </div>
          ) : (
            <div className="relative">
              <input
                value={patientQuery}
                onChange={(e) => setPatientQuery(e.target.value)}
                placeholder="Buscar paciente por nome…"
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
              />
              {patientResults.length > 0 && (
                <ul className="absolute z-10 mt-1 w-full rounded-lg border border-slate-200 bg-brand-surface shadow-sm">
                  {patientResults.map((p) => (
                    <li key={p.id}>
                      <button
                        type="button"
                        onClick={() => {
                          setSelectedPatient(p)
                          setPatientResults([])
                        }}
                        className="block w-full px-3 py-2 text-left text-sm hover:bg-slate-50"
                      >
                        {p.name}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </div>

        {professionals.length > 0 && (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Profissional</label>
            <select
              value={professionalId}
              onChange={(e) => setProfessionalId(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              <option value="">Sem profissional vinculado</option>
              {professionals.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Valor (R$)</label>
            <input
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="150,00"
              required
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Forma de pagamento</label>
            <select
              value={paymentMethod}
              onChange={(e) => setPaymentMethod(e.target.value as PaymentMethod)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              {Object.entries(PAYMENT_METHOD_LABEL).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Convênio (opcional)</label>
          <input
            value={insurancePlan}
            onChange={(e) => setInsurancePlan(e.target.value)}
            placeholder="Deixe em branco se particular"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Observações (opcional)</label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={2}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Registrando…' : 'Registrar Pagamento'}
        </button>
      </form>
    </Modal>
  )
}

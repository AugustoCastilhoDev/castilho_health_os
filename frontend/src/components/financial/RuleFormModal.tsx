import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { parseCurrencyToCents } from '../../lib/format'
import {
  FEE_DEDUCTION_LABEL,
  FINANCIAL_RULE_TYPE_LABEL,
  type FeeDeductionPolicy,
  type FinancialRuleType,
} from '../../lib/financial'
import type { FinancialRuleDTO } from '../../lib/api/types'

interface RuleFormModalProps {
  professionalId: string
  existingRule?: FinancialRuleDTO
  onClose: () => void
  onSaved: (rule: FinancialRuleDTO) => void
}

export function RuleFormModal({ professionalId, existingRule, onClose, onSaved }: RuleFormModalProps) {
  const { apiFetch } = useAuth()
  const isEditing = !!existingRule
  const [type, setType] = useState<FinancialRuleType>((existingRule?.type as FinancialRuleType) ?? 'PERCENTAGE')
  const [percentage, setPercentage] = useState(
    existingRule?.percentage != null ? String(existingRule.percentage * 100) : '',
  )
  const [fixedAmount, setFixedAmount] = useState(
    existingRule?.fixed_amount_cents != null ? (existingRule.fixed_amount_cents / 100).toFixed(2).replace('.', ',') : '',
  )
  const [procedureCode, setProcedureCode] = useState(existingRule?.procedure_code ?? '')
  const [insurancePlan, setInsurancePlan] = useState(existingRule?.insurance_plan ?? '')
  const [feeDeduction, setFeeDeduction] = useState<FeeDeductionPolicy>(
    (existingRule?.fee_deduction as FeeDeductionPolicy) ?? 'BEFORE_SPLIT',
  )
  const [priority, setPriority] = useState(existingRule?.priority ?? 0)
  const [isActive, setIsActive] = useState(existingRule?.is_active ?? true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    let percentageValue: number | null = null
    let fixedAmountCents: number | null = null
    if (type === 'PERCENTAGE') {
      const pct = Number(percentage.replace(',', '.'))
      if (!Number.isFinite(pct) || pct <= 0 || pct > 100) {
        setError('Informe um percentual entre 0 e 100.')
        return
      }
      percentageValue = pct / 100
    } else {
      fixedAmountCents = parseCurrencyToCents(fixedAmount)
      if (fixedAmountCents === null) {
        setError('Informe um valor fixo válido.')
        return
      }
    }

    setSubmitting(true)
    try {
      const body = {
        professional_id: professionalId,
        type,
        percentage: percentageValue,
        fixed_amount_cents: fixedAmountCents,
        procedure_code: procedureCode || undefined,
        insurance_plan: insurancePlan || undefined,
        fee_deduction: feeDeduction,
        priority,
        is_active: isActive,
      }
      const rule = isEditing
        ? await apiFetch<FinancialRuleDTO>(`/api/financial-rules/${existingRule.id}`, {
            method: 'PUT',
            body: JSON.stringify(body),
          })
        : await apiFetch<FinancialRuleDTO>('/api/financial-rules/', {
            method: 'POST',
            body: JSON.stringify(body),
          })
      onSaved(rule)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível salvar a regra.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={isEditing ? 'Editar Regra de Repasse' : 'Nova Regra de Repasse'} onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Tipo</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as FinancialRuleType)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            {Object.entries(FINANCIAL_RULE_TYPE_LABEL).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>

        {type === 'PERCENTAGE' ? (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Percentual (%)</label>
            <input
              value={percentage}
              onChange={(e) => setPercentage(e.target.value)}
              placeholder="70"
              required
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        ) : (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Valor fixo (R$)</label>
            <input
              value={fixedAmount}
              onChange={(e) => setFixedAmount(e.target.value)}
              placeholder="100,00"
              required
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        )}

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Procedimento (opcional)</label>
            <input
              value={procedureCode}
              onChange={(e) => setProcedureCode(e.target.value)}
              placeholder="Escopo geral se vazio"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Convênio (opcional)</label>
            <input
              value={insurancePlan}
              onChange={(e) => setInsurancePlan(e.target.value)}
              placeholder="Particular se vazio"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Dedução de taxa</label>
            <select
              value={feeDeduction}
              onChange={(e) => setFeeDeduction(e.target.value as FeeDeductionPolicy)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              {Object.entries(FEE_DEDUCTION_LABEL).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Prioridade</label>
            <input
              type="number"
              value={priority}
              onChange={(e) => setPriority(Number(e.target.value))}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        </div>

        {isEditing && (
          <label className="flex items-center gap-2 text-sm text-brand-text">
            <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
            Regra ativa
          </label>
        )}

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : isEditing ? 'Salvar Alterações' : 'Criar Regra'}
        </button>
      </form>
    </Modal>
  )
}

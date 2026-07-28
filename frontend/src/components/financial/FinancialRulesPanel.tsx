import { useEffect, useState } from 'react'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { formatCurrencyBRL } from '../../lib/format'
import { FEE_DEDUCTION_LABEL, FINANCIAL_RULE_TYPE_LABEL, type FeeDeductionPolicy, type FinancialRuleType } from '../../lib/financial'
import type { FinancialRuleDTO, UserDTO } from '../../lib/api/types'
import { useProfessionalScope } from '../../hooks/useProfessionalScope'

function ruleAmount(rule: FinancialRuleDTO): string {
  if (rule.type === 'PERCENTAGE' && rule.percentage != null) {
    return `${(rule.percentage * 100).toFixed(0)}%`
  }
  if (rule.fixed_amount_cents != null) {
    return formatCurrencyBRL(rule.fixed_amount_cents)
  }
  return '—'
}

export function FinancialRulesPanel() {
  const { apiFetch } = useAuth()
  const { professionals, professionalId, setProfessionalId, loading: loadingProfessionals } = useProfessionalScope()
  const [rules, setRules] = useState<FinancialRuleDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!professionalId) return
    let cancelled = false
    setLoading(true)
    apiFetch<FinancialRuleDTO[]>(`/api/financial-rules?professional_id=${professionalId}`)
      .then((data) => {
        if (!cancelled) setRules(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar as regras.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [professionalId, apiFetch])

  return (
    <div>
      {professionals.length > 1 && (
        <div className="mb-4 max-w-xs">
          <label className="mb-1 block text-sm font-medium text-brand-text">Profissional</label>
          <select
            value={professionalId ?? ''}
            onChange={(e) => setProfessionalId(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            {professionals.map((p: UserDTO) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
      )}

      {error && (
        <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
          {error}
        </p>
      )}

      <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
        {!loadingProfessionals && !loading && rules.length === 0 && (
          <p className="p-6 text-sm text-brand-text-muted">
            Nenhuma regra de repasse configurada para este profissional.
          </p>
        )}
        {rules.length > 0 && (
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
              <tr>
                <th className="px-6 py-3 font-medium">Tipo</th>
                <th className="px-6 py-3 font-medium">Valor</th>
                <th className="px-6 py-3 font-medium">Escopo</th>
                <th className="px-6 py-3 font-medium">Dedução de taxa</th>
                <th className="px-6 py-3 font-medium">Prioridade</th>
                <th className="px-6 py-3 font-medium">Situação</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {rules.map((rule) => (
                <tr key={rule.id}>
                  <td className="px-6 py-4 text-brand-text">{FINANCIAL_RULE_TYPE_LABEL[rule.type as FinancialRuleType] ?? rule.type}</td>
                  <td className="px-6 py-4 font-medium text-brand-text">{ruleAmount(rule)}</td>
                  <td className="px-6 py-4 text-brand-text-muted">
                    {rule.procedure_code ? `Procedimento: ${rule.procedure_code}` : null}
                    {rule.procedure_code && rule.insurance_plan ? ' · ' : null}
                    {rule.insurance_plan ? `Convênio: ${rule.insurance_plan}` : null}
                    {!rule.procedure_code && !rule.insurance_plan ? 'Geral (particular)' : null}
                  </td>
                  <td className="px-6 py-4 text-brand-text-muted">
                    {FEE_DEDUCTION_LABEL[rule.fee_deduction as FeeDeductionPolicy] ?? rule.fee_deduction}
                  </td>
                  <td className="px-6 py-4 text-brand-text-muted">{rule.priority}</td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${
                        rule.is_active
                          ? 'border-emerald-200 bg-emerald-50 text-brand-success-text'
                          : 'border-slate-300 bg-slate-100 text-slate-600'
                      }`}
                    >
                      {rule.is_active ? 'Ativa' : 'Inativa'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

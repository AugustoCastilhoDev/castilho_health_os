import { useEffect, useState } from 'react'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import {
  DECIDUOUS_LOWER,
  DECIDUOUS_UPPER,
  PERMANENT_LOWER,
  PERMANENT_UPPER,
  TOOTH_CONDITION_LABEL,
  TOOTH_CONDITION_STYLE,
  TOOTH_NO_ENTRY_STYLE,
  type ToothCondition,
} from '../../lib/odontograma'
import { OdontogramaEntryModal } from './OdontogramaEntryModal'
import { OdontogramaHistoryModal } from './OdontogramaHistoryModal'
import type { OdontogramaEntryDTO } from '../../lib/api/types'

const CAN_RECORD_ODONTOGRAMA_ROLES = new Set(['DENTIST'])

interface OdontogramaPanelProps {
  patientId: string
}

function ToothTile({
  number,
  condition,
  onClick,
}: {
  number: string
  condition?: ToothCondition
  onClick?: () => void
}) {
  const style = condition ? TOOTH_CONDITION_STYLE[condition] : TOOTH_NO_ENTRY_STYLE
  const label = condition ? TOOTH_CONDITION_LABEL[condition] : 'Sem registro'
  return (
    <button
      type="button"
      disabled={!onClick}
      onClick={onClick}
      title={`Dente ${number} — ${label}`}
      className={`flex h-10 items-center justify-center rounded-lg border text-xs font-semibold ${style} ${
        onClick ? 'cursor-pointer transition-colors hover:brightness-95' : 'cursor-default'
      }`}
    >
      {number}
    </button>
  )
}

function Arch({
  upper,
  lower,
  chartByTooth,
  onToothClick,
}: {
  upper: string[]
  lower: string[]
  chartByTooth: Map<string, ToothCondition>
  onToothClick?: (tooth: string) => void
}) {
  return (
    <div className="space-y-1.5">
      <div className="grid gap-1.5" style={{ gridTemplateColumns: `repeat(${upper.length}, minmax(0, 1fr))` }}>
        {upper.map((n) => (
          <ToothTile
            key={n}
            number={n}
            condition={chartByTooth.get(n)}
            onClick={onToothClick ? () => onToothClick(n) : undefined}
          />
        ))}
      </div>
      <div className="grid gap-1.5" style={{ gridTemplateColumns: `repeat(${lower.length}, minmax(0, 1fr))` }}>
        {lower.map((n) => (
          <ToothTile
            key={n}
            number={n}
            condition={chartByTooth.get(n)}
            onClick={onToothClick ? () => onToothClick(n) : undefined}
          />
        ))}
      </div>
    </div>
  )
}

export function OdontogramaPanel({ patientId }: OdontogramaPanelProps) {
  const { apiFetch, user } = useAuth()
  const [chart, setChart] = useState<OdontogramaEntryDTO[]>([])
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [selectedTooth, setSelectedTooth] = useState<string | null>(null)
  const [showHistory, setShowHistory] = useState(false)

  const canRecord = user ? CAN_RECORD_ODONTOGRAMA_ROLES.has(user.role) : false

  useEffect(() => {
    let cancelled = false
    apiFetch<OdontogramaEntryDTO[]>(`/api/patients/${patientId}/odontograma`)
      .then((data) => {
        if (!cancelled) setChart(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar o odontograma.')
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, patientId, reloadTick])

  const chartByTooth = new Map(chart.map((e) => [e.tooth_number, e.condition as ToothCondition]))

  return (
    <div className="rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-brand-text-muted">Odontograma</h2>
        <button
          type="button"
          onClick={() => setShowHistory(true)}
          className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
        >
          Histórico
        </button>
      </div>

      {error && <p className="mb-4 text-sm text-brand-alert-text">{error}</p>}

      <div className="space-y-6">
        <div>
          <p className="mb-2 text-xs font-medium text-brand-text-muted">Dentição permanente</p>
          <Arch
            upper={PERMANENT_UPPER}
            lower={PERMANENT_LOWER}
            chartByTooth={chartByTooth}
            onToothClick={canRecord ? setSelectedTooth : undefined}
          />
        </div>
        <div>
          <p className="mb-2 text-xs font-medium text-brand-text-muted">Dentição decídua</p>
          <Arch
            upper={DECIDUOUS_UPPER}
            lower={DECIDUOUS_LOWER}
            chartByTooth={chartByTooth}
            onToothClick={canRecord ? setSelectedTooth : undefined}
          />
        </div>
      </div>

      <div className="mt-5 flex flex-wrap gap-x-4 gap-y-1.5 border-t border-slate-100 pt-4 text-xs text-brand-text-muted">
        {Object.entries(TOOTH_CONDITION_LABEL).map(([value, label]) => (
          <span key={value} className="flex items-center gap-1.5">
            <span className={`h-3 w-3 rounded border ${TOOTH_CONDITION_STYLE[value as ToothCondition]}`} />
            {label}
          </span>
        ))}
      </div>

      {selectedTooth && (
        <OdontogramaEntryModal
          patientId={patientId}
          toothNumber={selectedTooth}
          onClose={() => setSelectedTooth(null)}
          onSaved={() => {
            setSelectedTooth(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}

      {showHistory && <OdontogramaHistoryModal patientId={patientId} onClose={() => setShowHistory(false)} />}
    </div>
  )
}

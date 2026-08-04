export type ToothCondition =
  | 'SAUDAVEL'
  | 'CARIE'
  | 'RESTAURADO'
  | 'AUSENTE'
  | 'CANAL'
  | 'COROA'
  | 'IMPLANTE'
  | 'FRATURADO'
  | 'A_EXTRAIR'

export const TOOTH_CONDITION_LABEL: Record<ToothCondition, string> = {
  SAUDAVEL: 'Saudável',
  CARIE: 'Cárie',
  RESTAURADO: 'Restaurado',
  AUSENTE: 'Ausente',
  CANAL: 'Canal (endodontia)',
  COROA: 'Coroa',
  IMPLANTE: 'Implante',
  FRATURADO: 'Fraturado',
  A_EXTRAIR: 'Indicação de extração',
}

// Border/background/text triplet per condition, same "badge" shape used
// elsewhere in the app (StockMovementHistoryModal, EstoquePage) — applied to
// each tooth tile in the chart grid.
export const TOOTH_CONDITION_STYLE: Record<ToothCondition, string> = {
  SAUDAVEL: 'border-emerald-300 bg-emerald-50 text-brand-success-text',
  CARIE: 'border-rose-300 bg-rose-50 text-brand-alert-text',
  RESTAURADO: 'border-sky-300 bg-sky-50 text-sky-700',
  AUSENTE: 'border-slate-300 bg-slate-200 text-slate-500 line-through',
  CANAL: 'border-amber-300 bg-amber-50 text-amber-700',
  COROA: 'border-violet-300 bg-violet-50 text-violet-700',
  IMPLANTE: 'border-indigo-300 bg-indigo-50 text-indigo-700',
  FRATURADO: 'border-red-300 bg-red-50 text-red-700',
  A_EXTRAIR: 'border-orange-300 bg-orange-50 text-orange-700',
}

// Tile style for a tooth with no entry yet — visually distinct from an
// explicit SAUDAVEL entry (which means "examined and confirmed healthy"),
// even though both read as "nothing wrong" at a glance.
export const TOOTH_NO_ENTRY_STYLE = 'border-slate-200 bg-white text-brand-text-muted'

// FDI notation, laid out the way a dental chart is conventionally drawn:
// quadrant 1 then 2 across the top (upper arch), quadrant 4 then 3 across
// the bottom (lower arch) — the patient's own right/left as seen by the
// professional facing them, not the professional's own right/left.
export const PERMANENT_UPPER = [
  '18', '17', '16', '15', '14', '13', '12', '11', '21', '22', '23', '24', '25', '26', '27', '28',
]
export const PERMANENT_LOWER = [
  '48', '47', '46', '45', '44', '43', '42', '41', '31', '32', '33', '34', '35', '36', '37', '38',
]
export const DECIDUOUS_UPPER = ['55', '54', '53', '52', '51', '61', '62', '63', '64', '65']
export const DECIDUOUS_LOWER = ['85', '84', '83', '82', '81', '71', '72', '73', '74', '75']

import type { ComponentType } from 'react'

interface StatCardProps {
  label: string
  value: string
  hint?: string
  icon: ComponentType<{ size?: number; className?: string }>
  accent?: 'action' | 'success'
}

const accentStyles = {
  action: 'bg-sky-50 text-brand-action',
  success: 'bg-emerald-50 text-brand-success-text',
} as const

export function StatCard({ label, value, hint, icon: Icon, accent = 'action' }: StatCardProps) {
  return (
    <div className="rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-brand-text-muted">{label}</p>
        <span className={`flex h-9 w-9 items-center justify-center rounded-lg ${accentStyles[accent]}`}>
          <Icon size={18} />
        </span>
      </div>
      <p className="mt-4 text-3xl font-semibold tracking-tight text-brand-text">{value}</p>
      {hint && <p className="mt-1 text-sm text-brand-text-muted">{hint}</p>}
    </div>
  )
}

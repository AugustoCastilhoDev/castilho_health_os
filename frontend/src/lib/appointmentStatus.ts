// Mirrors the backend's 7-state appointment lifecycle
// (internal/domain/models/appointment.go) so the frontend's status strings
// stay a direct match once this is wired to the real API.
export type AppointmentStatus =
  | 'SCHEDULED'
  | 'CONFIRMED'
  | 'WAITING'
  | 'IN_PROGRESS'
  | 'COMPLETED'
  | 'CANCELLED'
  | 'NO_SHOW'

export const APPOINTMENT_STATUS_LABEL: Record<AppointmentStatus, string> = {
  SCHEDULED: 'Agendado',
  CONFIRMED: 'Confirmado',
  WAITING: 'Aguardando',
  IN_PROGRESS: 'Em atendimento',
  COMPLETED: 'Concluído',
  CANCELLED: 'Cancelado',
  NO_SHOW: 'Não compareceu',
}

// Sky for anything still in the scheduling/action flow, emerald for the
// COMPLETED success state, rose for CANCELLED/NO_SHOW — the same three
// semantic families used everywhere else in the app.
export const APPOINTMENT_STATUS_STYLE: Record<AppointmentStatus, string> = {
  SCHEDULED: 'border-slate-300 bg-slate-100 text-slate-600',
  CONFIRMED: 'border-sky-200 bg-sky-50 text-brand-action',
  WAITING: 'border-sky-300 bg-sky-100 text-brand-action',
  IN_PROGRESS: 'border-brand-action bg-brand-action text-white',
  COMPLETED: 'border-emerald-200 bg-emerald-50 text-brand-success-text',
  CANCELLED: 'border-rose-200 bg-rose-50 text-brand-alert-text line-through opacity-75',
  NO_SHOW: 'border-rose-300 bg-rose-100 text-brand-alert-text',
}

export const APPOINTMENT_STATUS_DOT: Record<AppointmentStatus, string> = {
  SCHEDULED: 'bg-slate-400',
  CONFIRMED: 'bg-brand-action',
  WAITING: 'bg-brand-action',
  IN_PROGRESS: 'bg-brand-action',
  COMPLETED: 'bg-brand-success',
  CANCELLED: 'bg-brand-alert',
  NO_SHOW: 'bg-brand-alert',
}

// Mirrors validTransitions in internal/domain/models/appointment.go exactly
// — this is only used to decide which action buttons to show; the backend
// re-validates every transition regardless, so drift here just means a
// button that 400s, never an actual bypass of the state machine.
export const APPOINTMENT_NEXT_STATUSES: Record<AppointmentStatus, AppointmentStatus[]> = {
  SCHEDULED: ['CONFIRMED', 'CANCELLED', 'WAITING', 'NO_SHOW'],
  CONFIRMED: ['WAITING', 'CANCELLED', 'NO_SHOW'],
  WAITING: ['IN_PROGRESS', 'CANCELLED'],
  IN_PROGRESS: ['COMPLETED'],
  COMPLETED: [],
  CANCELLED: [],
  NO_SHOW: [],
}

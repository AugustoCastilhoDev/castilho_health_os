import { useState } from 'react'
import { Modal } from '../common/Modal'
import { StatusBadge } from '../common/StatusBadge'
import { formatDateLong, formatTime } from '../../lib/format'
import { APPOINTMENT_NEXT_STATUSES, APPOINTMENT_STATUS_LABEL, type AppointmentStatus } from '../../lib/appointmentStatus'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import type { AgendaAppointment } from './WeeklyAgendaGrid'

interface AppointmentDetailModalProps {
  appointment: AgendaAppointment
  onClose: () => void
  onUpdated: () => void
}

export function AppointmentDetailModal({ appointment, onClose, onUpdated }: AppointmentDetailModalProps) {
  const { apiFetch } = useAuth()
  const [pendingCancel, setPendingCancel] = useState(false)
  const [pendingComplete, setPendingComplete] = useState(false)
  const [reason, setReason] = useState('')
  const [cid, setCid] = useState(appointment.cid ?? '')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const nextStatuses = APPOINTMENT_NEXT_STATUSES[appointment.status]

  async function transition(to: AppointmentStatus, transitionReason = '', transitionCid = '') {
    setError(null)
    setSubmitting(true)
    try {
      await apiFetch(`/api/appointments/${appointment.id}/transition`, {
        method: 'POST',
        body: JSON.stringify({ to, reason: transitionReason, cid: transitionCid }),
      })
      onUpdated()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível atualizar o agendamento.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title="Agendamento" onClose={onClose}>
      <div className="space-y-4">
        <div>
          <p className="text-base font-semibold text-brand-text">{appointment.patientName}</p>
          <p className="text-sm text-brand-text-muted">
            {formatDateLong(appointment.scheduledAt)} · {formatTime(appointment.scheduledAt)} ·{' '}
            {appointment.durationMin} min
          </p>
        </div>

        <StatusBadge status={appointment.status} />

        {pendingCancel ? (
          <div className="space-y-2">
            <label className="block text-sm font-medium text-brand-text">Motivo do cancelamento</label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={2}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setPendingCancel(false)}
                className="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm text-brand-text hover:bg-slate-50"
              >
                Voltar
              </button>
              <button
                type="button"
                disabled={submitting || !reason.trim()}
                onClick={() => transition('CANCELLED', reason)}
                className="flex-1 rounded-lg bg-brand-alert px-3 py-2 text-sm font-semibold text-white hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Confirmar cancelamento
              </button>
            </div>
          </div>
        ) : pendingComplete ? (
          <div className="space-y-2">
            <label className="block text-sm font-medium text-brand-text">CID (opcional)</label>
            <input
              value={cid}
              onChange={(e) => setCid(e.target.value)}
              placeholder="Ex: F41.1 - Transtorno de ansiedade generalizada"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setPendingComplete(false)}
                className="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm text-brand-text hover:bg-slate-50"
              >
                Voltar
              </button>
              <button
                type="button"
                disabled={submitting}
                onClick={() => transition('COMPLETED', '', cid.trim())}
                className="flex-1 rounded-lg bg-brand-action px-3 py-2 text-sm font-semibold text-white hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
              >
                Confirmar conclusão
              </button>
            </div>
          </div>
        ) : (
          nextStatuses.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {nextStatuses.map((status) => (
                <button
                  key={status}
                  type="button"
                  disabled={submitting}
                  onClick={() => {
                    if (status === 'CANCELLED') setPendingCancel(true)
                    else if (status === 'COMPLETED') setPendingComplete(true)
                    else transition(status)
                  }}
                  className={
                    status === 'CANCELLED'
                      ? 'rounded-lg border border-rose-200 px-3 py-2 text-sm font-medium text-brand-alert-text hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-60'
                      : 'rounded-lg bg-brand-action px-3 py-2 text-sm font-medium text-white hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60'
                  }
                >
                  {APPOINTMENT_STATUS_LABEL[status]}
                </button>
              ))}
            </div>
          )
        )}

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}
      </div>
    </Modal>
  )
}

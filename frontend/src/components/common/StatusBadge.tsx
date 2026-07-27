import { APPOINTMENT_STATUS_LABEL, APPOINTMENT_STATUS_STYLE, type AppointmentStatus } from '../../lib/appointmentStatus'

export function StatusBadge({ status }: { status: AppointmentStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${APPOINTMENT_STATUS_STYLE[status]}`}
    >
      {APPOINTMENT_STATUS_LABEL[status]}
    </span>
  )
}

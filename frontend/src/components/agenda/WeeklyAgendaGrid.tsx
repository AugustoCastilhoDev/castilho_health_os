import { addDays, formatTime, formatWeekdayShort, isSameDay } from '../../lib/format'
import { APPOINTMENT_STATUS_LABEL, APPOINTMENT_STATUS_STYLE, type AppointmentStatus } from '../../lib/appointmentStatus'

export interface AgendaAppointment {
  id: string
  patientName: string
  scheduledAt: Date
  durationMin: number
  status: AppointmentStatus
}

interface WeeklyAgendaGridProps {
  weekStart: Date
  appointments: AgendaAppointment[]
  startHour?: number
  endHour?: number
  onAppointmentClick?: (appointment: AgendaAppointment) => void
}

const SLOT_MINUTES = 30
const SLOT_HEIGHT_PX = 44
const HEADER_HEIGHT_PX = 56

export function WeeklyAgendaGrid({
  weekStart,
  appointments,
  startHour = 8,
  endHour = 19,
  onAppointmentClick,
}: WeeklyAgendaGridProps) {
  const today = new Date()
  const days = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))
  const slots = ((endHour - startHour) * 60) / SLOT_MINUTES

  return (
    <div className="overflow-x-auto rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
      <div
        className="grid min-w-[840px]"
        style={{
          gridTemplateColumns: `64px repeat(7, minmax(0, 1fr))`,
          gridTemplateRows: `${HEADER_HEIGHT_PX}px repeat(${slots}, ${SLOT_HEIGHT_PX}px)`,
        }}
      >
        {/* Day headers */}
        <div
          className="sticky top-0 border-b border-slate-200 bg-brand-surface"
          style={{ gridColumn: 1, gridRow: 1 }}
        />
        {days.map((day, i) => {
          const todayCol = isSameDay(day, today)
          return (
            <div
              key={i}
              style={{ gridColumn: i + 2, gridRow: 1 }}
              className={`flex flex-col items-center justify-center gap-0.5 border-b border-l border-slate-200 py-2 ${
                todayCol ? 'bg-sky-50' : 'bg-brand-surface'
              }`}
            >
              <span className="text-xs font-medium uppercase tracking-wide text-brand-text-muted">
                {formatWeekdayShort(day)}
              </span>
              <span
                className={`flex h-7 w-7 items-center justify-center rounded-full text-sm font-semibold ${
                  todayCol ? 'bg-brand-action text-white' : 'text-brand-text'
                }`}
              >
                {day.getDate()}
              </span>
            </div>
          )
        })}

        {/* Time-of-day labels */}
        {Array.from({ length: slots }, (_, s) => {
          const minutesFromStart = s * SLOT_MINUTES
          const hour = startHour + Math.floor(minutesFromStart / 60)
          const isHourMark = minutesFromStart % 60 === 0
          return (
            <div
              key={`time-${s}`}
              style={{ gridColumn: 1, gridRow: s + 2 }}
              className="border-t border-slate-100 pr-2 text-right"
            >
              {isHourMark && (
                <span className="relative -top-2 inline-block text-xs text-brand-text-muted">
                  {String(hour).padStart(2, '0')}:00
                </span>
              )}
            </div>
          )
        })}

        {/* Background grid cells */}
        {days.map((day, dayIdx) =>
          Array.from({ length: slots }, (_, s) => (
            <div
              key={`cell-${dayIdx}-${s}`}
              style={{ gridColumn: dayIdx + 2, gridRow: s + 2 }}
              className={`border-l border-t border-slate-100 ${
                isSameDay(day, today) ? 'bg-sky-50/40' : ''
              }`}
            />
          )),
        )}

        {/* Appointments */}
        {appointments.map((appt) => {
          const dayIdx = days.findIndex((d) => isSameDay(d, appt.scheduledAt))
          if (dayIdx === -1) return null

          const minutesFromStart =
            appt.scheduledAt.getHours() * 60 + appt.scheduledAt.getMinutes() - startHour * 60
          if (minutesFromStart < 0 || minutesFromStart >= slots * SLOT_MINUTES) return null

          const rowStart = Math.floor(minutesFromStart / SLOT_MINUTES) + 2
          const rowSpan = Math.max(1, Math.round(appt.durationMin / SLOT_MINUTES))

          return (
            <button
              key={appt.id}
              type="button"
              onClick={() => onAppointmentClick?.(appt)}
              style={{ gridColumn: dayIdx + 2, gridRow: `${rowStart} / span ${rowSpan}` }}
              className={`z-10 m-0.5 overflow-hidden rounded-md border px-2 py-1 text-left text-xs leading-tight transition-shadow hover:shadow-md ${APPOINTMENT_STATUS_STYLE[appt.status]}`}
            >
              <p className="truncate font-semibold">
                {formatTime(appt.scheduledAt)} · {appt.patientName}
              </p>
              <p className="truncate opacity-80">{APPOINTMENT_STATUS_LABEL[appt.status]}</p>
            </button>
          )
        })}
      </div>
    </div>
  )
}

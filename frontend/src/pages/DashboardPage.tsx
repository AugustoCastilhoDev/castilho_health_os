import { useEffect, useState } from 'react'
import { Calendar, Clock, Wallet } from 'lucide-react'
import { DashboardHeader } from '../components/dashboard/DashboardHeader'
import { StatCard } from '../components/dashboard/StatCard'
import { useAuth } from '../lib/auth/AuthContext'
import { useProfessionalScope } from '../hooks/useProfessionalScope'
import { ApiError } from '../lib/api/client'
import type { AppointmentDTO, FinancialTransactionDTO } from '../lib/api/types'
import { formatCurrencyBRL, endOfDay, startOfDay } from '../lib/format'
import { roleLabel } from '../lib/roles'

interface Summary {
  appointmentsToday: number
  patientsWaiting: number
  revenueTodayCents: number
}

export function DashboardPage() {
  const { user, apiFetch } = useAuth()
  const { professionalId } = useProfessionalScope()
  const [summary, setSummary] = useState<Summary | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!professionalId) return
    let cancelled = false

    async function load() {
      setError(null)
      try {
        const now = new Date()
        const from = startOfDay(now).toISOString()
        const to = endOfDay(now).toISOString()
        const appointments = await apiFetch<AppointmentDTO[]>(
          `/api/appointments?professional_id=${professionalId}&from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
        )

        // No tenant-wide "revenue today" endpoint exists yet — sum each
        // appointment's paid PATIENT_PAYMENT transactions individually.
        // Fine at today's scale; a dedicated aggregate endpoint would be the
        // real fix once appointment volume grows.
        const txLists = await Promise.all(
          appointments.map((a) =>
            apiFetch<FinancialTransactionDTO[]>(`/api/financial-transactions/appointment/${a.id}`),
          ),
        )
        const revenueTodayCents = txLists
          .flat()
          .filter((tx) => tx.type === 'PATIENT_PAYMENT' && tx.status === 'PAID')
          .reduce((sum, tx) => sum + tx.net_amount_cents, 0)

        if (!cancelled) {
          setSummary({
            appointmentsToday: appointments.length,
            patientsWaiting: appointments.filter((a) => a.status === 'WAITING').length,
            revenueTodayCents,
          })
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar o dashboard.')
        }
      }
    }

    load()
    return () => {
      cancelled = true
    }
  }, [professionalId, apiFetch])

  return (
    <>
      <DashboardHeader
        professionalName={user?.name ?? '—'}
        professionalRole={user ? roleLabel(user.role) : '—'}
      />

      <main className="p-6">
        {error && (
          <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
            {error}
          </p>
        )}

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          <StatCard
            label="Consultas Hoje"
            value={summary ? String(summary.appointmentsToday) : '—'}
            hint="Agendamentos de hoje"
            icon={Calendar}
            accent="action"
          />
          <StatCard
            label="Pacientes na Recepção"
            value={summary ? String(summary.patientsWaiting) : '—'}
            hint="Aguardando atendimento"
            icon={Clock}
            accent="action"
          />
          <StatCard
            label="Faturamento do Dia"
            value={summary ? formatCurrencyBRL(summary.revenueTodayCents) : '—'}
            hint="Consultas + procedimentos pagos"
            icon={Wallet}
            accent="success"
          />
        </div>
      </main>
    </>
  )
}

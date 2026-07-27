import { Calendar, Clock, Wallet } from 'lucide-react'
import { DashboardHeader } from '../components/dashboard/DashboardHeader'
import { StatCard } from '../components/dashboard/StatCard'
import { formatCurrencyBRL } from '../lib/format'
import { mockSession } from '../lib/mockSession'

// Placeholder until wired to GET /api/appointments, /api/patients and
// /api/financial-transactions — shape mirrors what those endpoints already
// return today.
const mockSummary = {
  appointmentsToday: 12,
  appointmentsRemaining: 5,
  patientsWaiting: 3,
  revenueTodayCents: 348000,
}

export function DashboardPage() {
  return (
    <>
      <DashboardHeader
        professionalName={mockSession.professionalName}
        professionalRole={mockSession.professionalRole}
      />

      <main className="p-6">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          <StatCard
            label="Consultas Hoje"
            value={String(mockSummary.appointmentsToday)}
            hint={`${mockSummary.appointmentsRemaining} restantes`}
            icon={Calendar}
            accent="action"
          />
          <StatCard
            label="Pacientes na Recepção"
            value={String(mockSummary.patientsWaiting)}
            hint="Aguardando atendimento"
            icon={Clock}
            accent="action"
          />
          <StatCard
            label="Faturamento do Dia"
            value={formatCurrencyBRL(mockSummary.revenueTodayCents)}
            hint="Consultas + procedimentos pagos"
            icon={Wallet}
            accent="success"
          />
        </div>
      </main>
    </>
  )
}

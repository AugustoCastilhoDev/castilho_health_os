import { Calendar, Clock, Wallet } from 'lucide-react'
import { Sidebar } from './Sidebar'
import { DashboardHeader } from './DashboardHeader'
import { StatCard } from './StatCard'
import { formatCurrencyBRL } from '../../lib/format'

// Placeholder until the dashboard is wired to GET /api/appointments,
// /api/patients and /api/financial-transactions — shape mirrors what those
// endpoints already return today.
const mockSummary = {
  clinicName: 'Clínica Vida Plena',
  professionalName: 'Dra. Ana Souza',
  professionalRole: 'Clínico Geral',
  appointmentsToday: 12,
  appointmentsRemaining: 5,
  patientsWaiting: 3,
  revenueTodayCents: 348000,
}

export function Dashboard() {
  return (
    <div className="flex h-screen bg-brand-bg">
      <Sidebar clinicName={mockSummary.clinicName} />

      <div className="flex flex-1 flex-col overflow-y-auto">
        <DashboardHeader
          professionalName={mockSummary.professionalName}
          professionalRole={mockSummary.professionalRole}
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
      </div>
    </div>
  )
}

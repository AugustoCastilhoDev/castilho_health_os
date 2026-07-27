import { Cake, IdCard, Mail, Phone, Plus } from 'lucide-react'
import { PatientTimeline, type TimelineEntry } from '../components/prontuario/PatientTimeline'
import { calculateAge, initials } from '../lib/format'

// Placeholder until wired to GET /api/patients/:id and the appointment/
// encounter history endpoints — field names mirror models.Patient.
const mockPatient = {
  name: 'Beatriz Lima',
  document: '123.456.789-00',
  birthDate: new Date(1990, 4, 12),
  phone: '(11) 98877-6655',
  email: 'beatriz.lima@example.com',
  insurancePlan: 'Particular',
}

const mockTimeline: TimelineEntry[] = [
  {
    id: '1',
    date: new Date(2026, 6, 22),
    type: 'RETORNO',
    title: 'Retorno pós-procedimento',
    professionalName: 'Dra. Ana Souza',
    status: 'COMPLETED',
    notes:
      'Paciente relata melhora significativa dos sintomas. Sem intercorrências. Mantida a conduta anterior.',
  },
  {
    id: '2',
    date: new Date(2026, 6, 8),
    type: 'EXAME',
    title: 'Exames laboratoriais de rotina',
    professionalName: 'Dra. Ana Souza',
    status: 'COMPLETED',
    notes: 'Hemograma e glicemia dentro dos parâmetros normais. Solicitado repetir em 6 meses.',
  },
  {
    id: '3',
    date: new Date(2026, 5, 15),
    type: 'CONSULTA',
    title: 'Consulta de rotina',
    professionalName: 'Dra. Ana Souza',
    status: 'COMPLETED',
    notes:
      'Queixa principal: dor lombar leve. Prescrito anti-inflamatório por 5 dias e orientações posturais.',
  },
  {
    id: '4',
    date: new Date(2026, 4, 30),
    type: 'PROCEDIMENTO',
    title: 'Consulta agendada',
    professionalName: 'Dra. Ana Souza',
    status: 'NO_SHOW',
    notes: 'Paciente não compareceu e não justificou a ausência.',
  },
]

export function PatientRecordPage() {
  const age = calculateAge(mockPatient.birthDate)

  return (
    <>
      <header className="border-b border-slate-200 bg-brand-surface px-8 py-5">
        <p className="text-sm text-brand-text-muted">Prontuário Eletrônico</p>
        <h1 className="text-xl font-semibold text-brand-text">{mockPatient.name}</h1>
      </header>

      <main className="space-y-6 p-6">
        <div className="flex flex-wrap items-center justify-between gap-4 rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200">
          <div className="flex items-center gap-4">
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-slate-200 text-lg font-semibold text-brand-text">
              {initials(mockPatient.name)}
            </div>
            <div>
              <p className="text-base font-semibold text-brand-text">{mockPatient.name}</p>
              <p className="text-sm text-brand-text-muted">{mockPatient.insurancePlan}</p>
            </div>
          </div>

          <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-brand-text-muted">
            <span className="flex items-center gap-1.5">
              <Cake size={16} />
              {age} anos
            </span>
            <span className="flex items-center gap-1.5">
              <IdCard size={16} />
              {mockPatient.document}
            </span>
            <span className="flex items-center gap-1.5">
              <Phone size={16} />
              {mockPatient.phone}
            </span>
            <span className="flex items-center gap-1.5">
              <Mail size={16} />
              {mockPatient.email}
            </span>
          </div>

          <button
            type="button"
            className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
          >
            <Plus size={18} />
            Novo Registro
          </button>
        </div>

        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-brand-text-muted">
            Histórico
          </h2>
          <PatientTimeline entries={mockTimeline} />
        </div>
      </main>
    </>
  )
}

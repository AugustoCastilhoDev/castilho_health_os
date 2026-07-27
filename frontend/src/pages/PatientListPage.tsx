import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Search } from 'lucide-react'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import type { PatientDTO } from '../lib/api/types'

export function PatientListPage() {
  const { apiFetch } = useAuth()
  const [query, setQuery] = useState('')
  const [patients, setPatients] = useState<PatientDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    const handle = setTimeout(() => {
      apiFetch<PatientDTO[]>(`/api/patients?q=${encodeURIComponent(query)}&limit=50`)
        .then((data) => {
          if (!cancelled) setPatients(data ?? [])
        })
        .catch((err) => {
          if (!cancelled) {
            setError(err instanceof ApiError ? err.message : 'Não foi possível carregar os pacientes.')
          }
        })
        .finally(() => {
          if (!cancelled) setLoading(false)
        })
    }, 250)

    return () => {
      cancelled = true
      clearTimeout(handle)
    }
  }, [query, apiFetch])

  return (
    <>
      <header className="border-b border-slate-200 bg-brand-surface px-8 py-5">
        <p className="text-sm text-brand-text-muted">Pacientes</p>
        <h1 className="text-xl font-semibold text-brand-text">Buscar paciente</h1>
      </header>

      <main className="p-6">
        <div className="relative mb-4 max-w-md">
          <Search size={18} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-brand-text-muted" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Buscar por nome, documento ou telefone…"
            className="w-full rounded-lg border border-slate-300 py-2 pl-10 pr-3 text-sm text-brand-text focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        {error && (
          <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
            {error}
          </p>
        )}

        <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
          {!loading && patients.length === 0 && (
            <p className="p-6 text-sm text-brand-text-muted">Nenhum paciente encontrado.</p>
          )}
          <ul className="divide-y divide-slate-100">
            {patients.map((patient) => (
              <li key={patient.id}>
                <Link
                  to={`/pacientes/${patient.id}`}
                  className="flex items-center justify-between px-6 py-4 text-sm hover:bg-slate-50"
                >
                  <span className="font-medium text-brand-text">{patient.name}</span>
                  <span className="text-brand-text-muted">{patient.phone || patient.email || '—'}</span>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </main>
    </>
  )
}

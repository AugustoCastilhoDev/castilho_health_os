import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { TextField } from '../components/common/TextField'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import type { PatientDTO } from '../lib/api/types'

interface PatientFormState {
  name: string
  document: string
  birthDate: string
  phone: string
  email: string
  addressZip: string
  addressStreet: string
  addressCity: string
  addressState: string
}

const emptyForm: PatientFormState = {
  name: '',
  document: '',
  birthDate: '',
  phone: '',
  email: '',
  addressZip: '',
  addressStreet: '',
  addressCity: '',
  addressState: '',
}

function toFormState(patient: PatientDTO): PatientFormState {
  return {
    name: patient.name,
    document: patient.document ?? '',
    birthDate: patient.birth_date ? patient.birth_date.slice(0, 10) : '',
    phone: patient.phone ?? '',
    email: patient.email ?? '',
    addressZip: patient.address_zip ?? '',
    addressStreet: patient.address_street ?? '',
    addressCity: patient.address_city ?? '',
    addressState: patient.address_state ?? '',
  }
}

export function PatientFormPage() {
  const { patientId } = useParams<{ patientId: string }>()
  const isEditing = Boolean(patientId)
  const navigate = useNavigate()
  const { apiFetch } = useAuth()
  const [form, setForm] = useState<PatientFormState>(emptyForm)
  const [loading, setLoading] = useState(isEditing)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!patientId) return
    let cancelled = false
    apiFetch<PatientDTO>(`/api/patients/${patientId}`)
      .then((data) => {
        if (!cancelled) setForm(toFormState(data))
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar o paciente.')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [patientId, apiFetch])

  function set<K extends keyof PatientFormState>(key: K, value: string) {
    setForm((f) => ({ ...f, [key]: value }))
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    const body = {
      name: form.name,
      document: form.document,
      birth_date: form.birthDate || null,
      phone: form.phone,
      email: form.email,
      address_zip: form.addressZip,
      address_street: form.addressStreet,
      address_city: form.addressCity,
      address_state: form.addressState,
    }
    try {
      const patient = await apiFetch<PatientDTO>(isEditing ? `/api/patients/${patientId}` : '/api/patients/', {
        method: isEditing ? 'PUT' : 'POST',
        body: JSON.stringify(body),
      })
      navigate(`/pacientes/${patient.id}`, { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível salvar o paciente.')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <main className="p-6 text-sm text-brand-text-muted">Carregando…</main>

  return (
    <>
      <header className="border-b border-slate-200 bg-brand-surface px-8 py-5">
        <p className="text-sm text-brand-text-muted">Pacientes</p>
        <h1 className="text-xl font-semibold text-brand-text">
          {isEditing ? 'Editar paciente' : 'Novo paciente'}
        </h1>
      </header>

      <main className="p-6">
        <form
          onSubmit={handleSubmit}
          className="max-w-2xl space-y-6 rounded-xl bg-brand-surface p-6 shadow-sm ring-1 ring-slate-200"
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextField label="Nome" value={form.name} onChange={(v) => set('name', v)} required />
            <TextField label="CPF" value={form.document} onChange={(v) => set('document', v)} />
            <TextField
              label="Data de nascimento"
              type="date"
              value={form.birthDate}
              onChange={(v) => set('birthDate', v)}
            />
            <TextField label="Telefone" value={form.phone} onChange={(v) => set('phone', v)} />
            <TextField label="E-mail" type="email" value={form.email} onChange={(v) => set('email', v)} />
          </div>

          <div>
            <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-brand-text-muted">
              Endereço
            </h2>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <TextField label="CEP" value={form.addressZip} onChange={(v) => set('addressZip', v)} />
              <div className="sm:col-span-2">
                <TextField label="Rua" value={form.addressStreet} onChange={(v) => set('addressStreet', v)} />
              </div>
              <TextField label="Cidade" value={form.addressCity} onChange={(v) => set('addressCity', v)} />
            </div>
            <div className="mt-4 max-w-[120px]">
              <TextField label="UF" value={form.addressState} onChange={(v) => set('addressState', v)} />
            </div>
          </div>

          {error && <p className="text-sm text-brand-alert-text">{error}</p>}

          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => navigate(-1)}
              className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium text-brand-text hover:bg-slate-50"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
            >
              {submitting ? 'Salvando…' : 'Salvar'}
            </button>
          </div>
        </form>
      </main>
    </>
  )
}

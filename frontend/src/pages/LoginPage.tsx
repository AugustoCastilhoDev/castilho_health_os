import { useState, type FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { Activity } from 'lucide-react'
import { TextField } from '../components/common/TextField'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'

export function LoginPage() {
  const { login, token } = useAuth()
  const navigate = useNavigate()
  const [tenantSlug, setTenantSlug] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  if (token) return <Navigate to="/" replace />

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login(tenantSlug, email, password)
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível conectar à API.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-brand-bg px-4">
      <div className="w-full max-w-sm rounded-xl bg-brand-surface p-8 shadow-sm ring-1 ring-slate-200">
        <div className="mb-6 flex items-center justify-center gap-2">
          <Activity className="text-brand-action" size={28} />
          <span className="text-lg font-semibold text-brand-text">Castilho Health OS</span>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <TextField label="Clínica" value={tenantSlug} onChange={setTenantSlug} placeholder="minha-clinica" required />
          <TextField label="E-mail" type="email" value={email} onChange={setEmail} required />
          <TextField label="Senha" type="password" value={password} onChange={setPassword} required />

          {error && <p className="text-sm text-brand-alert-text">{error}</p>}

          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
          >
            {submitting ? 'Entrando…' : 'Entrar'}
          </button>
        </form>
      </div>
    </div>
  )
}

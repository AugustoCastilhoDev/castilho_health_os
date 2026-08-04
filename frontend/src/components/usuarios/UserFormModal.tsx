import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { DEFAULT_COUNCIL_TYPE_BY_ROLE, HEALTH_PROFESSIONAL_ROLES, ROLE_LABEL } from '../../lib/roles'
import type { UserDTO } from '../../lib/api/types'

interface UserFormModalProps {
  existingUser?: UserDTO
  onClose: () => void
  onSaved: (user: UserDTO) => void
}

function isHealthProfessionalRole(role: string): boolean {
  return HEALTH_PROFESSIONAL_ROLES.has(role)
}

export function UserFormModal({ existingUser, onClose, onSaved }: UserFormModalProps) {
  const { apiFetch } = useAuth()
  const isEditing = !!existingUser
  const [name, setName] = useState(existingUser?.name ?? '')
  const [email, setEmail] = useState(existingUser?.email ?? '')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState(existingUser?.role ?? 'RECEPTIONIST')
  const [isActive, setIsActive] = useState(existingUser?.is_active ?? true)
  const [councilType, setCouncilType] = useState(existingUser?.council_type ?? 'CRM')
  const [councilNumber, setCouncilNumber] = useState(existingUser?.council_number ?? '')
  const [councilState, setCouncilState] = useState(existingUser?.council_state ?? '')
  const [cpf, setCpf] = useState(existingUser?.cpf ?? '')
  const [birthDate, setBirthDate] = useState(existingUser?.birth_date ? existingUser.birth_date.slice(0, 10) : '')
  const [sex, setSex] = useState(existingUser?.sex ?? '')
  const [phone, setPhone] = useState(existingUser?.phone ?? '')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const isProfessional = isHealthProfessionalRole(role)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!name.trim() || !email.trim()) {
      setError('Preencha nome e e-mail.')
      return
    }
    if (!isEditing && password.length < 8) {
      setError('A senha deve ter ao menos 8 caracteres.')
      return
    }

    setSubmitting(true)
    try {
      const professionalFields = isProfessional
        ? {
            council_type: councilType || null,
            council_number: councilNumber || null,
            council_state: councilState || null,
            cpf: cpf || null,
            birth_date: birthDate || null,
            sex: sex || null,
            phone: phone || null,
          }
        : {
            council_type: null,
            council_number: null,
            council_state: null,
            cpf: null,
            birth_date: null,
            sex: null,
            phone: null,
          }

      const user = isEditing
        ? await apiFetch<UserDTO>(`/api/users/${existingUser.id}`, {
            method: 'PUT',
            body: JSON.stringify({ name, email, role, is_active: isActive, ...professionalFields }),
          })
        : await apiFetch<UserDTO>('/api/users', {
            method: 'POST',
            body: JSON.stringify({ name, email, password, role, ...professionalFields }),
          })
      onSaved(user)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível salvar o usuário.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={isEditing ? 'Editar Usuário' : 'Novo Usuário'} onClose={onClose} maxWidthClassName="max-w-lg">
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Nome</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">E-mail</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Papel</label>
            <select
              value={role}
              onChange={(e) => {
                const nextRole = e.target.value
                setRole(nextRole)
                // Pre-fill (not lock) the council type for the newly picked
                // role — psiquiatra defaults to CRM (same as médico),
                // psicólogo to CRP; still freely editable below.
                if (DEFAULT_COUNCIL_TYPE_BY_ROLE[nextRole]) {
                  setCouncilType(DEFAULT_COUNCIL_TYPE_BY_ROLE[nextRole])
                }
              }}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            >
              {Object.entries(ROLE_LABEL).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
        </div>

        {!isEditing && (
          <div>
            <label className="mb-1 block text-sm font-medium text-brand-text">Senha</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Mínimo 8 caracteres"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
            />
          </div>
        )}

        {isProfessional && (
          <div className="space-y-3 rounded-lg border border-slate-200 p-3">
            <p className="text-xs font-semibold uppercase tracking-wide text-brand-text-muted">
              Dados do profissional (registro no conselho e, quando aplicável, emissão de receita digital)
            </p>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">Conselho</label>
                <select
                  value={councilType}
                  onChange={(e) => setCouncilType(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                >
                  <option value="CRM">CRM</option>
                  <option value="CRO">CRO</option>
                  <option value="CRP">CRP</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">Número</label>
                <input
                  value={councilNumber}
                  onChange={(e) => setCouncilNumber(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">UF</label>
                <input
                  value={councilState}
                  onChange={(e) => setCouncilState(e.target.value.toUpperCase())}
                  maxLength={2}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm uppercase focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">CPF</label>
                <input
                  value={cpf}
                  onChange={(e) => setCpf(e.target.value.replace(/\D/g, '').slice(0, 11))}
                  placeholder="Somente números"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">Data de nascimento</label>
                <input
                  type="date"
                  value={birthDate}
                  onChange={(e) => setBirthDate(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">Sexo</label>
                <select
                  value={sex}
                  onChange={(e) => setSex(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                >
                  <option value="">Não informado</option>
                  <option value="M">Masculino</option>
                  <option value="F">Feminino</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-brand-text">Telefone</label>
                <input
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
                />
              </div>
            </div>
          </div>
        )}

        {isEditing && (
          <label className="flex items-center gap-2 text-sm text-brand-text">
            <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
            Usuário ativo
          </label>
        )}

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : isEditing ? 'Salvar Alterações' : 'Criar Usuário'}
        </button>
      </form>
    </Modal>
  )
}

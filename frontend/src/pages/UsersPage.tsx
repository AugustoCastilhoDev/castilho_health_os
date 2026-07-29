import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { useAuth } from '../lib/auth/AuthContext'
import { ApiError } from '../lib/api/client'
import { roleLabel } from '../lib/roles'
import { UserFormModal } from '../components/usuarios/UserFormModal'
import { ResetPasswordModal } from '../components/usuarios/ResetPasswordModal'
import type { UserDTO } from '../lib/api/types'

const CAN_MANAGE_USERS_ROLES = new Set(['TENANT_ADMIN'])

export function UsersPage() {
  const { apiFetch, user: currentUser } = useAuth()
  const canManage = currentUser ? CAN_MANAGE_USERS_ROLES.has(currentUser.role) : false
  const [users, setUsers] = useState<UserDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [editingUser, setEditingUser] = useState<UserDTO | 'new' | null>(null)
  const [resettingUser, setResettingUser] = useState<UserDTO | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    apiFetch<UserDTO[]>('/api/users')
      .then((data) => {
        if (!cancelled) setUsers(data ?? [])
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Não foi possível carregar os usuários.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [apiFetch, reloadTick])

  async function toggleActive(target: UserDTO) {
    setError(null)
    try {
      await apiFetch<UserDTO>(`/api/users/${target.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: target.name,
          email: target.email,
          role: target.role,
          is_active: !target.is_active,
          council_type: target.council_type ?? null,
          council_number: target.council_number ?? null,
          council_state: target.council_state ?? null,
          cpf: target.cpf ?? null,
          birth_date: target.birth_date ? target.birth_date.slice(0, 10) : null,
          sex: target.sex ?? null,
          phone: target.phone ?? null,
        }),
      })
      setReloadTick((t) => t + 1)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível atualizar o usuário.')
    }
  }

  return (
    <>
      <header className="flex items-center justify-between border-b border-slate-200 bg-brand-surface px-8 py-5">
        <div>
          <p className="text-sm text-brand-text-muted">Administração</p>
          <h1 className="text-xl font-semibold text-brand-text">Usuários</h1>
        </div>
        {canManage && (
          <button
            type="button"
            onClick={() => setEditingUser('new')}
            className="flex items-center gap-2 rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover"
          >
            <Plus size={18} />
            Novo Usuário
          </button>
        )}
      </header>

      <main className="p-6">
        {error && (
          <p className="mb-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-brand-alert-text">
            {error}
          </p>
        )}

        <div className="overflow-hidden rounded-xl bg-brand-surface shadow-sm ring-1 ring-slate-200">
          {!loading && users.length === 0 && (
            <p className="p-6 text-sm text-brand-text-muted">Nenhum usuário cadastrado.</p>
          )}
          {users.length > 0 && (
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-xs uppercase tracking-wide text-brand-text-muted">
                <tr>
                  <th className="px-6 py-3 font-medium">Nome</th>
                  <th className="px-6 py-3 font-medium">E-mail</th>
                  <th className="px-6 py-3 font-medium">Papel</th>
                  <th className="px-6 py-3 font-medium">Situação</th>
                  {canManage && <th className="px-6 py-3 font-medium" />}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {users.map((u) => (
                  <tr key={u.id}>
                    <td className="px-6 py-4 font-medium text-brand-text">{u.name}</td>
                    <td className="px-6 py-4 text-brand-text-muted">{u.email}</td>
                    <td className="px-6 py-4 text-brand-text-muted">{roleLabel(u.role)}</td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium ${
                          u.is_active
                            ? 'border-emerald-200 bg-emerald-50 text-brand-success-text'
                            : 'border-slate-300 bg-slate-100 text-slate-600'
                        }`}
                      >
                        {u.is_active ? 'Ativo' : 'Inativo'}
                      </span>
                    </td>
                    {canManage && (
                      <td className="px-6 py-4 text-right">
                        <div className="flex justify-end gap-2">
                          <button
                            type="button"
                            onClick={() => setEditingUser(u)}
                            className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                          >
                            Editar
                          </button>
                          <button
                            type="button"
                            onClick={() => setResettingUser(u)}
                            className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50"
                          >
                            Redefinir senha
                          </button>
                          <button
                            type="button"
                            onClick={() => toggleActive(u)}
                            disabled={u.id === currentUser?.id}
                            title={u.id === currentUser?.id ? 'Não é possível desativar o próprio usuário' : undefined}
                            className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-brand-text hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
                          >
                            {u.is_active ? 'Desativar' : 'Ativar'}
                          </button>
                        </div>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>

      {editingUser && (
        <UserFormModal
          existingUser={editingUser === 'new' ? undefined : editingUser}
          onClose={() => setEditingUser(null)}
          onSaved={() => {
            setEditingUser(null)
            setReloadTick((t) => t + 1)
          }}
        />
      )}

      {resettingUser && (
        <ResetPasswordModal
          targetUser={resettingUser}
          onClose={() => setResettingUser(null)}
          onDone={() => setResettingUser(null)}
        />
      )}
    </>
  )
}

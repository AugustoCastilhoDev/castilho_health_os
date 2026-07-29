import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import type { UserDTO } from '../../lib/api/types'

interface ResetPasswordModalProps {
  targetUser: UserDTO
  onClose: () => void
  onDone: () => void
}

export function ResetPasswordModal({ targetUser, onClose, onDone }: ResetPasswordModalProps) {
  const { apiFetch } = useAuth()
  const [newPassword, setNewPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (newPassword.length < 8) {
      setError('A senha deve ter ao menos 8 caracteres.')
      return
    }
    setSubmitting(true)
    try {
      await apiFetch(`/api/users/${targetUser.id}/reset-password`, {
        method: 'POST',
        body: JSON.stringify({ new_password: newPassword }),
      })
      onDone()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível redefinir a senha.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={`Redefinir senha de ${targetUser.name}`} onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Nova senha</label>
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Mínimo 8 caracteres"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : 'Redefinir Senha'}
        </button>
      </form>
    </Modal>
  )
}

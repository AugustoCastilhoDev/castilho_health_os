import { useState, type FormEvent } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import { TOOTH_CONDITION_LABEL, type ToothCondition } from '../../lib/odontograma'
import type { OdontogramaEntryDTO } from '../../lib/api/types'

interface OdontogramaEntryModalProps {
  patientId: string
  toothNumber: string
  onClose: () => void
  onSaved: (entry: OdontogramaEntryDTO) => void
}

export function OdontogramaEntryModal({ patientId, toothNumber, onClose, onSaved }: OdontogramaEntryModalProps) {
  const { apiFetch } = useAuth()
  const [condition, setCondition] = useState<ToothCondition>('CARIE')
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const entry = await apiFetch<OdontogramaEntryDTO>('/api/odontograma-entries', {
        method: 'POST',
        body: JSON.stringify({
          patient_id: patientId,
          tooth_number: toothNumber,
          condition,
          note: note.trim() || null,
        }),
      })
      onSaved(entry)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Não foi possível registrar o achado.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={`Dente ${toothNumber}`} onClose={onClose}>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Condição</label>
          <select
            value={condition}
            onChange={(e) => setCondition(e.target.value as ToothCondition)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          >
            {Object.entries(TOOTH_CONDITION_LABEL).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-brand-text">Nota (opcional)</label>
          <input
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Ex: dor à percussão, tratamento agendado…"
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-action focus:outline-none focus:ring-1 focus:ring-brand-action"
          />
        </div>

        {error && <p className="text-sm text-brand-alert-text">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-lg bg-brand-action px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-action-hover disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Salvando…' : 'Registrar'}
        </button>
      </form>
    </Modal>
  )
}

import { useEffect, useRef, useState } from 'react'
import { Modal } from '../common/Modal'
import { useAuth } from '../../lib/auth/AuthContext'
import { ApiError } from '../../lib/api/client'
import {
  loadMemedScript,
  removeMemedScript,
  type MemedModuleInitEvent,
  type MemedPrescriptionIssuedEvent,
} from '../../lib/memed'
import type { PatientDTO } from '../../lib/api/types'

interface IssuePrescriptionModalProps {
  patient: PatientDTO
  onClose: () => void
  onIssued: () => void
}

type Status = 'loading-token' | 'loading-widget' | 'ready' | 'error'

// IssuePrescriptionModal loads Memed's own widget (this app never renders
// prescription UI itself — see ROADMAP.md's "Backend nunca é dono do
// conteúdo da receita médica"). The widget takes over as a full-screen
// overlay once `newPrescription` is sent; closing this modal also tears the
// script down so a stale token/module never lingers across patients.
export function IssuePrescriptionModal({ patient, onClose, onIssued }: IssuePrescriptionModalProps) {
  const { apiFetch } = useAuth()
  const [status, setStatus] = useState<Status>('loading-token')
  const [error, setError] = useState<string | null>(null)
  const loggedRef = useRef(false)

  useEffect(() => {
    let cancelled = false

    async function start() {
      setStatus('loading-token')
      setError(null)
      try {
        const { token, script_url } = await apiFetch<{ token: string; script_url: string }>('/api/memed/token')
        if (cancelled) return
        setStatus('loading-widget')

        const mdHub = await loadMemedScript(script_url, token)
        if (cancelled) return

        mdHub.event.add('core:moduleInit', (payload) => {
          const moduleData = payload as MemedModuleInitEvent
          if (moduleData?.name !== 'plataforma.prescricao') return

          mdHub.command.send('plataforma.prescricao', 'setPaciente', {
            nome: patient.name,
            idExterno: patient.id,
            cpf: patient.document,
            telefone: patient.phone,
            email: patient.email,
          })

          mdHub.event.add('prescricaoImpressa', (payload) => {
            const data = payload as MemedPrescriptionIssuedEvent
            const memedPrescriptionId = String(data.id ?? data.prescriptionUuid ?? '')
            if (!memedPrescriptionId || loggedRef.current) return
            loggedRef.current = true
            apiFetch(`/api/patients/${patient.id}/memed-prescriptions`, {
              method: 'POST',
              body: JSON.stringify({ memed_prescription_id: memedPrescriptionId }),
            })
              .then(() => onIssued())
              .catch(() => {
                /* the prescription was already issued on Memed's side by this
                   point — a failed audit-log write shouldn't be presented to
                   the professional as a failed prescription */
              })
          })

          mdHub.command.send('plataforma.prescricao', 'newPrescription')
        })

        if (!cancelled) setStatus('ready')
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : 'Não foi possível carregar a Memed.')
          setStatus('error')
        }
      }
    }

    start()
    return () => {
      cancelled = true
      removeMemedScript()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [patient.id])

  return (
    <Modal title="Emitir Receita (Memed)" onClose={onClose}>
      <div className="space-y-4 text-sm text-brand-text-muted">
        {status === 'loading-token' && <p>Carregando credenciais…</p>}
        {status === 'loading-widget' && <p>Carregando o módulo de prescrição da Memed…</p>}
        {status === 'ready' && (
          <p>
            O módulo da Memed foi aberto para <strong className="text-brand-text">{patient.name}</strong>. Finalize a
            prescrição na janela da Memed — o registro de auditoria é salvo automaticamente aqui assim que ela for
            emitida.
          </p>
        )}
        {error && <p className="text-brand-alert-text">{error}</p>}
      </div>
    </Modal>
  )
}

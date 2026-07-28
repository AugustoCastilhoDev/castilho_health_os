// Thin wrapper around Memed's Sinapse Prescrição widget script. Memed's own
// script (loaded with a per-professional `data-token`, see
// IssuePrescriptionModal) injects `window.MdHub` as a side effect — there is
// no npm package, so this file just types the surface we actually call.
//
// Script loading itself was live-verified against Memed's public sandbox
// (window.MdHub does appear after the fix below). The event/command names
// used downstream (`core:moduleInit`, `prescricaoImpressa`, `setPaciente`,
// `newPrescription`) still come from Memed's public docs, not a session that
// got as far as actually clicking through to issue a prescription — see
// ROADMAP.md for what's left to verify.

export interface MemedModuleInitEvent {
  name: string
}

export interface MemedPrescriptionIssuedEvent {
  id?: string | number
  prescriptionUuid?: string
  [key: string]: unknown
}

interface MdHubGlobal {
  event: {
    add: (eventName: string, callback: (payload: unknown) => void) => void
  }
  command: {
    send: (module: string, command: string, payload?: unknown) => void
  }
  // Confirmed live: modules initialize (core:moduleInit fires) but stay
  // display:none until this is called explicitly — Memed's own "módulo não
  // carrega" troubleshooting doc calls this out as the fix for "only an
  // autofocusing block message displays".
  module: {
    show: (module: string) => void
  }
}

declare global {
  interface Window {
    MdHub?: MdHubGlobal
  }
}

const SCRIPT_ID = 'memed-sinapse-prescricao-script'
const MDHUB_POLL_INTERVAL_MS = 100
const MDHUB_POLL_TIMEOUT_MS = 15000

// window.MdHub isn't set synchronously by the script's `load` event — Memed's
// script keeps initializing asynchronously afterwards (confirmed live: the
// console logs "O serviço de métricas foi iniciado..."/"O módulo
// plataforma.sdk está em execução" appear after `load` fires), so checking
// window.MdHub once at `onload` reports it missing even though it shows up a
// moment later. Poll instead of trusting the load event.
function waitForMdHub(): Promise<MdHubGlobal> {
  return new Promise((resolve, reject) => {
    const start = Date.now()
    const check = () => {
      if (window.MdHub) {
        resolve(window.MdHub)
        return
      }
      if (Date.now() - start > MDHUB_POLL_TIMEOUT_MS) {
        reject(new Error('Memed script loaded but window.MdHub never appeared.'))
        return
      }
      setTimeout(check, MDHUB_POLL_INTERVAL_MS)
    }
    check()
  })
}

// Loads (or reuses) Memed's widget script for the given professional token.
export function loadMemedScript(scriptUrl: string, token: string): Promise<MdHubGlobal> {
  const existing = document.getElementById(SCRIPT_ID) as HTMLScriptElement | null
  if (existing && existing.dataset.token === token) {
    return waitForMdHub()
  }
  // A stale script (different professional/token) must be removed first —
  // Memed's script isn't designed to be loaded twice with different tokens.
  existing?.remove()

  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.id = SCRIPT_ID
    script.src = scriptUrl
    script.dataset.token = token
    script.async = true
    script.onload = () => {
      waitForMdHub().then(resolve, reject)
    }
    script.onerror = () => reject(new Error('Falha ao carregar o script da Memed.'))
    document.body.appendChild(script)
  })
}

export function removeMemedScript() {
  document.getElementById(SCRIPT_ID)?.remove()
}

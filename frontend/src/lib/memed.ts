// Thin wrapper around Memed's Sinapse Prescrição widget script. Memed's own
// script (loaded with a per-professional `data-token`, see
// IssuePrescriptionModal) injects `window.MdHub` as a side effect — there is
// no npm package, so this file just types the surface we actually call.
//
// NOTE: the exact event/command names below come from Memed's public docs
// (doc.memed.com.br) rather than a live sandbox session (no API key/secret
// pair was available yet at the time this was written — see ROADMAP.md).
// Re-verify this file end-to-end the first time real credentials are wired
// in, the same way the R2 upload flow was live-verified before shipping.

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
}

declare global {
  interface Window {
    MdHub?: MdHubGlobal
  }
}

const SCRIPT_ID = 'memed-sinapse-prescricao-script'

// Loads (or reuses) Memed's widget script for the given professional token.
// Resolves once window.MdHub exists — Memed's script attaches it
// synchronously with the `load` event, per their docs.
export function loadMemedScript(scriptUrl: string, token: string): Promise<MdHubGlobal> {
  return new Promise((resolve, reject) => {
    const existing = document.getElementById(SCRIPT_ID) as HTMLScriptElement | null
    if (existing && existing.dataset.token === token && window.MdHub) {
      resolve(window.MdHub)
      return
    }
    // A stale script (different professional/token) must be removed first —
    // Memed's script isn't designed to be loaded twice with different tokens.
    existing?.remove()

    const script = document.createElement('script')
    script.id = SCRIPT_ID
    script.src = scriptUrl
    script.dataset.token = token
    script.async = true
    script.onload = () => {
      if (window.MdHub) resolve(window.MdHub)
      else reject(new Error('Memed script loaded but window.MdHub was not found.'))
    }
    script.onerror = () => reject(new Error('Falha ao carregar o script da Memed.'))
    document.body.appendChild(script)
  })
}

export function removeMemedScript() {
  document.getElementById(SCRIPT_ID)?.remove()
}

import { X } from 'lucide-react'
import type { ReactNode } from 'react'

interface ModalProps {
  title: string
  onClose: () => void
  children: ReactNode
  maxWidthClassName?: string
}

export function Modal({ title, onClose, children, maxWidthClassName = 'max-w-md' }: ModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 px-4">
      <div className={`w-full ${maxWidthClassName} max-h-[90vh] overflow-y-auto rounded-xl bg-brand-surface p-6 shadow-lg`}>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-brand-text">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1 text-brand-text-muted hover:bg-slate-100 hover:text-brand-text"
            aria-label="Fechar"
          >
            <X size={20} />
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}

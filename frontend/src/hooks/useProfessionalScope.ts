import { useEffect, useState } from 'react'
import { useAuth } from '../lib/auth/AuthContext'
import type { UserDTO } from '../lib/api/types'

// The API only lists appointments "by professional" (there's no tenant-wide
// "every appointment today" endpoint yet). A logged-in DOCTOR/DENTIST scopes
// to themselves; anyone else (admin/receptionist/finance) needs to pick
// which professional's agenda to view, same as a real front-desk workflow.
export function useProfessionalScope() {
  const { user, apiFetch } = useAuth()
  const [professionals, setProfessionals] = useState<UserDTO[]>([])
  const [professionalId, setProfessionalId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!user) return

    if (user.role === 'DOCTOR' || user.role === 'DENTIST') {
      setProfessionals([user])
      setProfessionalId(user.id)
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    Promise.all([
      apiFetch<UserDTO[]>('/api/users?role=DOCTOR'),
      apiFetch<UserDTO[]>('/api/users?role=DENTIST'),
    ])
      .then(([doctors, dentists]) => {
        if (cancelled) return
        const all = [...doctors, ...dentists]
        setProfessionals(all)
        setProfessionalId((current) => current ?? all[0]?.id ?? null)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [user, apiFetch])

  return { professionals, professionalId, setProfessionalId, loading }
}

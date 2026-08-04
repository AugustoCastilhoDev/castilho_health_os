import { useEffect, useState } from 'react'
import { useAuth } from '../lib/auth/AuthContext'
import { HEALTH_PROFESSIONAL_ROLES } from '../lib/roles'
import type { UserDTO } from '../lib/api/types'

// The API only lists appointments "by professional" (there's no tenant-wide
// "every appointment today" endpoint yet). A logged-in health professional
// (DOCTOR/DENTIST/PSYCHOLOGIST/PSYCHIATRIST) scopes to themselves; anyone
// else (admin/receptionist/finance) needs to pick which professional's
// agenda to view, same as a real front-desk workflow.
export function useProfessionalScope() {
  const { user, apiFetch } = useAuth()
  const [professionals, setProfessionals] = useState<UserDTO[]>([])
  const [professionalId, setProfessionalId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!user) return

    if (HEALTH_PROFESSIONAL_ROLES.has(user.role)) {
      setProfessionals([user])
      setProfessionalId(user.id)
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    Promise.all([...HEALTH_PROFESSIONAL_ROLES].map((role) => apiFetch<UserDTO[]>(`/api/users?role=${role}`)))
      .then((lists) => {
        if (cancelled) return
        const all = lists.flat()
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

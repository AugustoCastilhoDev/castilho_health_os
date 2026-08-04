import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { apiRequest } from '../api/client'
import type { TenantDTO, UserDTO } from '../api/types'
import { decodeJwtClaims, isExpired, type JwtClaims } from '../jwt'

const TOKEN_STORAGE_KEY = 'chos_token'

interface AuthContextValue {
  token: string | null
  claims: JwtClaims | null
  user: UserDTO | null
  tenant: TenantDTO | null
  loading: boolean
  login: (tenantSlug: string, email: string, password: string) => Promise<void>
  logout: () => void
  apiFetch: <T>(path: string, init?: RequestInit) => Promise<T>
  refreshTenant: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_STORAGE_KEY))
  const [claims, setClaims] = useState<JwtClaims | null>(null)
  const [user, setUser] = useState<UserDTO | null>(null)
  const [tenant, setTenant] = useState<TenantDTO | null>(null)
  const [loading, setLoading] = useState(true)

  const apiFetch = useCallback(
    <T,>(path: string, init: RequestInit = {}) => apiRequest<T>(path, token, init),
    [token],
  )

  const clearSession = useCallback(() => {
    localStorage.removeItem(TOKEN_STORAGE_KEY)
    setToken(null)
    setClaims(null)
    setUser(null)
    setTenant(null)
  }, [])

  useEffect(() => {
    if (!token) {
      setLoading(false)
      return
    }
    let parsedClaims: JwtClaims
    try {
      parsedClaims = decodeJwtClaims(token)
    } catch {
      clearSession()
      setLoading(false)
      return
    }
    if (isExpired(parsedClaims)) {
      clearSession()
      setLoading(false)
      return
    }
    setClaims(parsedClaims)

    let cancelled = false
    setLoading(true)
    Promise.all([
      apiRequest<UserDTO>(`/api/users/${parsedClaims.user_id}`, token),
      apiRequest<TenantDTO>('/api/tenant', token),
    ])
      .then(([userData, tenantData]) => {
        if (cancelled) return
        setUser(userData)
        setTenant(tenantData)
      })
      .catch(() => {
        if (cancelled) return
        clearSession()
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  const login = useCallback(async (tenantSlug: string, email: string, password: string) => {
    const res = await apiRequest<{ access_token: string }>('/auth/login', null, {
      method: 'POST',
      body: JSON.stringify({ tenant_slug: tenantSlug, email, password }),
    })
    localStorage.setItem(TOKEN_STORAGE_KEY, res.access_token)
    setToken(res.access_token)
  }, [])

  const logout = useCallback(() => {
    clearSession()
  }, [clearSession])

  // Lets a settings screen pull the freshly-saved tenant (name, logo, ...)
  // back into context immediately — e.g. so the sidebar's clinic name
  // updates without requiring a full reload/re-login.
  const refreshTenant = useCallback(async () => {
    if (!token) return
    const tenantData = await apiRequest<TenantDTO>('/api/tenant', token)
    setTenant(tenantData)
  }, [token])

  const value = useMemo(
    () => ({ token, claims, user, tenant, loading, login, logout, apiFetch, refreshTenant }),
    [token, claims, user, tenant, loading, login, logout, apiFetch, refreshTenant],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider')
  return ctx
}

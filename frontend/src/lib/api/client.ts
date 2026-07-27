const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

// Thin fetch wrapper: JSON in/out, Bearer auth, and the backend's
// `{"error": "..."}` shape surfaced as a typed ApiError instead of a bare
// non-2xx Response the caller has to unwrap by hand.
export async function apiRequest<T>(
  path: string,
  token: string | null,
  init: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers,
    },
  })

  if (res.status === 204) return undefined as T

  const text = await res.text()
  const data = text ? JSON.parse(text) : null

  if (!res.ok) {
    const message =
      data && typeof data === 'object' && 'error' in data ? String(data.error) : `Erro ${res.status}`
    throw new ApiError(res.status, message)
  }

  return data as T
}

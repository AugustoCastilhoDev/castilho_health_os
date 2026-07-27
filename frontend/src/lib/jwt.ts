export interface JwtClaims {
  user_id: string
  tenant_id: string
  role: string
  exp: number
}

// Decodes the JWT payload for display purposes only (which tenant/role to
// show in the UI) — this never verifies the signature. The backend is the
// only party that needs to trust these claims; every request still carries
// the raw token and gets re-validated server-side.
export function decodeJwtClaims(token: string): JwtClaims {
  const [, payload] = token.split('.')
  let base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
  while (base64.length % 4) base64 += '='
  return JSON.parse(atob(base64)) as JwtClaims
}

export function isExpired(claims: JwtClaims): boolean {
  return Date.now() / 1000 >= claims.exp
}

// Placeholder for the authenticated session until login/auth is wired to
// the real API (JWT claims already carry this on the backend — see
// internal/auth/jwt.go — this just mirrors the shape for now).
export const mockSession = {
  clinicName: 'Clínica Vida Plena',
  professionalName: 'Dra. Ana Souza',
  professionalRole: 'Clínico Geral',
}

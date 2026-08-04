export const ROLE_LABEL: Record<string, string> = {
  TENANT_ADMIN: 'Administrador(a)',
  DOCTOR: 'Médico(a)',
  DENTIST: 'Dentista',
  PSYCHOLOGIST: 'Psicólogo(a)',
  PSYCHIATRIST: 'Psiquiatra',
  RECEPTIONIST: 'Recepção',
  FINANCE: 'Financeiro',
}

// Mirrors UserRole.IsHealthProfessional() (internal/domain/models/user.go).
export const HEALTH_PROFESSIONAL_ROLES = new Set(['DOCTOR', 'DENTIST', 'PSYCHOLOGIST', 'PSYCHIATRIST'])

// Mirrors UserRole.CanPrescribe() — narrower than HEALTH_PROFESSIONAL_ROLES:
// a psicólogo writes clinical notes but holds no prescribing council
// registration (CRP, not CRM/CRO).
export const PRESCRIBER_ROLES = new Set(['DOCTOR', 'DENTIST', 'PSYCHIATRIST'])

// Default council type per professional role, used to pre-fill (not lock)
// UserFormModal's "Conselho" select — psiquiatra is a physician (CRM) same
// as any other DOCTOR, psicólogo holds CRP instead.
export const DEFAULT_COUNCIL_TYPE_BY_ROLE: Record<string, string> = {
  DOCTOR: 'CRM',
  DENTIST: 'CRO',
  PSYCHIATRIST: 'CRM',
  PSYCHOLOGIST: 'CRP',
}

export function roleLabel(role: string): string {
  return ROLE_LABEL[role] ?? role
}

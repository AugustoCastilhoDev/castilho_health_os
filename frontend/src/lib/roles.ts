export const ROLE_LABEL: Record<string, string> = {
  TENANT_ADMIN: 'Administrador(a)',
  DOCTOR: 'Médico(a)',
  DENTIST: 'Dentista',
  RECEPTIONIST: 'Recepção',
  FINANCE: 'Financeiro',
}

export function roleLabel(role: string): string {
  return ROLE_LABEL[role] ?? role
}

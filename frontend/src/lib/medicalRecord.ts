// Mirrors the backend's MedicalRecordType enum (internal/domain/models/medical_record.go).
export type MedicalRecordType = 'MEDICA' | 'ODONTOLOGICA' | 'PSICOLOGICA' | 'PSIQUIATRICA'

export const RECORD_TYPE_LABEL: Record<MedicalRecordType, string> = {
  MEDICA: 'Evolução Médica',
  ODONTOLOGICA: 'Evolução Odontológica',
  PSICOLOGICA: 'Evolução Psicológica',
  PSIQUIATRICA: 'Evolução Psiquiátrica',
}

// Smart default for MedicalRecordFormModal's "Tipo" field — pre-selects the
// type matching the logged-in professional's role. Doesn't restrict the
// other options: e.g. a DOCTOR in a small clinic without a dedicated
// psiquiatra may still deliberately pick "Evolução Psiquiátrica".
export const DEFAULT_RECORD_TYPE_BY_ROLE: Record<string, MedicalRecordType> = {
  DOCTOR: 'MEDICA',
  DENTIST: 'ODONTOLOGICA',
  PSYCHOLOGIST: 'PSICOLOGICA',
  PSYCHIATRIST: 'PSIQUIATRICA',
}

// Mirrors DocumentTemplateType (internal/domain/models/document_template.go).
export type DocumentTemplateType = 'ATESTADO' | 'DECLARACAO' | 'LAUDO' | 'OUTRO'

export const DOCUMENT_TEMPLATE_TYPE_LABEL: Record<DocumentTemplateType, string> = {
  ATESTADO: 'Atestado',
  DECLARACAO: 'Declaração',
  LAUDO: 'Laudo',
  OUTRO: 'Outro',
}

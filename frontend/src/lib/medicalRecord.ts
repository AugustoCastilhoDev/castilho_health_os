// Mirrors the backend's MedicalRecordType enum (internal/domain/models/medical_record.go).
export type MedicalRecordType = 'MEDICA' | 'ODONTOLOGICA' | 'PSICOLOGICA' | 'PSIQUIATRICA'

export const RECORD_TYPE_LABEL: Record<MedicalRecordType, string> = {
  MEDICA: 'Evolução Médica',
  ODONTOLOGICA: 'Evolução Odontológica',
  PSICOLOGICA: 'Evolução Psicológica',
  PSIQUIATRICA: 'Evolução Psiquiátrica',
}

// Mirrors DocumentTemplateType (internal/domain/models/document_template.go).
export type DocumentTemplateType = 'ATESTADO' | 'DECLARACAO' | 'LAUDO' | 'OUTRO'

export const DOCUMENT_TEMPLATE_TYPE_LABEL: Record<DocumentTemplateType, string> = {
  ATESTADO: 'Atestado',
  DECLARACAO: 'Declaração',
  LAUDO: 'Laudo',
  OUTRO: 'Outro',
}

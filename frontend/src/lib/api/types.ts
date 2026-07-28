// Mirrors the backend's JSON responses (internal/domain/models) — field
// names and optionality follow the Go struct tags exactly so this stays a
// direct contract with the API, not a guess.

export interface TenantDTO {
  id: string
  name: string
  slug: string
  type: string
  document: string
  email: string
  phone: string
  is_active: boolean
}

export interface UserDTO {
  id: string
  tenant_id: string
  name: string
  email: string
  role: string
  is_active: boolean
  council_type?: string
  council_number?: string
  council_state?: string
}

export interface PatientDTO {
  id: string
  tenant_id: string
  name: string
  document?: string
  birth_date?: string
  phone?: string
  email?: string
  address_zip?: string
  address_street?: string
  address_city?: string
  address_state?: string
}

export interface AppointmentDTO {
  id: string
  tenant_id: string
  patient_id: string
  professional_id: string
  scheduled_at: string
  duration_min: number
  status: string
  confirmed_at?: string
  checked_in_at?: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
  no_show_at?: string
  cancellation_reason?: string
}

export interface FinancialTransactionDTO {
  id: string
  tenant_id: string
  created_at: string
  appointment_id?: string
  patient_id?: string
  professional_id?: string
  source_transaction_id?: string
  financial_rule_id?: string
  type: string
  status: string
  gross_amount_cents: number
  fee_amount_cents: number
  net_amount_cents: number
  payment_method?: string
  procedure_code?: string
  insurance_plan?: string
  due_date?: string
  paid_at?: string
  notes?: string
}

export interface FinancialTransactionListDTO {
  items: FinancialTransactionDTO[]
  total: number
  page: number
  page_size: number
}

export interface FinancialRuleDTO {
  id: string
  tenant_id: string
  professional_id: string
  type: string
  percentage?: number
  fixed_amount_cents?: number
  procedure_code?: string
  insurance_plan?: string
  fee_deduction: string
  priority: number
  is_active: boolean
}

export interface MedicalRecordDTO {
  id: string
  tenant_id: string
  created_at: string
  patient_id: string
  professional_id: string
  appointment_id?: string
  type: string
  content: string
  is_locked: boolean
  locked_at?: string
  locked_by_id?: string
}

export interface DocumentTemplateDTO {
  id: string
  tenant_id: string
  name: string
  type: string
  content: string
  is_active: boolean
}

export interface PatientDocumentDTO {
  id: string
  tenant_id: string
  created_at: string
  patient_id: string
  uploaded_by_id: string
  file_key: string
  file_name: string
  file_size: number
  mime_type: string
  description?: string
}

export interface MemedPrescriptionLogDTO {
  id: string
  tenant_id: string
  patient_id: string
  professional_id: string
  memed_prescription_id: string
  status: 'ISSUED' | 'CANCELLED'
  issued_at: string
}

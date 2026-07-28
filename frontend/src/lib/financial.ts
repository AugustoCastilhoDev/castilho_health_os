// Mirrors the backend's financial enums (internal/domain/models/financial.go).
export type TransactionType = 'PATIENT_PAYMENT' | 'PROFESSIONAL_PAYOUT'
export type TransactionStatus = 'PENDING' | 'PAID' | 'CANCELLED'
export type PaymentMethod = 'CASH' | 'DEBIT_CARD' | 'CREDIT_CARD' | 'PIX' | 'BANK_TRANSFER' | 'INSURANCE'
export type FinancialRuleType = 'PERCENTAGE' | 'FIXED_PER_APPOINTMENT' | 'FIXED_PER_PROCEDURE'
export type FeeDeductionPolicy = 'BEFORE_SPLIT' | 'AFTER_SPLIT'

export const TRANSACTION_TYPE_LABEL: Record<TransactionType, string> = {
  PATIENT_PAYMENT: 'Recebimento',
  PROFESSIONAL_PAYOUT: 'Repasse',
}

export const TRANSACTION_STATUS_LABEL: Record<TransactionStatus, string> = {
  PENDING: 'Pendente',
  PAID: 'Pago',
  CANCELLED: 'Cancelado',
}

// Same three semantic families used across the app: sky for still-open,
// emerald for the settled/success state, rose for cancelled.
export const TRANSACTION_STATUS_STYLE: Record<TransactionStatus, string> = {
  PENDING: 'border-sky-200 bg-sky-50 text-brand-action',
  PAID: 'border-emerald-200 bg-emerald-50 text-brand-success-text',
  CANCELLED: 'border-rose-200 bg-rose-50 text-brand-alert-text line-through opacity-75',
}

export const PAYMENT_METHOD_LABEL: Record<PaymentMethod, string> = {
  CASH: 'Dinheiro',
  DEBIT_CARD: 'Cartão de débito',
  CREDIT_CARD: 'Cartão de crédito',
  PIX: 'Pix',
  BANK_TRANSFER: 'Transferência',
  INSURANCE: 'Convênio',
}

export const FINANCIAL_RULE_TYPE_LABEL: Record<FinancialRuleType, string> = {
  PERCENTAGE: 'Percentual',
  FIXED_PER_APPOINTMENT: 'Fixo por atendimento',
  FIXED_PER_PROCEDURE: 'Fixo por procedimento',
}

export const FEE_DEDUCTION_LABEL: Record<FeeDeductionPolicy, string> = {
  BEFORE_SPLIT: 'Taxa deduzida antes do repasse',
  AFTER_SPLIT: 'Taxa deduzida da clínica',
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

type FinancialRuleRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, rule *models.FinancialRule) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialRule, error)
	Update(ctx context.Context, tenantID uuid.UUID, rule *models.FinancialRule) error
	ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID) ([]models.FinancialRule, error)
	// FindApplicable resolves which single rule governs a repasse calc: the
	// highest-priority active rule whose scope (procedure/insurance) either
	// matches what's passed or is a wildcard (nil on that dimension). The
	// actual percentage/fixed-amount math stays in a settlement service —
	// this only answers "which rule applies".
	FindApplicable(ctx context.Context, tenantID, professionalID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error)
}

type financialRuleRepository struct {
	db *gorm.DB
}

func NewFinancialRuleRepository(db *gorm.DB) FinancialRuleRepository {
	return &financialRuleRepository{db: db}
}

func (r *financialRuleRepository) Create(ctx context.Context, tenantID uuid.UUID, rule *models.FinancialRule) error {
	rule.TenantID = tenantID
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *financialRuleRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialRule, error) {
	var rule models.FinancialRule
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rule, nil
}

func (r *financialRuleRepository) Update(ctx context.Context, tenantID uuid.UUID, rule *models.FinancialRule) error {
	rule.TenantID = tenantID
	result := r.db.WithContext(ctx).
		Model(&models.FinancialRule{}).
		Where("tenant_id = ? AND id = ?", tenantID, rule.ID).
		Select("*").
		Omit("id", "tenant_id", "created_at").
		Updates(rule)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *financialRuleRepository) ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID) ([]models.FinancialRule, error) {
	var rules []models.FinancialRule
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND professional_id = ?", tenantID, professionalID).
		Order("priority DESC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *financialRuleRepository) FindApplicable(ctx context.Context, tenantID, professionalID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error) {
	var rule models.FinancialRule
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND professional_id = ? AND is_active = true", tenantID, professionalID).
		Where("procedure_code IS NULL OR procedure_code = ?", procedureCode).
		Where("insurance_plan IS NULL OR insurance_plan = ?", insurancePlan).
		Order("priority DESC").
		First(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rule, nil
}

// FinancialTransactionRepository stores both patient income and
// professional payouts (see models.FinancialTransaction). Settlement logic
// — deciding *when* to generate a payout — belongs to a service layer;
// FindPayoutBySource exists so that service can stay idempotent.
type FinancialTransactionRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialTransaction, error)
	MarkPaid(ctx context.Context, tenantID, id uuid.UUID, paidAt time.Time) error
	ListByAppointment(ctx context.Context, tenantID, appointmentID uuid.UUID) ([]models.FinancialTransaction, error)
	FindPayoutBySource(ctx context.Context, tenantID, sourceTransactionID uuid.UUID) (*models.FinancialTransaction, error)
	// ListByTenant backs the Financeiro screen's ledger view: no appointment
	// or professional is known upfront, so callers filter by Type/Status
	// instead. Returns the page of matching rows plus the total count (for
	// pagination) filtered by the same conditions.
	ListByTenant(ctx context.Context, tenantID uuid.UUID, filter TransactionFilter) ([]models.FinancialTransaction, int64, error)
}

// TransactionFilter narrows ListByTenant's query. A nil Type/Status means
// "don't filter on this dimension". Limit/Offset are always applied by the
// caller (FinancialService) after defaulting/capping Limit, so the
// repository itself never has to guess a sane default.
type TransactionFilter struct {
	Type   *models.TransactionType
	Status *models.TransactionStatus
	Limit  int
	Offset int
}

type financialTransactionRepository struct {
	db *gorm.DB
}

func NewFinancialTransactionRepository(db *gorm.DB) FinancialTransactionRepository {
	return &financialTransactionRepository{db: db}
}

func (r *financialTransactionRepository) Create(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error {
	tx.TenantID = tenantID
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *financialTransactionRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialTransaction, error) {
	var tx models.FinancialTransaction
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *financialTransactionRepository) MarkPaid(ctx context.Context, tenantID, id uuid.UUID, paidAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.FinancialTransaction{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]any{
			"status":  models.TransactionPaid,
			"paid_at": paidAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *financialTransactionRepository) ListByAppointment(ctx context.Context, tenantID, appointmentID uuid.UUID) ([]models.FinancialTransaction, error) {
	var txs []models.FinancialTransaction
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND appointment_id = ?", tenantID, appointmentID).
		Order("created_at").
		Find(&txs).Error
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *financialTransactionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, filter TransactionFilter) ([]models.FinancialTransaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.FinancialTransaction{}).Where("tenant_id = ?", tenantID)
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txs []models.FinancialTransaction
	err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&txs).Error
	if err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

func (r *financialTransactionRepository) FindPayoutBySource(ctx context.Context, tenantID, sourceTransactionID uuid.UUID) (*models.FinancialTransaction, error) {
	var tx models.FinancialTransaction
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND source_transaction_id = ? AND type = ?", tenantID, sourceTransactionID, models.TransactionProfessionalPayout).
		First(&tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

package service_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

// Hand-rolled function-field fakes for the repository interfaces — no
// mocking library needed for six small interfaces. Each test sets only the
// fields it needs; calling an unset field panics with a nil-func error,
// which is exactly the signal that a test forgot to stub a call it
// actually makes.

type fakeTenantRepo struct {
	createFn     func(ctx context.Context, tenant *models.Tenant) error
	findByIDFn   func(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	findBySlugFn func(ctx context.Context, slug string) (*models.Tenant, error)
	updateFn     func(ctx context.Context, tenant *models.Tenant) error
	listFn       func(ctx context.Context) ([]models.Tenant, error)
}

func (f *fakeTenantRepo) Create(ctx context.Context, tenant *models.Tenant) error {
	return f.createFn(ctx, tenant)
}
func (f *fakeTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return f.findByIDFn(ctx, id)
}
func (f *fakeTenantRepo) FindBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	return f.findBySlugFn(ctx, slug)
}
func (f *fakeTenantRepo) Update(ctx context.Context, tenant *models.Tenant) error {
	return f.updateFn(ctx, tenant)
}
func (f *fakeTenantRepo) List(ctx context.Context) ([]models.Tenant, error) {
	return f.listFn(ctx)
}

type fakeUserRepo struct {
	createFn      func(ctx context.Context, tenantID uuid.UUID, u *models.User) error
	findByIDFn    func(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error)
	findByEmailFn func(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error)
	updateFn      func(ctx context.Context, tenantID uuid.UUID, u *models.User) error
	deleteFn      func(ctx context.Context, tenantID, id uuid.UUID) error
	listFn        func(ctx context.Context, tenantID uuid.UUID, role *models.UserRole) ([]models.User, error)
}

func (f *fakeUserRepo) Create(ctx context.Context, tenantID uuid.UUID, u *models.User) error {
	return f.createFn(ctx, tenantID, u)
}
func (f *fakeUserRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	return f.findByIDFn(ctx, tenantID, id)
}
func (f *fakeUserRepo) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	return f.findByEmailFn(ctx, tenantID, email)
}
func (f *fakeUserRepo) Update(ctx context.Context, tenantID uuid.UUID, u *models.User) error {
	return f.updateFn(ctx, tenantID, u)
}
func (f *fakeUserRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return f.deleteFn(ctx, tenantID, id)
}
func (f *fakeUserRepo) List(ctx context.Context, tenantID uuid.UUID, role *models.UserRole) ([]models.User, error) {
	return f.listFn(ctx, tenantID, role)
}

type fakePatientRepo struct {
	createFn func(ctx context.Context, tenantID uuid.UUID, p *models.Patient) error
	findFn   func(ctx context.Context, tenantID, id uuid.UUID) (*models.Patient, error)
	updateFn func(ctx context.Context, tenantID uuid.UUID, p *models.Patient) error
	deleteFn func(ctx context.Context, tenantID, id uuid.UUID) error
	searchFn func(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]models.Patient, error)
}

func (f *fakePatientRepo) Create(ctx context.Context, tenantID uuid.UUID, p *models.Patient) error {
	return f.createFn(ctx, tenantID, p)
}
func (f *fakePatientRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Patient, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakePatientRepo) Update(ctx context.Context, tenantID uuid.UUID, p *models.Patient) error {
	return f.updateFn(ctx, tenantID, p)
}
func (f *fakePatientRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return f.deleteFn(ctx, tenantID, id)
}
func (f *fakePatientRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]models.Patient, error) {
	return f.searchFn(ctx, tenantID, query, limit, offset)
}

type fakeAppointmentRepo struct {
	createFn             func(ctx context.Context, tenantID uuid.UUID, a *models.Appointment) error
	findFn               func(ctx context.Context, tenantID, id uuid.UUID) (*models.Appointment, error)
	listByProfessionalFn func(ctx context.Context, tenantID, professionalID uuid.UUID, from, to time.Time) ([]models.Appointment, error)
	listByPatientFn      func(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.Appointment, error)
	transitionFn         func(ctx context.Context, tenantID, id uuid.UUID, to models.AppointmentStatus, changedByID uuid.UUID, reason string) (*models.Appointment, error)
}

func (f *fakeAppointmentRepo) Create(ctx context.Context, tenantID uuid.UUID, a *models.Appointment) error {
	return f.createFn(ctx, tenantID, a)
}
func (f *fakeAppointmentRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Appointment, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakeAppointmentRepo) ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID, from, to time.Time) ([]models.Appointment, error) {
	return f.listByProfessionalFn(ctx, tenantID, professionalID, from, to)
}
func (f *fakeAppointmentRepo) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.Appointment, error) {
	return f.listByPatientFn(ctx, tenantID, patientID)
}
func (f *fakeAppointmentRepo) TransitionStatus(ctx context.Context, tenantID, id uuid.UUID, to models.AppointmentStatus, changedByID uuid.UUID, reason string) (*models.Appointment, error) {
	return f.transitionFn(ctx, tenantID, id, to, changedByID, reason)
}

type fakeFinancialRuleRepo struct {
	createFn         func(ctx context.Context, tenantID uuid.UUID, r *models.FinancialRule) error
	findFn           func(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialRule, error)
	updateFn         func(ctx context.Context, tenantID uuid.UUID, r *models.FinancialRule) error
	listByProfFn     func(ctx context.Context, tenantID, professionalID uuid.UUID) ([]models.FinancialRule, error)
	findApplicableFn func(ctx context.Context, tenantID, professionalID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error)
}

func (f *fakeFinancialRuleRepo) Create(ctx context.Context, tenantID uuid.UUID, r *models.FinancialRule) error {
	return f.createFn(ctx, tenantID, r)
}
func (f *fakeFinancialRuleRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialRule, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakeFinancialRuleRepo) Update(ctx context.Context, tenantID uuid.UUID, r *models.FinancialRule) error {
	return f.updateFn(ctx, tenantID, r)
}
func (f *fakeFinancialRuleRepo) ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID) ([]models.FinancialRule, error) {
	return f.listByProfFn(ctx, tenantID, professionalID)
}
func (f *fakeFinancialRuleRepo) FindApplicable(ctx context.Context, tenantID, professionalID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error) {
	return f.findApplicableFn(ctx, tenantID, professionalID, procedureCode, insurancePlan)
}

type fakeFinancialTransactionRepo struct {
	createFn             func(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error
	findFn               func(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialTransaction, error)
	markPaidFn           func(ctx context.Context, tenantID, id uuid.UUID, paidAt time.Time) error
	listByAppointmentFn  func(ctx context.Context, tenantID, appointmentID uuid.UUID) ([]models.FinancialTransaction, error)
	findPayoutBySourceFn func(ctx context.Context, tenantID, sourceTransactionID uuid.UUID) (*models.FinancialTransaction, error)
	listByTenantFn       func(ctx context.Context, tenantID uuid.UUID, filter repository.TransactionFilter) ([]models.FinancialTransaction, int64, error)
}

func (f *fakeFinancialTransactionRepo) Create(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error {
	return f.createFn(ctx, tenantID, tx)
}
func (f *fakeFinancialTransactionRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialTransaction, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakeFinancialTransactionRepo) MarkPaid(ctx context.Context, tenantID, id uuid.UUID, paidAt time.Time) error {
	return f.markPaidFn(ctx, tenantID, id, paidAt)
}
func (f *fakeFinancialTransactionRepo) ListByAppointment(ctx context.Context, tenantID, appointmentID uuid.UUID) ([]models.FinancialTransaction, error) {
	return f.listByAppointmentFn(ctx, tenantID, appointmentID)
}
func (f *fakeFinancialTransactionRepo) FindPayoutBySource(ctx context.Context, tenantID, sourceTransactionID uuid.UUID) (*models.FinancialTransaction, error) {
	return f.findPayoutBySourceFn(ctx, tenantID, sourceTransactionID)
}
func (f *fakeFinancialTransactionRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, filter repository.TransactionFilter) ([]models.FinancialTransaction, int64, error) {
	return f.listByTenantFn(ctx, tenantID, filter)
}

type fakeMedicalRecordRepo struct {
	createFn        func(ctx context.Context, tenantID uuid.UUID, r *models.MedicalRecord) error
	findFn          func(ctx context.Context, tenantID, id uuid.UUID) (*models.MedicalRecord, error)
	listByPatientFn func(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MedicalRecord, error)
	updateFn        func(ctx context.Context, tenantID uuid.UUID, r *models.MedicalRecord) error
	lockFn          func(ctx context.Context, tenantID, id, lockedByID uuid.UUID) (*models.MedicalRecord, error)
}

func (f *fakeMedicalRecordRepo) Create(ctx context.Context, tenantID uuid.UUID, r *models.MedicalRecord) error {
	return f.createFn(ctx, tenantID, r)
}
func (f *fakeMedicalRecordRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.MedicalRecord, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakeMedicalRecordRepo) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MedicalRecord, error) {
	return f.listByPatientFn(ctx, tenantID, patientID)
}
func (f *fakeMedicalRecordRepo) Update(ctx context.Context, tenantID uuid.UUID, r *models.MedicalRecord) error {
	return f.updateFn(ctx, tenantID, r)
}
func (f *fakeMedicalRecordRepo) Lock(ctx context.Context, tenantID, id, lockedByID uuid.UUID) (*models.MedicalRecord, error) {
	return f.lockFn(ctx, tenantID, id, lockedByID)
}

type fakeDocumentTemplateRepo struct {
	createFn       func(ctx context.Context, tenantID uuid.UUID, t *models.DocumentTemplate) error
	findFn         func(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error)
	updateFn       func(ctx context.Context, tenantID uuid.UUID, t *models.DocumentTemplate) error
	listByTenantFn func(ctx context.Context, tenantID uuid.UUID) ([]models.DocumentTemplate, error)
}

func (f *fakeDocumentTemplateRepo) Create(ctx context.Context, tenantID uuid.UUID, t *models.DocumentTemplate) error {
	return f.createFn(ctx, tenantID, t)
}
func (f *fakeDocumentTemplateRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakeDocumentTemplateRepo) Update(ctx context.Context, tenantID uuid.UUID, t *models.DocumentTemplate) error {
	return f.updateFn(ctx, tenantID, t)
}
func (f *fakeDocumentTemplateRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.DocumentTemplate, error) {
	return f.listByTenantFn(ctx, tenantID)
}

type fakePatientDocumentRepo struct {
	createFn        func(ctx context.Context, tenantID uuid.UUID, d *models.PatientDocument) error
	findFn          func(ctx context.Context, tenantID, id uuid.UUID) (*models.PatientDocument, error)
	listByPatientFn func(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.PatientDocument, error)
	deleteFn        func(ctx context.Context, tenantID, id uuid.UUID) error
}

func (f *fakePatientDocumentRepo) Create(ctx context.Context, tenantID uuid.UUID, d *models.PatientDocument) error {
	return f.createFn(ctx, tenantID, d)
}
func (f *fakePatientDocumentRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.PatientDocument, error) {
	return f.findFn(ctx, tenantID, id)
}
func (f *fakePatientDocumentRepo) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.PatientDocument, error) {
	return f.listByPatientFn(ctx, tenantID, patientID)
}
func (f *fakePatientDocumentRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return f.deleteFn(ctx, tenantID, id)
}

type fakeMemedPrescriptionLogRepo struct {
	createFn          func(ctx context.Context, tenantID uuid.UUID, l *models.MemedPrescriptionLog) error
	findByMemedIDFn   func(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error)
	listByPatientFn   func(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MemedPrescriptionLog, error)
	cancelByMemedIDFn func(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) error
}

func (f *fakeMemedPrescriptionLogRepo) Create(ctx context.Context, tenantID uuid.UUID, l *models.MemedPrescriptionLog) error {
	return f.createFn(ctx, tenantID, l)
}
func (f *fakeMemedPrescriptionLogRepo) FindByMemedID(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error) {
	return f.findByMemedIDFn(ctx, tenantID, memedPrescriptionID)
}
func (f *fakeMemedPrescriptionLogRepo) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MemedPrescriptionLog, error) {
	return f.listByPatientFn(ctx, tenantID, patientID)
}
func (f *fakeMemedPrescriptionLogRepo) CancelByMemedID(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) error {
	return f.cancelByMemedIDFn(ctx, tenantID, memedPrescriptionID)
}

type fakeMemedClient struct {
	fetchOrRegisterTokenFn func(ctx context.Context, p service.MemedPrescriber) (string, error)
}

func (f *fakeMemedClient) FetchOrRegisterToken(ctx context.Context, p service.MemedPrescriber) (string, error) {
	return f.fetchOrRegisterTokenFn(ctx, p)
}

type fakeObjectStorage struct {
	presignUploadFn   func(ctx context.Context, fileKey, contentType string, expiresIn time.Duration) (string, error)
	presignDownloadFn func(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error)
	deleteObjectFn    func(ctx context.Context, fileKey string) error
}

func (f *fakeObjectStorage) PresignUpload(ctx context.Context, fileKey, contentType string, expiresIn time.Duration) (string, error) {
	return f.presignUploadFn(ctx, fileKey, contentType, expiresIn)
}
func (f *fakeObjectStorage) PresignDownload(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error) {
	return f.presignDownloadFn(ctx, fileKey, expiresIn)
}
func (f *fakeObjectStorage) DeleteObject(ctx context.Context, fileKey string) error {
	return f.deleteObjectFn(ctx, fileKey)
}

type fakeStockItemRepo struct {
	createFn   func(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error
	findByIDFn func(ctx context.Context, tenantID, id uuid.UUID) (*models.StockItem, error)
	updateFn   func(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error
	listFn     func(ctx context.Context, tenantID uuid.UUID) ([]models.StockItem, error)
}

func (f *fakeStockItemRepo) Create(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error {
	return f.createFn(ctx, tenantID, item)
}
func (f *fakeStockItemRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.StockItem, error) {
	return f.findByIDFn(ctx, tenantID, id)
}
func (f *fakeStockItemRepo) Update(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error {
	return f.updateFn(ctx, tenantID, item)
}
func (f *fakeStockItemRepo) List(ctx context.Context, tenantID uuid.UUID) ([]models.StockItem, error) {
	return f.listFn(ctx, tenantID)
}

type fakeStockMovementRepo struct {
	recordMovementFn func(ctx context.Context, tenantID uuid.UUID, movement *models.StockMovement) (*models.StockItem, error)
	listByItemFn     func(ctx context.Context, tenantID, itemID uuid.UUID) ([]models.StockMovement, error)
}

func (f *fakeStockMovementRepo) RecordMovement(ctx context.Context, tenantID uuid.UUID, movement *models.StockMovement) (*models.StockItem, error) {
	return f.recordMovementFn(ctx, tenantID, movement)
}
func (f *fakeStockMovementRepo) ListByItem(ctx context.Context, tenantID, itemID uuid.UUID) ([]models.StockMovement, error) {
	return f.listByItemFn(ctx, tenantID, itemID)
}

type fakeOdontogramaRepo struct {
	createFn        func(ctx context.Context, tenantID uuid.UUID, entry *models.OdontogramaEntry) error
	listByPatientFn func(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error)
	currentChartFn  func(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error)
}

func (f *fakeOdontogramaRepo) Create(ctx context.Context, tenantID uuid.UUID, entry *models.OdontogramaEntry) error {
	return f.createFn(ctx, tenantID, entry)
}
func (f *fakeOdontogramaRepo) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error) {
	return f.listByPatientFn(ctx, tenantID, patientID)
}
func (f *fakeOdontogramaRepo) CurrentChart(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error) {
	return f.currentChartFn(ctx, tenantID, patientID)
}

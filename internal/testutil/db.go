// Package testutil provides shared helpers for integration tests that need
// a real Postgres connection (the local docker-compose db). It's only ever
// imported from _test.go files — nothing in cmd/ or the production import
// graph depends on it.
package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/config"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/infra/db"
)

// ConnectDB opens the local Postgres used by docker-compose, skipping the
// calling test (not failing it) if it's unreachable — an integration test
// shouldn't break `go test ./...` for someone who hasn't run
// `docker compose up db` yet.
func ConnectDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("skipping integration test: config: %v", err)
	}
	gdb, err := db.Connect(cfg)
	if err != nil {
		t.Skipf("skipping integration test: db unreachable (is `docker compose up -d db` running?): %v", err)
	}
	return gdb
}

// NewTenant creates and persists a throwaway Tenant with a random
// slug/document/email, and registers cleanup to hard-delete it — along
// with anything a test hangs off it by TenantID — when the test ends.
func NewTenant(t *testing.T, gdb *gorm.DB) *models.Tenant {
	t.Helper()
	suffix := uuid.NewString()[:8]
	tenant := &models.Tenant{
		Name:     "Test Tenant " + suffix,
		Slug:     "test-" + uuid.NewString(),
		Type:     models.TenantTypeMista,
		Document: uuid.NewString()[:14],
		Email:    "tenant-" + suffix + "@example.com",
		IsActive: true,
	}
	if err := gdb.WithContext(context.Background()).Create(tenant).Error; err != nil {
		t.Fatalf("create test tenant: %v", err)
	}
	t.Cleanup(func() {
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.StockMovement{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.StockItem{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.MemedPrescriptionLog{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.PatientDocument{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.MedicalRecord{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.DocumentTemplate{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.FinancialTransaction{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.FinancialRule{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.AppointmentStatusLog{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.Appointment{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.Patient{})
		gdb.Unscoped().Where("tenant_id = ?", tenant.ID).Delete(&models.User{})
		gdb.Unscoped().Where("id = ?", tenant.ID).Delete(&models.Tenant{})
	})
	return tenant
}

func NewUser(t *testing.T, gdb *gorm.DB, tenantID uuid.UUID, role models.UserRole) *models.User {
	t.Helper()
	user := &models.User{
		TenantModel:  models.TenantModel{TenantID: tenantID},
		Name:         "Test User " + uuid.NewString()[:8],
		Email:        uuid.NewString() + "@example.com",
		PasswordHash: "test-hash",
		Role:         role,
		IsActive:     true,
	}
	if err := gdb.WithContext(context.Background()).Create(user).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return user
}

func NewPatient(t *testing.T, gdb *gorm.DB, tenantID uuid.UUID) *models.Patient {
	t.Helper()
	patient := &models.Patient{
		TenantModel: models.TenantModel{TenantID: tenantID},
		Name:        "Test Patient " + uuid.NewString()[:8],
	}
	if err := gdb.WithContext(context.Background()).Create(patient).Error; err != nil {
		t.Fatalf("create test patient: %v", err)
	}
	return patient
}

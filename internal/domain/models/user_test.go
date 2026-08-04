package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/castilho/health-os/internal/domain/models"
)

func TestUserRole_IsHealthProfessional(t *testing.T) {
	cases := []struct {
		role models.UserRole
		want bool
	}{
		{models.RoleDoctor, true},
		{models.RoleDentist, true},
		{models.RolePsychologist, true},
		{models.RolePsychiatrist, true},
		{models.RoleTenantAdmin, false},
		{models.RoleReceptionist, false},
		{models.RoleFinance, false},
	}
	for _, tc := range cases {
		t.Run(string(tc.role), func(t *testing.T) {
			assert.Equal(t, tc.want, tc.role.IsHealthProfessional())
		})
	}
}

func TestUserRole_CanPrescribe(t *testing.T) {
	cases := []struct {
		role models.UserRole
		want bool
	}{
		{models.RoleDoctor, true},
		{models.RoleDentist, true},
		{models.RolePsychiatrist, true},
		// PSYCHOLOGIST is a health professional but not a prescriber — see
		// CanPrescribe's doc comment.
		{models.RolePsychologist, false},
		{models.RoleTenantAdmin, false},
		{models.RoleReceptionist, false},
		{models.RoleFinance, false},
	}
	for _, tc := range cases {
		t.Run(string(tc.role), func(t *testing.T) {
			assert.Equal(t, tc.want, tc.role.CanPrescribe())
		})
	}
}

package models

import "time"

// UserSex is deliberately just M/F, not a broader enum — the only consumer
// is Memed's prescriber-registration API, which only accepts those two
// values for the (optional) sexo field.
type UserSex string

const (
	UserSexMale   UserSex = "M"
	UserSexFemale UserSex = "F"
)

// UserRole is the coarse-grained RBAC role. Fine-grained permission
// overrides (if a clinic ever needs them) belong in a separate
// role_permissions table added when that need actually shows up — a fixed
// enum is enough to ship the first version of every module described in
// scope.
type UserRole string

const (
	RoleTenantAdmin  UserRole = "TENANT_ADMIN"
	RoleDoctor       UserRole = "DOCTOR"
	RoleDentist      UserRole = "DENTIST"
	RoleReceptionist UserRole = "RECEPTIONIST"
	RoleFinance      UserRole = "FINANCE"
	RolePsychologist UserRole = "PSYCHOLOGIST"
	RolePsychiatrist UserRole = "PSYCHIATRIST"
)

// IsHealthProfessional reports whether this role can be the ProfessionalID
// on an Appointment / FinancialRule.
func (r UserRole) IsHealthProfessional() bool {
	return r == RoleDoctor || r == RoleDentist || r == RolePsychologist || r == RolePsychiatrist
}

// CanPrescribe reports whether this role legally holds a prescribing
// council registration (CRM/CRO) and may issue a prescription via Memed.
// PSYCHOLOGIST is a health professional (IsHealthProfessional) but is not a
// prescriber — psicólogos are a distinct profession in Brazil (CRP), unlike
// psiquiatras, who are physicians (CRM) same as any other DOCTOR.
func (r UserRole) CanPrescribe() bool {
	return r == RoleDoctor || r == RoleDentist || r == RolePsychiatrist
}

// User represents any person who logs in: health professionals (doctors,
// dentists) and staff (receptionists, finance, admins). Splitting
// professionals into their own table was considered and rejected for v1 —
// every role shares login/auth/RBAC concerns, and only a handful of extra
// nullable columns are professional-only.
type User struct {
	TenantModel
	Name         string   `gorm:"type:varchar(255);not null" json:"name"`
	Email        string   `gorm:"type:varchar(255);not null;index" json:"email"`
	PasswordHash string   `gorm:"type:varchar(255);not null" json:"-"`
	Role         UserRole `gorm:"type:varchar(20);not null;index" json:"role"`
	IsActive     bool     `gorm:"not null;default:true" json:"is_active"`

	// Professional council registration. Populated only when
	// Role.IsHealthProfessional(); nil otherwise.
	CouncilType   *string `gorm:"type:varchar(10)" json:"council_type,omitempty"` // "CRM" | "CRO" | "CRP"
	CouncilNumber *string `gorm:"type:varchar(20)" json:"council_number,omitempty"`
	CouncilState  *string `gorm:"type:varchar(2)" json:"council_state,omitempty"`

	// CPF and BirthDate are mandatory fields on Memed's prescriber
	// registration API (only relevant for a CanPrescribe role); Phone/Sex
	// are optional there. All nil until an admin fills them in for a health
	// professional — kept available to every IsHealthProfessional role
	// (including PSYCHOLOGIST, who never calls Memed) since council
	// registration is general professional record-keeping, not solely a
	// Memed prerequisite.
	CPF       *string    `gorm:"type:varchar(11)" json:"cpf,omitempty"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	Sex       *UserSex   `gorm:"type:varchar(1)" json:"sex,omitempty"`
	Phone     *string    `gorm:"type:varchar(20)" json:"phone,omitempty"`

	FinancialRules []FinancialRule `gorm:"foreignKey:ProfessionalID" json:"-"`
}

func (User) TableName() string { return "users" }

// Email is unique per tenant, not globally. Because that constraint spans
// the embedded TenantID field, it's added as a raw migration statement
// rather than a GORM struct tag (a tag here would collide across every
// other TenantModel-embedding table, since Postgres index names are unique
// per schema):
//
//	CREATE UNIQUE INDEX idx_users_tenant_email ON users (tenant_id, email) WHERE deleted_at IS NULL;

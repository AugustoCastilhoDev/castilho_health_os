-- Two new health-professional roles (see internal/domain/models/user.go):
-- PSYCHOLOGIST (CRP, cannot prescribe) and PSYCHIATRIST (CRM, can prescribe
-- — a physician same as DOCTOR, just tracked as its own role for
-- reporting/filtering). Postgres CHECK constraints can't be altered
-- in-place, so the old one is dropped and recreated with the wider list.
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('TENANT_ADMIN', 'DOCTOR', 'DENTIST', 'RECEPTIONIST', 'FINANCE', 'PSYCHOLOGIST', 'PSYCHIATRIST'));

-- Free-text diagnosis code (see doc comments on MedicalRecord.CID /
-- Appointment.CID) — not validated against an official CID-10 table in v1.
ALTER TABLE medical_records ADD COLUMN cid VARCHAR(255);
ALTER TABLE appointments ADD COLUMN cid VARCHAR(255);

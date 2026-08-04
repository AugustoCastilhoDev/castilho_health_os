-- Clinic letterhead data (see internal/domain/models/tenant.go) — all
-- nullable, since a tenant that hasn't filled these in yet must keep
-- generating documents fine (letterhead blocks just render without a
-- logo/address, see internal/pdf/generate.go).
ALTER TABLE tenants ADD COLUMN address_street VARCHAR(255);
ALTER TABLE tenants ADD COLUMN address_city   VARCHAR(100);
ALTER TABLE tenants ADD COLUMN address_state  VARCHAR(2);
ALTER TABLE tenants ADD COLUMN address_zip    VARCHAR(10);
-- R2 object key, same pattern as patient_documents.file_key — never a
-- public URL, always resolved through a presigned URL on demand.
ALTER TABLE tenants ADD COLUMN logo_key       VARCHAR(500);

-- Per-template toggles for the optional PDF layout blocks (see
-- internal/domain/models/document_template.go). Default TRUE so existing
-- templates keep their current "just title + body" look only if the admin
-- explicitly unchecks a block — the common case (new templates) is meant
-- to look complete out of the box.
ALTER TABLE document_templates ADD COLUMN include_header    BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE document_templates ADD COLUMN include_footer    BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE document_templates ADD COLUMN include_signature BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE document_templates ADD COLUMN include_stamp     BOOLEAN NOT NULL DEFAULT TRUE;

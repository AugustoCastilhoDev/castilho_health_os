ALTER TABLE document_templates DROP COLUMN IF EXISTS include_stamp;
ALTER TABLE document_templates DROP COLUMN IF EXISTS include_signature;
ALTER TABLE document_templates DROP COLUMN IF EXISTS include_footer;
ALTER TABLE document_templates DROP COLUMN IF EXISTS include_header;

ALTER TABLE tenants DROP COLUMN IF EXISTS logo_key;
ALTER TABLE tenants DROP COLUMN IF EXISTS address_zip;
ALTER TABLE tenants DROP COLUMN IF EXISTS address_state;
ALTER TABLE tenants DROP COLUMN IF EXISTS address_city;
ALTER TABLE tenants DROP COLUMN IF EXISTS address_street;

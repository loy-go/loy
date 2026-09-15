-- +goose Up
-- SQL in this section is executed after the migration is applied.
CREATE TABLE IF NOT EXISTS organizations (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    name TEXT NOT NULL,
    plan TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_organizations_org_id ON organizations (org_id);
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations FORCE ROW LEVEL SECURITY;
CREATE POLICY organizations_tenant_isolation_policy ON organizations
    FOR ALL
    TO public
    USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS organizations CASCADE;

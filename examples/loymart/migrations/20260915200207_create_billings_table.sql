-- +goose Up
-- SQL in this section is executed after the migration is applied.
CREATE TABLE IF NOT EXISTS billings (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_billings_org_id ON billings (org_id);
ALTER TABLE billings ENABLE ROW LEVEL SECURITY;
ALTER TABLE billings FORCE ROW LEVEL SECURITY;
CREATE POLICY billings_tenant_isolation_policy ON billings
    FOR ALL
    TO public
    USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS billings CASCADE;

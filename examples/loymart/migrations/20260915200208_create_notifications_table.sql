-- +goose Up
-- SQL in this section is executed after the migration is applied.
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    channel TEXT NOT NULL,
    recipient TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notifications_org_id ON notifications (org_id);
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications FORCE ROW LEVEL SECURITY;
CREATE POLICY notifications_tenant_isolation_policy ON notifications
    FOR ALL
    TO public
    USING (org_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS notifications CASCADE;

CREATE TABLE IF NOT EXISTS audit_logs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type  VARCHAR(100) NOT NULL,
    entity_id    UUID         NOT NULL,
    action       VARCHAR(50)  NOT NULL,
    actor_id     UUID         REFERENCES users(id),
    old_values   JSONB,
    new_values   JSONB,
    changes      JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_actor  ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_time   ON audit_logs(created_at);

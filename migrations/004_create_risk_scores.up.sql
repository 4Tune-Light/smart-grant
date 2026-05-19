CREATE TABLE IF NOT EXISTS risk_scores (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id  UUID         NOT NULL REFERENCES proposals(id),
    risk_level   VARCHAR(50)  NOT NULL CHECK (risk_level IN ('low', 'medium', 'high')),
    confidence   REAL         NOT NULL DEFAULT 0.0 CHECK (confidence >= 0 AND confidence <= 1),
    features     JSONB,
    details      JSONB,
    model_version VARCHAR(50) NOT NULL DEFAULT 'c4.5-v1',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_risk_scores_proposal ON risk_scores(proposal_id);
CREATE INDEX idx_risk_scores_level ON risk_scores(risk_level);

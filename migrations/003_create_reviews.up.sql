CREATE TABLE IF NOT EXISTS reviews (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id  UUID         NOT NULL REFERENCES proposals(id),
    reviewer_id  UUID         NOT NULL REFERENCES users(id),
    score        INT          NOT NULL CHECK (score >= 0 AND score <= 100),
    comment      TEXT,
    status       VARCHAR(50)  NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE(proposal_id, reviewer_id)
);

CREATE INDEX idx_reviews_proposal ON reviews(proposal_id);
CREATE INDEX idx_reviews_reviewer ON reviews(reviewer_id);

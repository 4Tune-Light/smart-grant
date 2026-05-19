CREATE TABLE IF NOT EXISTS proposals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id  UUID         NOT NULL REFERENCES users(id),
    title         VARCHAR(255) NOT NULL,
    description   TEXT         NOT NULL,
    nominal_amount NUMERIC(15,2) NOT NULL CHECK (nominal_amount > 0),
    organization  VARCHAR(255) NOT NULL,
    status        VARCHAR(50)  NOT NULL DEFAULT 'draft'
                  CHECK (status IN ('draft', 'submitted', 'in_review', 'approved', 'rejected', 'revision_needed')),
    version       INT          NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_proposals_applicant ON proposals(applicant_id);
CREATE INDEX idx_proposals_status ON proposals(status);

CREATE TABLE IF NOT EXISTS proposal_documents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id  UUID         NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
    filename     VARCHAR(255) NOT NULL,
    file_path    TEXT         NOT NULL,
    mime_type    VARCHAR(100) NOT NULL,
    file_size    BIGINT       NOT NULL DEFAULT 0,
    uploaded_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_proposal_documents_proposal ON proposal_documents(proposal_id);

CREATE TABLE IF NOT EXISTS proposal_versions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id    UUID         NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
    version_number INT          NOT NULL,
    snapshot       JSONB        NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE(proposal_id, version_number)
);

CREATE INDEX idx_proposal_versions_proposal ON proposal_versions(proposal_id);

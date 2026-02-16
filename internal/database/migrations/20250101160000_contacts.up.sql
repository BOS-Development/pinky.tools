BEGIN;

CREATE TABLE contacts (
    id BIGSERIAL PRIMARY KEY,
    requester_user_id BIGINT NOT NULL REFERENCES users(id),
    recipient_user_id BIGINT NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT contacts_unique_pair UNIQUE (requester_user_id, recipient_user_id),
    CONSTRAINT contacts_no_self CHECK (requester_user_id != recipient_user_id)
);

CREATE INDEX idx_contacts_requester ON contacts(requester_user_id);
CREATE INDEX idx_contacts_recipient ON contacts(recipient_user_id);
CREATE INDEX idx_contacts_status ON contacts(status);

COMMIT;

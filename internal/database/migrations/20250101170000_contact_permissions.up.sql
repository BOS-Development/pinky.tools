BEGIN;

CREATE TABLE contact_permissions (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    granting_user_id BIGINT NOT NULL REFERENCES users(id),
    receiving_user_id BIGINT NOT NULL REFERENCES users(id),
    service_type VARCHAR(50) NOT NULL,
    can_access BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT permission_unique_grant UNIQUE (contact_id, granting_user_id, receiving_user_id, service_type)
);

CREATE INDEX idx_permission_contact ON contact_permissions(contact_id);
CREATE INDEX idx_permission_receiving ON contact_permissions(receiving_user_id, service_type, can_access);

COMMIT;

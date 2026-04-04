CREATE TABLE IF NOT EXISTS documents(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT Now(),
    updated_at TIMESTAMPTZ DEFAULT Now(),
    CONSTRAINT fk_user_documents FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);
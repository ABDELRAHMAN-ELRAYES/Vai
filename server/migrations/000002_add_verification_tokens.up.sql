CREATE TABLE IF NOT EXISTS verification_tokens(
    user_id uuid NOT NULL,
    token bytea NOT NULL,
    expired_at TIMESTAMPTZ NOT NULL ,

    CONSTRAINT pk_user_verification_token FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
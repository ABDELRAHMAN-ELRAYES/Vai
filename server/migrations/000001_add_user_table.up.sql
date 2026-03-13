CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users(
    id uuid primary key DEFAULT gen_random_uuid(),
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email citext UNIQUE NOT NULL,
    password bytea NOT NULL,
    created_at timestamp(0) WITH TIME ZONE  NOT NULL DEFAULT NOW(),
    is_active BOOLEAN DEFAULT FALSE
);
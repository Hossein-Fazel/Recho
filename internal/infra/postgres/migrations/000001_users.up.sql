CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    username      CITEXT       NOT NULL UNIQUE,
    
    display_name  VARCHAR(100),
    avatar_key    TEXT,
    bio           VARCHAR(250),

    password_hash TEXT          NOT NULL,

    last_seen     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT username_length CHECK (char_length(username) BETWEEN 3 AND 30)
);

CREATE INDEX users_username_search_idx ON users (username citext_pattern_ops);
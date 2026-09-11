CREATE TYPE message_type AS ENUM (
    'text'
    -- 'voice',
    -- 'video',
    -- 'picture'
    -- etc
);

CREATE TABLE conversations (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    last_message_id         BIGINT,
    last_message_type       message_type,
    last_message_text       TEXT,
    last_message_created_at TIMESTAMPTZ,
    message_id_counter      BIGINT      NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_updated_at ON conversations(updated_at DESC);

CREATE TABLE direct_conversations (
    conversation_id UUID        PRIMARY KEY REFERENCES conversations(id) ON DELETE CASCADE,

    user_one_id     UUID        NOT NULL REFERENCES users(id),

    user_two_id     UUID        NOT NULL REFERENCES users(id),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT direct_conversations_different_users CHECK (user_one_id <> user_two_id),
    CONSTRAINT direct_conversations_unique_pair     UNIQUE (user_one_id, user_two_id)
);


CREATE INDEX idx_direct_conversations_user_one ON direct_conversations(user_one_id);

CREATE INDEX idx_direct_conversations_user_two ON direct_conversations(user_two_id);


CREATE TABLE groups (
    conversation_id UUID PRIMARY KEY REFERENCES conversations(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    avatar_url      TEXT,
    invite_code     VARCHAR(100) UNIQUE,
    bio             VARCHAR(250),
    created_by      UUID         NOT NULL REFERENCES users(id),

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TYPE group_member_role AS ENUM (
    'owner',
    'admin',
    'member'
);

CREATE TABLE group_members (
    group_id  UUID              NOT NULL REFERENCES groups(conversation_id) ON DELETE CASCADE,
    user_id   UUID              NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role      group_member_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ       NOT NULL DEFAULT NOW(),

    PRIMARY KEY (group_id, user_id)
);
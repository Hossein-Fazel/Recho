CREATE TABLE conversations (
    id UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE direct_conversations (
    conversation_id UUID PRIMARY KEY
        REFERENCES conversations(id)
        ON DELETE CASCADE,

    user_one_id UUID NOT NULL
        REFERENCES users(id),

    user_two_id UUID NOT NULL
        REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT direct_conversations_different_users
        CHECK (user_one_id <> user_two_id),

    CONSTRAINT direct_conversations_unique_pair
        UNIQUE (user_one_id, user_two_id)
);


CREATE INDEX idx_direct_conversations_user_one
    ON direct_conversations(user_one_id);

CREATE INDEX idx_direct_conversations_user_two
    ON direct_conversations(user_two_id);


CREATE TABLE groups (
    conversation_id UUID PRIMARY KEY
        REFERENCES conversations(id)
        ON DELETE CASCADE,

    name        VARCHAR(100) NOT NULL,
    avatar_url  TEXT,
    invite_code VARCHAR(100) UNIQUE,

    created_by UUID NOT NULL
        REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE group_member_role AS ENUM (
    'owner',
    'admin',
    'member'
);

CREATE TABLE group_members (
    group_id UUID NOT NULL REFERENCES groups(conversation_id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role      group_member_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (group_id, user_id)
);
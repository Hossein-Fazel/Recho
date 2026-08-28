CREATE TABLE conversations (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    last_message_id         UUID
    last_message_content    TEXT
    last_message_created_at TIMESTAMPTZ;
    last_message_id         BIGINT      NOT NULL DEFAULT 0;
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_updated_at ON conversations(updated_at DESC);

CREATE OR REPLACE FUNCTION update_conversation_on_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET
        updated_at = NEW.created_at,
        last_message_id = NEW.id,
        last_message_content = NEW.content,
        last_message_created_at = NEW.created_at
    WHERE id = NEW.conversation_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_messages_update_conversation
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION update_conversation_on_message();

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
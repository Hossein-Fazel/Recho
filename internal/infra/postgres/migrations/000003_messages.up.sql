CREATE TABLE messages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    sender_id       UUID        NOT NULL REFERENCES users(id),

    content         TEXT        NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT message_content_not_empty CHECK (length(trim(content)) > 0)
);

CREATE INDEX idx_messages_conversation_created_at ON messages(conversation_id, created_at DESC);
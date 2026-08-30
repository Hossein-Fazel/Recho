CREATE TABLE messages (
    message_id      BIGINT      NOT NULL,
    conversation_id UUID        NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    sender_id       UUID        NOT NULL REFERENCES users(id),

    content         TEXT        NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (conversation_id, message_id),
    CONSTRAINT message_content_not_empty CHECK (length(trim(content)) > 0)
);

CREATE INDEX idx_messages_conversation_created_at ON messages(conversation_id, created_at DESC);

CREATE OR REPLACE FUNCTION set_message_id()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET message_id_counter = message_id_counter + 1
    WHERE id = NEW.conversation_id
    RETURNING message_id_counter INTO NEW.message_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_set_message_id
BEFORE INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION set_message_id();

CREATE OR REPLACE FUNCTION update_conversation_on_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET
        updated_at = NEW.created_at,
        last_message_id = NEW.message_id,
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
CREATE TABLE messages (
    message_id      BIGINT       NOT NULL,
    conversation_id UUID         NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID         NOT NULL REFERENCES users(id),
    type            message_type NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    PRIMARY KEY (conversation_id, message_id)
);

CREATE INDEX idx_messages_conversation_created_at ON messages(conversation_id, created_at DESC);

CREATE TABLE text_messages (
    conversation_id UUID   NOT NULL,
    message_id      BIGINT NOT NULL,
    content         TEXT   NOT NULL,

    PRIMARY KEY (conversation_id, message_id),
    FOREIGN KEY (conversation_id, message_id) REFERENCES messages(conversation_id, message_id) ON DELETE CASCADE,

    CONSTRAINT message_text_content_not_empty CHECK (length(trim(content)) > 0)
);

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

CREATE OR REPLACE FUNCTION refresh_last_message_on_delete()
RETURNS TRIGGER AS $$
DECLARE
    latest_message_id BIGINT;
    latest_message_type message_type;
    latest_message_text TEXT;
    latest_message_created_at TIMESTAMPTZ;
BEGIN
    -- Only refresh if the deleted message was the last message.
    IF OLD.message_id <> (
        SELECT last_message_id
        FROM conversations
        WHERE id = OLD.conversation_id
    ) THEN
        RETURN OLD;
    END IF;

    SELECT
        m.message_id,
        m.type,
        m.created_at
    INTO
        latest_message_id,
        latest_message_type,
        latest_message_created_at
    FROM messages m
    WHERE m.conversation_id = OLD.conversation_id
    ORDER BY m.message_id DESC
    LIMIT 1;

    -- FOUND is a special PL/pgSQL variable that is set to
    -- true if the previous `SELECT INTO` found a row and
    -- false otherwise.
    IF FOUND THEN

        IF latest_message_type = 'text' THEN
            SELECT content
            INTO latest_message_text
            FROM text_messages
            WHERE conversation_id = OLD.conversation_id AND message_id = latest_message_id;
        ELSE
            latest_message_text := NULL;
        END IF;

        UPDATE conversations
        SET
            last_message_id = latest_message_id,
            last_message_type = latest_message_type,
            last_message_text = latest_message_text,
            last_message_created_at = latest_message_created_at,
            updated_at = latest_message_created_at
        WHERE id = OLD.conversation_id;

    -- All message in this conversation was deleted
    ELSE

        UPDATE conversations
        SET
            last_message_id = NULL,
            last_message_type = NULL,
            last_message_text = NULL,
            last_message_created_at = NULL,
            updated_at = NOW()
        WHERE id = OLD.conversation_id;

    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_messages_refresh_last_message
AFTER DELETE ON messages
FOR EACH ROW
EXECUTE FUNCTION refresh_last_message_on_delete();

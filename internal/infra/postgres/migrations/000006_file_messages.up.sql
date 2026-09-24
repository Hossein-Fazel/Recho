ALTER TYPE message_type RENAME TO message_type_old;

CREATE TYPE message_type AS ENUM (
    'text',
    'file'
);

ALTER TABLE conversations
    ALTER COLUMN last_message_type TYPE message_type
    USING last_message_type::text::message_type;

ALTER TABLE messages
    ALTER COLUMN type TYPE message_type
    USING type::text::message_type;

DROP TYPE message_type_old;

CREATE TABLE file_messages (
    conversation_id UUID         NOT NULL,
    message_id      BIGINT       NOT NULL,
    file_key        TEXT         NOT NULL,
    category        TEXT         NOT NULL,
    content_type    TEXT         NOT NULL,
    size_bytes      BIGINT       NOT NULL,
    file_name       TEXT         NOT NULL DEFAULT '',
    caption         TEXT         NOT NULL DEFAULT '',

    PRIMARY KEY (conversation_id, message_id),
    FOREIGN KEY (conversation_id, message_id) REFERENCES messages(conversation_id, message_id) ON DELETE CASCADE,

    CONSTRAINT file_message_category_not_empty CHECK (length(trim(category)) > 0),
    CONSTRAINT file_message_content_type_not_empty CHECK (length(trim(content_type)) > 0),
    CONSTRAINT file_message_size_positive CHECK (size_bytes > 0),
    CONSTRAINT file_message_key_not_empty CHECK (length(trim(file_key)) > 0)
);

CREATE INDEX idx_file_messages_key ON file_messages(file_key);

CREATE OR REPLACE FUNCTION refresh_last_message_on_delete()
RETURNS TRIGGER AS $$
DECLARE
    latest_message_id BIGINT;
    latest_message_type message_type;
    latest_message_text TEXT;
    latest_message_created_at TIMESTAMPTZ;
BEGIN
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

    IF FOUND THEN
        IF latest_message_type = 'text' THEN
            SELECT content INTO latest_message_text
            FROM text_messages
            WHERE conversation_id = OLD.conversation_id AND message_id = latest_message_id;
        ELSIF latest_message_type = 'file' THEN
            SELECT caption INTO latest_message_text
            FROM file_messages
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

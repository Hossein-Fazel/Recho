DROP TABLE IF EXISTS file_messages;

ALTER TYPE message_type RENAME TO message_type_with_file;

CREATE TYPE message_type AS ENUM (
    'text'
);

ALTER TABLE conversations
    ALTER COLUMN last_message_type TYPE message_type
    USING last_message_type::text::message_type;

ALTER TABLE messages
    ALTER COLUMN type TYPE message_type
    USING type::text::message_type;

DROP TYPE message_type_with_file;

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

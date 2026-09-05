CREATE OR REPLACE FUNCTION refresh_last_message_on_delete()
RETURNS TRIGGER AS $$
DECLARE
    latest messages%ROWTYPE;
BEGIN
    IF OLD.message_id <> (SELECT last_message_id
                          FROM conversations
                          WHERE id = OLD.conversation_id) THEN
        RETURN OLD;
    END IF;

    SELECT message_id, content, created_at
    INTO latest.message_id, latest.content, latest.created_at
    FROM messages
    WHERE conversation_id = OLD.conversation_id
    ORDER BY message_id DESC
    LIMIT 1;

    -- FOUND is a special PL/pgSQL variable that is set to
    -- true if the previous `SELECT INTO` found a row and
    -- false otherwise.
    IF FOUND THEN
        UPDATE conversations
        SET last_message_id         = latest.message_id,
            last_message_content    = latest.content,
            last_message_created_at = latest.created_at,
            updated_at              = latest.created_at
        WHERE id = OLD.conversation_id;
    ELSE
        UPDATE conversations
        SET last_message_id         = NULL,
            last_message_content    = NULL,
            last_message_created_at = NULL
        WHERE id = OLD.conversation_id;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_messages_refresh_last_message
AFTER DELETE ON messages
FOR EACH ROW
EXECUTE FUNCTION refresh_last_message_on_delete();

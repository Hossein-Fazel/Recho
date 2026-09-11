DROP TRIGGER IF EXISTS trg_set_message_id ON messages;
DROP FUNCTION IF EXISTS set_message_id();
DROP FUNCTION IF EXISTS update_conversation_on_message(
    UUID,
    BIGINT,
    message_type,
    TEXT,
    TIMESTAMPTZ
);
DROP FUNCTION IF EXISTS refresh_last_message_on_delete();
DROP TABLE IF EXISTS text_message;
DROP TABLE IF EXISTS messages;
ALTER TABLE conversations
ADD COLUMN last_message_id UUID;

ALTER TABLE conversations
ADD COLUMN last_message_content TEXT;

ALTER TABLE conversations
ADD COLUMN last_message_created_at TIMESTAMPTZ;


CREATE INDEX idx_conversations_updated_at
ON conversations(updated_at DESC);

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
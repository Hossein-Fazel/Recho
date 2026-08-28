DROP INDEX IF EXISTS idx_conversations_updated_at;

ALTER TABLE conversations
DROP COLUMN IF EXISTS last_message_created_at;

ALTER TABLE conversations
DROP COLUMN IF EXISTS last_message_content;

ALTER TABLE conversations
DROP COLUMN IF EXISTS last_message_id;

DROP TRIGGER IF EXISTS trg_messages_update_conversation
ON messages;

DROP FUNCTION IF EXISTS update_conversation_on_message();
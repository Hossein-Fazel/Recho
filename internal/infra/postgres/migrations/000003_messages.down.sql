DROP TRIGGER IF EXISTS trg_set_message_id ON messages;
DROP TRIGGER IF EXISTS trg_messages_refresh_last_message ON messages;
DROP FUNCTION IF EXISTS set_message_id();
DROP FUNCTION IF EXISTS refresh_last_message_on_delete();
DROP TABLE IF EXISTS text_messages;
DROP TABLE IF EXISTS file_messages;
DROP TABLE IF EXISTS messages;

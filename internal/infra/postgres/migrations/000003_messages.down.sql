DROP TABLE messages;
DROP TRIGGER IF EXISTS trg_set_message_id ON messages;
DROP FUNCTION IF EXISTS set_message_id();
DROP TABLE group_members;
DROP TABLE groups;
DROP TABLE direct_conversations;
DROP TABLE conversations;
DROP TYPE group_member_role;
DROP TRIGGER IF EXISTS trg_messages_update_conversation ON messages;
DROP FUNCTION IF EXISTS update_conversation_on_message();
package models

import (
	"common/db"
	"common/log"
	"proto/chat"
)

//Migrate executes additional migrations
func Migrate() error {

	db.New().Model(&Member{}).
		AddForeignKey("conversation_id", "conversations(id)", "CASCADE", "RESTRICT")
	db.New().Model(&Message{}).
		AddForeignKey("conversation_id", "conversations(id)", "CASCADE", "RESTRICT")

	db.New().Model(&MessagePart{}).
		AddForeignKey("message_id", "messages(id)", "CASCADE", "RESTRICT")

	db.New().Exec("ALTER TABLE conversations ALTER COLUMN status SET DEFAULT 'new'")
	db.New().Model(&Conversation{}).Where("status IS NULL or status = ''").Update("status", "active")
	db.New().Exec("DROP INDEX IF EXISTS idx_messages_instagram_id")
	db.New().Exec("DROP INDEX IF EXISTS messages_instagram_id_key")
	db.New().Exec("DROP INDEX IF EXISTS unique_message_id")
	db.New().Exec("CREATE UNIQUE INDEX unique_message_per_conv_id ON messages(instagram_id, conversation_id) WHERE (instagram_id != '' AND instagram_id IS NOT NULL)")
	if db.HasColumn(&Conversation{}, "direct_sync") {
		tx := db.NewTransaction()
		tx.Model(&Conversation{}).Where("direct_sync").UpdateColumn("sync_status", chat.SyncStatus_SYNCED)
		tx.Model(&Conversation{}).Where("NOT direct_sync").UpdateColumn("sync_status", chat.SyncStatus_NONE)
		tx.Model(&Conversation{}).DropColumn("direct_sync")
		tx.Commit()
	}
	return membersKeys()
}

func membersKeys() error {
	if !db.HasColumn(&Member{}, "id") {
		return nil
	}
	tx := db.NewTransaction()
	log.Error(tx.Exec(
		"UPDATE messages msg SET user_id = mbr.user_id FROM members mbr WHERE mbr.id = msg.member_id AND (msg.user_id = 0 OR msg.user_id IS NULL)",
	).Error)
	log.Error(tx.Exec(
		"ALTER TABLE members DROP CONSTRAINT members_pkey CASCADE",
	).Error)
	log.Error(tx.Exec(
		"DROP INDEX IF EXISTS once_per_conv CASCADE",
	).Error)
	log.Error(tx.Exec(
		"ALTER TABLE members ADD PRIMARY KEY (conversation_id, user_id)",
	).Error)
	log.Error(tx.Model(&Member{}).AddIndex("idx_members_conversation_id", "conversation_id").Error)
	log.Error(tx.Model(&Member{}).AddIndex("idx_members_user_id", "user_id").Error)
	log.Error(tx.Model(&Member{}).DropColumn("id").Error)
	log.Error(tx.Model(&Message{}).DropColumn("member_id").Error)
	log.Error(tx.Exec(
		"ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_member_id_members_id_foreign",
	).Error)
	log.Error(tx.Exec(
		"ALTER TABLE messages ADD CONSTRAINT message_autor_foreign FOREIGN KEY (conversation_id, user_id) REFERENCES members (conversation_id, user_id) ON DELETE CASCADE",
	).Error)

	return tx.Commit().Error
}

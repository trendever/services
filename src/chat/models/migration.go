package models

import (
	"common/db"
	"proto/chat"
)

//Migrate executes additional migrations
func Migrate() error {

	db.New().Model(&Member{}).
		AddForeignKey("conversation_id", "conversations(id)", "CASCADE", "RESTRICT")
	db.New().Model(&Message{}).
		AddForeignKey("conversation_id", "conversations(id)", "CASCADE", "RESTRICT").
		AddForeignKey("member_id", "members(id)", "CASCADE", "RESTRICT")

	db.New().Model(&MessagePart{}).
		AddForeignKey("message_id", "messages(id)", "CASCADE", "RESTRICT")

	db.New().Exec("ALTER TABLE conversations ALTER COLUMN status SET DEFAULT 'new'")
	db.New().Model(&Conversation{}).Where("status IS NULL or status = ''").Update("status", "active")
	db.New().Model(&Member{}).AddUniqueIndex("once_per_conv", "user_id", "conversation_id")
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
	return nil
}

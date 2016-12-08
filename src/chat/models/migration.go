package models

import "github.com/jinzhu/gorm"

//Migrate executes additional migrations
func Migrate(db *gorm.DB) error {

	db.Model(&Member{}).
		AddForeignKey("conversation_id", "conversations(id)", "CASCADE", "RESTRICT")
	db.Model(&Message{}).
		AddForeignKey("conversation_id", "conversations(id)", "CASCADE", "RESTRICT").
		AddForeignKey("member_id", "members(id)", "CASCADE", "RESTRICT")

	db.Model(&MessagePart{}).
		AddForeignKey("message_id", "messages(id)", "CASCADE", "RESTRICT")

	db.Exec("ALTER TABLE conversations ALTER COLUMN status SET DEFAULT 'new'")
	db.Model(&Conversation{}).Where("status IS NULL or status = ''").Update("status", "active")
	db.New().Model(&Member{}).AddUniqueIndex("once_per_conv", "user_id", "conversation_id")
	db.New().Exec("DROP INDEX IF EXISTS idx_messages_instagram_id")
	db.New().Exec("DROP INDEX IF EXISTS messages_instagram_id_key")
	db.New().Exec("CREATE UNIQUE INDEX unique_message_id ON messages(instagram_id) WHERE (instagram_id != '' AND instagram_id IS NOT NULL)")
	return nil
}

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
	return nil
}

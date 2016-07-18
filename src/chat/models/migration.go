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
	return nil
}

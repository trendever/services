package models

import (
	"fetcher/conf"
	"proto/bot"
	"utils/db"
	"utils/log"
)

var modelsList = []interface{}{
	&Activity{},
	&ThreadInfo{},
	&DirectRequest{},
}

// old-style message types, only for migration
const (
	SendMessageRequest RequestType = iota
	ShareMediaRequest
	CreateThreadRequest
)

var typeMap = map[RequestType]bot.MessageType{
	SendMessageRequest:  bot.MessageType_Text,
	ShareMediaRequest:   bot.MessageType_MediaShare,
	CreateThreadRequest: bot.MessageType_CreateThread,
}

// AutoMigrate used models
func AutoMigrate(drop bool) error {
	// initialize database
	db.Init(&conf.GetSettings().DB)

	if drop {
		err := db.New().DropTableIfExists(modelsList...).Error
		if err != nil {
			return err
		}

		log.Warn("Drop Tables: success.")
	}

	err := db.New().AutoMigrate(modelsList...).Error
	if err != nil {
		return err
	}

	if db.HasColumn(&DirectRequest{}, "text") {
		db.New().Model(&DirectRequest{}).DropColumn("text")
	}
	if db.HasColumn(&DirectRequest{}, "type") {
		tx := db.NewTransaction()
		for old, cur := range typeMap {
			log.Error(tx.Model(&DirectRequest{}).Where("type = ?", old).UpdateColumn("kind", cur).Error)
		}
		err := tx.Commit().Error
		if err != nil {
			return err
		}
	}

	log.Info("Migration: success.")

	return nil
}

package main

import (
	"common/db"
	"utils/nats"
)

var settings = Settings{}

// Settings for this service
type Settings struct {
	DB             *db.Settings
	Debug          bool
	SentryDSN      string
	Listen         string
	InstagramDebug bool
	Nats           nats.Config
	RPC            struct {
		Core     string
		Telegram string
	}
	UserURLTemplate string
	DefaultProxy    string
}

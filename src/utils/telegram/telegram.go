package telegram

import (
	"github.com/tucnak/telebot"
)

var (
	bot  *telebot.Bot
	chat chatDestination
)

type chatDestination string

//Destination returns destination
func (chd chatDestination) Destination() string {
	return string(chd)
}

//Init initializes connection
func Init(token, chatRoom string) error {
	newbot, err := telebot.NewBot(token)
	if err != nil {
		return err
	}
	bot = newbot

	chat = chatDestination(chatRoom)

	return nil
}

//Notify sends a message to a chat
func Notify(msg string) error {
	return bot.SendMessage(chat, msg, nil)
}

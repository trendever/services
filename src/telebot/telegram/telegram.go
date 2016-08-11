package telegram

import (
	"fmt"
	"github.com/tucnak/telebot"
	"telebot/conf"
	"time"
)

// Telegram defines telegram sender
type Telegram struct {
	bot   *telebot.Bot
	chats map[string]chatDestination
}

type chatDestination string

func (dest chatDestination) Destination() string {
	return string(dest)
}

// Init initializes the bot
func Init(token string, rooms []conf.Room) (*Telegram, error) {
	bot, err := telebot.NewBot(token)
	if err != nil {
		return nil, err
	}

	telegram := &Telegram{bot: bot, chats: make(map[string]chatDestination)}

	for _, room := range rooms {
		telegram.chats[room.Name] = chatDestination(room.Room)
	}

	return telegram, nil
}

// Notify sends message to chat named dst
func (t *Telegram) Notify(dst, msg string) error {
	chat, ok := t.chats[dst]
	if !ok {
		return fmt.Errorf("Chat %v not defined in config", dst)
	}

	var (
		err     error
		retries = 15
	)

	for {
		err = t.bot.SendMessage(chat, msg, nil)
		if err != nil {
			retries--
			if retries == 0 {
				return err
			}
			time.Sleep(time.Second * 300)
		} else {
			return nil
		}
	}

}

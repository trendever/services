package telegram

import (
	"github.com/tucnak/telebot"
	"telebot/conf"
	"time"
	"utils/log"
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
func (t *Telegram) Notify(dst, msg string) {
	chat, ok := t.chats[dst]
	if !ok {
		log.Errorf("Chat %v not defined in config", dst)
		return
	}

	var (
		err     error
		retries = 15
	)

	for {
		err = t.bot.SendMessage(chat, msg, nil)
		if err != nil {
			log.Errorf("failed to send message to channel '%v': %+v", chat, err)
			retries--
			if retries == 0 {
				log.Errorf("to many errors in row, message dropped")
				return
			}
			time.Sleep(time.Second * 300)
		} else {
			return
		}
	}

}

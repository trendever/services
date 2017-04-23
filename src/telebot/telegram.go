package main

import (
	"github.com/tucnak/telebot"
	"strings"
	"time"
	"utils/log"
)

func init() {
	RegisterHandler("/start", helpHandler)
	RegisterHandler("/help", helpHandler)
}

// Telegram defines telegram sender
type Telegram struct {
	bot   *telebot.Bot
	chats map[string]chatDestination
}

type chatDestination string

func (dest chatDestination) Destination() string {
	return string(dest)
}

type messageHandler func(bot *telebot.Bot, msg *telebot.Message)

var handlers = map[string]messageHandler{}

func RegisterHandler(name string, handler messageHandler) {
	handlers[name] = handler
}

func InitBot(token string, rooms []Room) (*Telegram, error) {
	bot, err := telebot.NewBot(token)
	if err != nil {
		return nil, err
	}

	telegram := &Telegram{bot: bot, chats: make(map[string]chatDestination)}

	for _, room := range rooms {
		telegram.chats[room.Name] = chatDestination(room.Room)
	}

	go telegram.Listen()

	return telegram, nil
}

func helpHandler(bot *telebot.Bot, msg *telebot.Message) {
	log.Error(bot.SendMessage(msg.Chat, settings.Messages.Help, nil))
}

func (t *Telegram) Listen() {
	messages := make(chan telebot.Message, 100)
	t.bot.Listen(messages, 1*time.Second)

	for message := range messages {
		if !message.IsPersonal() {
			continue
		}
		split := strings.SplitAfterN(message.Text, " ", 2)
		log.Debug("%v from %v", split[0], message.Chat.Username)

		handler, ok := handlers[split[0]]
		if !ok {
			helpHandler(t.bot, &message)
			continue
		}
		handler(t.bot, &message)
	}
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

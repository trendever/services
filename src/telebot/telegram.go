package main

import (
	"fmt"
	"github.com/tucnak/telebot"
	"proto/telegram"
	"strconv"
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
func (t *Telegram) Notify(req *telegram.NotifyMessageRequest) (err error, retry bool) {
	var dest chatDestination
	if req.ChatId != 0 {
		dest = chatDestination(strconv.FormatUint(req.ChatId, 10))
	} else {
		var ok bool
		dest, ok = t.chats[req.Channel]
		if !ok {
			err := fmt.Errorf("chat %v not defined in config", req.Channel)
			log.Error(err)
			return err, false
		}
	}
	err = t.bot.SendMessage(dest, req.Message, nil)
	if err == nil {
		return nil, false
	}
	// ugly check, but probably it is shortest way to determinate kind of error without changes in telebot lib
	return err, !strings.HasPrefix(err.Error(), "telebot:")
}

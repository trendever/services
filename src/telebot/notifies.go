package main

import (
	"github.com/tucnak/telebot"
	"proto/core"
	"strings"
	"utils/log"
	"utils/rpc"
)

func init() {
	RegisterHandler("/subscribe", subscribeHandler)
	RegisterHandler("/unsubscribe", unsubscribeHandler)
}

func subscribeHandler(bot *telebot.Bot, msg *telebot.Message) {
	split := strings.Split(msg.Text, " ")
	if len(split) != 1 {
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.Help, nil))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	reply, err := userServer.AddTelegram(ctx, &core.AddTelegramRequest{
		Username:       split[0],
		ChatId:         uint64(msg.Chat.ID),
		SubsricberName: msg.Chat.Username,
	})
	if err != nil {
		log.Errorf("failed to add telegram: %v", err)
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.ExternalError, nil))
		return
	}
	switch reply.Error {
	case "":
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.Subscribed, nil))
	case "user not found":
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.UserNotFound, nil))
	default:
		log.Errorf("failed to add telegram: %v", reply.Error)
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.ExternalError, nil))
	}
}

func unsubscribeHandler(bot *telebot.Bot, msg *telebot.Message) {
	split := strings.SplitAfter(msg.Text, " ")
	if len(split) != 2 {
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.Help, nil))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	reply, err := userServer.DelTelegram(ctx, &core.DelTelegramRequest{
		Username: split[1],
		ChatId:   uint64(msg.Chat.ID),
	})
	if err != nil {
		log.Errorf("failed to del telegram: %v", err)
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.ExternalError, nil))
		return
	}
	switch reply.Error {
	case "":
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.Unsubscribed, nil))
	case "user not found":
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.UserNotFound, nil))
	default:
		log.Errorf("failed to del telegram: %v", reply.Error)
		log.Error(bot.SendMessage(msg.Chat, settings.Messages.ExternalError, nil))
	}
}

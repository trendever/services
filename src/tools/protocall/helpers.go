package main

// This file is auto-generated
// See generate-helpers.sh for the reference
// Do not try to edit this manually

import "proto/auth"
import "proto/bot"
import "proto/chat"
import "proto/checker"
import "proto/core"
import "proto/mail"
import "proto/payment"
import "proto/push"
import "proto/sms"
import "proto/trendcoin"

var services map[string]interface{}

func connect() {
	services = map[string]interface{}{

		"AuthService":       auth.NewAuthServiceClient(conn),
		"FetcherService":    bot.NewFetcherServiceClient(conn),
		"SaveTrendService":  bot.NewSaveTrendServiceClient(conn),
		"TelegramService":   bot.NewTelegramServiceClient(conn),
		"ChatService":       chat.NewChatServiceClient(conn),
		"Notifier":          chat.NewNotifierClient(conn),
		"CheckerService":    checker.NewCheckerServiceClient(conn),
		"LeadService":       core.NewLeadServiceClient(conn),
		"ProductService":    core.NewProductServiceClient(conn),
		"PushTokensService": core.NewPushTokensServiceClient(conn),
		"ShopCardService":   core.NewShopCardServiceClient(conn),
		"ShopService":       core.NewShopServiceClient(conn),
		"TagService":        core.NewTagServiceClient(conn),
		"UserService":       core.NewUserServiceClient(conn),
		"MailService":       mail.NewMailServiceClient(conn),
		"PaymentService":    payment.NewPaymentServiceClient(conn),
		"PushService":       push.NewPushServiceClient(conn),
		"SmsService":        sms.NewSmsServiceClient(conn),
		"TrendcoinService":  trendcoin.NewTrendcoinServiceClient(conn),
	}
}

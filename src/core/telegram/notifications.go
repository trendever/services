package telegram

import (
	"core/api"
	"core/conf"
	"core/models"
	"fmt"
)

func init() {
	models.NotifyUserCreated = NotifyUserCreated
}

// NotifyUserCreated notifies about user creation
func NotifyUserCreated(u *models.User) {

	api.NotifyByTelegram(api.TelegramChannelNewUser,
		fmt.Sprintf(
			`New user %v registered
			%v`,
			u.Stringify(),
			fmt.Sprintf("%v/qor/users/%v", conf.GetSettings().SiteURL, u.ID),
		),
	)
}

// NotifyProductCreated notifies about product creation
func NotifyProductCreated(p *models.Product) {

	api.NotifyByTelegram(api.TelegramChannelNewProduct,
		fmt.Sprintf(
			"%v added %v by %v\n"+ // [scout] added [product_code] by [shop]
				"%v\n"+ // [instagram_link]
				"%v", // [qor_link]
			p.MentionedBy.Stringify(), p.Code, p.Shop.Stringify(),
			p.InstagramLink,
			fmt.Sprintf("%v/qor/products/%v", conf.GetSettings().SiteURL, p.ID),
		),
	)
}

// NotifyLeadCreated notifies about lead creation
func NotifyLeadCreated(l *models.Lead, p *models.Product, realInstLink string) {

	if p.Shop.ID == 0 && p.ShopID > 0 {
		if shop, err := models.GetShopByID(p.ShopID); err == nil {
			p.Shop = *shop
		}
	}

	api.NotifyByTelegram(api.TelegramChannelNewLead,
		fmt.Sprintf(
			"%v ordered %v by %v from %v\n, comment: '%v'\n"+ // [client] ordered [product_code] in [shop] from [wantit or website] comment: '[comment]'
				"%v\n"+ // [website_link]
				"%v\n"+ // [instgram_repost_link]
				"%v", // [qor_link]
			// first line
			l.Customer.Stringify(),
			p.Code,
			p.Shop.Stringify(),
			l.Source,
			l.Comment,
			// the rest
			fmt.Sprintf("%v/chat/%v", conf.GetSettings().SiteURL, l.ID),
			realInstLink,
			fmt.Sprintf("%v/qor/orders/%v", conf.GetSettings().SiteURL, l.ID),
		),
	)
}

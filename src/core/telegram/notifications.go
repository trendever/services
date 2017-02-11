package telegram

import (
	"core/api"
	"core/conf"
	"core/models"
	"fmt"
	"proto/core"
)

func init() {
	models.NotifyUserCreated = NotifyUserCreated
}

// NotifyUserCreated notifies about user creation
func NotifyUserCreated(u *models.User) {

	api.NotifyByTelegram(api.TelegramChannelNewUser,
		fmt.Sprintf(
			`New user %v registered
			%v
			#%v`,
			u.Stringify(),
			fmt.Sprintf("%v/qor/users/%v", conf.GetSettings().SiteURL, u.ID),
			u.Source,
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

var actionText = map[core.LeadAction]string{
	core.LeadAction_BUY:  "ordered",
	core.LeadAction_INFO: "requested info about",
	core.LeadAction_SKIP: "skiped product",
}

// NotifyLeadCreated notifies about lead creation
func NotifyLeadCreated(l *models.Lead, p *models.Product, realInstLink string, action core.LeadAction) {

	if p.Shop.ID == 0 && p.ShopID > 0 {
		if shop, err := models.GetShopByID(p.ShopID); err == nil {
			p.Shop = *shop
		}
	}
	api.NotifyByTelegram(api.TelegramChannelNewLead,
		fmt.Sprintf(
			"%v %v %v by %v from %v, comment: '%v'\n"+ // [client] [action] [product_code] in [shop] from [wantit or website] comment: '[comment]'
				"%v\n"+ // [website_link]
				"%v\n"+ // [instgram_repost_link]
				"%v\n"+ // [qor_link]
				"#%v", // [tag for search]
			// first line
			l.Customer.Stringify(),
			actionText[action],
			p.Code,
			p.Shop.Stringify(),
			l.Source,
			l.Comment,
			// the rest
			fmt.Sprintf("%v/chat/%v", conf.GetSettings().SiteURL, l.ID),
			realInstLink,
			fmt.Sprintf("%v/qor/orders/%v", conf.GetSettings().SiteURL, l.ID),
			actionText[action],
		),
	)
}

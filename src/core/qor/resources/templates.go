package resources

import (
	"core/models"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"strconv"
)

func init() {
	addResource(&models.SMSTemplate{}, &admin.Config{
		Name: "SMS Templates",
		Menu: []string{"Settings"},
	}, initSMSTemplateResource)
	addResource(&models.EmailTemplate{}, &admin.Config{
		Name: "Email templates",
		Menu: []string{"Settings"},
	}, initEmailTemplateResource)
	addResource(&models.PushTemplate{}, &admin.Config{
		Name: "Push templates",
		Menu: []string{"Settings"},
	}, initPushTemplateResource)
	addResource(&models.ChatTemplate{}, &admin.Config{
		Name: "Chat templates",
		Menu: []string{"Settings"},
	}, initChatTemplateResource)
}

func initSMSTemplateResource(sms *admin.Resource) {
	sms.Meta(&admin.Meta{
		Name:       "TemplateID",
		Type:       "select_one",
		Collection: models.TemplatesList["sms"],
	})
	sms.Meta(&admin.Meta{
		Name: "Message",
		Type: "text",
	})
	sms.IndexAttrs(
		"TemplateID", "TemplateName",
	)
	sms.SearchAttrs(
		"TemplateName", "TemplateID",
	)
	attrs := []*admin.Section{
		{
			Title: "Template settings",
			Rows: [][]string{
				{"TemplateID", "TemplateName"},
			},
		},
		{
			Title: "Message",
			Rows: [][]string{
				{"Message"},
			},
		},
	}
	sms.NewAttrs(attrs)
	sms.EditAttrs(attrs)
}

func initEmailTemplateResource(email *admin.Resource) {
	// body textarea
	email.Meta(&admin.Meta{
		Name: "Body",
		Type: "rich_editor",
	})
	email.Meta(&admin.Meta{
		Name:       "TemplateID",
		Type:       "select_one",
		Collection: models.TemplatesList["email"],
	})
	email.IndexAttrs(
		"TemplateID", "TemplateName", "Subject",
	)
	email.SearchAttrs(
		"TemplateName", "TemplateName", "Subject",
	)
	attrs := []*admin.Section{
		{
			Title: "Template settings",
			Rows: [][]string{
				{"TemplateID", "TemplateName"},
			},
		},
		{
			Title: "Message",
			Rows: [][]string{
				{"From", "Subject"},
				{"Body"},
			},
		},
	}
	email.NewAttrs(attrs)
	email.EditAttrs(attrs)
}

func initPushTemplateResource(push *admin.Resource) {
	push.Meta(&admin.Meta{
		Name:       "TemplateID",
		Type:       "select_one",
		Collection: models.TemplatesList["push"],
	})
	push.Meta(&admin.Meta{
		Name: "Body",
		Type: "text",
	})
	push.IndexAttrs(
		"TemplateID", "TemplateName", "Title",
	)
	push.SearchAttrs(
		"TemplateName", "TemplateID",
	)
	attrs := []*admin.Section{
		{
			Title: "Template settings",
			Rows: [][]string{
				{"TemplateID", "TemplateName"},
			},
		},
		{
			Title: "Payload",
			Rows: [][]string{
				{"Title"},
				{"Body"},
			},
		},
	}
	push.NewAttrs(attrs)
	push.EditAttrs(attrs)
}

func initChatTemplateResource(chat *admin.Resource) {
	chat.Meta(&admin.Meta{
		Name:       "Group",
		Type:       "select_one",
		Collection: models.TemplatesList["chat"],
	})
	chat.Meta(&admin.Meta{
		Name: "Product",
		Config: &admin.SelectOneConfig{
			AllowBlank: true,
			Collection: func(this interface{}, ctx *qor.Context) (results [][]string) {
				var res []models.Product
				ctx.GetDB().
					Joins("LEFT JOIN products_shops ON products_product.shop_id = products_shops.id").
					Where("products_shops.supplier_id = ?", models.SystemUser.ID).
					Find(&res)
				for _, p := range res {
					results = append(
						results,
						[]string{strconv.FormatUint(uint64(p.ID), 10), p.Stringify()},
					)
				}
				return
			},
		},
	})

	chat.Scope(&admin.Scope{
		Name:  "Default",
		Group: "Scope",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_default = ?", true)
		},
	})
	chat.Scope(&admin.Scope{
		Name:  "Specific",
		Group: "Scope",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_default = ?", false)
		},
	})

	chat.IndexAttrs("TemplateName", "Group", "IsDefault", "Product")
	chat.SearchAttrs("TemplateName", "Group", "IsDefault", "ProductID")

	attrs := []*admin.Section{
		{
			Rows: [][]string{
				{"TemplateName", "Group"},
				{"IsDefault", "Product"},
				{"Cases"},
			},
		},
	}
	chat.NewAttrs(attrs)
	chat.EditAttrs(attrs)

	caseRes := chat.Meta(&admin.Meta{Name: "Cases"}).Resource
	caseRes.Meta(&admin.Meta{
		Name:       "Source",
		Type:       "select_one",
		Collection: models.LeadSources,
	})
	caseRes.Meta(&admin.Meta{
		Name:      "IfSupplierIsRegistered",
		FieldName: "ForSuppliersWithNotices",
	})
	attrs = []*admin.Section{
		{
			Rows: [][]string{
				{"Source"},
				{"ForNewUsers", "ForSuppliersWithNotices"},
				{"Messages"},
			},
		},
	}
	caseRes.NewAttrs(attrs)
	caseRes.EditAttrs(attrs)

	msgRes := caseRes.Meta(&admin.Meta{Name: "Messages"}).Resource
	msgRes.Meta(&admin.Meta{Name: "Text", Type: "text"})
}

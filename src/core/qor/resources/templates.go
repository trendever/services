package resources

import (
	"core/models"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

func init() {
	addOnQorInitCallback(addTemplateResource)
}

func addTemplateResource(a *admin.Admin) {

	sms := a.AddResource(&models.SMSTemplate{}, &admin.Config{
		Name: "SMS Templates",
		Menu: []string{"Settings"},
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

	email := a.AddResource(&models.EmailTemplate{}, &admin.Config{
		Name: "Email templates",
		Menu: []string{"Settings"},
	})
	// body textarea
	email.Meta(&admin.Meta{
		Name: "Body",
		Type: "rich_editor",
	})
	email.IndexAttrs(
		"TemplateID", "TemplateName", "Subject",
	)
	email.SearchAttrs(
		"TemplateName", "TemplateName", "Subject",
	)
	attrs = []*admin.Section{
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

	chat := a.AddResource(&models.ChatTemplate{}, &admin.Config{
		Name: "Chat templates",
		Menu: []string{"Settings"},
	})
	chat.Meta(&admin.Meta{
		Name: "Message",
		Type: "text",
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

	chat.IndexAttrs(
		"Group", "Order", "TemplateName",
		"OnAction", "IsDefault", "Product",
	)
	chat.SearchAttrs(
		"TemplateName", "Group", "IsDefault", "ProductID", "Message",
	)

	attrs = []*admin.Section{
		{
			Title: "Template settings",
			Rows: [][]string{
				{"TemplateName"},
				{"Group", "Order"},
				{"IsDefault", "Product"},
			},
		},
		{
			Title: "Message",
			Rows: [][]string{
				{"Message"},
			},
		},
	}
	chat.NewAttrs(attrs)
	chat.EditAttrs(attrs)
}

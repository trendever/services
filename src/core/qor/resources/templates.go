package resources

import (
	"core/models"
	"github.com/qor/admin"
)

func init() {
	addOnQorInitCallback(addTemplateResource)
}

func addTemplateResource(a *admin.Admin) {

	sms := a.AddResource(&models.SMSTemplate{}, &admin.Config{
		Name: "SMS Templates",
		Menu: []string{"Settings"},
	})

	email := a.AddResource(&models.EmailTemplate{}, &admin.Config{
		Name: "Email templates",
		Menu: []string{"Settings"},
	})

	all := map[string]*admin.Resource{
		"SMS":   sms,
		"Email": email,
	}

	// all available models
	var collection [][]string
	for _, model := range models.TemplateModels {
		collection = append(collection, []string{model, model})
	}

	// body textarea
	email.Meta(&admin.Meta{
		Name: "Body",
		Type: "rich_editor",
	})

	sms.Meta(&admin.Meta{
		Name: "Message",
		Type: "text",
	})

	for _, tpl := range all {

		// appliable model
		tpl.Meta(&admin.Meta{
			Name:       "ModelName",
			Type:       "select_one",
			Collection: collection,
		})

		tpl.IndexAttrs(
			"TemplateID", "TemplateName", "ModelName", "Subject",
		)

		tpl.EditAttrs(tpl.NewAttrs())
	}

	sms.SearchAttrs(
		"TemplateName", "Subject", "To", "Message",
	)

	email.SearchAttrs(
		"TemplateName", "Subject", "Body", "From", "To",
	)

	email.NewAttrs(
		&admin.Section{
			Title: "Template settings",
			Rows: [][]string{
				{"TemplateID"},
				{"TemplateName", "ModelName"},
				{"Preloads"},
			}},
		&admin.Section{
			Title: "Message",
			Rows: [][]string{
				{"From", "To"},
				{"Subject"},
				{"Body"},
			}},
	)

	sms.NewAttrs(
		&admin.Section{
			Title: "Template settings",
			Rows: [][]string{
				{"TemplateID"},
				{"TemplateName", "ModelName"},
				{"Preloads"},
			}},
		&admin.Section{
			Title: "Message",
			Rows: [][]string{
				{"To"},
				{"Message"},
			}},
	)

}

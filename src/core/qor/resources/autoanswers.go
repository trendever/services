package resources

import (
	"core/models"
	"github.com/qor/admin"
)

func init() {
	addResource(&models.AutoAnswer{}, &admin.Config{
		Name: "AutoAnswers",
		Menu: []string{"Settings"},
	}, answersInit)
}

func answersInit(res *admin.Resource) {
	res.SearchAttrs(
		"Name",
	)
	res.IndexAttrs(
		"ID", "Name", "Language",
	)
	res.Meta(&admin.Meta{
		Name:       "Language",
		Type:       "select_one",
		Collection: models.AnswersSupportedLanguages,
	})
	attrs := &admin.Section{
		Rows: [][]string{
			{"Name", "Language"},
			{"Dictionary"},
			{"Template"},
		},
	}
	res.NewAttrs(attrs)
	res.EditAttrs(attrs)
}

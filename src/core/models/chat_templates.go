package models

import (
	"database/sql"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/jinzhu/gorm"
	"github.com/qor/sorting"
	"github.com/qor/validations"
	"proto/chat"
	"strings"
)

type ChatTemplate struct {
	ID uint `gorm:"primary_key"`

	// Template fields
	TemplateName string
	// We want to send more then one message per action
	// so chat templates have group instead of TemplateID
	Group string
	// Specific product(if any)
	ProductID sql.NullInt64
	Product   Product
	// When true this template will be used if there is no specific templates
	IsDefault bool

	Cases []ChatTemplateCase `gorm:"ForeignKey:TemplateID"`
}

type ChatTemplateCase struct {
	ID         uint `gorm:"primary_key"`
	TemplateID uint `gorm:"index"`
	// lead source with which this template can be used
	Source                  string
	ForNewUsers             bool
	ForSuppliersWithNotices bool

	Messages []ChatTemplateMessage `gorm:"ForeignKey:CaseID"`
}

type ChatTemplateMessage struct {
	ID     uint `gorm:"primary_key"`
	CaseID uint `gorm:"index"`
	sorting.Sorting
	Text string `gorm:"type:text"`
	Data string `gorm:"type:text"`
}

// Validate fields
func (t ChatTemplate) Validate(db *gorm.DB) {
	if t.TemplateName == "" {
		db.AddError(validations.NewError(t, "TemplateName", "Template name can not be empty"))
	}
	if !templateIDRegexp.MatchString(t.Group) {
		db.AddError(validations.NewError(t, "Group", "Incorrect template group"))
	}

	if t.IsDefault && t.Product.ID != 0 {
		db.AddError(validations.NewError(
			t, "ProductID", "Default templates should not be product-specific",
		))
	}
	if !t.IsDefault && t.Product.ID == 0 {
		db.AddError(validations.NewError(
			t, "ProductID", "Non-default templates should be specific for product",
		))
	}
}

func (c ChatTemplateCase) Validate(db *gorm.DB) {
	knownSource := false
	for _, s := range LeadSources {
		if s == c.Source {
			knownSource = true
			break
		}
	}
	if !knownSource {
		db.AddError(validations.NewError(c, "Source", "Unknown source"))
	}

	var tmp ChatTemplateCase
	ret := db.
		Where("source = ?", c.Source).
		Where("for_new_users = ?", c.ForNewUsers).
		Where("for_suppliers_with_notices = ?", c.ForSuppliersWithNotices).
		Where("template_id = ?", c.TemplateID).
		Where("id <> ?", c.ID).
		First(&tmp)
	if !ret.RecordNotFound() {
		db.AddError(validations.NewError(c, "", "Identical cases detected"))
	}
}

func (m ChatTemplateMessage) Validate(db *gorm.DB) {
	if strings.Trim(m.Text, " \t\n") == "" {
		db.AddError(validations.NewError(m, "Text", "blank message text"))
	}
	_, err := pongo2.FromString(m.Text)
	if err != nil {
		db.AddError(validations.NewError(
			m,
			"Text",
			fmt.Sprintf("failed to compile template: %v", err),
		))
	}
	_, err = pongo2.FromString(m.Data)
	if err != nil {
		db.AddError(validations.NewError(
			m,
			"ImageURL",
			fmt.Sprintf("failed to compile template: %v", err),
		))
	}
}

// Execute returns ready-to-use message parts([]*chat.MessagePart)
func (t *ChatTemplateMessage) Execute(ctx interface{}) (interface{}, error) {
	text, err := applyTemplate(t.Text, ctx)
	if err != nil {
		return nil, err
	}
	data, err := applyTemplate(t.Data, ctx)
	if err != nil {
		return nil, err
	}
	parts := make([]*chat.MessagePart, 0, 2)
	text = strings.Trim(text, " \t\n")
	if text != "" {
		parts = append(parts, &chat.MessagePart{
			Content:  text,
			MimeType: "text/plain",
		})
	}
	data = strings.Trim(data, " \t\n")
	if data != "" {
		parts = append(parts, &chat.MessagePart{
			Content:  data,
			MimeType: "text/data",
		})
	}
	return parts, nil
}

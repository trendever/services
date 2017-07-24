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
	ID uint64 `gorm:"primary_key"`

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
	// Specific lead source or 'any'
	Source string

	Messages       []ChatTemplateMessage `gorm:"ForeignKey:TemplateID"`
	MessagesSorter sorting.SortableCollection
}

type ChatTemplateMessage struct {
	ID         uint `gorm:"primary_key"`
	TemplateID uint `gorm:"index"`
	sorting.Sorting
	Text string `gorm:"type:text"`
	Data string `gorm:"type:text"`
}

// Validate fields
func (t ChatTemplate) Validate(db *gorm.DB) {
	if t.TemplateName == "" {
		db.AddError(validations.NewError(t, "TemplateName", "Template name can not be empty"))
	}

	var ok bool
	for _, group := range TemplatesList["chat"] {
		if t.Group == group {
			ok = true
			break
		}
	}
	if !ok {
		db.AddError(validations.NewError(t, "Group", "Unknown template group"))
	}

	scope := db.New().Model(&ChatTemplate{}).Where(`"group" = ?`, t.Group).Where("id != ?", t.ID)
	if t.Product.ID != 0 {
		scope = scope.Where("product_id = ?", t.Product.ID)
	} else {
		scope = scope.Where("product_id IS NULL")
	}
	var count uint
	scope.Count(&count)
	if count > 0 {
		db.AddError(validations.NewError(
			t, "ProductID", "Template with this group and prodcut already exists",
		))
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

func (m ChatTemplateMessage) Validate(db *gorm.DB) {
	if strings.Trim(m.Text, " \t\r\n") == "" {
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
			"Data",
			fmt.Sprintf("failed to compile template: %v", err),
		))
	}
}

// Execute returns ready-to-use message parts([]*chat.MessagePart)
func (t *ChatTemplateMessage) Execute(ctx interface{}) (interface{}, error) {
	text, err := applyTemplate(t.Text, ctx, false)
	if err != nil {
		return nil, err
	}
	data, err := applyTemplate(t.Data, ctx, false)
	if err != nil {
		return nil, err
	}
	parts := make([]*chat.MessagePart, 0, 2)
	if text != "" {
		parts = append(parts, &chat.MessagePart{
			Content:  text,
			MimeType: "text/plain",
		})
	}
	if data != "" {
		parts = append(parts, &chat.MessagePart{
			Content:  data,
			MimeType: "text/data",
		})
	}
	return parts, nil
}

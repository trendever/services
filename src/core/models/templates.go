package models

import (
	"core/conf"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"gopkg.in/flosch/pongo2.v3"
	"regexp"
	"utils/log"
)

// TemplateModels contains all other models that can be passed to a template
var TemplateModels = []string{
	"Lead",
}

// TemplateTypes contains types that correspond to send action types
var TemplateTypes = []string{
	"SMS",
	"Email",
	"Chat",
}

// TemplateTypes contains types that correspond to available send actions

// templateIDRegexp is used to match only correct templates: alphanumeric symbols and _
var templateIDRegexp = regexp.MustCompile("^[[:word:]]+$")

// BaseTemplate model
type BaseNotifierTemplate struct {
	gorm.Model

	// Template fields
	TemplateName string
	TemplateID   string `gorm:"unique_index"`
}

// EmailTemplate object
type EmailTemplate struct {
	BaseNotifierTemplate
	EmailMessage
}

// EmailMessage object
type EmailMessage struct {
	From    string `gorm:"type:text"` // For example: "Hello trendever <hello@trendever.com>"
	Subject string `gorm:"type:text"`
	Body    string `gorm:"type:text"`
}

// SMSTemplate object
type SMSTemplate struct {
	BaseNotifierTemplate
	Message string `gorm:"type:text"`
}

type ChatTemplate struct {
	gorm.Model

	// Template fields
	TemplateName string
	// We want to send more then one message per action
	// so chat templates have group instead of TemplateID
	Group string
	// Order of sending if there is more then one template with same id
	Order   int64
	Message string `gorm:"type:text"`
	// Specific product(if any)
	ProductID sql.NullInt64
	Product   Product
	// When true this template will be used if there is no specific templates
	IsDefault               bool
	ForSuppliersWithNotices bool
}

// Template is a common interface for all the models
type Template interface {
	Execute(interface{}) (interface{}, error)
}

//GetFrom returns from
func (em *EmailMessage) GetFrom() string {
	return em.From
}

//GetSubject returns subject
func (em *EmailMessage) GetSubject() string {
	return em.Subject
}

//GetMessage returns message
func (em *EmailMessage) GetMessage() string {
	return em.Body
}

// TableName for gorm
func (t EmailTemplate) TableName() string {
	return "settings_templates_email"
}

// TableName for gorm
func (t SMSTemplate) TableName() string {
	return "settings_templates_sms"
}

// TableName for gorm
func (t ChatTemplate) TableName() string {
	return "settings_templates_chat"
}

// Validate fields
func (t BaseNotifierTemplate) Validate(db *gorm.DB) {
	if t.TemplateName == "" {
		db.AddError(validations.NewError(t, "TemplateName", "Template name can not be empty"))
	}

	if !templateIDRegexp.MatchString(t.TemplateID) {
		db.AddError(validations.NewError(t, "TemplateID", "Incorrect template id"))
	}
}

// Validate fields
func (t EmailTemplate) Validate(db *gorm.DB) {
	t.BaseNotifierTemplate.Validate(db)
	sources := map[string]string{
		"From":    t.From,
		"Subject": t.Subject,
		"Body":    t.Body,
	}
	for column, str := range sources {
		_, err := pongo2.FromString(str)
		if err != nil {
			db.AddError(validations.NewError(
				t,
				column,
				fmt.Sprintf("Failed to compile template: %v", err),
			))
		}
	}
}

// Validate fields
func (t SMSTemplate) Validate(db *gorm.DB) {
	t.BaseNotifierTemplate.Validate(db)
	_, err := pongo2.FromString(t.Message)
	if err != nil {
		db.AddError(validations.NewError(
			t,
			"Message",
			fmt.Sprintf("Failed to compile template: %v", err),
		))
	}
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

// Execute returns MessageFields object with ready-to-use fields
func (t *EmailTemplate) Execute(ctx interface{}) (interface{}, error) {

	subject, err := applyTemplate(t.Subject, ctx)
	if err != nil {
		return nil, err
	}

	body, err := applyTemplate(t.Body, ctx)
	if err != nil {
		return nil, err
	}

	from, err := applyTemplate(t.From, ctx)
	if err != nil {
		return nil, err
	}

	return &EmailMessage{
		Subject: subject,
		Body:    body,
		From:    from,
	}, nil
}

// Execute returns ready-to-use message text
func (t *SMSTemplate) Execute(ctx interface{}) (interface{}, error) {
	return applyTemplate(t.Message, ctx)
}

// Execute returns ready-to-use message text
func (t *ChatTemplate) Execute(ctx interface{}) (interface{}, error) {
	return applyTemplate(t.Message, ctx)
}

// applyTemplate applies template from string to ctx and returns result
func applyTemplate(template string, ctx interface{}) (string, error) {
	tmpl, err := pongo2.FromString(template)
	if err != nil {
		return "", err
	}

	out, err := tmpl.Execute(pongo2.Context{
		"object":   ctx,
		"settings": conf.GetSettings(),
	})
	if err != nil {
		return "", err
	}

	return out, nil
}

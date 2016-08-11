package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"gopkg.in/flosch/pongo2.v3"
	"regexp"
	"strings"
)

// Template is a common interface for all the models
type Template interface {
	Execute(interface{}) (interface{}, error)
}

// domain -> []group/id
var TemplatesList = map[string][]string{}

func RegisterTemplate(domain, name string) error {
	if domain == "" || name == "" {
		return errors.New("domain and name of template should not be empty")
	}
	sub, ok := TemplatesList[domain]
	if !ok {
		TemplatesList[domain] = []string{name}
		return nil
	}
	for _, t := range sub {
		if name == t {
			return fmt.Errorf("template %v:%v alteady registred", domain, name)
		}
	}
	TemplatesList[domain] = append(sub, name)
	return nil
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

// applyTemplate applies template from string to ctx and returns result
func applyTemplate(template string, ctx interface{}) (string, error) {
	tmpl, err := pongo2.FromString(template)
	if err != nil {
		return "", err
	}

	arg, ok := ctx.(map[string]interface{})
	if !ok {
		arg = map[string]interface{}{"object": ctx}
	}
	out, err := tmpl.Execute(pongo2.Context(arg))
	if err != nil {
		return "", err
	}
	return strings.Trim(out, " \t\n"), nil
}

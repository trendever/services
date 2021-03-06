package models

import (
	"common/db"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"gopkg.in/flosch/pongo2.v3"
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

var notifyDomains = []string{"email", "sms", "telegram"}

func RegisterNotifyTemplates(names ...string) error {
	for _, name := range names {
		for _, domain := range notifyDomains {
			err := RegisterTemplate(domain, name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// @TODO too many types with minor differences, refactor it somehow?

// TemplateTypes contains types that correspond to available send actions

// templateIDRegexp is used to match only correct templates: alphanumeric symbols and _
var templateIDRegexp = regexp.MustCompile("^[[:word:]]+$")

// BaseTemplate model
type BaseNotifierTemplate struct {
	ID uint64 `gorm:"primary_key"`
	// @CHECK for what do we need name if TemplateID is unique?
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

type TelegramTemplate struct {
	BaseNotifierTemplate
	Message string `gorm:"type:text"`
}

// generic templates
type OtherTemplate struct {
	BaseNotifierTemplate
	Text string `gorm:"type:text"`
}

// TableName for gorm
func (t EmailTemplate) TableName() string {
	return "settings_templates_email"
}

// TableName for gorm
func (t SMSTemplate) TableName() string {
	return "settings_templates_sms"
}

func (t TelegramTemplate) TableName() string {
	return "settings_templates_telegram"
}

func (t OtherTemplate) TableName() string {
	return "settings_templates_other"
}

func GetOther(id string) (*OtherTemplate, error) {
	template := &OtherTemplate{}
	ret := db.New().Find(template, "template_id = ?", id)
	return template, ret.Error
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

func (t OtherTemplate) Validate(db *gorm.DB) {
	t.BaseNotifierTemplate.Validate(db)
	_, err := pongo2.FromString(t.Text)
	if err != nil {
		db.AddError(validations.NewError(
			t,
			"Text",
			fmt.Sprintf("Failed to compile template: %v", err),
		))
	}
}

func (t TelegramTemplate) Validate(db *gorm.DB) {
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

	subject, err := applyTemplate(t.Subject, ctx, false)
	if err != nil {
		return nil, err
	}

	body, err := applyTemplate(t.Body, ctx, true)
	if err != nil {
		return nil, err
	}

	from, err := applyTemplate(t.From, ctx, false)
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
	return applyTemplate(t.Message, ctx, false)
}

func (t *TelegramTemplate) Execute(ctx interface{}) (interface{}, error) {
	return applyTemplate(t.Message, ctx, false)
}

func (t *OtherTemplate) Execute(ctx interface{}) (interface{}, error) {
	return applyTemplate(t.Text, ctx, false)
}

func randomArgument(args ...*pongo2.Value) *pongo2.Value {
	if len(args) == 0 {
		return pongo2.AsValue("")
	}
	return pongo2.AsValue(args[rand.Intn(len(args))])
}

// applyTemplate applies template from string to ctx and returns result
func applyTemplate(template string, ctx interface{}, escape bool) (string, error) {
	// there is no other way to disable autoescape on global level... Looks very dirty
	if !escape {
		template = "{% autoescape off %}" + template + "{% endautoescape %}"
	}
	tmpl, err := pongo2.FromString(template)
	if err != nil {
		return "", err
	}

	arg, ok := ctx.(map[string]interface{})
	if !ok {
		arg = map[string]interface{}{"object": ctx}
	}
	arg["rand_arg"] = randomArgument
	out, err := tmpl.Execute(pongo2.Context(arg))
	if err != nil {
		return "", err
	}
	return strings.Trim(out, " \r\t\n"), nil
}

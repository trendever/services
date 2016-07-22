package models

import (
	"core/conf"
	"core/db"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"gopkg.in/flosch/pongo2.v3"
	"regexp"
	"strings"
)

// TemplateModels contains all other models that can be passed to a template
var TemplateModels = []string{
	"Lead",
}

// TemplateTypes contains types that correspond to send action types
var TemplateTypes = []string{
	"SMS",
	"Email",
}

// TemplateTypes contains types that correspond to available send actions

// templateIDRegexp is used to match only correct templates: alphanumeric symbols and _
var templateIDRegexp = regexp.MustCompile("^[[:word:]]+$")

// BaseTemplate model
type BaseTemplate struct {
	gorm.Model

	// Template fields
	TemplateName string
	TemplateID   string `gorm:"unique"`

	// model name to which this template can be applied
	ModelName string `gorm:"not null"`

	// comma-separated list of db Preloads that are needed for template
	Preloads string
}

// EmailTemplate object
type EmailTemplate struct {
	BaseTemplate
	EmailMessage
}

// EmailMessage object
type EmailMessage struct {
	From    string `gorm:"type:text"` // For example: "Hello trendever <hello@trendever.com>"
	To      string `gorm:"type:text"` // comma separated recievers
	Subject string `gorm:"type:text"`
	Body    string `gorm:"type:text"`
}

// SMSTemplate object
type SMSTemplate struct {
	BaseTemplate
	SMSMessage
}

// SMSMessage object
type SMSMessage struct {
	To      string `gorm:"type:text"` // For example: "+79991234242"
	Message string `gorm:"type:text"` // comma separated recievers
}

// Template is a common interface for all the models
type Template interface {
	GetPreloads() string
	Parse(interface{}) (interface{}, error)
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

//GetTo returns to
func (em *EmailMessage) GetTo() string {
	return em.To
}

//GetTo returns sms to
func (sm *SMSMessage) GetTo() string {
	return sm.To
}

//GetMessage returns sms message
func (sm *SMSMessage) GetMessage() string {
	return sm.Message
}

// TableName for gorm
func (t EmailTemplate) TableName() string {
	return "settings_templates_email"
}

// GetPreloads returns needed prelads
func (t *BaseTemplate) GetPreloads() string {
	return t.Preloads
}

// TableName for gorm
func (t SMSTemplate) TableName() string {
	return "settings_templates_sms"
}

// Validate fields
// @TODO: use validator
func (t BaseTemplate) Validate(db *gorm.DB) {
	if t.TemplateName == "" {
		db.AddError(validations.NewError(t, "TemplateName", "Template name can not be empty"))
	}

	if !templateIDRegexp.MatchString(t.TemplateID) {
		db.AddError(validations.NewError(t, "TemplateID", "Incorrect template id"))
	}

	if t.ModelName == "" {
		db.AddError(validations.NewError(t, "ModelName", "Model name can not be empty"))
	}
}

// Validate fields
func (t EmailTemplate) Validate(db *gorm.DB) {
	t.BaseTemplate.Validate(db)
}

// Validate fields
func (t SMSTemplate) Validate(db *gorm.DB) {
	t.BaseTemplate.Validate(db)
}

// Parse returns MessageFields object with ready-to-use fields
func (t *EmailTemplate) Parse(ctx interface{}) (interface{}, error) {

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

	to, err := applyTemplate(t.To, ctx)
	if err != nil {
		return nil, err
	}

	return &EmailMessage{
		Subject: subject,
		Body:    body,
		From:    from,
		To:      strings.TrimSpace(to),
	}, nil
}

// Parse returns MessageFields object with ready-to-use fields
func (t *SMSTemplate) Parse(ctx interface{}) (interface{}, error) {

	message, err := applyTemplate(t.Message, ctx)
	if err != nil {
		return nil, err
	}

	to, err := applyTemplate(t.To, ctx)
	if err != nil {
		return nil, err
	}

	return &SMSMessage{
		To:      to,
		Message: message,
	}, nil
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

// EmailTemplateCollection returns 2dim string slice:
// Each element contains template code and template pretty name
// Only templates that are appliable to template type model and appliable to model templateFor are returned
func EmailTemplateCollection(templateFor string) ([][]string, error) {

	var templates []EmailTemplate
	err := db.
		New().
		Where("model_name = ?", templateFor).
		Find(&templates).
		Error

	if err != nil {
		return nil, err
	}

	var out [][]string
	for _, tmpl := range templates {
		out = append(out, []string{tmpl.TemplateID, tmpl.TemplateName})
	}

	return out, nil
}

// SMSTemplateCollection returns 2dim string slice:
// Each element contains template code and template pretty name
// Only templates that are appliable to template type model and appliable to model templateFor are returned
func SMSTemplateCollection(templateFor string) ([][]string, error) {

	var templates []SMSTemplate
	err := db.
		New().
		Where("model_name = ?", templateFor).
		Find(&templates).
		Error

	if err != nil {
		return nil, err
	}

	var out [][]string
	for _, tmpl := range templates {
		out = append(out, []string{tmpl.TemplateID, tmpl.TemplateName})
	}

	return out, nil
}

// FindEmailTemplateByID returns a template with specified TemplateID (string!)
func FindEmailTemplateByID(name string) (Template, error) {

	var template EmailTemplate
	err := db.New().Where("template_id = ?", name).Find(&template).Error

	return &template, err
}

// FindSMSTemplateByID returns a template with specified TemplateID (string!)
func FindSMSTemplateByID(name string) (Template, error) {

	var template SMSTemplate
	err := db.New().Where("template_id = ?", name).Find(&template).Error

	return &template, err
}

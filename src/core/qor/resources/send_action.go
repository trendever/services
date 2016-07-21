package resources

import (
	"core/db"
	"core/models"
	"core/utils"
	"fmt"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"utils/log"
)

// send form parameters
type selectTemplateArgument struct {
	Template string
}

// VisibleFunc is used to specify action visibility scope
type VisibleFunc func(object interface{}, ctx *admin.Context, actionName string) bool

// available actions
const (
	ActionSendSMS   = "Send SMS"
	ActionSendEmail = "Send Email"
)

// Our action resource
var selectRes map[string]*admin.Resource

var actions = []string{
	ActionSendEmail,
	ActionSendSMS,
}

// Action: SMS or email
func createSelectTemplate(a *admin.Admin, action string) *admin.Resource {
	// define a resource for selecting template in GUI
	res := a.NewResource(&selectTemplateArgument{})
	res.Meta(&admin.Meta{
		Name: "Template",
		Type: "select_one",

		// create a collection with filtered by model templates
		Collection: func(this interface{}, ctx *qor.Context) [][]string {
			// Load templates for Lead
			var (
				collection [][]string
				err        error
			)

			switch action {
			case ActionSendEmail:
				collection, err = models.EmailTemplateCollection("Lead")
			case ActionSendSMS:
				collection, err = models.SMSTemplateCollection("Lead")
			}

			if err != nil {
				log.Error(err)
				return [][]string{}
			}
			return collection
		},
	})

	return res
}

func initResources(a *admin.Admin) {
	selectRes = make(map[string]*admin.Resource)
	for _, act := range actions {
		selectRes[act] = createSelectTemplate(a, act)
	}
}

// addSendActions adds "send" action to resource with model name. User selects template and presses "send" button
func addSendActions(model string, a *admin.Admin, res *admin.Resource, visible VisibleFunc) {

	if selectRes == nil {
		initResources(a)
	}

	for i := range actions {

		// work-around unconvitient ranges in Go
		var action = actions[i]

		// send email resource
		res.Action(&admin.Action{
			Name:  action,
			Modes: []string{"index", "menu_item"},

			Resource: selectRes[action],
			Handle: func(arg *admin.ActionArgument) error {
				return handleResourceAction(arg, action)
			},

			Visible: func(record interface{}, context *admin.Context) bool {
				return visible(record, context, action)
			},
		})
	}
}

func handleResourceAction(arg *admin.ActionArgument, action string) error {

	var (
		template models.Template
		err      error
		argument = arg.Argument.(*selectTemplateArgument)
	)

	// firstly, find a template
	switch action {
	case ActionSendEmail:
		template, err = models.FindEmailTemplateByID(argument.Template)
	case ActionSendSMS:
		template, err = models.FindSMSTemplateByID(argument.Template)
	}

	if err != nil {
		return err
	}

	// in bulk-edit mode multiple records can be selected
	for _, object := range arg.FindSelectedRecords() {
		err := handleSendAction(template, object, arg.Context)
		if err != nil {
			log.Debug("Action %v error: %v!", action, err)
			return err
		}
	}

	return nil

}

// handle sending (using send sendFunc) object with template; log to context (optional)
func handleSendAction(template models.Template, object interface{}, ctx *admin.Context) error {

	err := preloadObject(template.GetPreloads(), object)
	if err != nil {
		logError(ctx, err)
		return err
	}

	msg, err := template.Parse(object)
	if err != nil {
		logError(ctx, err)
		return err
	}

	// run an action!
	err = sendMessage(ctx, msg)
	if err != nil {
		logError(ctx, err)
		return err
	}

	return err
}

// sendMessage sends message using a corresponding transport
func sendMessage(ctx *admin.Context, msg interface{}) error {

	switch fields := msg.(type) {
	case *models.EmailMessage:

		err := utils.SendEmail(fields)

		if err == nil {
			var msg string
			// @TODO: remove this workaround once upstream is fixed
			if len(fields.Body) < 5000 {
				msg = fields.Body
			} else {
				msg = fmt.Sprintf("Sent e-mail message (`%s`)", fields.Subject)
			}

			err = utils.AddActivity(
				ctx,
				msg, // activity message
				fmt.Sprintf("Email sent to: %s", fields.To), // activity note
			)
		}

		return err

	case *models.SMSMessage:

		err := utils.SendSMS(fields)

		if err == nil {
			err = utils.AddActivity(
				ctx,
				fields.Message,                            // activity message
				fmt.Sprintf("SMS sent to: %s", fields.To), // activity note
			)
		}

		return err
	}

	return fmt.Errorf("Unknown message type")
}

// log unsuccessful sending
func logError(context *admin.Context, err error) {
	if context != nil && err != nil {
		utils.AddActivity(context, logErrorActivityMsg(err), "")
	}
}

// generate an activity HTML for unsuccessful email sending
func logErrorActivityMsg(err error) string {
	return fmt.Sprintf(`Failed to sending an email: <br/>
		<pre>%v</pre>
	`, err.Error())
}

// preloads a comma separated fields of object (must be ptr to gorm model)
func preloadObject(preloads string, object interface{}) error {
	req := db.New()
	for _, field := range utils.SplitAndTrim(preloads) {
		req = req.Preload(field)
	}

	// do preload
	err := req.Find(object).Error
	return err
}

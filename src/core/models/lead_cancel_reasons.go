package models

import (
	"encoding/json"
	"fmt"
	"proto/chat"

	"github.com/flosch/pongo2"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
)

// LeadCancelReason is template model for possible lead cancel reasons
type LeadCancelReason struct {
	ID   uint64 `gorm:"primary_key"`
	Name string `gorm:"text"`
	// will be send in chat on change CancelReason to that one
	Template string `gorm:"text"`
}

// Stringify generates friendly name
func (r LeadCancelReason) Stringify() string {
	return r.Name
}

// Validate check if template is correct
func (r *LeadCancelReason) Validate(db *gorm.DB) {
	if r.Name == "" {
		db.AddError(validations.NewError(r, "Next", "blank reason text"))
	}
	_, err := pongo2.FromString(r.Template)
	if err != nil {
		db.AddError(validations.NewError(
			r,
			"Template",
			fmt.Sprintf("failed to compile template: %v", err),
		))
	}
}

// GenChatMessage generates chat msg about lead notification
func (r *LeadCancelReason) GenChatMessage(lead *Lead, mover *User) (*chat.Message, error) {

	message, err := applyTemplate(r.Template, map[string]interface{}{
		"lead":   lead,
		"mover":  mover,
		"reason": r,
	}, false)

	if err != nil {
		return nil, err
	}

	if message == "" {
		// template is empty
		// @CHECK if it's needed pls
		return nil, nil
	}

	msg := map[string]string{
		"what":   "lead_cancel",
		"reason": message,
	}

	json, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &chat.Message{
		UserId: uint64(SystemUser.ID),
		Parts: []*chat.MessagePart{
			{
				Content:  string(json),
				MimeType: "json/status",
			},
		},
	}, nil

}

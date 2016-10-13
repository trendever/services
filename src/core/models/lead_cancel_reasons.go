package models

import (
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
)

type LeadCancelReason struct {
	ID   uint64 `gorm:"primary_key"`
	Name string `gorm:"text"`
	// will be send in chat on change CancelReason to that one
	Template string `gorm:"text"`
}

func (r LeadCancelReason) Stringify() string {
	return r.Name
}

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

func (r *LeadCancelReason) GenChatMessage(lead *Lead, mover *User) (string, error) {
	return applyTemplate(r.Template, map[string]interface{}{
		"lead":   lead,
		"mover":  mover,
		"reason": r,
	}, false)
}

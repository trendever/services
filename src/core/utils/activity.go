package utils

import (
	"github.com/qor/activity"
	"github.com/qor/admin"
)

// AddActivity creates a new activity for this context
func AddActivity(ctx *admin.Context, message, note string) error {

	result, err := ctx.FindOne()
	if err != nil {
		return err
	}

	return activity.CreateActivity(ctx, &activity.QorActivity{
		Content: message,
		Note:    note,
	}, result)
}

package resources

import (
	"core/models"
	"github.com/qor/admin"
)

func init() {
	addOnQorInitCallback(addCancelReasonResource)
}

func addCancelReasonResource(a *admin.Admin) {
	res := a.AddResource(&models.LeadCancelReason{}, &admin.Config{
		Name: "LeadCancelReasons",
		Menu: []string{"Products"},
	})

	res.SearchAttrs(
		"Name",
	)
	res.IndexAttrs(
		"ID", "Name",
	)
	//tag.NewAttrs(tag.IndexAttrs())
	//tag.EditAttrs(tag.IndexAttrs())
	//tag.ShowAttrs(tag.IndexAttrs())
}

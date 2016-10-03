package resources

import (
	"core/models"
	"github.com/qor/admin"
)

func init() {
	addResource(&models.LeadCancelReason{}, &admin.Config{
		Name: "LeadCancelReasons",
		Menu: []string{"Products"},
	}, reasonsInit)
}

func reasonsInit(res *admin.Resource) {
	res.SearchAttrs(
		"Name",
	)
	res.IndexAttrs(
		"ID", "Name", "Template",
	)
}

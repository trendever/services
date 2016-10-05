package resources

import (
	"core/models"
	"github.com/qor/admin"
)

func init() {
	addResource(&models.Tag{}, &admin.Config{
		Name: "Tags",
		Menu: []string{"Products"},
	}, initTagResource)
}

func initTagResource(tag *admin.Resource) {
	tag.SearchAttrs(
		"Name",
	)
	tag.IndexAttrs(
		"ID", "Name", "Main", "Hidden", "Group",
	)
	tag.NewAttrs(tag.IndexAttrs())
	tag.EditAttrs(tag.IndexAttrs())
	tag.ShowAttrs(tag.IndexAttrs())

	grp := tag.GetAdmin().AddResource(&models.TagGroup{}, &admin.Config{
		Name: "Product Groups",
		Menu: []string{"Products"},
	})

	grp.SearchAttrs("Name")
}

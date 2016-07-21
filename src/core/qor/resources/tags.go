package resources

import (
	"core/models"
	"github.com/qor/admin"
	"github.com/trendever/ajaxor"
)

func init() {
	addOnQorInitCallback(addTagResource)
}

func addTagResource(a *admin.Admin) {
	tag := a.AddResource(&models.Tag{}, &admin.Config{
		Name: "Tags",
		Menu: []string{"Products"},
	})

	ajaxor.Meta(tag, &admin.Meta{
		Name: "Group",
		Type: "select_one",
	})

	tag.SearchAttrs(
		"Name",
	)
	tag.IndexAttrs(
		"Name", "Main", "Hidden", "Group",
	)
	tag.NewAttrs(tag.IndexAttrs())
	tag.EditAttrs(tag.IndexAttrs())
	tag.ShowAttrs(tag.IndexAttrs())

	grp := a.AddResource(&models.TagGroup{}, &admin.Config{
		Name: "Product Groups",
		Menu: []string{"Products"},
	})

	grp.SearchAttrs("Name")

}

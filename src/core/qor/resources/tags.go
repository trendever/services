package resources

import (
	"core/models"
	"github.com/qor/admin"
	"github.com/trendever/ajaxor"
)

func init() {
	addResource(&models.Tag{}, &admin.Config{
		Name: "Tags",
		Menu: []string{"Products"},
	}, initTagResource)
}

func initTagResource(tag *admin.Resource) {

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

	grp := tag.GetAdmin().AddResource(&models.TagGroup{}, &admin.Config{
		Name: "Product Groups",
		Menu: []string{"Products"},
	})

	grp.SearchAttrs("Name")

}

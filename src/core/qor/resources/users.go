package resources

import (
	"core/conf"
	"core/models"
	"core/qor/filters"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

func init() {
	addOnQorInitCallback(addUserResource)
}

func addUserResource(a *admin.Admin) {
	res := a.AddResource(models.User{}, &admin.Config{Name: "Users"})

	res.Meta(&admin.Meta{
		Name: "Caption",
		Type: "text",
	})

	res.SearchAttrs(
		"Name", "Email", "Phone", "Website", "InstagramId",
		"InstagramUsername", "InstagramFullname",
	)

	res.IndexAttrs(
		"Name", "InstagramUsername", "InstagramCaption", "Email", "Phone",
	)
	res.ShowAttrs(
		&admin.Section{
			Title: "Profile",
			Rows: [][]string{
				{"CreatedAt"},
				{"Name"},
				{"Email", "Phone"},
				{"Website"},
				{"IsAdmin", "IsSeller", "SuperSeller", "IsScout"},
				{"Caption"},
				{"Slogan"},
			},
		},
		&admin.Section{
			Title: "Instagram",
			Rows: [][]string{
				{"InstagramUsername", "InstagramFullname"},
				{"InstagramAvatarURL"},
			},
		},
	)

	if conf.GetSettings().Debug {
		res.NewAttrs(res.ShowAttrs())
	}

	res.EditAttrs(res.ShowAttrs(), "-CreatedAt")

	filters.MetaFilter(res, "CreatedAt", "gt")
	filters.MetaFilter(res, "CreatedAt", "lt")

	res.Scope(&admin.Scope{
		Name:  "With phone",
		Group: "Type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("char_length(phone) > 0")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "With instagram profile",
		Group: "Type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("char_length(instagram_username) > 0")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "With name",
		Group: "Type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("char_length(name) > 0")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Scouts",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_scout = true")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Admins",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_admin = true")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Sellers",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_seller = true")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Super Sellers",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("super_seller = true")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Users with orders",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Joins(`INNER JOIN products_leads as pl ON pl.id =
			  (SELECT id from products_leads WHERE users_user.id = products_leads.customer_id AND products_leads.deleted_at is NULL LIMIT 1)`)
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Customers",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.
				Where("is_seller = false").
				Where("is_admin = false").
				Where("is_scout = false").
				Where("super_seller = false")
		},
	})
}

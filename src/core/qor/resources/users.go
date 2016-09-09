package resources

import (
	"core/conf"
	"core/models"
	"core/qor/filters"
	"errors"
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
		"ID", "Name", "InstagramUsername", "InstagramCaption",
		"Email", "Phone", "Confirmed",
	)
	res.ShowAttrs(
		&admin.Section{
			Title: "Profile",
			Rows: [][]string{
				{"CreatedAt"},
				{"Name"},
				{"Email", "Phone"},
				{"Website"},
				{"IsAdmin", "IsSeller", "SuperSeller", "IsScout", "Confirmed"},
				{"Caption"},
				{"Slogan"},
			},
		},
		&admin.Section{
			Title: "Instagram",
			Rows: [][]string{
				{"InstagramUsername", "InstagramFullname"},
				{"InstagramCaption"},
				{"InstagramAvatarURL"},
			},
		},
	)

	if conf.GetSettings().Debug {
		res.NewAttrs(res.ShowAttrs())
	}

	res.EditAttrs(res.ShowAttrs(), "-CreatedAt", "-Confirmed")

	filters.MetaFilter(res, "CreatedAt", "gt")
	filters.MetaFilter(res, "CreatedAt", "lt")

	res.Scope(&admin.Scope{
		Name: "Only confirmed users",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("confirmed = ?", true)
		},
	})

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
		Name:  "With email",
		Group: "Type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("char_length(email) > 0")
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
		Name:  "Suppliers",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.
				Joins("JOIN products_shops shop ON users_user.id = shop.supplier_id AND shop.deleted_at IS NULL").
				Where("EXISTS (SELECT 1 FROM products_product product WHERE product.shop_id = shop.id AND product.deleted_at IS NULL)")
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Customers",
		Group: "Role",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.
				Joins("LEFT JOIN products_shops shop ON users_user.id = shop.supplier_id AND shop.deleted_at IS NULL").
				Where("shop.id IS NULL OR NOT EXISTS (SELECT 1 FROM products_product product WHERE product.shop_id = shop.id AND product.deleted_at IS NULL)")
		},
	})

	res.Action(&admin.Action{
		Name:  "Check instagram",
		Modes: []string{"index", "menu_item"},
		Handle: func(arg *admin.ActionArgument) error {
			for _, record := range arg.FindSelectedRecords() {
				user, ok := record.(models.User)
				if !ok {
					return errors.New("unexpected record type")
				}
				user.CheckInstagram()
			}
			return nil
		},
	})
}

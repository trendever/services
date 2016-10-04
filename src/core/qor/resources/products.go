package resources

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/trendever/ajaxor"

	"core/models"
	"utils/db"
)

func init() {
	addResource(models.Product{}, &admin.Config{
		Name: "Products",
		Menu: []string{"Products"},
	}, initProductResource)

}

func initProductResource(res *admin.Resource) {
	itemRes := res.GetAdmin().AddResource(models.ProductItem{}, &admin.Config{
		Name:      "ProductItem",
		Invisible: true,
	})

	res.Meta(&admin.Meta{
		Name:     "Items",
		Type:     "collection_edit",
		Resource: itemRes,
	})

	res.Meta(&admin.Meta{Name: "Code", Type: "string"})
	res.Meta(&admin.Meta{Name: "Img", FieldName: "InstagramImageURL", Type: "image"})
	res.Meta(&admin.Meta{Name: "InstagramImageCaption", Type: "text"})
	res.Meta(&admin.Meta{Name: "CreatedAt", Type: "date"})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Shop",
		Type: "select_one",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "MentionedBy",
		Type: "select_one",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Tags",
		Type: "select_many",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name:      "Scout",
		FieldName: "MentionedBy",
		Type:      "select_one",
		Collection: func(this interface{}, ctx *qor.Context) [][]string {

			searchCtx := ctx.Clone()

			searchCtx.SetDB(ctx.GetDB().
				Joins("JOIN products_product as pr" +
					" ON pr.mentioned_by_id = users_user.id AND pr.deleted_at IS NULL AND users_user.is_scout = true").
				Group("users_user.id").
				Having("COUNT(pr.id) > 0").
				Order("COUNT(pr.id) DESC"),
			)

			return res.GetMeta("MentionedBy").Config.(interface {
				GetCollection(value interface{}, context *admin.Context) [][]string
			}).GetCollection(this, &admin.Context{Context: ctx})
		},
	})

	ajaxor.Meta(res, &admin.Meta{
		Name:      "ShopSearch",
		Label:     "Shop",
		FieldName: "Shop",
		Type:      "select_one",
		Collection: func(this interface{}, ctx *qor.Context) [][]string {

			searchCtx := ctx.Clone()

			searchCtx.SetDB(ctx.GetDB().
				Joins("JOIN products_product as pr ON pr.shop_id = products_shops.id AND pr.deleted_at IS NULL").
				Group("products_shops.id").
				Having("COUNT(pr.id) > 0").
				Order("COUNT(pr.id) DESC"),
			)

			return res.GetMeta("Shop").Config.(interface {
				GetCollection(value interface{}, context *admin.Context) [][]string
			}).GetCollection(this, &admin.Context{Context: ctx})
		},
	})

	res.SearchAttrs(
		"Code", "Title", "InstagramLink", "Shop", "MentionedBy",
	)
	res.IndexAttrs(
		"ID", "InstagramImageURL", "Code", "Title",
		"IsSale", "Shop", "MentionedBy", "Tags",
	)
	res.EditAttrs(
		&admin.Section{
			Title: "Product information",
			Rows: [][]string{
				{"Code", "Title"},
				{"IsSale"},
			},
		},
		&admin.Section{
			Title: "Instagram",
			Rows: [][]string{
				{"InstagramImageURL"},
				{"InstagramLink"},
				{"InstagramImageCaption"},
				{"InstagramPublishedAt"},
				{"Shop", "MentionedBy"},
			},
		},
		&admin.Section{
			Title: "Items",
			Rows: [][]string{
				{"Items"},
			},
		},
	)
	res.NewAttrs(res.EditAttrs())
	res.ShowAttrs(
		&admin.Section{
			Title: "Product information",
			Rows: [][]string{
				{"Code", "Title"},
				{"CreatedAt"},
				{"IsSale"},
			},
		},
		&admin.Section{
			Title: "Instagram",
			Rows: [][]string{
				{"InstagramImageURL"},
				{"InstagramLink"},
				{"InstagramImageCaption"},
				{"InstagramPublishedAt"},
				{"Shop", "MentionedBy"},
			},
		},
		&admin.Section{
			Title: "Items",
			Rows: [][]string{
				{"Items"},
			},
		},
	)

	res.Scope(&admin.Scope{
		Name:  "Only on sale",
		Group: "Type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_sale = ?", true)
		},
	})

	res.Scope(&admin.Scope{
		Name:  "Not on sale",
		Group: "Type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("is_sale = ?", false)
		},
	})

	res.Scope(&admin.Scope{
		Name: "Only Scouts' products",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Joins(`
				LEFT JOIN users_user as mentioned_by
				ON mentioned_by.id = products_product.mentioned_by_id`).
				Where("mentioned_by.is_scout = ?", true)
		},
	})

	// embed productItem stuff
	ajaxor.Meta(itemRes, &admin.Meta{
		Name: "Tags",
		Type: "select_many",
	})

	itemRes.EditAttrs(
		"Name",
		"Price", "DiscountPrice",
		"Tags",
	)
	itemRes.NewAttrs(itemRes.EditAttrs())

	itemRes.SearchAttrs("Name", "Tags")

	res.UseTheme("filter-workaround")
	for name, act := range map[string]string{"from": ">", "to": "<"} {
		op := act
		res.Filter(&admin.Filter{
			Name:  "created_at_" + name,
			Label: "Created At " + name,
			Handler: func(scope *gorm.DB, arg *admin.FilterArgument) *gorm.DB {
				metaValue := arg.Value.Get("Value")
				if metaValue == nil {
					return scope
				}
				return scope.Where(fmt.Sprintf("products_product.created_at %v ?", op), metaValue.Value)
			},
			Type: "date",
		})
	}

	res.Filter(&admin.Filter{
		Label:  "Scout",
		Name:   "MentionedBy",
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Users")},
	})
	res.Filter(&admin.Filter{
		Name:   "Shop",
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Shops")},
	})

	res.Filter(&admin.Filter{
		Name: "Tag",
		Handler: func(scope *gorm.DB, arg *admin.FilterArgument) *gorm.DB {
			metaValue := arg.Value.Get("Value")
			if metaValue == nil {
				return scope
			}
			return scope.Joins(`
				INNER JOIN products_product_item_tags as tagrel ON
				tagrel.product_id = products_product.id AND tagrel.tag_id = ?
			`, metaValue.Value)
		},
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Tags")},
	})

	type userArg struct {
		UserID uint64
		User   models.User
	}
	userArgRes := res.GetAdmin().NewResource(&userArg{})
	ajaxor.Meta(userArgRes, &admin.Meta{
		Name: "User",
		Type: "select_one",
	})

	res.Action(&admin.Action{
		Name: "Set supplier",
		Handle: func(argument *admin.ActionArgument) error {
			arg, ok := argument.Argument.(*userArg)
			if !ok {
				return errors.New("unxepected argument type")
			}
			if arg.User.ID == 0 {
				return nil
			}
			shopID, _, err := models.FindOrCreateShopForSupplier(&arg.User, true)
			if err != nil {
				return err
			}
			for _, record := range argument.FindSelectedRecords() {
				product, ok := record.(models.Product)
				if !ok {
					return errors.New("unxepected record type")
				}
				product.ShopID = uint(shopID)
				err := db.New().Save(&product).Error
				if err != nil {
					return err
				}
			}
			return nil
		},
		Resource: userArgRes,
		Modes:    []string{"show", "menu_item"},
	})
}

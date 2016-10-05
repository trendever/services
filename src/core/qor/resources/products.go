package resources

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/trendever/ajaxor"

	"core/models"
	"core/qor/filters"
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

	filters.SetDateFilters(res, "CreatedAt")

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
		Name: "Tags",
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
	userArgRes.Meta(&admin.Meta{
		Name:   "User",
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Users")},
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

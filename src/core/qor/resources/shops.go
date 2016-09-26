package resources

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/trendever/ajaxor"
	"reflect"
	"utils/log"

	"core/models"
	"core/qor/filters"
)

func init() {
	addOnQorInitCallback(addShopResource)
}

func addShopResource(a *admin.Admin) {
	res := a.AddResource(models.Shop{}, &admin.Config{
		Name: "Shops",
	})

	// @TODO: make image URL editable
	//res.Meta(&admin.Meta{Name: "Img", FieldName: "InstagramAvatarURL", Type: "image"})
	res.Meta(&admin.Meta{Name: "InstagramCaption", Type: "text"})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Supplier",
		Type: "select_one",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Sellers",
		Type: "select_many",
	})

	ajaxor.Meta(res, &admin.Meta{ //filters Sellers only to set ones
		Name:      "SellersOnly",
		FieldName: "Sellers",
		Label:     "Sellers",
		Type:      "select_many",
		Collection: func(this interface{}, ctx *qor.Context) [][]string {

			searchCtx := ctx.Clone()

			searchCtx.SetDB(ctx.GetDB().Where("users_user.is_seller = true"))

			return res.GetMeta("Sellers").GetCollection(this, searchCtx)
		},
	})

	res.Meta(&admin.Meta{
		Name: "ShippingRules",
		Type: "text",
	})

	res.Meta(&admin.Meta{
		Name: "PaymentRules",
		Type: "text",
	})

	res.Meta(&admin.Meta{
		Name: "Caption",
		Type: "text",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Tags",
		Type: "select_many",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name:      "TagsSearch",
		FieldName: "Tags",
		Label:     "Tags",
		Type:      "select_many",
		Collection: func(this interface{}, ctx *qor.Context) [][]string {

			searchCtx := ctx.Clone()

			searchCtx.SetDB(ctx.GetDB().
				Joins("JOIN products_shops_tags as tagrel ON products_tag.id = tagrel.tag_id").
				Group("products_tag.id,tagrel.tag_id").
				Having("COUNT(tagrel.tag_id) > 0").
				Order("COUNT(tagrel.tag_id) DESC"),
			)

			return res.GetMeta("Tags").GetCollection(this, searchCtx)
		},
	})

	res.Action(&admin.Action{
		Name:  "Delete with products",
		Modes: []string{"show", "menu_item"},
		Handle: func(arg *admin.ActionArgument) error {

			// we work in transcation: either everything transists to the new state, either nothing
			tx := arg.Context.GetDB().Begin()

			for _, record := range arg.FindSelectedRecords() {

				shop, ok := record.(models.Shop)
				if !ok {
					err := fmt.Errorf("Got incorrect record type (%v), should NOT normally happen", reflect.TypeOf(record))
					log.Error(err)
					return err
				}

				if err := tx.Where("shop_id = ?", shop.ID).Delete(models.Product{}).Error; err != nil {
					log.Warn("wtf", err)
					return err
				}

				if err := tx.Where("id = ?", shop.ID).Delete(models.Shop{}).Error; err != nil {
					log.Warn("wtf", err)
					return err
				}
			}

			tx.Commit()
			return nil
		},
	})

	res.SearchAttrs(
		"InstagramUsername", "Supplier.Name", "InstagramCaption", "InstagramFullname", "InstagramWebsite",
	)
	res.IndexAttrs(
		"InstagramUsername", "InstagramCaption", "InstagramFullname", "Supplier", "SupplierLastLogin", "Tags",
	)

	res.EditAttrs(
		&admin.Section{
			Title: "Instagram",
			Rows: [][]string{
				{"InstagramID"},
				{"InstagramCaption"},
				{"InstagramUsername"},
				{"InstagramFullname"},
				{"InstagramWebsite"},
			},
		},
		&admin.Section{
			Title: "Shop",
			Rows: [][]string{
				{"CreatedAt"},
				{"Tags"},
				{"Supplier", "Sellers"},
				{"Caption"},
				{"Slogan"},
				{"ShippingRules"},
				{"PaymentRules", "Cards"},
				{"NotifySupplier"},
			},
		},
	)

	res.NewAttrs(res.EditAttrs())
	res.ShowAttrs(res.EditAttrs())

	filters.MetaFilter(res, "TagsSearch", "eq")
	res.Filter(&admin.Filter{
		Name: "tags_id_eq",
		Handler: func(fieldName, query string, scope *gorm.DB, context *qor.Context) *gorm.DB {
			return scope.Joins(`
				INNER JOIN products_shops_tags as tagrel ON 
				tagrel.shop_id = id AND tagrel.tag_id = ?
			`, query)
		},
	})

	filters.MetaFilter(res, "CreatedAt", "gt")
	filters.MetaFilter(res, "CreatedAt", "lt")
}

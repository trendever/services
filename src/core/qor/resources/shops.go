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
)

func init() {
	addResource(models.Shop{}, &admin.Config{
		Name: "Shops",
	}, initShopResource)
}

func initShopResource(res *admin.Resource) {
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

			return res.GetMeta("Sellers").Config.(interface {
				GetCollection(value interface{}, context *admin.Context) [][]string
			}).GetCollection(this, &admin.Context{Context: ctx})
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

			return res.GetMeta("Tags").Config.(interface {
				GetCollection(value interface{}, context *admin.Context) [][]string
			}).GetCollection(this, &admin.Context{Context: ctx})
		},
	})

	noteRes := res.Meta(&admin.Meta{Name: "Notes"}).Resource
	noteRes.Meta(&admin.Meta{Name: "Text", Type: "text"})

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
		"Supplier.InstagramUsername", "Supplier.Name", "Supplier.InstagramCaption", "Supplier.InstagramFullname", "Supplier.InstagramWebsite",
	)
	res.IndexAttrs(
		"ID", "InstagramUsername", "Supplier", "SupplierLastLogin", "Tags",
	)

	res.EditAttrs(
		&admin.Section{
			Title: "Shop",
			Rows: [][]string{
				{"InstagramUsername", "InstagramWebsite"},
				{"CreatedAt"},
				{"Tags"},
				{"Supplier", "Sellers"},
				{"Caption"},
				{"Slogan"},
				{"ShippingRules"},
				{"PaymentRules"},
				{"Cards"},
				{"NotifySupplier"},
			},
		},
		"Notes",
	)

	res.ShowAttrs(
		&admin.Section{
			Title: "Shop",
			Rows: [][]string{
				{"InstagramUsername", "InstagramWebsite"},
				{"CreatedAt"},
				{"Tags"},
				{"Supplier", "Sellers"},
				{"Caption"},
				{"Slogan"},
				{"ShippingRules"},
				{"PaymentRules"},
				{"Cards"},
				{"NotifySupplier"},
			},
		},
		"Notes",
	)

	res.NewAttrs(res.EditAttrs())
	res.ShowAttrs(res.EditAttrs())

	res.Filter(&admin.Filter{
		Name: "Tag",
		Handler: func(scope *gorm.DB, arg *admin.FilterArgument) *gorm.DB {
			metaValue := arg.Value.Get("Value")
			if metaValue == nil {
				return scope
			}
			return scope.Joins(`
				INNER JOIN products_shops_tags as tagrel ON
				tagrel.shop_id = id AND tagrel.tag_id = ?
			`, metaValue.Value)
		},
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Tags")},
	})

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
				return scope.Where(fmt.Sprintf("products_shops.created_at %v ?", op), metaValue.Value)
			},
			Type: "date",
		})
	}
}

package resources

import (
	"common/log"
	"core/models"
	"core/qor/filters"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"reflect"
)

func init() {
	addResource(models.Shop{}, &admin.Config{
		Name: "Shops",
	}, initShopResource)
}

func initShopResource(res *admin.Resource) {
	// @TODO meta with collection will not use pagination. Any workaround?
	// there was many places where it was used with ajaxor to filter search results
	//res.Meta(&admin.Meta{
	//	Name: "Sellers",
	//	Type: "select_many",
	//	Collection: func(_ interface{}, ctx *qor.Context) [][]string {
	//		var values []*models.User
	//		err := ctx.GetDB().Where("users_user.is_seller = true").Find(&values).Error
	//		if err != nil {
	//			log.Errorf("failed to select values from db: %v", err)
	//			return [][]string{}
	//		}
	//		ret := makeCollection(values, ctx.GetDB())
	//		return ret
	//	},
	//})

	res.Meta(&admin.Meta{
		Name: "ShippingRules",
		Type: "text",
	})

	res.Meta(&admin.Meta{
		Name: "PaymentRules",
		Type: "text",
	})

	res.Meta(&admin.Meta{
		Name: "WorkingTime",
		Type: "text",
	})

	res.Meta(&admin.Meta{
		Name:      "Name",
		FieldName: "InstagramUsername",
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
		"Supplier.InstagramUsername", "Supplier.Name", "Supplier.InstagramCaption", "Supplier.InstagramFullname", "Supplier.InstagramWebsite", "Location",
	)
	res.IndexAttrs(
		"ID", "InstagramUsername", "Supplier", "SupplierLastLogin", "Suspended", "Tags",
	)

	res.EditAttrs(
		&admin.Section{
			Title: "Shop",
			Rows: [][]string{
				{"InstagramUsername", "InstagramWebsite"},
				{"CreatedAt"},
				{"Tags"},
				{"Supplier", "Sellers"},
				{"Plan", "PlanExpiresAt"},
				{"AutoRenewal", "Suspended"},
				{"Caption"},
				{"Slogan"},
				{"Location"},
				{"WorkingTime"},
				{"ShippingRules"},
				{"PaymentRules"},
				{"Cards"},
				{"NotifySupplier", "SeparateLeads"},
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
				{"Location"},
				{"WorkingTime"},
				{"ShippingRules"},
				{"PaymentRules"},
				{"Cards"},
				{"NotifySupplier", "SeparateLeads"},
			},
		},
		"Notes",
	)

	res.NewAttrs(res.EditAttrs())
	res.ShowAttrs(res.EditAttrs())

	res.Filter(&admin.Filter{
		Name: "Tags",
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

	filters.SetDateFilters(res, "CreatedAt")
}

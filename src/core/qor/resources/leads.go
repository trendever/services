package resources

import (
	"common/db"
	"common/log"
	"core/conf"
	"core/models"
	"core/qor/filters"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

func init() {
	addResource(&models.Lead{}, &admin.Config{
		Name: "Orders",
		Menu: []string{"Products"},
	}, initLeadResource)
}

func initLeadResource(res *admin.Resource) {

	res.Meta(&admin.Meta{
		Name: "State", Type: "select_one",
		Collection: models.GetLeadStates(),
	})

	itemRes := res.GetAdmin().NewResource(models.ProductItem{}, &admin.Config{
		Name: "ProductItem",
		Menu: []string{},
	})
	// we will use custom SearchHandler but panic will be raised without this line(qor bug)...
	itemRes.SearchAttrs("Name")
	itemRes.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
		return context.GetDB().Where(`
		lower(products_product_item.name) LIKE lower(?)
		OR EXISTS (
			SELECT 1 FROM products_product product
			WHERE product.id = products_product_item.product_id AND product.deleted_at IS NULL
			AND product.code LIKE lower(?)
		)`, "%"+keyword+"%", keyword+"%")
	}
	res.Meta(&admin.Meta{
		Name: "ProductItems",
		Config: &admin.SelectManyConfig{
			RemoteDataResource: itemRes,
		},
	})

	res.SearchAttrs(
		"ID", "Name", "Source", "Customer.Name", "Comment",
	)
	res.IndexAttrs(
		"ID", "CreatedAt", "Customer", "Shop", "Name", "Source", "ProductItems", "State", "CancelReason",
	)
	res.ShowAttrs(
		&admin.Section{
			Title: "Order information",
			Rows: [][]string{
				{"ID"},
				{"CreatedAt"},
				{"Source", "Comment"},
				{"Shop", "Customer"},
				{"ProductItems"},
				{"CancelReason"},
				{"StatusComment"},
			},
		},
	)

	// creating lead manually only for debugging purposes`
	if conf.GetSettings().Debug {
		res.EditAttrs(res.ShowAttrs())
		res.NewAttrs(res.ShowAttrs())
	}

	// Add state scopes
	for _, state := range models.GetLeadStates() {
		var stateCopy = state
		res.Scope(&admin.Scope{
			Name:  state,
			Group: "State Filter",
			Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				return db.Where("products_leads.state = ?", stateCopy)
			},
		})
	}

	addTransitionActions(res.GetAdmin(), res)

	type productArg struct {
		ProductID uint64
		Product   models.Product
	}
	argRes := res.GetAdmin().NewResource(&productArg{})

	res.Action(&admin.Action{
		Name: "Add product",
		Handle: func(argument *admin.ActionArgument) error {
			arg, ok := argument.Argument.(*productArg)
			if !ok {
				return errors.New("unxepected argument type")
			}
			if arg.Product.ID == 0 {
				return nil
			}
			err := db.New().Model(&arg.Product).Related(&arg.Product.Items).Error
			if err != nil {
				return err
			}
			for _, record := range argument.FindSelectedRecords() {
				lead, ok := record.(*models.Lead)
				if !ok {
					return errors.New("unxepected record type")
				}
				_, err := models.AppendLeadItems(lead, arg.Product.Items)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Resource: argRes,
		Modes:    []string{"show", "menu_item"},
	})

	filters.SetDateFilters(res, "CreatedAt")

	res.Filter(&admin.Filter{
		Name:   "Customer",
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Users")},
	})

	res.Filter(&admin.Filter{
		Name:   "Shop",
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Shops")},
	})

	res.Filter(&admin.Filter{
		Name: "Products",
		Handler: func(scope *gorm.DB, arg *admin.FilterArgument) *gorm.DB {
			metaValue := arg.Value.Get("Value")
			if metaValue == nil {
				return scope
			}
			return scope.Where(
				`EXISTS (
					SELECT 1 FROM products_leads_items related
					JOIN products_product_item item
					ON related.product_item_id = item.id
					WHERE item.product_id = ? AND related.lead_id = products_leads.id
				)`,
				metaValue.Value)
		},
		Config: &admin.SelectOneConfig{RemoteDataResource: res.GetAdmin().GetResource("Products")},
	})
}

// and typical actions for changing order state
func addTransitionActions(a *admin.Admin, res *admin.Resource) {
	type gotEmailArgument struct {
		Email string
	}

	// Add actions that trigger LeadState events
	for i := range models.GetLeadEvents() {
		var ev = models.GetLeadEvents()[i] // copy event so we can use it async

		res.Action(&admin.Action{
			Name:  ev.Name,
			Modes: []string{"index", "menu_item"},
			// that is what called when user clicks action
			Handle: func(arg *admin.ActionArgument) error {
				mover, _ := arg.Context.CurrentUser.(*models.User)

				for _, order := range arg.FindSelectedRecords() {
					lead := order.(*models.Lead)
					err := db.New().Preload("Shop").Preload("Shop.Supplier").Preload("Customer").First(lead).Error
					if err != nil {
						log.Error(err)
						return err
					}
					err = lead.TriggerEvent(ev.Name, "", 0, mover)
					if err != nil {
						log.Error(err)
						return err
					}
				}

				return nil
			},
			// that defines if action is visible
			Visible: func(record interface{}, context *admin.Context) bool {
				lead := record.(*models.Lead)
				return models.LeadEventPossible(ev.Name, lead.State)
			},
		})
	}

	// Supplier scopes
	res.Scope(&admin.Scope{
		Name:  "With phone",
		Group: "Supplier type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Joins("INNER JOIN products_shops as ps ON ps.id = products_leads.shop_id").
				Joins("INNER JOIN users_user as pu ON pu.id = ps.supplier_id").
				Where("char_length(pu.phone) > 0")
		},
	})
	res.Scope(&admin.Scope{
		Name:  "Without phone",
		Group: "Supplier type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Joins("INNER JOIN products_shops as ps ON ps.id = products_leads.shop_id").
				Joins("INNER JOIN users_user as pu ON pu.id = ps.supplier_id").
				Where("char_length(pu.phone) = 0")
		},
	})
	res.Scope(&admin.Scope{
		Name:  "With email",
		Group: "Supplier type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Joins("INNER JOIN products_shops as ps ON ps.id = products_leads.shop_id").
				Joins("INNER JOIN users_user as pu ON pu.id = ps.supplier_id").
				Where("char_length(pu.email) > 0")
		},
	})
	res.Scope(&admin.Scope{
		Name:  "Without email",
		Group: "Supplier type",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Joins("INNER JOIN products_shops as ps ON ps.id = products_leads.shop_id").
				Joins("INNER JOIN users_user as pu ON pu.id = ps.supplier_id").
				Where("char_length(pu.email) = 0")
		},
	})
	res.Scope(&admin.Scope{
		Name:  "With comment",
		Group: "Comment",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("char_length(products_leads.comment) > 0")
		},
	})
	res.Scope(&admin.Scope{
		Name:  "No comment",
		Group: "Comment",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("char_length(products_leads.comment) = 0")
		},
	})

	// Source scopes
	for _, s := range models.LeadSources {
		res.Scope(&admin.Scope{
			Name:  "From " + s,
			Group: "Source",
			Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				return db.Where("products_leads.source = ?", s)
			},
		})
	}
}

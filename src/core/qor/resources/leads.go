package resources

import (
	"core/conf"
	"core/models"
	"core/qor/filters"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/activity"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/trendever/ajaxor"
	"utils/log"
)

func init() {
	addOnQorInitCallback(addLeadResource)
}

type leadEvent struct {
	Resource *admin.Resource
	Handler  func(*admin.ActionArgument, *gorm.DB, interface{}) error
}

func addLeadResource(a *admin.Admin) {
	res := a.AddResource(&models.Lead{}, &admin.Config{
		Name: "Orders",
		Menu: []string{"Products"},
	})

	res.Meta(&admin.Meta{
		Name: "State", Type: "select_one",
		Collection: models.GetLeadStates(),
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Customer",
		Type: "select_one",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "Shop",
		Type: "select_one",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name: "ProductItems",
		Type: "select_many",
	})

	ajaxor.Meta(res, &admin.Meta{
		Name:      "CustomerSearch",
		Label:     "Customer",
		FieldName: "Customer",
		Type:      "select_one",
		Collection: func(this interface{}, ctx *qor.Context) [][]string {

			searchCtx := ctx.Clone()

			searchCtx.SetDB(ctx.GetDB().
				Joins("JOIN products_leads as lead ON lead.customer_id = users_user.id AND lead.deleted_at IS NULL").
				Group("users_user.id").
				Having("COUNT(lead.id) > 0").
				Order("COUNT(lead.id) DESC"),
			)

			return res.GetMeta("Customer").GetCollection(this, searchCtx)
		},
	})

	filters.MetaFilter(res, "CustomerSearch", "eq")

	res.SearchAttrs(
		"ID", "Name", "Source", "Customer.Name", "Comment",
	)
	res.IndexAttrs(
		"ID", "CreatedAt", "Customer", "Name", "Source", "ProductItems", "State",
	)
	res.ShowAttrs(
		&admin.Section{
			Title: "Order information",
			Rows: [][]string{
				{"ID"},
				{"CreatedAt"},
				{"Source", "Comment"},
				{"Shop"},
				{"ProductItems", "Customer"},
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

	activity.Register(res) // register order activity
	addTransitionActions(a, res)

	// add various send actions
	addSendActions("Lead", a, res, func(object interface{}, ctx *admin.Context, actionName string) bool {

		switch actionName {
		case ActionSendEmail:
			return false
		case ActionSendSMS:
			return true
		default:
			return false
		}
	})
}

// and typical actions for changing order state
func addTransitionActions(a *admin.Admin, res *admin.Resource) {
	type gotEmailArgument struct {
		Email string
	}

	// helper map that allows to add custom action resources and handlers without unneeded copy&paste
	events := map[string]leadEvent{
	//models.LeadEventGotEmail: leadEvent{
	//	Resource: a.NewResource(&gotEmailArgument{}),
	//	Handler: func(arg *admin.ActionArgument, db *gorm.DB, record interface{}) error {
	//		lead := record.(*models.Lead)
	//		argument := arg.Argument.(*gotEmailArgument)
	//
	//		// get user (qor won't preload it to lead)
	//		user, err := models.FindUserByID(lead.CustomerID)
	//		if err != nil {
	//			log.Error(err)
	//			return err
	//		}
	//
	//		log.Printf("Editing user email %v", user)
	//		user.Email = argument.Email
	//
	//		err = db.Save(&user).Error
	//		if err != nil {
	//			return err
	//		}
	//
	//		return nil
	//	},
	//},
	}

	// Add actions that trigger LeadState events
	for i := range models.GetLeadEvents() {
		var ev = models.GetLeadEvents()[i] // copy event so we can use it async

		res.Action(&admin.Action{
			Name:  ev.Name,
			Modes: []string{"index", "menu_item"},

			// exploit map default value here
			Resource: events[ev.Name].Resource,

			// that is what called when user clicks action
			Handle: func(arg *admin.ActionArgument) error {

				// we work in transcation: either everything transists to the new state, either nothing
				tx := arg.Context.GetDB().Begin()

				for _, order := range arg.FindSelectedRecords() {
					lead := order.(*models.Lead)
					log.Debug("Starting processing order %v", lead)

					// run handler if exists
					if handler := events[ev.Name].Handler; handler != nil {
						err := handler(arg, tx, order)
						if err != nil {
							tx.Rollback()
							log.Error(err)
							return err
						}
					}

					// then, trigger an event using qor/transition state machine instance
					err := models.LeadState.Trigger(ev.Name, lead, tx)
					if err != nil {
						tx.Rollback()
						log.Error(err)
						return err
					}

					// save everything
					err = tx.Select("state").Save(order).Error
					if err != nil {
						tx.Rollback()
						log.Error(err)
						return err
					}
				}

				tx.Commit()
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
	res.Scope(&admin.Scope{
		Name:  "From website",
		Group: "Source",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("products_leads.source = ?", "website")
		},
	})
	res.Scope(&admin.Scope{
		Name:  "From instagram",
		Group: "Source",
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("products_leads.source like ?", "@%")
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
				Joins("JOIN products_leads as pl ON pl.shop_id = products_shops.id AND pl.deleted_at IS NULL").
				Group("products_shops.id").
				Having("COUNT(pl.id) > 0").
				Order("COUNT(pl.id) DESC"),
			)

			return res.GetMeta("Shop").GetCollection(this, searchCtx)
		},
	})

	// filters
	filters.MetaFilter(res, "CreatedAt", "gt")
	filters.MetaFilter(res, "CreatedAt", "lt")
	filters.MetaFilter(res, "ShopSearch", "eq")

	// @QORBUG
	// workaround due to bug in qor
	for op, act := range map[string]string{"gt": ">", "lt": "<"} {
		var actcp = act

		res.Filter(&admin.Filter{
			Name: "created_at_" + op,
			Handler: func(fieldName, query string, scope *gorm.DB, context *qor.Context) *gorm.DB {
				return scope.Where(fmt.Sprintf("products_leads.created_at %v ?", actcp), query)
			},
		})
	}

}

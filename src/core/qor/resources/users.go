package resources

import (
	"core/api"
	"core/conf"
	"core/models"
	"core/qor/filters"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/trendever/ajaxor"
	"proto/trendcoin"
	"utils/rpc"
)

func init() {
	addResource(models.User{}, &admin.Config{Name: "Users"}, initUserResource)
}

func initUserResource(res *admin.Resource) {
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
		"Email", "Phone", "Balance", "Confirmed", "LastLogin",
	)
	res.ShowAttrs(
		&admin.Section{
			Title: "Profile",
			Rows: [][]string{
				{"CreatedAt", "LastLogin"},
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

	filters.SetDateFilters(res, "CreatedAt")

	type refillArg struct {
		Amount  uint64
		Comment string
	}
	refillArgRes := res.GetAdmin().NewResource(&refillArg{})
	res.Action(&admin.Action{
		Name:     "Refill coins",
		Resource: refillArgRes,
		Modes:    []string{"show", "menu_item"},
		Handle: func(argument *admin.ActionArgument) error {
			arg, ok := argument.Argument.(*refillArg)
			if !ok {
				return errors.New("unxepected argument type")
			}
			transactions := []*trendcoin.TransactionData{}
			mover := argument.Context.CurrentUser.(*models.User)
			reason := fmt.Sprintf(
				"User %v(%v) trigger refill action in qor, comment: '%v'",
				mover.ID, mover.GetName(), arg.Comment,
			)
			for _, record := range argument.FindSelectedRecords() {
				user, ok := record.(models.User)
				if !ok {
					return errors.New("unxepected record type")
				}
				transactions = append(transactions, &trendcoin.TransactionData{
					Destination:    uint64(user.ID),
					Amount:         arg.Amount,
					AllowEmptySide: true,
					Reason:         reason,
				})
			}
			return performTransactions(transactions...)
		},
	})
	type writeOffArg struct {
		Amount      uint64
		AllowCredit bool
		Comment     string
	}
	writeOffArgRes := res.GetAdmin().NewResource(&writeOffArg{})
	res.Action(&admin.Action{
		Name:     "Write-off coins",
		Resource: writeOffArgRes,
		Modes:    []string{"show", "menu_item"},
		Handle: func(argument *admin.ActionArgument) error {
			arg, ok := argument.Argument.(*writeOffArg)
			if !ok {
				return errors.New("unxepected argument type")
			}
			transactions := []*trendcoin.TransactionData{}
			mover := argument.Context.CurrentUser.(*models.User)
			reason := fmt.Sprintf(
				"User %v(%v) trigger write-off action in qor, comment: '%v'",
				mover.ID, mover.GetName(), arg.Comment,
			)
			for _, record := range argument.FindSelectedRecords() {
				user, ok := record.(models.User)
				if !ok {
					return errors.New("unxepected record type")
				}
				transactions = append(transactions, &trendcoin.TransactionData{
					Source:         uint64(user.ID),
					Amount:         arg.Amount,
					AllowCredit:    arg.AllowCredit,
					AllowEmptySide: true,
					Reason:         reason,
				})
			}
			return performTransactions(transactions...)
		},
	})

	type transferArg struct {
		DestinationID uint64
		Destination   models.User
		Amount        uint64
		AllowCredit   bool
		Comment       string
	}
	transferArgRes := res.GetAdmin().NewResource(&transferArg{})
	// @TODO resource in context should have all this themes...
	// we need somehow set it ajaxor probably
	res.UseTheme("select2.min")
	res.UseTheme("ajaxor")
	ajaxor.Meta(transferArgRes, &admin.Meta{
		Name: "Destination",
		Type: "select_one",
	})
	res.Action(&admin.Action{
		Name:     "Transfer coins",
		Resource: transferArgRes,
		Modes:    []string{"show", "menu_item"},
		Handle: func(argument *admin.ActionArgument) error {
			arg, ok := argument.Argument.(*transferArg)
			fmt.Printf("transfer arg: %v\n", arg)
			if !ok {
				return errors.New("unxepected argument type")
			}
			transactions := []*trendcoin.TransactionData{}
			mover := argument.Context.CurrentUser.(*models.User)
			reason := fmt.Sprintf(
				"User %v(%v) trigger transfer action in qor, comment: '%v'",
				mover.ID, mover.GetName(), arg.Comment,
			)
			for _, record := range argument.FindSelectedRecords() {
				user, ok := record.(models.User)
				if !ok {
					return errors.New("unxepected record type")
				}
				transactions = append(transactions, &trendcoin.TransactionData{
					Source:      uint64(user.ID),
					Destination: uint64(arg.Destination.ID),
					Amount:      arg.Amount,
					AllowCredit: arg.AllowCredit,
					Reason:      reason,
				})
			}
			return performTransactions(transactions...)
		},
	})
}

func performTransactions(transactions ...*trendcoin.TransactionData) error {
	// @TODO add local checks after service tests
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := api.TrendcoinServiceClient.MakeTransactions(
		ctx,
		&trendcoin.MakeTransactionsRequest{Transactions: transactions},
	)
	if err != nil {
		return err
	}
	if res.Error != "" {
		return errors.New(res.Error)
	}
	return nil
}

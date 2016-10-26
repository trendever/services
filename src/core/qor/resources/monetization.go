package resources

import (
	"core/models"
	"github.com/qor/admin"
	"proto/payment"
)

var currencyCollection []string

func init() {
	for _, name := range payment.Currency_name {
		currencyCollection = append(currencyCollection, name)
	}

	addResource(models.MonetizationPlan{}, &admin.Config{
		Name: "Plans",
		Menu: []string{"Monetization"},
	}, initPlansResource)
	addResource(models.CoinsOffer{}, &admin.Config{
		Name: "CoinsPackages",
		Menu: []string{"Monetization"},
	}, initCoinsPackagesResource)
}

func initPlansResource(res *admin.Resource) {
	res.Meta(&admin.Meta{
		Name:       "PrimaryCurrency",
		Type:       "select_one",
		Collection: currencyCollection,
	})
	res.Meta(&admin.Meta{
		Name: "About",
		Type: "text",
	})
	res.Meta(&admin.Meta{
		Name:      "CoinsExchangeRateForCommissionCharge",
		FieldName: "CoinsExchangeRate",
	})
	res.IndexAttrs("-About", "-CoinsExchangeRateForCommissionCharge")
	attrs := &admin.Section{
		Rows: [][]string{
			{"Name", "PrimaryCurrency"},
			{"About"},
			{"SubscriptionPeriod", "SubscriptionPrice"},
			{"TransactionCommission", "CoinsExchangeRateForCommissionCharge"},
			{"Public"},
		},
	}
	res.NewAttrs(attrs)
	res.EditAttrs(attrs)
}

func initCoinsPackagesResource(res *admin.Resource) {
	res.Meta(&admin.Meta{
		Name:       "Currency",
		Type:       "select_one",
		Collection: currencyCollection,
	})
	res.Meta(&admin.Meta{
		Name:      "CoinsInPackage",
		FieldName: "Amount",
	})
	attrs := &admin.Section{
		Rows: [][]string{
			{"Currency"},
			{"Amount", "Price"},
		},
	}
	res.NewAttrs(attrs)
	res.EditAttrs(attrs)
}

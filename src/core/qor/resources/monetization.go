package resources

import (
	"core/models"
	"github.com/qor/admin"
	"proto/payment"
)

func init() {
	addResource(models.MonetizationPlan{}, &admin.Config{
		Name: "MonetizationPlan",
	}, initMonetizationResource)
}

func initMonetizationResource(res *admin.Resource) {
	var currencyCollection []string
	for _, name := range payment.Currency_name {
		currencyCollection = append(currencyCollection, name)
	}
	res.Meta(&admin.Meta{
		Name:       "PrimaryCurrency",
		Type:       "select_one",
		Collection: currencyCollection,
	})
	res.Meta(&admin.Meta{
		Name: "About",
		Type: "text",
	})
	res.IndexAttrs("-About", "-CoinsOffers")
	offersRes := res.Meta(&admin.Meta{Name: "CoinsOffers"}).Resource
	attrs := &admin.Section{
		Rows: [][]string{
			{"Amount", "Price"},
		},
	}
	offersRes.NewAttrs(attrs)
	offersRes.EditAttrs(attrs)
}

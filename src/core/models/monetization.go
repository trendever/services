package models

import (
	"proto/core"
	"utils/db"
)

type MonetizationPlan struct {
	db.Model
	Name               string
	About              string `gorm:"text"`
	PrimaryCurrency    string
	SubscriptionPeriod uint64
	// in coins, zero if plan has no subscription fee
	SubscriptionPrice uint64
	// zero for commission-free plans
	TransactionCommission float64
	// for commission charge
	CoinsExchangeRate float64

	Public bool
}

type CoinsOffer struct {
	ID       uint64 `gorm:"primary_key"`
	Amount   uint64
	Currency string
	Price    uint64
}

func (offer CoinsOffer) Encode() *core.CoinsOffer {
	return &core.CoinsOffer{
		Id:       offer.ID,
		Amount:   offer.Amount,
		Price:    offer.Price,
		Currency: offer.Currency,
	}
}

func (plan MonetizationPlan) Encode() *core.MonezationPlan {
	ret := &core.MonezationPlan{
		Id:                    plan.ID,
		Name:                  plan.Name,
		About:                 plan.About,
		PrimaryCurrency:       plan.PrimaryCurrency,
		SubscriptionPeriod:    plan.SubscriptionPeriod,
		SubscriptionPrice:     plan.SubscriptionPrice,
		TransactionCommission: plan.TransactionCommission,
		CoinsExchangeRate:     plan.CoinsExchangeRate,
		Public:                plan.Public,
	}
	return ret
}

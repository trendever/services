package models

import (
	"github.com/jinzhu/gorm"
	"proto/core"
)

type MonetizationPlan struct {
	gorm.Model
	Name              string
	About             string `gorm:"text"`
	PrimaryCurrency   string
	SubscriptionPrice uint64
	CommissionRate    float64
	CoinsExchangeRate float64
	Public            bool

	CoinsOffers []CoinsOffer `gorm:"ForeignKey:PlanID"`
}

type CoinsOffer struct {
	ID     uint64 `gorm:"primary_key"`
	PlanID uint64
	Amount uint64
	// in primary currency of plan
	Price uint64
}

func (offer CoinsOffer) Encode() *core.ConsOffer {
	return &core.ConsOffer{
		Id:     offer.ID,
		Amount: offer.Amount,
		Price:  offer.Price,
	}
}

func (plan MonetizationPlan) Encode() *core.MonezationPlan {
	ret := &core.MonezationPlan{
		Id:                uint64(plan.ID),
		Name:              plan.Name,
		About:             plan.About,
		PrimaryCurrency:   plan.PrimaryCurrency,
		SubscriptionPrice: plan.SubscriptionPrice,
		CommissionRate:    plan.CommissionRate,
		CoinsExchangeRate: plan.CoinsExchangeRate,
		Public:            plan.Public,
	}
	for _, offer := range plan.CoinsOffers {
		ret.ConsOffers = append(ret.ConsOffers, offer.Encode())
	}
	return ret
}

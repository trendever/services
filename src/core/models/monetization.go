package models

import (
	"core/conf"
	"fmt"
	"proto/core"
	"time"
	"utils/db"
	"utils/log"
)

// duration multiplier for plans periods
var PlansBaseDuration = time.Minute

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

func (plan *MonetizationPlan) AfterCommit() {
	if plan.ID == InitialPlan.ID {
		InitialPlan = *plan
	}
}

var defaultIniPlan = MonetizationPlan{
	Name:               "init",
	About:              "default initial plan",
	PrimaryCurrency:    "RUB",
	SubscriptionPeriod: 7,
}

var InitialPlan MonetizationPlan

func InitializeMonetization() error {
	var err error
	PlansBaseDuration, err = time.ParseDuration(conf.GetSettings().Monetization.PlansBaseDuration)
	if err != nil {
		return fmt.Errorf("failed to parce PlansBaseDuration: %v", err)
	}
	name := conf.GetSettings().Monetization.InitialPlanName
	if name == "" {
		name = "init"
	}
	res := db.New().First(&InitialPlan, "name = ?", name)
	if res.RecordNotFound() {
		log.Warn("Initial plan with name %v not found, creating dalault one", name)
		InitialPlan = defaultIniPlan
		InitialPlan.Name = name
		return db.New().Create(&InitialPlan).Error
	}
	return res.Error
}

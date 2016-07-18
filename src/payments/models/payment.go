package models

import "github.com/jinzhu/gorm"

type Payment struct {
	gorm.Model
	LeadID uint64
	//order number in external payment system
	OrderID            string
	ShopCardNumber     string
	CustomerCardNumber string
	Amount             int
}

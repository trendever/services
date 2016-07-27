package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"proto/core"
)

// ShopCard contains payment card info
type ShopCard struct {
	gorm.Model

	ShopID uint `gorm:"index"`

	Name   string `gorm:"not null"`
	Number string `gorm:"not null"`
}

// ShopCards is a collection of ShopCard
type ShopCards []ShopCard

// TableName for this shop
func (c ShopCard) TableName() string {
	return "products_shops_cards"
}

// CardRepository is access layer for cards (used for mocks)
type CardRepository interface {
	GetShopSupplierID(shopID uint) (uint, error)
	GetShopSellers(shopID uint) ([]uint, error)
	CreateCard(card *ShopCard) error
	GetCardByID(id uint) (*ShopCard, error)
	DeleteCardByID(id uint) error
	GetCardsForShop(shopID uint) ([]ShopCard, error)
	GetUserByID(id uint) (*User, error)
}

// CardRepositoryImpl is database access to cards
type CardRepositoryImpl struct {
	DB *gorm.DB
}

var (
	errHasNoPerm = errors.New("Has no permissions to get this shop cards")
)

// =*=
// Database Access mockable stuff
// =*=

// GetShopSupplierID get shop supplier by ID
func (r CardRepositoryImpl) GetShopSupplierID(shopID uint) (uint, error) {

	var result []uint

	err := r.DB.Model(Shop{}).Where("id = ?", shopID).Pluck("supplier_id", &result).Error

	if len(result) < 1 {
		return 0, err
	}

	return result[0], err
}

// GetCardByID returns card with given ID
func (r CardRepositoryImpl) GetCardByID(id uint) (*ShopCard, error) {
	var card ShopCard

	err := r.DB.
		Where("id = ?", id).
		Find(&card).
		Error

	return &card, err
}

// DeleteCardByID deletes a card
func (r CardRepositoryImpl) DeleteCardByID(id uint) error {
	return r.DB.
		Where("id = ?", id).
		Delete(&ShopCard{}).
		Error
}

// GetCardsForShop deletes a card
func (r CardRepositoryImpl) GetCardsForShop(shopID uint) (res []ShopCard, err error) {

	err = r.DB.
		Where("shop_id = ?", shopID).
		Find(&res).
		Error

	return
}

// CreateCard creates a given card
func (r CardRepositoryImpl) CreateCard(card *ShopCard) error {

	return r.DB.
		Create(card).
		Error
}

// GetUserByID is a wrapper for user creation
func (r CardRepositoryImpl) GetUserByID(id uint) (*User, error) {

	return GetUserByID(id)
}

// GetShopSellers get shop supplier by ID
func (r CardRepositoryImpl) GetShopSellers(shopID uint) ([]uint, error) {

	var result []uint

	err := r.DB.Table("products_shops_sellers").Where("shop_id = ?", shopID).Pluck("seller_id", &result).Error

	return result, err
}

// =*=
// Logic stuff part
// =*=

// HasShopPermission checks if user can access shop cards
func HasShopPermission(r CardRepository, userID, shopID uint) error {

	// check if user is superseller
	{
		user, err := r.GetUserByID(userID)
		if err != nil {
			return err
		}

		if user.SuperSeller {
			return nil
		}
	}

	// check if user is supplier
	{
		// check shop supplier
		supplierID, err := r.GetShopSupplierID(shopID)
		if err != nil {
			return err
		}

		if supplierID == userID {
			return nil
		}
	}

	// check if user is seller
	{

		sellerIds, err := r.GetShopSellers(shopID)

		// check shop sellers
		if err != nil {
			return err
		}

		for _, id := range sellerIds {
			if id == userID {
				return nil
			}
		}
	}

	return errHasNoPerm
}

// CreateCard creates a card for the shop
func CreateCard(r CardRepository, userID uint, card ShopCard) (uint, error) {

	err := HasShopPermission(r, userID, card.ShopID)
	if err != nil {
		return 0, err
	}

	err = r.CreateCard(&card)

	return card.ID, err
}

// DeleteCard deletes card with ID
func DeleteCard(r CardRepository, userID, cardID uint) error {

	card, err := r.GetCardByID(cardID)
	if err != nil {
		return err
	}

	err = HasShopPermission(r, userID, card.ShopID)
	if err != nil {
		return err
	}

	return r.DeleteCardByID(cardID)
}

// GetCardsFor gets cards for given shops
func GetCardsFor(r CardRepository, userID, shopID uint) ([]ShopCard, error) {

	err := HasShopPermission(r, userID, shopID)
	if err != nil {
		return nil, err
	}

	return r.GetCardsForShop(shopID)

}

// =*=
// RPC copypasta-like stuff (@TODO: gen it?)
// =*=

//Encode converts collection of cards to proto
func (c ShopCards) Encode() []*core.ShopCard {
	out := make([]*core.ShopCard, 0, len(c))
	for _, card := range c {
		out = append(out, card.Encode())
	}
	return out
}

//Encode shop card
func (c ShopCard) Encode() *core.ShopCard {
	return &core.ShopCard{
		Id:     uint64(c.ID),
		Name:   c.Name,
		Number: c.Number,
	}
}

//Decode shop card
func (c ShopCard) Decode(p *core.ShopCard) ShopCard {
	if p == nil {
		return c
	}

	return ShopCard{
		Model: gorm.Model{
			ID: uint(p.Id),
		},
		ShopID: uint(p.ShopId),
		Name:   p.Name,
		Number: p.Number,
	}
}

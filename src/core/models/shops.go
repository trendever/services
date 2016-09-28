package models

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"proto/core"
	"strings"
	"time"
	"utils/db"
	"utils/nats"
)

// Shop model defines virtual "Shop"
//  shop is instagram user used for selling items
type Shop struct {
	gorm.Model
	// @TODO it is just shop local name now, rename it later
	InstagramUsername string `gorm:"index"`
	// Supplier is a real user who act his responsible for shop
	SupplierID        uint
	Supplier          User
	SupplierLastLogin time.Time
	oldSupplier       uint

	ShippingRules    string `gorm:"type:text"`
	PaymentRules     string `gorm:"type:text"`
	InstagramWebsite string `gorm:"type:text"`

	// Sell managers, assigned to this shop
	Sellers []*User `gorm:"many2many:products_shops_sellers;"`

	// Shop's payment cards
	Cards []ShopCard

	// Shop private tags; internal use only
	Tags           []Tag `gorm:"many2many:products_shops_tags;"`
	NotifySupplier bool

	// it's better to keep them outside main struct to avoid load surplus data
	Notes []ShopNote `gorm:"ForeignKey:ShopID"`
}

type ShopNote struct {
	ID     uint64 `gorm:"primary_key"`
	ShopID uint64
	Text   string `gorm:"text"`
}

// TableName for this shop
func (s Shop) TableName() string {
	return "products_shops"
}

// ResourceName for this model
func (s Shop) ResourceName() string {
	return "Shop"
}

// Stringify returns human-friendly name
func (s Shop) Stringify() string {
	return fmt.Sprintf("%s shop", s.InstagramUsername)
}

//Encode converts Shop to core.Shop
func (s Shop) Encode() *core.Shop {
	return &core.Shop{
		Id:                 int64(s.ID),
		InstagramId:        s.Supplier.InstagramID,
		InstagramUsername:  s.Supplier.InstagramUsername,
		InstagramFullname:  s.Supplier.InstagramFullname,
		InstagramAvatarUrl: s.Supplier.InstagramAvatarURL,
		InstagramCaption:   s.Supplier.InstagramCaption,
		InstagramWebsite:   s.InstagramWebsite,
		ShippingRules:      s.ShippingRules,
		PaymentRules:       s.PaymentRules,
		Caption:            s.Supplier.Caption,
		Slogan:             s.Supplier.Slogan,
		Sellers:            Users(s.Sellers).PublicEncode(),
		Supplier:           s.Supplier.PublicEncode(),
		//don't remove, useful when not fully preloaded
		SupplierId: int64(s.SupplierID),
		AvatarUrl:  s.Supplier.AvatarURL,
		Available:  s.NotifySupplier,
	}
}

//Decode converts core.Shop to Shop
func (s Shop) Decode(cs *core.Shop) Shop {
	if cs == nil {
		return s
	}

	return Shop{
		Model: gorm.Model{
			ID: uint(cs.Id),
		},
		InstagramUsername: cs.InstagramUsername,
		SupplierID:        uint(cs.SupplierId),
		ShippingRules:     cs.ShippingRules,
		PaymentRules:      cs.PaymentRules,
		NotifySupplier:    cs.Available,
	}
}

//gorm callbacks
func (s *Shop) BeforeSave(db *gorm.DB) {
	s.InstagramUsername = strings.ToLower(s.InstagramUsername)
}

func (s *Shop) BeforeUpdate() error {
	err := db.New().Model(&s).Select("supplier_id").Where("id = ?", s.ID).Row().Scan(&s.oldSupplier)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to load old supplier for shop %v: %v", s.ID, err)
	}
	return nil
}

func (s *Shop) AfterUpdate() error {
	if s.oldSupplier != 0 && s.oldSupplier != s.SupplierID {
		err := s.onSupplierChanged()
		if err != nil {
			return err
		}
	}
	go nats.Publish("core.shop.flush", s.ID)
	return nil
}

func (s *Shop) AfterDelete() {
	go nats.Publish("core.shop.flush", s.ID)
}

func (s *Shop) onSupplierChanged() error {
	var conversations []uint64
	err := db.New().
		Model(&Lead{}).
		Select("conversation_id").
		Where("shop_id = ?", s.ID).
		Where("state NOT IN ('EMPTY','NEW')").
		Pluck("conversation_id", &conversations).
		Error
	if err != nil {
		return fmt.Errorf("failed to load related conversations list for shop %v: %v", s.ID, err)
	}
	if s.Supplier.ID == 0 {
		err := db.New().First(&s.Supplier, "id = ?", s.SupplierID)
		return fmt.Errorf("failed to load supplier for shop %v: %v", s.ID, err)
	}
	for _, conversation := range conversations {
		err = joinChat(conversation, chat.MemberRole_SUPPLIER, &s.Supplier)
		if err != nil {
			return fmt.Errorf("failed to add supplier to chat %v: %v", conversation, err)
		}
	}
	return nil
}

//GetShopByID returns Shop by ID
func GetShopByID(shopID uint) (*Shop, error) {
	shop := &Shop{}
	err := db.New().Preload("Supplier").Find(shop, shopID).Error
	return shop, err
}

//GetShopByInstagramName returns Shop by shop instagram username
func GetShopByInstagramName(name string) (*Shop, error) {
	shop := &Shop{}
	err := db.New().Preload("Supplier").Scopes(InstagramUsernameScope(name)).Find(shop).Error
	return shop, err
}

//FindShopIDByInstagramName returns only shop ID by shop instagram username
func FindShopIDByInstagramName(name string) (id uint, err error) {
	shop, err := GetShopByInstagramName(name)
	return shop.ID, err
}

//GetSellersByShopID returns shop sellers by shop ID
func GetSellersByShopID(id uint) ([]*User, error) {
	shop := &Shop{}
	err := db.New().Preload("Sellers").Find(shop, id).Error
	return shop.Sellers, err
}

//GetShopsIDWhereUserIsSeller returns shops ids where user is a seller
func GetShopsIDWhereUserIsSeller(userID uint) (out []uint64, err error) {
	err = db.New().Table("products_shops_sellers").Where("user_id = ?", userID).Pluck("shop_id", &out).Error
	return
}

//GetShopsIDWhereUserIsSupplier returns shops ids where user is a supplier
func GetShopsIDWhereUserIsSupplier(userID uint) (out []uint64, err error) {
	err = db.New().Table("products_shops").Where("supplier_id = ?", userID).Select("id").Pluck("id", &out).Error
	return
}

func FindOrCreateShopForSupplier(supplier *User, recreateDeleted bool) (shopID uint64, deleted bool, err error) {
	scope := db.New().Where("supplier_id = ?", supplier.ID)
	if !recreateDeleted {
		scope = scope.Unscoped().Order("deleted_at IS NOT NULL")
	}
	var shop Shop
	res := scope.First(&shop)
	if res.Error != nil && !res.RecordNotFound() {
		return 0, false, fmt.Errorf("failed to check fot existing shop: %v", res.Error)
	}
	if shop.ID != 0 {
		if shop.DeletedAt == nil {
			return uint64(shop.ID), false, nil
		}
		if !recreateDeleted {
			return uint64(shop.ID), true, nil
		}
	}
	new := Shop{
		InstagramUsername: supplier.GetName(),
		SupplierID:        supplier.ID,
	}
	err = db.New().Save(&new).Error
	if err != nil {
		return 0, false, fmt.Errorf("failed to save created shop: %v", err)
	}
	return uint64(new.ID), false, nil
}

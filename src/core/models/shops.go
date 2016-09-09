package models

import (
	"core/api"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"proto/core"
	"strings"
	"utils/db"
)

// Shop model defines virtual "Shop"
//  shop is instagram user used for selling items
type Shop struct {
	gorm.Model

	// Instagram fields
	InstagramID        uint64 `gorm:"index"`
	InstagramUsername  string `gorm:"index"`
	InstagramFullname  string
	InstagramAvatarURL string
	InstagramWebsite   string
	AvatarURL          string `gorm:"text"`

	// instagram calls this `biography`. can be really long
	InstagramCaption string `gorm:"type:text"`

	// Supplier is a real user who act his responsible for shop
	SupplierID  uint
	Supplier    User
	oldSupplier uint

	ShippingRules string `gorm:"type:text"`
	PaymentRules  string `gorm:"type:text"`
	Caption       string `gorm:"type:text"`
	Slogan        string

	// Sell managers, assigned to this shop
	Sellers []*User `gorm:"many2many:products_shops_sellers;"`

	// Shop's payment cards
	Cards []ShopCard

	// Shop private tags; internal use only
	Tags           []Tag `gorm:"many2many:products_shops_tags;"`
	NotifySupplier bool
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
		InstagramId:        s.InstagramID,
		InstagramUsername:  s.InstagramUsername,
		InstagramFullname:  s.InstagramFullname,
		InstagramAvatarUrl: s.InstagramAvatarURL,
		InstagramCaption:   s.InstagramCaption,
		InstagramWebsite:   s.InstagramWebsite,
		ShippingRules:      s.ShippingRules,
		PaymentRules:       s.PaymentRules,
		Caption:            s.Caption,
		Slogan:             s.Slogan,
		Sellers:            Users(s.Sellers).PublicEncode(),
		Supplier:           s.Supplier.PublicEncode(),
		//don't remove, useful when not fully preloaded
		SupplierId: int64(s.SupplierID),
		AvatarUrl:  s.AvatarURL,
		Available:  s.NotifySupplier,
	}
}

//Decode converts core.Shop to Shop
func (s Shop) Decode(cs *core.Shop) Shop {
	// @CHECK: why was that necessary?
	if cs == nil {
		return s
	}

	return Shop{
		Model: gorm.Model{
			ID: uint(cs.Id),
		},

		InstagramID:        cs.InstagramId,
		InstagramUsername:  cs.InstagramUsername,
		InstagramFullname:  cs.InstagramFullname,
		InstagramAvatarURL: cs.InstagramAvatarUrl,
		InstagramCaption:   cs.InstagramCaption,
		InstagramWebsite:   cs.InstagramWebsite,
		AvatarURL:          cs.AvatarUrl,
		SupplierID:         uint(cs.SupplierId),
		ShippingRules:      cs.ShippingRules,
		PaymentRules:       cs.PaymentRules,
		Caption:            cs.Caption,
		Slogan:             cs.Slogan,
		NotifySupplier:     cs.Available,
	}
}

//gorm callbacks
func (s *Shop) BeforeSave(db *gorm.DB) {
	s.InstagramUsername = strings.ToLower(s.InstagramUsername)
}

func (s *Shop) BeforeUpdate() error {
	err := db.New().Model(&s).Select("supplier_id").Where("id = ?", s.ID).Row().Scan(&s.oldSupplier)
	if err != nil {
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
	go api.Publish("core.shop.flush", s.ID)
	return nil
}

func (s *Shop) AfterDelete() {
	go api.Publish("core.shop.flush", s.ID)
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

//CreateNewShop creates a new shop, and if needed a new supplier
func CreateNewShop(model *Shop) error {

	if model.SupplierID == 0 {
		supplier, err := createNewSupplier(model)
		if err != nil {
			return err
		}
		model.SupplierID = supplier.ID
	}

	if err := db.New().Create(model).Error; err != nil {
		return err
	}

	return nil
}

func createNewSupplier(model *Shop) (*User, error) {
	supplier := &User{}
	//check, maybe we already have user with this instagram_id
	if scope := db.New().Find(supplier, "instagram_id = ?", model.InstagramID); scope.Error != nil && !scope.RecordNotFound() {
		return supplier, scope.Error
	}
	//if not, create new one
	if db.New().NewRecord(supplier) {
		supplier = &User{
			InstagramID:        model.InstagramID,
			InstagramUsername:  model.InstagramUsername,
			InstagramFullname:  model.InstagramFullname,
			InstagramAvatarURL: model.InstagramAvatarURL,
			InstagramCaption:   model.InstagramCaption,
		}

		if err := db.New().Create(supplier).Error; err != nil {
			return supplier, err
		}
	}
	return supplier, nil
}

//GetShopByID returns Shop by ID
func GetShopByID(shopID uint) (*Shop, error) {
	shop := &Shop{}
	err := db.New().Find(shop, shopID).Error
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

// CheckShopByID returns false only if shop exists and deleted
func CheckShopByID(id uint) (bool, error) {
	// check shop existance
	var count int
	err := db.New().
		Model(Shop{}).
		Where("deleted_at is not null").
		Where("id = ?", id).
		Count(&count).
		Error
	return count == 0, err
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

func FindOrCreateShopForSupplier(supplier *User) (shopID uint64, err error) {
	existing, err := GetShopsIDWhereUserIsSupplier(supplier.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to check fot existing shop: %v", err)
	}
	if len(existing) != 0 {
		return existing[0], nil
	}
	new := Shop{
		InstagramID:        supplier.InstagramID,
		InstagramUsername:  supplier.InstagramUsername,
		InstagramFullname:  supplier.InstagramFullname,
		InstagramAvatarURL: supplier.InstagramAvatarURL,
		InstagramCaption:   supplier.InstagramCaption,
		SupplierID:         supplier.ID,
		AvatarURL:          supplier.AvatarURL,
	}
	err = db.New().Save(&new).Error
	if err != nil {
		return 0, fmt.Errorf("failed to save created shop: %v", err)
	}
	return uint64(new.ID), nil
}

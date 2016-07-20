package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/core"
	"core/db"
	"strings"
)

// Shop model defines virtual "Shop"
//  shop is instagram user used for selling items
type Shop struct {
	gorm.Model

	// Instagram fields
	InstagramID        uint64
	InstagramUsername  string `gorm:"index"`
	InstagramFullname  string
	InstagramAvatarURL string
	InstagramWebsite   string
	AvatarURL          string `gorm:"text"`

	// instagram calls this `biography`. can be really long
	InstagramCaption string `gorm:"type:text"`

	// Supplier is a real user who act his responsible for shop
	SupplierID uint
	Supplier   User

	ShippingRules string `gorm:"type:text"`
	PaymentRules  string `gorm:"type:text"`
	Caption       string `gorm:"type:text"`
	Slogan        string

	// Sell managers, assigned to this shop
	Sellers []User `gorm:"many2many:products_shops_sellers;"`

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
	}
}

//BeforeSave gorm callback
func (s *Shop) BeforeSave(db *gorm.DB) {
	s.InstagramUsername = strings.ToLower(s.InstagramUsername)
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
func GetSellersByShopID(id uint) ([]User, error) {
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
func GetShopsIDWhereUserIsSeller(userID uint) (out []uint, err error) {
	err = db.New().Table("products_shops_sellers").Where("user_id = ?", userID).Pluck("shop_id", &out).Error
	return
}

//GetShopsIDWhereUserIsSupplier returns shops ids where user is a supplier
func GetShopsIDWhereUserIsSupplier(userID uint) (out []uint, err error) {
	err = db.New().Table("products_shops").Where("supplier_id = ?", userID).Select("id").Pluck("id", &out).Error
	return
}

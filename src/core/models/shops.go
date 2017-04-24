package models

import (
	"database/sql"
	"fmt"
	"proto/chat"
	"proto/core"
	"strings"
	"time"
	"utils/db"
	"utils/nats"

	"github.com/jinzhu/gorm"
)

const (
	SupplierChangedTopic = "shop_supplier_changed"
)

func init() {
	notifyTopics := []string{
		SupplierChangedTopic,
	}
	for _, t := range notifyTopics {
		RegisterNotifyTemplate(t)
	}
}

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

	// @TODO separate data somehow, there is no need to carry it all the time
	Location         string `gorm:"type:text"`
	WorkingTime      string `gorm:"type:text"`
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
	// if true new leads will not be joined with existing one
	SeparateLeads bool

	// monetization related stuff
	PlanID uint64
	Plan   MonetizationPlan
	// used to prevent repetitive actions
	LastPlanUpdate time.Time
	// zero time for plans without expiration
	PlanExpiresAt time.Time
	// if true subscription will be renewed automatically after expiration
	AutoRenewal bool
	// there may be some time between subscription expiration and autorenewal,
	// so  any logic should use field below instead of compare expiration time with current one.
	// true if shop was no active monetization plan
	Suspended bool

	// it's better to keep them outside main struct to avoid load surplus data
	Notes []ShopNote `gorm:"ForeignKey:ShopID"`
}

// ShopNote model for keeping notes about shops in qor
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

	if strings.HasSuffix(strings.ToLower(s.InstagramUsername), "shop") {
		return s.InstagramUsername
	}

	return fmt.Sprintf("%s shop", s.InstagramUsername)
}

//Encode converts Shop to core.Shop
func (s Shop) Encode() *core.Shop {
	ret := &core.Shop{
		Id:                 int64(s.ID),
		InstagramId:        s.Supplier.InstagramID,
		InstagramUsername:  s.Supplier.InstagramUsername,
		InstagramFullname:  s.Supplier.InstagramFullname,
		InstagramAvatarUrl: s.Supplier.InstagramAvatarURL,
		InstagramCaption:   s.Supplier.InstagramCaption,
		InstagramWebsite:   s.InstagramWebsite,
		Location:           s.Location,
		WorkingTime:        s.WorkingTime,
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

		CreatedAt: uint64(s.CreatedAt.Unix()),

		PlanId:      s.PlanID,
		Suspended:   s.Suspended,
		AutoRenewal: s.AutoRenewal,
	}
	if !s.PlanExpiresAt.IsZero() {
		ret.PlanExpiresAt = s.PlanExpiresAt.Unix()
	}
	return ret
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

// BeforeSave gorm callbacks
func (s *Shop) BeforeCreate(db *gorm.DB) {
	s.InstagramUsername = strings.ToLower(s.InstagramUsername)
	s.PlanID = InitialPlan.ID
	if InitialPlan.SubscriptionPeriod != 0 {
		s.PlanExpiresAt = time.Now().Add(PlansBaseDuration * time.Duration(InitialPlan.SubscriptionPeriod))
	}
}

func (s *Shop) BeforeSave(db *gorm.DB) {
	s.InstagramUsername = strings.ToLower(s.InstagramUsername)
}

// BeforeUpdate hook
func (s *Shop) BeforeUpdate() error {
	err := db.New().Model(&s).Select("supplier_id").Where("id = ?", s.ID).Row().Scan(&s.oldSupplier)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to load old supplier for shop %v: %v", s.ID, err)
	}
	return nil
}

// AfterUpdate hook
func (s *Shop) AfterUpdate() error {
	if s.oldSupplier != 0 && s.oldSupplier != s.SupplierID {
		err := s.onSupplierChanged()
		if err != nil {
			return err
		}
		go GetNotifier().NotifyUserByID(
			uint64(s.oldSupplier),
			SupplierChangedTopic,
			map[string]interface{}{"shop": s},
		)
	}
	go nats.Publish("core.shop.flush", s.ID)
	return nil
}

// AfterDelete hook
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

// sellers should be preloaded
func (s *Shop) HasSeller(userID uint) bool {
	for _, seller := range s.Sellers {
		if seller.ID == userID {
			return true
		}
	}
	return false
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

func GetShopProductsCount(shopID uint64) (count uint64, err error) {
	err = db.New().Model(Product{}).Where("shop_id = ?", shopID).Count(&count).Error
	return
}

// FindOrCreateShopForSupplier func
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

// FindOrCreateAttachedShop func
func FindOrCreateAttachedShop(supplierID uint64, shopInstagramUsername string) (shopID uint64, err error) {
	err = SetInstagramForUser(supplierID, shopInstagramUsername)
	if err != nil {
		return 0, fmt.Errorf("failed to update instagram username: %v", err)
	}

	var shop Shop
	res := db.New().
		Or("instagram_username = ?", shopInstagramUsername).
		Or("supplier_id = ?", supplierID).
		First(&shop)

	if res.Error != nil && !res.RecordNotFound() {
		return 0, fmt.Errorf("failed to check fot existing shop: %v", res.Error)
	}

	switch {
	case shop.ID == 0:
		shop = Shop{
			InstagramUsername: shopInstagramUsername,
			SupplierID:        uint(supplierID),
		}

	case uint64(shop.SupplierID) != supplierID:
		shop.SupplierID = uint(supplierID)

	// @CHECK do we realy need to change it?
	case shop.InstagramUsername != shopInstagramUsername:
		shop.InstagramUsername = shopInstagramUsername
	// everything totally match, just return shop id
	default:
		return uint64(shop.ID), nil
	}

	err = db.New().Save(&shop).Error
	if err != nil {
		return 0, fmt.Errorf("failed to save shop: %v", err)
	}
	return uint64(shop.ID), nil
}

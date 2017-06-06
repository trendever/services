package models

import (
	"database/sql"
	"fmt"
	"proto/chat"
	"proto/core"
	"savetrend/tumbmap"
	"time"
	"utils/db"
	"utils/nats"

	"github.com/jinzhu/gorm"
)

const productTable = "products_product"

// Product model
type Product struct {
	gorm.Model

	Code   string `gorm:"not null"`
	Title  string `gorm:"text"`
	IsSale bool

	InstagramImageCaption string `gorm:"type:text"`
	InstagramImageID      string
	InstagramImageHeight  uint
	InstagramImageWidth   uint
	InstagramImageURL     string `gorm:"text"`
	InstagramLink         string
	InstagramPublishedAt  time.Time
	InstagramLikesCount   int
	InstagramImages       []ImageCandidate

	WebShopURL string `gorm:"type:text"`

	// Product shop
	ShopID uint `gorm:"index:shops_index"`
	Shop   Shop
	// for update related chats in update callbacks
	oldShop uint

	// Product mentioner
	MentionedByID uint `gorm:"index:mentioners_index"`
	MentionedBy   User

	LikedBy []*User `gorm:"many2many:users_products"`

	Items []ProductItem

	Tags []Tag `gorm:"many2many:products_product_item_tags;"`
}

//Products array of Product
type Products []*Product

//ProductFilter product filter for ProductSearcher
type ProductFilter struct {
	ShopID     uint64
	UserID     uint64
	FromID     uint64
	IsSaleOnly bool
	Limit      int
	Offset     int
	Direction  bool
	Keyword    string
	Tags       []int64
}

//ProductSearcher interface for search products
type ProductSearcher interface {
	//Search returns array of ids of products
	Search(filter ProductFilter) ([]uint, error)
}

type productSearcher struct {
	db *gorm.DB
}

//NewProductSearcher returns new ProductSearcher
func NewProductSearcher(db *gorm.DB) ProductSearcher {
	return &productSearcher{
		db: db,
	}
}

// TableName for this model
func (p Product) TableName() string {
	return productTable
}

//gorm callbacks
func (p *Product) BeforeUpdate() error {
	err := db.New().Model(&p).Select("shop_id").Where("id = ?", p.ID).Row().Scan(&p.oldShop)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to load old shop for product %v: %v", p.ID, err)
	}
	return nil
}

func (p Product) AfterUpdate() error {
	if p.oldShop != 0 && p.ShopID != p.oldShop {
		err := p.onShopChanged()
		if err != nil {
			return err
		}
	}
	go nats.Publish("core.product.flush", p.ID)
	return nil
}

func (p Product) AfterDelete() {
	go nats.Publish("core.product.flush", p.ID)
}

// this method should update all related data
// however change shop for product with already created leads is bad idea
// especially if those leads contains multiple products
func (p Product) onShopChanged() error {
	var info []struct {
		ConversationID uint64
		State          string
	}
	err := db.New().
		Raw(`
		UPDATE products_leads SET shop_id = ?
		WHERE deleted_at IS NULL
		AND EXISTS (
			SELECT 1 FROM products_leads_items related
			JOIN products_product_item item
				ON related.product_item_id = item.id
			WHERE item.product_id = ?
				AND related.lead_id = products_leads.id
				AND item.deleted_at IS NULL
		)
		RETURNING conversation_id, state`, p.ShopID, p.ID).
		Scan(&info).
		Error
	if err != nil {
		return fmt.Errorf("failed to update related leads for product %v: %v", p.ID, err)
	}
	err = db.New().Preload("Supplier").Preload("Sellers").First(&p.Shop, "id = ?", p.ShopID).Error
	if err != nil {
		return fmt.Errorf("failed to load shop %v: %v", p.ShopID, err)
	}
	for _, lead := range info {
		if lead.State == "NEW" || lead.State == "EMPTY" {
			continue
		}
		err = joinChat(lead.ConversationID, chat.MemberRole_SUPPLIER, &p.Shop.Supplier)
		if err != nil {
			return fmt.Errorf("failed to add supplier to chat %v: %v", lead.ConversationID, err)
		}
		err = joinChat(lead.ConversationID, chat.MemberRole_SELLER, p.Shop.Sellers...)
		if err != nil {
			return fmt.Errorf("failed to add sellers to chat %v: %v", lead.ConversationID, err)
		}
	}
	return nil
}

// Validate fields
// @TODO: make good and full checks
func (p Product) Validate(db *gorm.DB) {
	// @TODO: do smth with product code. save in on PostSave hook?
}

// Stringify generates product name for this model. Should not be empty!
func (p Product) Stringify() string {
	out := p.Code

	if p.Title != "" {
		out += fmt.Sprintf(" (%s)", p.Title)
	}

	return out
}

func (p Product) SmallestImg() *ImageCandidate {
	if len(p.InstagramImages) == 0 {
		return nil
	}
	ret := &p.InstagramImages[0]
	var minSize uint = 100500
	for i := range p.InstagramImages {
		img := &p.InstagramImages[i]
		info, ok := tumbmap.ThumbByName[img.Name]
		if !ok {
			continue
		}
		if info.Size < minSize {
			ret = img
			minSize = info.Size
		}
	}
	return ret
}

// ImageCandidate contains instagram image info
// @CHECK any seance in this compose key?
type ImageCandidate struct {
	ID        uint       `gorm:"primary_key"`
	ProductID uint       `gorm:"primary_key;index:products_index"`
	UpdatedAt time.Time  `gorm:"index"`
	DeletedAt *time.Time `gorm:"index"`
	URL       string     `gorm:"text"`
	Name      string     // small string; no need of text type
}

//ImageCandidates array of ProductItems
type ImageCandidates []ImageCandidate

// TableName for this model
func (p ImageCandidate) TableName() string {
	return "products_product_images"
}

// ProductItem child model
type ProductItem struct {
	gorm.Model

	ProductID int64  `gorm:"index"`
	Name      string `gorm:"text"`

	Price, DiscountPrice uint64

	Tags []Tag `gorm:"many2many:products_product_item_tags;"`
}

//ProductItems array of ProductItems
type ProductItems []ProductItem

// TableName for this model
func (p ProductItem) TableName() string {
	return "products_product_item"
}

// Stringify returns human-friendly item id name
func (p ProductItem) Stringify() string {
	switch {
	case p.Name != "":
		return p.Name
	default:
		return fmt.Sprintf("Item ID=%d", p.ID)
	}
}

//ToLeadInfoItem converts to core.ProductItem
func (p ProductItem) ToLeadInfoItem() *core.ProductItem {
	item := &core.ProductItem{
		Id:            int64(p.ID),
		Name:          p.Name,
		Price:         p.Price,
		DiscountPrice: p.DiscountPrice,
	}

	return item
}

// GetURLValue makes qor links point to Product, not ProductItem
func (p ProductItem) GetURLValue() interface{} {
	// return parent stub
	return Product{Model: gorm.Model{ID: uint(p.ProductID)}}
}

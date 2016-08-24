package models

import (
	"core/api"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/core"
	"time"
)

const productTable = "products_product"

// Product model
type Product struct {
	gorm.Model

	Code   string `gorm:"not null"`
	Title  string `gorm:"text"`
	IsSale bool

	InstagramImageCaption string `gorm:"type:text"`
	InstagramImageID      string `gorm:"unique"`
	InstagramImageHeight  uint
	InstagramImageWidth   uint
	InstagramImageURL     string `gorm:"text"`
	InstagramLink         string
	InstagramPublishedAt  time.Time
	InstagramLikesCount   int
	InstagramImages       []ImageCandidate

	// Product shop
	ShopID uint `gorm:"index:shops_index"`
	Shop   Shop

	// Product mentioner
	MentionedByID uint `gorm:"index:mentioners_index"`
	MentionedBy   User

	LikedBy []User `gorm:"many2many:users_products"`

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

//AfterUpdate is a gorm callback
func (p Product) AfterUpdate() {
	go api.Publish("core.product.flush", p.ID)
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

// ResourceName returns qor resource name
func (p ProductItem) ResourceName() string {
	return "ProductItem"
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

package models

import (
	"common/db"
	"core/conf"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"proto/core"
	"savetrend/tumbmap"
	"time"
	"utils/mandible"
	"utils/nats"
	"utils/product_code"
)

const productTable = "products_product"

var thumbsConfig = []mandible.Thumbnail{
	{
		Name:   "XL",
		Width:  1080,
		Height: 1080,
		Shape:  "thumb",
	},
	{
		Name:   "L",
		Width:  750,
		Height: 7500,
		Shape:  "thumb",
	},
	{
		Name:   "M_square",
		Width:  480,
		Height: 480,
		Shape:  "square",
	},
	{
		Name:   "S_square",
		Width:  306,
		Height: 306,
		Shape:  "square",
	},
}

// Product model
type Product struct {
	gorm.Model

	Code   string `gorm:"not null"`
	Title  string `gorm:"text"`
	IsSale bool

	// @TODO garbage data all around:
	// * InstagramImageID can be determined by InstagramLink and vice versa
	// * original image size is available in "Max" ImageCandidate
	// * InstagramLikesCount updates never
	// * most of "Instagram" prefixes are useless and annoying
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
	// it is send after main chat templates
	ChatMessage string `gorm:"type:text"`

	// Product shop
	ShopID uint `gorm:"index:shops_index"`
	Shop   Shop `gorm:"save_associations:false"`
	// for update related chats in update callbacks
	oldShop uint

	// Product mentioner
	MentionedByID uint `gorm:"index:mentioners_index"`
	MentionedBy   User `gorm:"save_associations:false"`

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

func (p *Product) BeforeSave() {
	if p.MentionedBy.ID != 0 {
		p.MentionedByID = p.MentionedBy.ID
	}
	if p.Shop.ID != 0 {
		p.ShopID = p.Shop.ID
	}
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

func (p Product) ImageURL() string {
	for _, c := range p.InstagramImages {
		if c.Name == "Max" {
			return c.URL
		}
	}
	if len(p.InstagramImages) != 0 {
		return p.InstagramImages[0].URL
	}
	return ""
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

func (p *Product) UpdateImage(url string) error {
	if p.InstagramImageURL == url {
		return nil
	}
	if p.ID == 0 {
		return errors.New("zero product id")
	}
	if url == "" {
		tx := db.NewTransaction()
		err := tx.Delete(&ImageCandidate{}, "product_id = ?", p.ID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Model(&p).Update(map[string]interface{}{
			"InstagramImageURL":    "",
			"InstagramImageHeight": 0,
			"InstagramImageWidth":  0,
		}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		p.InstagramImages = nil
		return tx.Commit().Error
	}

	resp, err := mandible.New(conf.GetSettings().MandibleURL, thumbsConfig...).DoRequest("url", url)
	if err != nil {
		return err
	}
	imgs := []ImageCandidate{
		{
			Name: "Max",
			URL:  resp.Link,
		},
	}
	for name, url := range resp.Thumbs {
		imgs = append(imgs, ImageCandidate{
			Name: name,
			URL:  url,
		})
	}

	tx := db.NewTransaction()

	err = tx.Delete(&ImageCandidate{}, "product_id = ?", p.ID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(p).Association("InstagramImages").Append(&imgs).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&p).Update(map[string]interface{}{
		"InstagramImageURL":    url,
		"InstagramImageHeight": resp.Height,
		"InstagramImageWidth":  resp.Width,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// saves updated product to database without risk to modify internal data of related objects
// if editroID in not zero, checks whether user has permissions to modify this product
// if restricted is true shop and mentioner will not be changed
//
// passed object will be filled with actual data
func UpdateProduct(updated *Product, editorID uint64, restricted bool) error {
	if updated.ID == 0 {
		return errors.New("zero product id")
	}

	product, err := GetProductByID(uint64(updated.ID), "Shop", "MentionedBy", "Items", "InstagramImages")
	if err != nil {
		return err
	}

	if editorID != 0 {
		ok, err := IsUserSupplierOrSeller(editorID, uint64(product.ShopID))
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("forbidden")
		}
	}

	if product.InstagramLink != updated.InstagramLink {
		updated.InstagramImageID, err = product_code.ParsePostURL(updated.InstagramLink)
		if err != nil {
			return errors.New("invalid instagram link")
		}
	}

	// should not be updated here
	updated.LikedBy = nil
	// gorm tends to fuck everything up by saving nested objects recursive, tag data may be changed
	// we will keep tags here and deal with them later
	allTags := make([][]uint, len(updated.Items))
	for i := range updated.Items {
		itemTags := make([]uint, len(updated.Items[i].Tags))
		for j, tag := range updated.Items[i].Tags {
			if tag.ID == 0 {
				return errors.New("zero tag id")
			}
			itemTags[j] = tag.ID
		}
		allTags[i] = itemTags
		updated.Items[i].Tags = nil
	}

	tx := db.NewTransaction()
	if restricted {
		updated.Shop = product.Shop
		updated.MentionedBy = product.MentionedBy
		// disallow item relinking from other products
		itemsMap := map[uint]bool{}
		for _, item := range product.Items {
			itemsMap[item.ID] = true
		}
		for _, item := range updated.Items {
			if item.ID != 0 && !itemsMap[item.ID] {
				return errors.New("item relinking is disallowed")
			}
		}
	}

	if product.InstagramImageURL == updated.InstagramImageURL {
		updated.InstagramImages = product.InstagramImages
	} else {
		img := updated.InstagramImageURL
		updated.InstagramImageURL = product.InstagramImageURL
		err = updated.UpdateImage(img)
		if err != nil {
			return fmt.Errorf("failed to update image: %v", err)
		}
	}

	err = tx.Save(&updated).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// we have ids for all items now, time to update tags. raw sql anyone?
	itemIDs := make([]uint, len(updated.Items))
	values := ""
	for i, item := range updated.Items {
		itemIDs[i] = item.ID
		for _, tagID := range allTags[i] {
			values = values + fmt.Sprintf("(%v, %v), ", item.ID, tagID)
		}
	}
	err = tx.Exec("DELETE FROM products_product_item_tags WHERE product_item_id in (?)", itemIDs).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if values != "" {
		// strip extra ", "
		values = values[:len(values)-2]
		err = tx.Exec("INSERT INTO products_product_item_tags (product_item_id, tag_id) VALUES " + values).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Model(&updated).Preload("Tags").Association("Items").Find(&updated.Items)
	err = tx.Commit().Error
	if err != nil {
		return errors.New("transaction failed")
	}
	updated.LikedBy = product.LikedBy
	return nil
}

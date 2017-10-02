package models

import (
	"common/db"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

//GetProductByID returns product with preloaded models
func GetProductByID(productID uint64, preloads ...string) (*Product, error) {
	product := &Product{}
	scope := db.New()
	for _, column := range preloads {
		scope = scope.Preload(column)
	}
	scope = scope.Find(product, productID)
	return product, scope.Error
}

//GetProductsByIDs returns products by products ids
func GetProductsByIDs(ids []uint, direction bool) (Products, error) {
	query := DefaultProductQuery().Where("id in (?)", ids)
	if direction {
		query = query.Order("created_at asc")
	} else {
		query = query.Order("created_at desc")
	}
	products := Products{}
	err := query.Find(&products).Error
	return products, err
}

func (p productSearcher) Search(filter ProductFilter) ([]uint, error) {
	ids := []uint{}

	query := p.db.Model(Product{}).
		Limit(filter.Limit)

	if filter.IsSaleOnly {
		query = query.Where("is_sale = ?", filter.IsSaleOnly)
	}

	switch {
	case filter.Offset > 0:
		query = query.Offset(filter.Offset)
	case filter.FromID > 0:
		if filter.Direction {
			query = query.Where(productTable+".id > ?", filter.FromID)
		} else {
			query = query.Where(productTable+".id < ?", filter.FromID)
		}
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if len(filter.Tags) > 0 {
		query = applyTags(query, filter.Tags)
	}

	if filter.Keyword != "" {
		query = applySearch(query, filter.Keyword)
	}

	if filter.Direction {
		query = query.Order(productTable + ".created_at asc")
	} else {
		query = query.Order(productTable + ".created_at desc")
	}

	// first select -- get user likes
	if filter.UserID > 0 {
		idsLiked, err := selectRows(
			query.Joins(fmt.Sprintf(`LEFT JOIN %s AS up ON up.product_id = %s.id AND up.user_id = ? AND up.deleted_at IS NULL`,
				// table names (%s)
				usersProductsTable, productTable),
				// "?" expr
				filter.UserID).
				Where("up.user_id = ?", filter.UserID),
		)
		if err != nil {
			return nil, err
		}

		ids = append(ids, idsLiked...)
	}

	if filter.ShopID > 0 {
		query = query.Where(productTable+".shop_id = ?", filter.ShopID)
	}

	if filter.UserID > 0 && filter.ShopID == 0 {
		// do nothing
		// @TODO make it beautiful
	} else {
		// second select -- get everything else (owned products, normal search, etc)
		idsFin, err := selectRows(query)
		if err != nil {
			return nil, err
		}
		ids = append(ids, idsFin...)
	}

	return ids, nil
}

// scan for id -- createdAt tuple; for use in previous func
func selectRows(db *gorm.DB) ([]uint, error) {

	ids := []uint{}

	rows, err := db.Select("DISTINCT products_product.id,products_product.created_at").Rows()
	if err != nil {
		return nil, err
	}

	var (
		id        uint
		createdAt time.Time
	)

	for rows.Next() {
		err = rows.Scan(&id, &createdAt)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil

}

// query modifier that filters products only to ones that own every tag in tag_ids
func applyTags(db *gorm.DB, tagIds []int64) *gorm.DB {

	db = db.Joins("INNER JOIN products_product_item as it ON (products_product.id = it.product_id)")

	for i, tagID := range tagIds {
		relname := fmt.Sprintf("tagrel_%v", i)
		db = db.Joins(
			fmt.Sprintf(
				"INNER JOIN products_product_item_tags as %v ON (%v.product_item_id = it.id AND %v.tag_id = ?)",
				relname, relname, relname),
			tagID,
		)
	}

	return db
}

func applySearch(db *gorm.DB, query string) *gorm.DB {

	// @TODO: good fulltext search including items
	search := "%" + query + "%" //@TODO: Security impact?
	return db.
		Where("(products_product.title LIKE ? OR products_product.code LIKE ?)", search, search)
}

// DefaultProductQuery returns default query for searching products
func DefaultProductQuery() *gorm.DB {
	return db.New().
		Preload("Items").
		Preload("Items.Tags").  // preload will load this fields into each found product
		Preload("MentionedBy"). //otherwise, they are initialized with default values
		Preload("InstagramImages").
		Preload("Shop")
}

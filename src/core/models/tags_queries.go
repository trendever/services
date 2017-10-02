package models

import (
	"common/db"
)

// return tags from items that have all passed tags. Passed tags themselves are excluded
func RelatedTags(tags []int64, limit int) ([]Tag, error) {
	var result []Tag
	err := db.New().
		Limit(limit).
		Joins("JOIN products_tag_group grp ON (grp.id = products_tag.group_id AND grp.name != products_tag.name)").
		Where(`
		products_tag.id IN (
			SELECT tag_id FROM products_product_item_tags WHERE product_item_id IN (
				SELECT tagged.id FROM (
					SELECT product_item_id AS id FROM products_product_item_tags
					WHERE tag_id IN (?)
					GROUP BY product_item_id HAVING COUNT(1) = ?
				) AS tagged
				JOIN products_product_item item
				ON item.id = tagged.id AND item.deleted_at IS NULL
				JOIN products_product product
				ON product.id = item.product_id AND product.deleted_at IS NULL AND product.is_sale
			) AND tag_id NOT IN (?)
		)
		`, tags, len(tags), tags).
		Find(&result).
		Error

	return result, err

}

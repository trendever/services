package models

import (
	"fmt"
	"utils/db"
)

// get all tags that each product from relatedProducts(tags) has
// tags from tags slice are excluded
// plus main tags
func relatedTagIDs(tags []int64, limit int) ([]int64, error) {

	related, err := relatedTagsIds(tags, limit)

	return related, err
}

func relatedTagsIds(tags []int64, limit int) ([]int64, error) {

	var related []int64

	query := db.New()

	// generate tag joins
	var relname string
	for i, tagID := range tags {
		relname = fmt.Sprintf("tagrel_%v", i)
		query = query.Joins(
			fmt.Sprintf(
				"INNER JOIN products_product_item_tags as %v ON (%v.product_item_id = it.id AND %v.tag_id = ?)",
				relname, relname, relname),
			tagID,
		)
	}

	// final joins: only new and not hidden tags
	query = query.
		Joins("INNER JOIN products_product_item_tags as finrel"+
			fmt.Sprintf(" ON (%v.product_item_id = finrel.product_item_id AND finrel.tag_id NOT IN (?))", relname),
			tags,
		).
		Joins("INNER JOIN products_tag as pt ON (pt.id = finrel.tag_id AND pt.main = false AND hidden = false)").
		Joins("INNER JOIN products_product as pr ON (pr.id = it.product_id AND is_sale = true AND pr.deleted_at is null)").
		// filter out groups
		Joins("INNER JOIN products_tag_group as grp ON (grp.id = pt.group_id AND grp.name != pt.name)")

	rows, err := query.
		Limit(limit).
		Table("products_product_item as it").
		Group("pt.id").
		// sort by tag weight
		Order("pt.position").
		// smash copies
		Select("DISTINCT pt.id,pt.position").
		Rows()

	if err != nil {
		return nil, err
	}

	var id, position int64
	for rows.Next() {
		err = rows.Scan(&id, &position)
		if err != nil {
			return nil, err
		}

		related = append(related, id)
	}

	return related, err

}

// mainTagIds returns main tag ID
func mainTagIds(ignore []int64, limit int) ([]int64, error) {
	var result []int64
	err := db.New().
		Model(&Tag{}).
		Where("main = ?", true).
		Where("hidden = ?", false).
		Where("id NOT IN (?)", ignore).
		Limit(limit).
		Pluck("id", &result).
		Error

	return result, err
}

// MainTags func returns main tags
func MainTags(limit int) ([]Tag, error) {

	var result []Tag
	err := db.New().
		Where("main = ?", true).
		Where("hidden = ?", false).
		Limit(limit).
		Find(&result).
		Error

	return result, err

}

// RelatedTags func returns main tags
func RelatedTags(tags []int64, limit int) ([]Tag, error) {

	ids, err := relatedTagIDs(tags, limit)
	if err != nil {
		return nil, err
	}

	var result []Tag
	err = db.New().
		Limit(limit).
		Where("id in (?)", ids).
		Find(&result).
		Error

	return result, err

}

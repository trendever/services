package models

import (
	"fmt"
	"core/db"
)

//Migrate runs migrations
func Migrate() error {

	db.New().Model(&UsersProducts{}).AddForeignKey("user_id", "users_user(id)", "CASCADE", "RESTRICT")
	db.New().Model(&UsersProducts{}).AddForeignKey("product_id", "products_product(id)", "CASCADE", "RESTRICT")

	db.New().Model(&Product{}).AddForeignKey("mentioned_by_id", "users_user(id)", "CASCADE", "RESTRICT")
	db.New().Model(&Product{}).AddForeignKey("shop_id", "products_shops(id)", "CASCADE", "RESTRICT")

	db.New().Model(&ProductItem{}).AddForeignKey("product_id", "products_product(id)", "CASCADE", "RESTRICT")

	db.New().Model(&Lead{}).AddForeignKey("shop_id", "products_shops(id)", "CASCADE", "RESTRICT")
	db.New().Model(&Lead{}).AddForeignKey("customer_id", "users_user(id)", "CASCADE", "RESTRICT")

	db.New().Model(&Shop{}).AddForeignKey("supplier_id", "users_user(id)", "CASCADE", "RESTRICT")

	db.New().Model(&Tag{}).AddForeignKey("group_id", "products_tag_group(id)", "CASCADE", "RESTRICT")

	db.New().Table("products_leads_items").AddForeignKey("lead_id", "products_leads(id)", "CASCADE", "RESTRICT")
	db.New().Table("products_leads_items").AddForeignKey("product_item_id", "products_product_item(id)", "CASCADE", "RESTRICT")

	db.New().Table("products_product_item_tags").AddForeignKey("product_item_id", "products_product_item(id)", "CASCADE", "RESTRICT")
	db.New().Table("products_product_item_tags").AddForeignKey("tag_id", "products_tag(id)", "CASCADE", "RESTRICT")

	db.New().Table("products_shops_tags").AddForeignKey("shop_id", "products_shops(id)", "CASCADE", "RESTRICT")
	db.New().Table("products_shops_tags").AddForeignKey("tag_id", "products_tag(id)", "CASCADE", "RESTRICT")

	db.New().Table("products_shops_sellers").AddForeignKey("shop_id", "products_shops(id)", "CASCADE", "RESTRICT")
	db.New().Table("products_shops_sellers").AddForeignKey("user_id", "users_user(id)", "CASCADE", "RESTRICT")

	migrateTagrel()
	isSellerLabelMigrate()

	db.New().Exec("ALTER TABLE products_product ALTER COLUMN instagram_image_url TYPE text;")

	for _, col := range []string{"is_seller", "is_admin", "is_scout", "super_seller"} {
		db.New().Exec(fmt.Sprintf("UPDATE users_user SET %v = false WHERE %v is null", col, col))
	}

	db.New().Table("products_product_images").AddForeignKey("product_id", "products_product(id)", "CASCADE", "RESTRICT")

	db.New().Exec("UPDATE products_leads SET chat_updated_at=updated_at WHERE chat_updated_at IS NULL")

	db.New().Model(&ImageCandidate{}).AddIndex("idx_products_product_images_product_id", "product_id")

	return nil
}

func migrateTagrel() {

	err := db.New().Exec("ALTER TABLE products_product_item_tags ADD COLUMN product_id integer").Error

	if err == nil {
		db.New().Table("products_product_item_tags").AddForeignKey("product_id", "products_product(id)", "CASCADE", "RESTRICT")
		db.New().Exec(
			`
			UPDATE products_product_item_tags as tagrel
				SET product_id = p.product_id 
				FROM products_product_item as p 
				WHERE p.id = tagrel.product_item_id
				`,
		)
		db.New().Exec(
			`
CREATE OR REPLACE FUNCTION products_product_tagrel_product_id_set () RETURNS trigger 
LANGUAGE  plpgsql AS '
BEGIN
NEW.product_id = (SELECT product_id FROM products_product_item WHERE id=NEW.product_item_id);
RETURN NEW;
END;
';

CREATE TRIGGER products_product_tagrel_trigger
BEFORE INSERT ON products_product_item_tags FOR EACH ROW
EXECUTE PROCEDURE products_product_tagrel_product_id_set();
`,
		)
	}
}

func isSellerLabelMigrate() {

	var count int
	err := db.New().Model(User{}).Where("is_seller = true").Count(&count).Error

	if err == nil && count == 0 {

		var customers []uint

		err := db.New().Model(&User{}).Joins("INNER JOIN products_shops_sellers as ps ON ps.user_id = users_user.id").Pluck("id", &customers).Error

		if err == nil && len(customers) > 0 {
			db.New().Model(&User{}).Where("id in (?)", customers).Update("is_seller", true)
		}
	}
}

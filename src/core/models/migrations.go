package models

import (
	"fmt"
	"time"
	"utils/db"
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

	db.New().Model(&Product{}).AddIndex("idx_products_product_instagram_image_id", "instagram_image_id")

	db.New().Model(&PushToken{}).AddForeignKey("user_id", "users_user(id)", "CASCADE", "RESTRICT")

	// i'm somewhat unsure if drop something here is good idea
	db.New().Model(&EmailTemplate{}).
		DropColumn("model_name").DropColumn("preloads").DropColumn("to")
	db.New().Model(&SMSTemplate{}).
		DropColumn("model_name").DropColumn("preloads").DropColumn("to")

	db.New().Model(&Lead{}).Where("source LIKE '@%'").Update("source", "wantit")

	db.New().Exec("DROP INDEX IF EXISTS unique_active_lead")
	db.New().Model(&Shop{}).Where("separate_leads IS NULL").UpdateColumn("separate_leads", "false")

	db.New().Model(&Lead{}).AddForeignKey("cancel_reason_id", "lead_cancel_reasons(id)", "SET NULL", "RESTRICT")

	db.New().Exec("UPDATE users_user SET confirmed = (phone != '') WHERE confirmed IS NULL")

	relationsIndices()

	db.New().Model(&ChatTemplateMessage{}).AddForeignKey("template_id", "chat_templates(id)", "CASCADE", "RESTRICT")

	if db.HasColumn(&ChatTemplateMessage{}, "case_id") {
		db.New().Exec(`
			UPDATE chat_template_messages msg
			SET template_id = c.template_id
			FROM chat_template_cases c WHERE c.id = msg.case_id
		`)
		db.New().Model(&ChatTemplateMessage{}).DropColumn("case_id")
	}

	db.New().Model(&ChatTemplate{}).Where("source IS NULL OR source = ''").UpdateColumn("source", "any")

	err := setInitialPlan()
	if err != nil {
		return err
	}

	if db.HasColumn(&EmailTemplate{}, "deleted_at") {
		for _, table := range []interface{}{&PushTemplate{}, &SMSTemplate{}, &EmailTemplate{}} {
			db.New().Delete(table, "deleted_at IS NOT NULL")
			db.New().Model(table).DropColumn("deleted_at").DropColumn("created_at").DropColumn("updated_at")
		}
	}

	return nil
}

func setInitialPlan() error {
	err := InitializeMonetization()
	if err != nil {
		return fmt.Errorf("failed to init monetization: %v", err)
	}
	updateMap := map[string]interface{}{
		"plan_id":          InitialPlan.ID,
		"suspended":        false,
		"auto_renewal":     false,
		"last_plan_update": time.Now(),
		"plan_expires_at":  time.Time{},
	}
	if InitialPlan.SubscriptionPeriod != 0 {
		updateMap["plan_expires_at"] = time.Now().Add(PlansBaseDuration * time.Duration(InitialPlan.SubscriptionPeriod))
	}
	return db.New().Model(&Shop{}).Where("plan_id IS NULL OR plan_id = 0").UpdateColumns(updateMap).Error
}

func relationsIndices() {
	db.New().Model(&Product{}).AddIndex("shops_index", "shop_id")
	db.New().Model(&Product{}).AddIndex("mentioners_index", "mentioner_id")
	// already exist with another name
	//db.New().Model(&ProductItem{}).AddIndex("products_index", "product_id")
	db.New().Model(&ImageCandidate{}).AddIndex("products_index", "product_id")
}

func migrateTagrel() {

	err := db.New().Exec("ALTER TABLE products_product_item_tags ADD COLUMN product_item_id integer").Error

	if err == nil {
		db.New().Table("products_product_item_tags").AddForeignKey("product_item_id", "products_product_item(id)", "CASCADE", "RESTRICT")

		db.New().Exec(`
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
		`)
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

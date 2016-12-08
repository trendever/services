package sync

import (
	"database/sql"
	"elasticsync/config"
	"elasticsync/models"
	"fmt"
	"github.com/lib/pq"
	"gopkg.in/olivere/elastic.v3"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"utils/db"
	ewrapper "utils/elastic"
	"utils/log"
)

type SyncStrategy int

const (
	// basic strategy
	// it compares update time of everything related to document in db with last sync time
	// on large data set that will lead to heavy db load, but it's main way to handle out-of-sync situations
	RangeStrategy SyncStrategy = iota
	// event-based strategy
	// it updates indexed documents on demand(flush events on nats)
	// there is much less db load, but index should be synchronized from start
	QueueStrategy
)

var strategyMap = map[SyncStrategy]func() (SyncStrategy, error){
	RangeStrategy: RangeIndex,
	QueueStrategy: QueueIndex,
}

func Loop() {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	conf := config.Get()

	var err error
	strategy := RangeStrategy
	for {
		strategy, err = strategyMap[strategy]()
		if err != nil {
			log.Error(err)
		}

		select {
		case <-time.After(time.Millisecond * time.Duration(conf.Delay)):
		case <-interrupt:
			log.Info("elasticsync service stopped")
			os.Exit(0)
		}
	}
}

func RangeIndex() (nextStrategy SyncStrategy, err error) {
	nextStrategy = RangeStrategy
	// @NOTICE we have no direct information about adding/deleting tags
	// but items should be touched as well if this action performed in qor or with gorm relations in general
	rows, err := db.New().Raw(`
SELECT product.id
FROM
	products_product product
	LEFT JOIN products_shops shop ON product.shop_id = shop.id
	LEFT JOIN users_user supplier ON shop.supplier_id = supplier.id
	LEFT JOIN users_user mentioner ON product.mentioned_by_id = mentioner.id
	LEFT JOIN elastic_product_meta index ON product.id = index.id
WHERE
	index.id IS NULL
	OR product.updated_at > index.source_updated_at
	OR product.deleted_at > index.source_updated_at
	OR shop.updated_at > index.source_updated_at
	OR shop.deleted_at > index.source_updated_at
	OR supplier.updated_at > index.source_updated_at
	OR supplier.deleted_at > index.source_updated_at
	OR mentioner.updated_at > index.source_updated_at
	OR mentioner.deleted_at > index.source_updated_at
UNION
SELECT item.product_id
FROM
	products_product_item item
	LEFT JOIN elastic_product_meta index ON item.product_id = index.id
WHERE
	index.id IS NULL
	OR item.updated_at > index.source_updated_at
	OR item.deleted_at > index.source_updated_at
UNION
SELECT itag.product_id
FROM
	products_product_item_tags itag
	JOIN products_tag tag ON tag.id = itag.tag_id
	LEFT JOIN elastic_product_meta index ON itag.product_id = index.id
WHERE
	index.id IS NULL
	OR tag.updated_at > index.source_updated_at
	OR tag.deleted_at > index.source_updated_at
LIMIT ?
	`, config.Get().ChunkSize).Rows()
	if err != nil {
		return RangeStrategy, err
	}
	defer rows.Close()

	ids := make([]uint64, 0, config.Get().ChunkSize)
	var tmp uint64
	for rows.Next() {
		rows.Scan(&tmp)
		ids = append(ids, tmp)
	}
	log.Debug("got %v ids with range query", len(ids))

	if len(ids) > 0 {
		var products map[uint64]*models.ElasticProduct
		products, err = LoadProducts(ids)
		if err != nil {
			return
		}
		err = IndexProducts(products)
		if err != nil {
			return
		}
	}

	if len(ids) < config.Get().ChunkSize {
		// everything should be synchronized, switching to event-driven strategy
		nextStrategy = QueueStrategy
	}
	return
}

func QueueIndex() (SyncStrategy, error) {
	// @TODO
	return RangeStrategy, nil
}

func LoadProducts(ids []uint64) (products map[uint64]*models.ElasticProduct, err error) {
	rows, err := db.New().Raw(`
SELECT
	product.id, product.updated_at, product.deleted_at,
	product.code, product.title, product.instagram_image_caption, product.is_sale,
	shop.id, shop.updated_at, shop.deleted_at, shop.location,
	supplier.id, supplier.updated_at, supplier.deleted_at,
	supplier.instagram_username, supplier.instagram_fullname,
	mentioner.id, mentioner.updated_at, mentioner.deleted_at,
	mentioner.name, mentioner.instagram_username, mentioner.instagram_fullname,
	index.version
FROM
	products_product product
	LEFT JOIN products_shops shop ON product.shop_id = shop.id
	LEFT JOIN users_user supplier ON shop.supplier_id = supplier.id
	LEFT JOIN users_user mentioner ON product.mentioned_by_id = mentioner.id
	LEFT JOIN elastic_product_meta index ON product.id = index.id
WHERE
	product.id in (?)`, ids).Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	products = make(map[uint64]*models.ElasticProduct)
	times := make([]pq.NullTime, 8, 8)
	var shop_id, mentioner_id, version sql.NullInt64
	for rows.Next() {
		p := &models.ElasticProduct{}
		var alter_name string
		err = db.NilScan(rows,
			&p.Meta.ID, &times[0], &times[1],
			&p.Data.Code, &p.Data.Title, &p.Data.Caption, &p.Data.Sale,
			&shop_id, &times[2], &times[3], &p.Data.Shop.Location,
			&p.Data.Shop.Supplier, &times[4], &times[5],
			&p.Data.Shop.Name, &p.Data.Shop.FullName,
			&mentioner_id, &times[6], &times[7],
			&p.Data.Mentioner.Name, &alter_name, &p.Data.Mentioner.FullName,
			&version,
		)
		if err != nil {
			return
		}
		p.Data.ID = p.Meta.ID
		p.Data.Shop.ID = uint64(shop_id.Int64)
		p.Data.Mentioner.ID = uint64(mentioner_id.Int64)
		p.Meta.SourceUpdatedAt = maxNullTime(times)
		p.Meta.Version = int(version.Int64)
		// product was deleted softly
		if times[1].Valid {
			p.Deleted = true
		}
		if p.Data.Mentioner.Name == "" {
			p.Data.Mentioner.Name = alter_name
		}
		products[p.Meta.ID] = p
	}
	err = LoadProductsItems(ids, products)
	if err != nil {
		return
	}
	err = LoadProductsTags(ids, products)
	if err != nil {
		return
	}
	err = LoadProductsImages(ids, products)

	return
}

func maxNullTime(arr []pq.NullTime) (max time.Time) {
	for _, t := range arr {
		if t.Valid {
			if max.Before(t.Time) {
				max = t.Time
			}
		}
	}
	return
}

func LoadProductsTags(product_ids []uint64, products map[uint64]*models.ElasticProduct) error {
	rows, err := db.New().
		Select("tag.id, tag.name, tag.updated_at, tag.deleted_at, tag.hidden, related.product_id, related.product_item_id").
		Table("products_product_item_tags related").
		Joins("JOIN products_tag tag ON related.tag_id = tag.id").
		Where("related.product_id in (?)", product_ids).
		Order("related.product_id, related.product_item_id").
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var tag struct {
		models.ElasticTag
		UpdatedAt pq.NullTime
		DeletedAt pq.NullTime
		Hidden    bool
		ProductID uint64
		ItemID    uint64
	}

	var prodCur *models.ElasticProduct = products[product_ids[0]]
	// current item index
	var itemCur int
	for rows.Next() {
		err := db.NilScan(rows,
			&tag.ID, &tag.Name,
			&tag.UpdatedAt, &tag.DeletedAt,
			&tag.Hidden, &tag.ProductID, &tag.ItemID,
		)
		if err != nil {
			return err
		}
		if tag.ProductID != prodCur.Meta.ID {
			prodCur = products[tag.ProductID]
			itemCur = 0
		}
		for tag.ItemID != prodCur.Data.Items[itemCur].ID {
			itemCur++
			if itemCur == len(prodCur.Data.Items) {
				return fmt.Errorf("inconsistent data: missing item %v in product %v", tag.ItemID, prodCur.Data.ID)
			}
		}
		if prodCur.Meta.SourceUpdatedAt.Before(tag.DeletedAt.Time) {
			prodCur.Meta.SourceUpdatedAt = tag.DeletedAt.Time
		}
		if prodCur.Meta.SourceUpdatedAt.Before(tag.UpdatedAt.Time) {
			prodCur.Meta.SourceUpdatedAt = tag.UpdatedAt.Time
		}
		if !tag.Hidden {
			prodCur.Data.Items[itemCur].Tags = append(prodCur.Data.Items[itemCur].Tags, tag.ElasticTag)
		}
	}
	return nil
}

func LoadProductsItems(product_ids []uint64, products map[uint64]*models.ElasticProduct) error {
	rows, err := db.New().
		Select("id, product_id, updated_at, deleted_at, name, price, discount_price").
		Table("products_product_item").
		Where("product_id in (?)", product_ids).
		Order("product_id, id").
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var itemMeta struct {
		UpdatedAt pq.NullTime
		DeletedAt pq.NullTime
		ProductID uint64
	}

	var cur *models.ElasticProduct = products[product_ids[0]]
	for rows.Next() {
		item := &models.ElasticProductItem{}
		err = db.NilScan(rows,
			&item.ID, &itemMeta.ProductID, &itemMeta.UpdatedAt, &itemMeta.DeletedAt,
			&item.Name, &item.Price, &item.DiscountPrice,
		)
		if err != nil {
			return err
		}
		if itemMeta.ProductID != cur.Meta.ID {
			cur = products[itemMeta.ProductID]
		}
		if cur.Meta.SourceUpdatedAt.Before(itemMeta.DeletedAt.Time) {
			cur.Meta.SourceUpdatedAt = itemMeta.DeletedAt.Time
		}
		if cur.Meta.SourceUpdatedAt.Before(itemMeta.UpdatedAt.Time) {
			cur.Meta.SourceUpdatedAt = itemMeta.UpdatedAt.Time
		}
		cur.Data.Items = append(cur.Data.Items, item)
	}
	return nil
}

func LoadProductsImages(product_ids []uint64, products map[uint64]*models.ElasticProduct) error {
	rows, err := db.New().
		Select("product_id, updated_at, deleted_at, url, name").
		Table("products_product_images").
		Where("product_id in (?)", product_ids).
		Order("product_id").
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var image struct {
		models.ElasticProductImage
		UpdatedAt pq.NullTime
		DeletedAt pq.NullTime
		ProductID uint64
	}

	var cur *models.ElasticProduct = products[product_ids[0]]
	for rows.Next() {
		err = db.NilScan(rows,
			&image.ProductID, &image.UpdatedAt, &image.DeletedAt,
			&image.URL, &image.Name,
		)
		if err != nil {
			return err
		}
		if image.ProductID != cur.Meta.ID {
			cur = products[image.ProductID]
		}
		if cur.Meta.SourceUpdatedAt.Before(image.DeletedAt.Time) {
			cur.Meta.SourceUpdatedAt = image.DeletedAt.Time
		}
		if cur.Meta.SourceUpdatedAt.Before(image.UpdatedAt.Time) {
			cur.Meta.SourceUpdatedAt = image.UpdatedAt.Time
		}
		cur.Data.Images = append(cur.Data.Images, image.ElasticProductImage)
	}
	return nil
}

func IndexProducts(products map[uint64]*models.ElasticProduct) error {
	el := ewrapper.Cli()
	// @TODO version control
	bulk := el.Bulk().Index("products").Type("product")
	for id, p := range products {
		if p.Deleted {
			bulk.Add(elastic.NewBulkDeleteRequest().Id(strconv.FormatUint(id, 10)))
		} else {
			bulk.Add(elastic.NewBulkIndexRequest().Id(strconv.FormatUint(id, 10)).Doc(&p.Data))
		}
	}
	res, err := bulk.Do()
	if err != nil {
		return err
	}
	successCount, err := saveMeta(products, res)
	log.Debug("%v documents were indexed/deleted successefuly", successCount)
	return err
}

func saveMeta(products map[uint64]*models.ElasticProduct, res *elastic.BulkResponse) (successCount uint64, err error) {
	var placeholders string
	arguments := []interface{}{}
	for _, chunk := range res.Items {
		for action, item := range chunk {
			if (item.Status < 200 || item.Status >= 300) &&
				!(item.Status == 404 && action == "delete") {

				err = &elastic.Error{Status: item.Status, Details: item.Error}
				continue
			}
			id, _ := strconv.ParseUint(item.Id, 10, 64)
			meta := &products[id].Meta
			if action == "delete" {
				meta.Version = -1
			} else {
				meta.Version = item.Version
			}
			placeholders += "(?, ?, ?),\n"
			arguments = append(arguments, meta.ID, meta.Version, meta.SourceUpdatedAt)
			successCount++
		}
	}
	if err != nil {
		log.Errorf("at last one action failed, last error: %v", err)
	}
	if len(arguments) == 0 {
		return
	}

	// @TODO replace all this with one upsert after upgrade to postgres 9.5+
	tx := db.New().Begin()
	defer tx.Commit()
	err = tx.Exec(`
CREATE TEMPORARY TABLE IF NOT EXISTS new_meta
(id bigint, version integer, source_updated_at timestamp with time zone)
ON COMMIT DELETE ROWS;
	`).Error
	if err != nil {
		return
	}
	err = tx.Exec("LOCK TABLE elastic_product_meta IN EXCLUSIVE MODE").Error
	if err != nil {
		return
	}
	err = tx.Exec(
		"INSERT INTO new_meta(id, version, source_updated_at) VALUES"+placeholders[:len(placeholders)-2]+";",
		arguments...,
	).Error
	if err != nil {
		return
	}
	err = tx.Exec(`
UPDATE elastic_product_meta
SET version = new_meta.version, source_updated_at = new_meta.source_updated_at
FROM new_meta
WHERE new_meta.id = elastic_product_meta.id
	`).Error
	if err != nil {
		return
	}
	err = tx.Exec(`
INSERT INTO elastic_product_meta
SELECT new_meta.id, new_meta.version, new_meta.source_updated_at
FROM new_meta
LEFT OUTER JOIN elastic_product_meta ON (elastic_product_meta.id = new_meta.id)
WHERE elastic_product_meta.id IS NULL;
	`).Error
	return
}

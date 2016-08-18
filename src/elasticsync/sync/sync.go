package sync

import (
	"elasticsync/config"
	"elasticsync/models"
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
		case <-interrupt:
			log.Info("elasticsync service stopped")
			os.Exit(0)
		default:
			time.Sleep(time.Second * time.Duration(conf.Delay))
		}
	}
}

func RangeIndex() (nextStrategy SyncStrategy, err error) {
	nextStrategy = RangeStrategy
	// @CHECK can postgres handle limit on union effective?
	rows, err := db.New().Raw(`
SELECT product.id
FROM
	products_product product
	LEFT JOIN products_shops shop ON product.shop_id = shop.id
	LEFT JOIN users_user mentioner ON product.mentioned_by_id = mentioner.id
	LEFT JOIN elastic_product_indices index ON product.id = index.product_id
WHERE
	index.product_id IS NULL
	OR product.updated_at > index.source_updated_at
	OR product.deleted_at > index.source_updated_at
	OR shop.updated_at > index.source_updated_at
	OR shop.deleted_at > index.source_updated_at
	OR mentioner.updated_at > index.source_updated_at
	OR mentioner.deleted_at > index.source_updated_at
UNION
SELECT item.product_id
FROM
	products_product_item item
	LEFT JOIN elastic_product_indices index ON item.product_id = index.product_id
WHERE
	index.product_id IS NULL
	OR item.updated_at > index.source_updated_at
	OR item.deleted_at > index.source_updated_at
UNION
SELECT itag.product_id
FROM
	products_product_item_tags itag
	JOIN products_tag tag ON tag.id = itag.tag_id
	LEFT JOIN elastic_product_indices index ON itag.product_id = index.product_id
WHERE
	index.product_id IS NULL
	OR tag.updated_at > index.source_updated_at
	OR tag.deleted_at > index.source_updated_at
LIMIT ?
	`, config.Get().ChunkSize).Rows()
	if err != nil {
		return RangeStrategy, err
	}
	defer rows.Close()

	var ids []uint64
	var tmp uint64
	for rows.Next() {
		rows.Scan(&tmp)
		ids = append(ids, tmp)
	}
	products, err := LoadProducts(ids)
	if err != nil {
		return
	}
	err = IndexProducts(products)
	if err != nil {
		return
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
	// @TODO
	return products, nil
}

func IndexProducts(products map[uint64]*models.ElasticProduct) error {
	el := ewrapper.Cli()
	// @TODO versions, deletions
	bulk := el.Bulk().Index("products").Type("product")
	for id, p := range products {
		bulk.Add(elastic.NewBulkIndexRequest().Id(strconv.FormatUint(id, 10)).Doc(&p.ElasticProductData))
	}
	res, bulkErr := bulk.Do()
	var succeeded []models.ElasticProductIndex
	for _, item := range res.Succeeded() {
		id, _ := strconv.ParseUint(item.Id, 10, 64)
		p := products[id]
		p.Version = item.Version
		succeeded = append(succeeded, p.ElasticProductIndex)
	}
	if succeeded != nil {
		err := db.New().Save(succeeded).Error
		if err != nil {
			return err
		}
	}
	return bulkErr
}

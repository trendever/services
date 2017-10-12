package models

import (
	"common/db"
	"common/log"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"strconv"
	ewrapper "utils/elastic"
)

var indices = []struct {
	name   string
	body   string
	dbMeta interface{}
}{
	{
		name:   "products",
		body:   ProductIndex,
		dbMeta: &ElasticProductMeta{},
	},
}

func Migrate(drop bool) {
	db := db.New()
	el := ewrapper.Cli()

	for _, index := range indices {
		var oldDate uint64
		ret, err := el.IndexGetSettings(index.name).Do()
		if err != nil {
			e, ok := err.(*elastic.Error)
			if !ok || e.Details.Type != "index_not_found_exception" {
				log.Fatal(err)
			}
		} else {
			oldDate, err = strconv.ParseUint(ret[index.name].Settings["index"].(map[string]interface{})["creation_date"].(string), 10, 64)
		}

		if drop || oldDate < IndexUpdatedAt {
			db.DropTableIfExists(index.dbMeta)
			for _, index := range indices {
				el.DeleteIndex(index.name).Do()
			}
		}
		if err = db.AutoMigrate(index.dbMeta).Error; err != nil {
			log.Fatal(err)
		}
		_, err = el.CreateIndex(index.name).BodyString(fmt.Sprintf(index.body, IndexUpdatedAt)).Do()
		if err != nil {
			e, ok := err.(*elastic.Error)
			if !ok || e.Details.Type != "index_already_exists_exception" {
				log.Fatal(err)
			}
		}
	}

	db.New().Model(&ElasticProductMeta{}).AddForeignKey("id", "products_product(id)", "CASCADE", "RESTRICT")
}

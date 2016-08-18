package models

import (
	"gopkg.in/olivere/elastic.v3"
	"utils/db"
	ewrapper "utils/elastic"
	"utils/log"
)

var dbModels = []interface{}{
	&ElasticProductIndex{},
}

var elasticIndexes = []struct {
	name string
	body string
}{
	{
		name: "products",
		body: ProductIndex,
	},
}

func Migrate(drop bool) {
	db := db.New()
	el := ewrapper.Cli()
	if drop {
		log.Warn("Droping tables")
		db.DropTableIfExists(dbModels)
		log.Warn("Droping indexes")
		for _, index := range elasticIndexes {
			el.DeleteIndex(index.name).Do()
		}
	}

	if err := db.AutoMigrate(dbModels...).Error; err != nil {
		log.Fatal(err)
	}
	for _, index := range elasticIndexes {
		_, err := el.CreateIndex(index.name).BodyString(index.body).Do()
		if err != nil {
			e, ok := err.(*elastic.Error)
			if !ok || e.Details.Type != "index_already_exists_exception" {
				log.Fatal(err)
			}
		}
	}
}

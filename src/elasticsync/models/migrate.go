package models

import (
	"utils/db"
	"utils/elastic"
	"utils/log"
)

var dbModels = []interface{}{
	&ElasticProductIndexed{},
}

var elasticIndexes = []struct {
	name string
	json string
}{
	{
		name: "products",
		json: ProductIndex,
	},
}

func Migrate(drop bool) {
	db := db.New()
	el := elastic.Cli()
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
		_, err := el.CreateIndex(index.name).BodyString(index.json).Do()
		if err != nil {
			log.Fatal(err)
		}
	}
}

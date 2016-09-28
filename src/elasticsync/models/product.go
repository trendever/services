package models

import "time"

// last index update time
// it will be included as creation date in index settings
// index will be recreated if old data is lower than this
// unix time, ms
const IndexUpdatedAt = 1475067569 * 1000
const ProductIndex = `{
"settings": {
	"analysis": {
		"filter": {
			"ru_stop": {
				"type": "stop",
				"stopwords": "_russian_"
			},
			"ru_stemmer": {
				"type": "stemmer",
				"language": "russian"
			}
		},
		"analyzer": {
			"default": {
				"char_filter": ["html_strip"],
				"tokenizer": "standard",
				"filter": [
					"lowercase",
					"ru_stop",
					"ru_stemmer"
				]
			}
		}
	},
	"creation_date" : %v
},
"mappings": {
	"product": {
		"properties": {
			"id": {
				"type": "long",
				"include_in_all": false
			},
			"code": {
				"type": "string",
				"index": "not_analyzed"
			},
			"title": {
				"type": "string"
			},
			"caption": {
				"type": "string"
			},
			"sale": {
				"type": "boolean"
			},
			"shop": {
				"properties": {
					"id": {
						"type": "long",
						"include_in_all": false
					},
					"name": {
						"type": "string",
						"index": "not_analyzed"
					},
					"full_name": {
						"type": "string",
						"index": "not_analyzed"
					}
				}
			},
			"mentioner": {
				"properties": {
					"id": {
						"type": "long",
						"include_in_all": false
					},
					"name": {
						"type": "string",
						"index": "not_analyzed"
					},
					"full_name": {
						"type": "string",
						"index": "not_analyzed"
					}
				}
			},
			"items": {
				"type": "nested",
				"properties": {
					"name": {
						"type": "string"
					},
					"price": {
						"type": "long",
						"include_in_all": false
					},
					"discount_price": {
						"type": "long",
						"include_in_all": false
					},
					"tags": {
						"properties": {
							"id": {
								"type": "long",
								"include_in_all": false
							},
							"name": {
								"type": "string",
								"fields": {
									"raw": {
										"type": "string",
										"index": "not_analyzed",
										"include_in_all": false
									}
								}
							}
						}
					}
				}
			},
			"images": {
				"include_in_all": false,
				"properties": {
					"url": {
						"type": "string",
						"index": "not_analyzed"
					},
					"name": {
						"type": "string",
						"index": "not_analyzed"
					}
				}
			}
		}
	}
}
}`

type ElasticTag struct {
	ID   uint64 `json:"id"`
	Name string `json:"name,omitempty"`
}
type ElasticProductItem struct {
	ID            uint64       `json:"-"`
	Name          string       `json:"name,omitempty"`
	Price         uint64       `json:"price"`
	DiscountPrice uint64       `json:"discount_price"`
	Tags          []ElasticTag `json:"tags,omitempty"`
}
type ElasticProductImage struct {
	URL  string `json:"url,omitempty"`
	Name string `json:"name,omitempty"`
}

type ElasticProductData struct {
	ID      uint64 `json:"id"`
	Code    string `json:"code"`
	Title   string `json:"title,omitempty"`
	Caption string `json:"caption,omitempty"`
	Sale    bool   `json:"sale,omitempty"`
	Shop    struct {
		ID       uint64 `json:"id"`
		Name     string `json:"name,omitempty"`
		FullName string `json:"full_name,omitempty"`
	} `json:"shop,omitempty"`
	Mentioner struct {
		ID       uint64 `json:"id"`
		Name     string `json:"name,omitempty"`
		FullName string `json:"full_name,omitempty"`
	} `json:"mentioner,omitempty"`
	Items  []*ElasticProductItem `json:"items,omitempty"`
	Images []ElasticProductImage `json:"images,omitempty"`
}

// represent relation table in db
type ElasticProductMeta struct {
	// product id
	ID uint64 `gorm:"primary_key"`
	// current elastic version of document
	// -1 if product was deleted
	Version         int       `gorm:"not null"`
	SourceUpdatedAt time.Time `gorm:"index;not null"`
}

type ElasticProduct struct {
	Meta    ElasticProductMeta
	Data    ElasticProductData
	Deleted bool
}

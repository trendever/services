package models

import "time"

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
	}
},
"mappings": {
	"product": {
		"properties": {
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
					}
				}
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
			},
			"items": {
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

type ElasticProductData struct {
	Code    string `json:"code"`
	Title   string `json:"title,omitempty"`
	Caption string `json:"caption,omitempty"`
	Sale    bool   `json:"sale,omitempty"`
	Shop    struct {
		ID   uint64 `json:"id"`
		Name string `json:"name,omitempty"`
	} `json:"shop,omitempty"`
	Mentioner struct {
		ID   uint64 `json:"id"`
		Name string `json:"name,omitempty"`
	} `json:"mentioner,omitempty"`
	// tags from all items
	Tags []struct {
		ID   uint64 `json:"id"`
		Name string `json:"name,omitempty"`
	} `json:"tags,omitempty"`
	Items []struct {
		Name          string `json:"name,omitempty"`
		Price         uint64 `json:"price"`
		DiscountPrice uint64 `json:"discount_price"`
	} `json:"items,omitempty"`
	Images []struct {
		URL  string `json:"url,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"images,omitempty"`
}

// represent relation table in db
type ElasticProductIndex struct {
	ID uint64 `gorm:"primary_key"`
	// current elastic version of document
	// -1 if product was deleted
	Version         int
	SourceUpdatedAt time.Time `gorm:"index"`
}

type ElasticProduct struct {
	ElasticProductData
	ElasticProductIndex
}

package cache

import (
	"proto/core"
	"time"
	"utils/log"
)

//Product puts to cache product
func Product(id int64, result *core.ProductSearchResult) {
	key := getProductTagKey(id)
	log.Error(
		PutV(key, result, time.Minute*30),
	)
	AddTags(getProductTagKey(id), time.Minute*60, key)
}

func getProductTagKey(id int64) string {
	return idKey("product", id)
}

func getProductTags(product *core.Product) (out []string) {
	if product.Supplier.Id > 0 {
		out = append(out, idKey("shop", product.Supplier.Id))
	}

	if product.Supplier.Id > 0 {
		out = append(out, idKey("shop", product.Supplier.Id))
	}

	return
}

//GetProduct gets a product from the cache
func GetProduct(id int64) *core.ProductSearchResult {
	return getCachedProducts(getProductTagKey(id))
}

//FlushProduct removes all cached results related with this product id
func FlushProduct(id int64) {
	key := getProductTagKey(id)
	keys := GetTags(key)
	if len(keys) > 0 {
		Delete(keys...)
	}
	Delete(key)
}

package cache

import (
	"proto/core"
	"time"
	"utils/log"
)

// SaveGetProduct puts to cache product
func SaveGetProduct(response *core.ProductSearchResult) {

	if len(response.Result) != 1 {
		return
	}

	product := response.Result[0]
	id := product.Id

	key := getProductTagKey(id)
	log.Error(
		PutV(key, response, time.Minute*30),
	)

	tags := getProductTags(product)
	log.Debug("Tags for this product: %v", tags)
	AddTags(key, time.Minute*60, tags...)
}

//GetProduct gets a product from the cache
func GetProduct(id int64) *core.ProductSearchResult {
	return getCachedProducts(getProductTagKey(id))
}

func getProductTagKey(id int64) string {
	return idKey("product", id)
}

func getProductTags(product *core.Product) (out []string) {
	if product.Supplier.Id > 0 {
		out = append(out, idKey("shop", product.Supplier.Id))
	}

	if product.Mentioned.Id > 0 {
		out = append(out, idKey("user", product.Mentioned.Id))
	}

	return
}

//FlushProduct removes all cached results related with this product id
func FlushProduct(id int64) {
	log.Debug("Flushing product %v", id)
	flush(getProductTagKey(id))
}

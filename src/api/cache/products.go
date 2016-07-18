package cache

import (
	"common"
	"crypto/md5"
	"fmt"
	"proto/core"
	"utils/log"
	"strconv"
	"strings"
	"time"
)

//GetSearchProductCacheKey returns a cache key for the search request
func GetSearchProductCacheKey(r *core.SearchProductRequest) string {
	common.Ints64(r.Tags)
	p := []string{
		r.Keyword,
		fmt.Sprint(r.Tags),
		strconv.FormatInt(r.Limit, 10),
		strconv.FormatInt(r.GetOffset(), 10),
		strconv.FormatUint(r.GetFromId(), 10),
		fmt.Sprint(r.IsSaleOnly),
		fmt.Sprint(r.OffsetDirection),
		strconv.FormatUint(r.GetUserId(), 10),
		strconv.FormatUint(r.GetShopId(), 10),
		r.GetInstagramName(),
	}

	return fmt.Sprintf("products.search.result.%x", md5.Sum([]byte(strings.Join(p, "_"))))
}

//GetProductCacheKey returns a cache key for the product id
func GetProductCacheKey(id int64) string {
	return fmt.Sprintf("products.id.%v", id)
}

//CacheSearchResults puts to cache search results
func CacheSearchResults(req *core.SearchProductRequest, results *core.ProductSearchResult) {
	ttl := time.Minute
	if req.GetFromId() > 0 {
		ttl = ttl * 30
	}
	key := GetSearchProductCacheKey(req)
	log.Error(
		Put(key, results, ttl),
	)

	for _, p := range results.Result {
		AddTags(fmt.Sprintf("product.%v", p.Id), time.Minute*60, key)
	}

	if req.GetInstagramName() != "" {
		AddTags(getUsernameTagKey(req.GetInstagramName()), time.Minute*60, key)
	}
}

//CacheProduct puts to cache product
func CacheProduct(id int64, result *core.ProductSearchResult) {
	key := GetProductCacheKey(id)
	log.Error(
		Put(key, result, time.Minute*30),
	)
	AddTags(getProductTagKey(id), time.Minute*60, key)
}

func getProductTagKey(id int64) string {
	return fmt.Sprintf("product.%v", id)
}

func getUsernameTagKey(username string) string {
	return fmt.Sprintf("instagramname.%v", username)
}

func getCachedProducts(key string) *core.ProductSearchResult {
	results := &core.ProductSearchResult{}
	log.Error(GetV(key, results))
	if len(results.Result) == 0 {
		return nil
	}
	return results
}

//GetCachedSearch gets search results from the cache
func GetCachedSearch(req *core.SearchProductRequest) *core.ProductSearchResult {
	return getCachedProducts(GetSearchProductCacheKey(req))
}

//GetCachedProduct gets a product from the cache
func GetCachedProduct(id int64) *core.ProductSearchResult {
	return getCachedProducts(GetProductCacheKey(id))
}

//FlushProductCache removes all cached results related with this product id
func FlushProductCache(id int64) {
	key := getProductTagKey(id)
	keys := GetTags(key)
	if len(keys) > 0 {
		Delete(keys...)
	}
	Delete(key)
}

func FlushUserCache(instagramname string) {
	key := getUsernameTagKey(instagramname)
	keys := GetTags(key)
	if len(keys) > 0 {
		Delete(keys...)
	}
	Delete(key)
}

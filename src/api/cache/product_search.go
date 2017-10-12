package cache

import (
	"common/log"
	"crypto/md5"
	"fmt"
	"proto/core"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Int64Slice) Sort() { sort.Sort(p) }

// Ints sorts a slice of ints in increasing order.
func Ints64(a []int64) { sort.Sort(Int64Slice(a)) }

//GetSearchProductKey returns a cache key for the search request
func GetSearchProductKey(r *core.SearchProductRequest) string {
	Ints64(r.Tags)
	p := []string{
		r.Keyword,
		fmt.Sprint(r.Tags),
		strconv.FormatInt(r.Limit, 10),
		strconv.FormatInt(r.GetOffset(), 10),
		strconv.FormatUint(r.GetFromId(), 10),
		fmt.Sprint(r.IsSaleOnly),
		fmt.Sprint(r.OffsetDirection),
		strconv.FormatUint(r.UserId, 10),
		strconv.FormatUint(r.ShopId, 10),
		r.InstagramName,
	}

	// @TODO: why not just encode request and get md5 of that?
	// that way seems to be faster; but no time to check if I break smth with that

	return fmt.Sprintf("products.search.result.%x", md5.Sum([]byte(strings.Join(p, "_"))))
}

//SearchResults puts to cache search results
func SearchResults(req *core.SearchProductRequest, results *core.ProductSearchResult) {
	ttl := time.Minute
	if req.GetFromId() > 0 {
		ttl = ttl * 30
	}
	key := GetSearchProductKey(req)
	log.Error(
		PutV(key, results, ttl),
	)

	for _, p := range results.Result {
		AddTags(key, time.Minute*60, getProductTags(p)...)
	}
}

func getCachedProducts(key string) *core.ProductSearchResult {
	results := &core.ProductSearchResult{}
	log.Error(GetV(key, results))
	if len(results.Result) == 0 {
		return nil
	}
	return results
}

//GetSearch gets search results from the cache
func GetSearch(req *core.SearchProductRequest) *core.ProductSearchResult {
	return getCachedProducts(GetSearchProductKey(req))
}

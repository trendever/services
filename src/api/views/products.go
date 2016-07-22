package views

import (
	"api/api"
	"api/soso"
	"net/http"

	"api/cache"
	"errors"
	"proto/core"
	"utils/log"
	"utils/rpc"
)

const SearchDefaultLimit = 9

var productServiceClient = core.NewProductServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"retrieve", "product", RetrieveProduct},
		soso.Route{"search", "product", SearchProduct},
		soso.Route{"like", "product", LikeProduct},
	)
}

// Parameters:
//
//   * id (optional uint) returns one product with this id
func RetrieveProduct(c *soso.Context) {
	req := c.RequestMap

	// Select one product parameters
	id, ok := req["id"].(float64)

	if !ok {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("id is required"))
		return
	}

	var (
		response *core.ProductSearchResult
		err      error
	)

	if response = cache.GetCachedProduct(int64(id)); response == nil {
		// Context is responsible for timeouts
		ctx, cancel := rpc.DefaultContext()
		defer cancel()

		response, err = productServiceClient.GetProduct(ctx, &core.GetProductRequest{
			SearchBy: &core.GetProductRequest_Id{int64(id)},
		})

		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}

		cache.CacheProduct(int64(id), response)
	}

	resp := map[string]interface{}{
		"object_list": response.Result,
		"count":       len(response.Result),
	}

	c.SuccessResponse(resp)
}

// Parameters:
//
//   * limit (optional uint; default is 10) limits number of entries
//   * offset (optional uint; default is 0) returns them with an offset
//   * from_id (optional uint; default is 0) returns products from this id
//   * direction (optional bool; default is false) true - asc, false - desc
//   * q (optional string) full-text search query
//   * tags (optional uint array) list of tags ID to find associated
func SearchProduct(c *soso.Context) {
	req := c.RequestMap

	request := &core.SearchProductRequest{
		IsSaleOnly: true,
	}

	// Search parameters
	limit, ok := req["limit"].(float64)
	if !ok || limit < 0 || limit > 30 {
		limit = float64(SearchDefaultLimit)
	}
	request.Limit = int64(limit)

	offset, ok := req["offset"].(float64)
	if !ok || offset < 0 {
		offset = 0
	}

	from_id, ok := req["from_id"].(float64)

	switch {
	case offset > 0:
		request.OffsetBy = &core.SearchProductRequest_Offset{int64(offset)}
	case from_id > 0:
		request.OffsetBy = &core.SearchProductRequest_FromId{uint64(from_id)}
	}

	if value, ok := req["user_id"].(float64); ok {
		request.FeedBy = &core.SearchProductRequest_UserId{UserId: uint64(value)}
	}

	if value, ok := req["instagram_name"].(string); ok {
		request.FeedBy = &core.SearchProductRequest_InstagramName{InstagramName: value}
	}

	if value, ok := req["shop_id"].(float64); ok {
		request.FeedBy = &core.SearchProductRequest_ShopId{ShopId: uint64(value)}
	}

	if request.FeedBy != nil {
		request.IsSaleOnly = false
	}

	// We don't want panic if q is not a string
	// so we provide empty _ variable
	search, _ := req["q"].(string)

	request.Keyword = search

	var tags []int64
	if tags_in, ok := req["tags"].([]interface{}); ok {
		tags = getIntArr(tags_in)
	}

	request.Tags = tags

	if value, ok := req["direction"].(bool); ok {
		request.OffsetDirection = value
	}

	var (
		response *core.ProductSearchResult
		err      error
	)

	if response = cache.GetCachedSearch(request); response == nil {
		// Context is responsible for timeouts
		ctx, cancel := rpc.DefaultContext()
		defer cancel()

		response, err = productServiceClient.SearchProducts(ctx, request)

		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}

		if len(response.Result) > 0 {
			cache.CacheSearchResults(request, response)
		}
	}

	resp := map[string]interface{}{
		"object_list": response.Result,
		"count":       len(response.Result),
	}

	c.SuccessResponse(resp)
}

func LikeProduct(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	request := &core.LikeProductRequest{
		UserId: c.Token.UID,
	}

	if value, ok := req["product_id"].(float64); ok {
		request.ProductId = uint64(value)
	}

	if value, ok := req["like"].(bool); ok {
		request.Like = value
	}

	log.Debug("%v", req)

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := productServiceClient.LikeProduct(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "ok",
	})
}

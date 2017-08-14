package views

import (
	"api/api"
	"api/cache"
	"api/conf"
	"api/soso"
	"encoding/json"
	"errors"
	"gopkg.in/olivere/elastic.v3"
	"net/http"
	"proto/core"
	ewrapper "utils/elastic"
	"utils/log"
	"utils/mandible"
	"utils/rpc"
)

const SearchDefaultLimit = 9

var productServiceClient = core.NewProductServiceClient(api.CoreConn)
var productImagesUploader = mandible.New(conf.GetSettings().MandibleURL, []mandible.Thumbnail{
	{
		Name:   "XL",
		Width:  1080,
		Height: 1080,
		Shape:  "thumb",
	},
	{
		Name:   "L",
		Width:  750,
		Height: 7500,
		Shape:  "thumb",
	},
	{
		Name:   "M_square",
		Width:  480,
		Height: 480,
		Shape:  "square",
	},
	{
		Name:   "S_square",
		Width:  306,
		Height: 306,
		Shape:  "square",
	},
}...)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"product", "retrieve", RetrieveProduct},
		soso.Route{"product", "search", SearchProduct},
		soso.Route{"product", "like", LikeProduct},
		soso.Route{"product", "get_specials", GetSpecialProducts},
		soso.Route{"product", "elastic_search", ElasticSearch},
		soso.Route{"product", "get_liked_by", GetLikedBy},
		soso.Route{"product", "lastid", GetLastProductID},
		soso.Route{"product", "delete", DelProduct},
		soso.Route{"product", "edit", EditProduct},
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

	if response = cache.GetProduct(int64(id)); response == nil {
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

		cache.SaveGetProduct(response)
	}

	resp := map[string]interface{}{
		"object_list": response.Result,
		"count":       len(response.Result),
	}

	c.SuccessResponse(resp)
}

// SearchProduct Parameters:
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
		request.UserId = uint64(value)
	}

	if value, ok := req["instagram_name"].(string); ok {
		request.InstagramName = value
	}

	if value, ok := req["shop_id"].(float64); ok {
		request.ShopId = uint64(value)
	}

	if request.UserId > 0 || request.InstagramName != "" || request.ShopId > 0 {
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

	if response = cache.GetSearch(request); response == nil {
		// Context is responsible for timeouts
		ctx, cancel := rpc.DefaultContext()
		defer cancel()

		response, err = productServiceClient.SearchProducts(ctx, request)

		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}

		if len(response.Result) > 0 {
			cache.SearchResults(request, response)
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

func GetSpecialProducts(c *soso.Context) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := productServiceClient.GetSpecialProducts(ctx, &core.GetSpecialProductsRequest{})
	if err == nil && res.Err != "" {
		err = errors.New(res.Err)
	}
	if err != nil {
		log.Errorf("Failed to get special products list: %v", err)
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"special_products": res.List,
	})
}

//  Parameters:
//
//   * limit (optional uint; default is 9) limits number of entries
//   * offset (optional uint; default is 0) returns them with an offset
//   * q (optional string) full-text search query
//   * tags (optional uint array) list of tags ID to find associated
//
//   any of options below will disable isSale filter
//   * shop_id (optional uint) if presented search only in specified shop
//   * mentioner_id (optional uint) if presented search only products with this mentioner or liked by him
//     Products where mentioner_id is supplier will be filtered
func ElasticSearch(c *soso.Context) {
	req := c.RequestMap
	// Search parameters
	limit_f, _ := req["limit"].(float64)
	limit := int(limit_f)
	switch {
	case limit <= 0:
		limit = SearchDefaultLimit
	case limit_f > 30:
		limit = 30
	}

	offset_f, _ := req["offset"].(float64)
	offset := int(offset_f)
	switch {
	case offset < 0:
		offset = 0
	// default index.max_result_window = 10000. Nobody will scroll so deep i guess
	case offset+limit > 10000:
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("too large offset"))
		return
	}

	query := elastic.NewBoolQuery()

	if value, _ := req["mentioner_id"].(float64); value > 0 {
		liked_ids, err := getLikedBy(uint64(value))
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}
		if len(liked_ids) != 0 {
			query.Filter(
				elastic.NewBoolQuery().Should(
					elastic.NewTermQuery("mentioner.id", uint64(value)),
					elastic.NewTermsQuery("_id", liked_ids),
				),
			)
		} else {
			query.Filter(elastic.NewTermQuery("mentioner.id", uint64(value)))
		}
		query.MustNot(elastic.NewTermQuery("shop.supplier", uint64(value)))
	}

	if value, _ := req["shop_id"].(float64); value > 0 {
		query.Filter(elastic.NewTermQuery("shop.id", uint64(value)))
	} else if value, _ := req["include_not_on_sale"].(bool); !value {
		query.Filter(elastic.NewTermQuery("sale", true))
	}

	var tags []int64
	if tags_in, ok := req["tags"].([]interface{}); ok {
		tags = getIntArr(tags_in)
	}
	if len(tags) != 0 {
		tagsBool := elastic.NewBoolQuery()
		for _, tag := range tags {
			tagsBool.Must(elastic.NewTermQuery("items.tags.id", uint64(tag)))
		}
		query.Filter(elastic.NewNestedQuery("items", tagsBool))
	}

	eCli := ewrapper.Cli()
	search := eCli.Search().Index("products").Type("product").From(offset).Size(limit)

	text, _ := req["query"].(string)
	if text != "" {
		query.Must(elastic.NewMatchQuery("_all", text).Analyzer("search"))
	} else {
		search = search.Sort("id", false)
	}

	res, err := search.Query(query).Do()
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	type product struct {
		ID   string           `json:"id"`
		Data *json.RawMessage `json:"data"`
	}
	hits := []product{}
	for _, hit := range res.Hits.Hits {
		hits = append(hits, product{
			ID:   hit.Id,
			Data: hit.Source,
		})
	}
	c.SuccessResponse(hits)
}

// returns slice of product ids that are liked by user
func getLikedBy(user_id uint64) ([]uint64, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := productServiceClient.GetLikedBy(ctx, &core.GetLikedByRequest{user_id})
	return res.ProductIds, err
}

func GetLikedBy(c *soso.Context) {
	userID, ok := c.RequestMap["user_id"].(float64)
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("user_id undefined"))
		return
	}
	ids, err := getLikedBy(uint64(userID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"products": ids,
	})
}

func GetLastProductID(c *soso.Context) {
	shopID, ok := c.RequestMap["shop_id"].(float64)
	if !ok || shopID <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("shop_id is null"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := productServiceClient.GetLastProductID(ctx, &core.GetLastProductIDRequest{
		ShopId: uint64(shopID),
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"id": res.Id,
	})
}

func DelProduct(c *soso.Context) {

	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	productID, _ := req["product_id"].(float64)
	if productID <= 0 {
		c.ErrorResponse(403, soso.LevelError, errors.New("Incorrect product ID"))
		return
	}

	{
		// get related product
		ctx, cancel := rpc.DefaultContext()
		defer cancel()

		res, err := productServiceClient.GetProduct(ctx, &core.GetProductRequest{
			SearchBy: &core.GetProductRequest_Id{int64(productID)},
		})

		if err != nil {
			c.ErrorResponse(404, soso.LevelError, err)
			return
		}

		if len(res.Result) != 1 {
			c.ErrorResponse(404, soso.LevelError, errors.New("Product not found"))
			return

		}

		if res.Result[0].Supplier.SupplierId <= 0 || uint64(res.Result[0].Supplier.SupplierId) != c.Token.UID {
			c.ErrorResponse(403, soso.LevelError, errors.New("Only shop supplier allowed to do that"))
			return
		}
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	response, err := productServiceClient.DelProduct(ctx, &core.DelProductRequest{
		ProductId: uint64(productID),
	})

	if err != nil {
		c.ErrorResponse(503, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success": response.Success,
	})
}

func EditProduct(c *soso.Context, arg *struct {
	ID           uint64              `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	InstagramURL string              `json:"instagram_url"`
	ChatMessage  string              `json:"chat_message"`
	IsSale       bool                `json:"is_sale"`
	ImageURL     string              `json:"image_url"`
	Items        []*core.ProductItem `json:"items"`
	//Items        struct {
	//	ID            uint64   `json:"id"`
	//	Name          string   `json:"name"`
	//	Price         uint64   `json:"price"`
	//	DiscountPrice uint64   `json:"discount_price"`
	//	Tags          []uint64 `json:"tags"`
	//} `json:"items"`
}) {
	if c.Token == nil {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := core.EditProductRequest{
		EditorId: c.Token.UID,
		Product: &core.Product{
			Id:    int64(arg.ID),
			Title: arg.Title,
			InstagramImageCaption: arg.Description,
			ChatMessage:           arg.ChatMessage,
			IsSale:                arg.IsSale,
			Items:                 arg.Items,
		},
	}
	// @CHECK i think it's better to upload it here
	// but that may lead to unlinked images if product will not be saved for whatever reason
	// or useless work if link is actuality same as before
	if arg.ImageURL != "" {
		resp, err := productImagesUploader.DoRequest("url", arg.ImageURL)
		if err != nil {
			log.Warn("failed to upload image product image '%v': %v", arg.ImageURL, err)
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("failed to upload image"))
			return
		}
		imgs := []*core.ImageCandidate{
			{
				Name: "Max",
				Url:  resp.Link,
			},
		}
		for name, url := range resp.Thumbs {
			imgs = append(imgs, &core.ImageCandidate{
				Name: name,
				Url:  url,
			})
		}
		req.Product.InstagramImageUrl = arg.ImageURL
		req.Product.InstagramImages = imgs
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := productServiceClient.EditProduct(ctx, &req)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if resp.Forbidden {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("permission denied"))
	}
	if resp.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(resp.Error))
		return
	}

	c.SuccessResponse(resp)
}

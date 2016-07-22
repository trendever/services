package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"

	"proto/core"
	"utils/rpc"
)

var tagServiceClient = core.NewTagServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"retrieve", "tag", RetrieveTag},
	)
}

// Parameters:
//   * is_main (optional bool; default false) return tags with is_main field set
//     * limit (optional int; default 15) how much to return
//   * tags (optional int array) return related to this search tags
func RetrieveTag(c *soso.Context) {

	req := c.RequestMap

	is_main, _ := req["is_main"].(bool)

	limit, ok := req["limit"].(float64)
	if !ok || limit < 0 || limit > 60 {
		limit = 15
	}

	var search_tags []int64
	// Convert []interface{} input array to good []int64 slice
	if tags_in, ok := req["tags"].([]interface{}); ok {
		search_tags = getIntArr(tags_in)
	}

	var (
		err    error
		result *core.TagSearchResult
	)

	// timeouts
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	switch {
	case is_main: // retrieve main tags using RPC
		result, err = tagServiceClient.GetMainTags(ctx, &core.GetMainTagsRequest{
			Limit: int64(limit),
		})
	case len(search_tags) > 0: // retrieve related tags
		result, err = tagServiceClient.GetRelatedTags(ctx, &core.GetRelatedTagsRequest{
			Tags:  search_tags,
			Limit: int64(limit),
		})
	default: //don't know what to do, throw an error
		err = errors.New("Unknown request mode")
	}

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"object_list": result.Result,
		"count":       len(result.Result),
	})
}

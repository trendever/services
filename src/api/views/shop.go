package views

import (
	"errors"
	"proto/core"
	"utils/rpc"
	"net/http"
	"api/api"
	"api/soso"
)

var shopServiceClient = core.NewShopServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "shop", GetShopProfile},
	)
}

func GetShopProfile(c *soso.Context) {
	req := c.RequestMap
	request := &core.ShopProfileRequest{}

	if value, ok := req["instagram_name"].(string); ok {
		request.SearchBy = &core.ShopProfileRequest_InstagramName{InstagramName: value}
	}

	if value, ok := req["shop_id"].(float64); ok {
		request.SearchBy = &core.ShopProfileRequest_Id{Id: uint64(value)}
	}

	if request.SearchBy == nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("instagram_name or shop_id are required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := shopServiceClient.GetShopProfile(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"shop": resp.Shop,
	})
}

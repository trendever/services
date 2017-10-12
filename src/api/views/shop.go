package views

import (
	"api/api"
	"common/soso"
	"errors"
	"net/http"
	"proto/core"
	"utils/rpc"
)

var shopServiceClient = core.NewShopServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"shop", "retrieve", GetShopProfile},
		soso.Route{"shop", "create", createShop},
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

	c.SuccessResponse(resp)
}

func createShop(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("User not authorized"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := shopServiceClient.FindOrCreateShopForSupplier(ctx, &core.FindOrCreateShopForSupplierRequest{
		SupplierId:      c.Token.UID,
		RecreateDeleted: true,
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if resp.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(resp.Error))
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"shop_id": resp.ShopId,
	})
}

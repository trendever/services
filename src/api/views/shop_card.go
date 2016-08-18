package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/core"
	"utils/rpc"
)

var shopCardServiceClient = core.NewShopCardServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "card", GetShopCards},
		soso.Route{"create", "card", CreateCard},
		soso.Route{"delete", "card", DeleteCard},
	)
}

// GetShopCards returns cards for given shop
func GetShopCards(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	// Check ShopId correctness
	shopID, _ := req["shop_id"].(float64)

	// Launch RPC req
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := shopCardServiceClient.GetCards(ctx, &core.GetCardsRequest{
		UserId: c.Token.UID,
		ShopId: uint64(shopID),
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"cards": res.Cards,
	})
}

// DeleteCard deletes existent card by id
func DeleteCard(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	cardID, ok := req["card_id"].(float64)

	if !ok || cardID <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Incorrect card id"))
		return
	}

	// Launch RPC req
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := shopCardServiceClient.DeleteCard(ctx, &core.DeleteCardRequest{
		Id:     uint64(cardID),
		UserId: c.Token.UID,
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success": true,
	})
}

// CreateCard by shop id, name and number
func CreateCard(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	// int params
	shopID, _ := req["shop_id"].(float64)

	// string params
	cardName, okCardName := req["card_name"].(string)
	cardNumber, okCardNumber := req["card_number"].(string)

	if !okCardName || !okCardNumber || cardName == "" || cardNumber == "" {

		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Bad parameters"))
		return
	}

	// Launch RPC req
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := shopCardServiceClient.CreateCard(ctx, &core.CreateCardRequest{
		Card: &core.ShopCard{
			Name:   cardName,
			Number: cardNumber,
			ShopId: uint64(shopID),
			UserId: c.Token.UID,
		},
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success": true,
		"id":      res.Id,
	})
}

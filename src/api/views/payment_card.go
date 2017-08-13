package views

import (
	"api/soso"
	"errors"
	"net/http"
	"proto/payment"
	"utils/rpc"
)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"paymentcard", "retrieve", GetPaymentCards},
		soso.Route{"paymentcard", "create", CreatePaymentCard},
		soso.Route{"paymentcard", "delete", DeletePaymentCard},
	)
}

const useGw = "payture_ewallet"

// GetPaymentCards returns cards for given shop
func GetPaymentCards(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	// Launch RPC req
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := paymentServiceClient.GetCards(ctx, &payment.GetCardsRequest{
		Gateway: useGw,
		User: &payment.UserInfo{
			UserId: c.Token.UID,
			Ip:     c.RemoteIP,
		},
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error > 0 {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.ErrorMessage))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"cards": res.Cards,
	})
}

// DeletePaymentCard deletes existent card by id
func DeletePaymentCard(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	cardID, ok := req["card_id"].(string)

	if !ok || cardID <= "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Incorrect card id"))
		return
	}

	// Launch RPC req
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := paymentServiceClient.DelCard(ctx, &payment.DelCardRequest{
		Gateway: useGw,
		CardId:  cardID,
		User: &payment.UserInfo{
			UserId: c.Token.UID,
			Ip:     c.RemoteIP,
		},
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error > 0 {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.ErrorMessage))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success": true,
	})
}

// CreatePaymentCard by shop id, name and number
func CreatePaymentCard(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	// Launch RPC req
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := paymentServiceClient.AddCard(ctx, &payment.AddCardRequest{
		Gateway: useGw,
		User: &payment.UserInfo{
			UserId: c.Token.UID,
			Ip:     c.RemoteIP,
		},
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error > 0 {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.ErrorMessage))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success":      true,
		"redirect_url": res.RedirectUrl,
	})
}

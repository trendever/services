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
		soso.Route{"retrieve", "paymentcard", GetPaymentCards},
		soso.Route{"create", "paymentcard", CreatePaymentCard},
		soso.Route{"delete", "paymentcard", DeletePaymentCard},
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

	_, err := paymentServiceClient.DelCard(ctx, &payment.DelCardRequest{
		CardId: cardID,
		User: &payment.UserInfo{
			UserId: c.Token.UID,
			Ip:     c.RemoteIP,
		},
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
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
		User: &payment.UserInfo{
			UserId: c.Token.UID,
			Ip:     c.RemoteIP,
		},
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success":      true,
		"redirect_url": res.RedirectUrl,
	})
}

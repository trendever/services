package views

import (
	"errors"
	"net/http"

	"api/api"
	"api/soso"
	"proto/core"
	"proto/payment"
	"utils/rpc"
)

var paymentServiceClient = payment.NewPaymentServiceClient(api.PaymentsConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"create", "order", CreateOrder},
		soso.Route{"create", "payment", CreateOrder},
	)
}

var _ = rpc.DefaultContext

// GetShopCards returns cards for given shop
func CreateOrder(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	amount, _ := req["amount"].(float64)
	leadID, _ := req["lead"].(float64)

	currency, _ := req["currency"].(float64)
	_, currencyOK := payment.Currency_name[int32(currency)]

	// validated in payments service
	shopCardNumber, _ := req["card"].(string)

	if amount <= 0 || leadID <= 0 || !currencyOK {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("Incorrect parameter"))
		return
	}

	conversationID, err := getAdminnedConversationID(c.Token.UID, uint64(leadID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
	}

	// now -- create the order
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := paymentServiceClient.CreateOrder(ctx, &payment.CreateOrderRequest{
		Amount:   uint64(amount),
		Currency: payment.Currency(currency),

		LeadId:         uint64(leadID),
		UserId:         c.Token.UID,
		ConversationId: conversationID,

		ShopCardNumber: shopCardNumber,
	})

	if err != nil {
		if resp.Error > 0 { // ignore RPC errors
			c.Response.ResponseMap = map[string]interface{}{
				"ErrorCode":    resp.Error,
				"ErrorMessage": err,
			}
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{})
}

func CreatePayment(c *soso.Context) {

}

func getAdminnedConversationID(userID, leadID uint64) (uint64, error) {

	info, err := getLeadInfo(userID, leadID)
	if err != nil {
		return 0, err
	}

	// must be owner in his chat
	switch info.UserRole {
	case core.LeadUserRole_SELLER, core.LeadUserRole_SUPER_SELLER, core.LeadUserRole_SUPPLIER:
		// ok
	default:

		return 0, errors.New("Access denied; should be admin of this chat")
	}

	return info.ConversationId, nil
}

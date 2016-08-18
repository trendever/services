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
		soso.Route{"create", "payment", CreatePayment},
	)
}

var _ = rpc.DefaultContext

// CreateOrder for given summ, card number and leadID
func CreateOrder(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	amount, _ := req["amount"].(float64)
	leadID, _ := req["lead_id"].(float64)

	currency, _ := req["currency"].(float64)
	_, currencyOK := payment.Currency_name[int32(currency)]

	// validated in payments service
	shopCardNumber, _ := req["card"].(string)

	if amount <= 0 || leadID <= 0 || !currencyOK || shopCardNumber == "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("Incorrect parameter"))
		return
	}

	conversationID, role, err := getConversationID(c.Token.UID, uint64(leadID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	direction, err := paymentDirection(role, true)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	// now -- create the order
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := paymentServiceClient.CreateOrder(ctx, &payment.CreateOrderRequest{
		Amount:   uint64(amount),
		Currency: payment.Currency(currency),

		LeadId:         uint64(leadID),
		Direction:      direction,
		UserId:         c.Token.UID,
		ConversationId: conversationID,
		ShopCardNumber: shopCardNumber,
	})

	if err != nil {
		if resp != nil && resp.Error > 0 { // ignore RPC errors
			c.Response.ResponseMap = map[string]interface{}{
				"ErrorCode":    resp.Error,
				"ErrorMessage": err,
			}
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"id": resp.Id,
	})
}

// CreatePayment for given order
func CreatePayment(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	// We have to transfer lead ID to make sure only associated user can create payment
	// Malicious request can be sent and the following check will succeed
	// However, payments service must check if reqeusted leadID is equal to the CreateOrder one
	// This scheme help both avoid calling core from payments and guarantee security

	// In other words
	//  api checks if user is customer in lead(LeadID)
	//  payments checks if leadID is connected to pay(payID)
	//  -> user is checked to have access to pay

	payID, _ := req["id"].(float64)
	leadID, _ := req["lead_id"].(float64)

	if leadID <= 0 || payID <= 0 {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("Incorrect parameter"))
		return
	}

	_, role, err := getConversationID(c.Token.UID, uint64(leadID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	// must be owner in his chat
	direction, err := paymentDirection(role, true)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	ip, err := getIP(c.Session)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	// now -- create the order
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := paymentServiceClient.BuyOrder(ctx, &payment.BuyOrderRequest{
		PayId:     uint64(payID),
		LeadId:    uint64(leadID),
		Direction: direction,
		Ip:        ip,
	})

	if err != nil {
		if resp != nil && resp.Error > 0 { // ignore RPC errors
			c.Response.ResponseMap = map[string]interface{}{
				"ErrorCode":    resp.Error,
				"ErrorMessage": err,
			}
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"redirect_url": resp.RedirectUrl,
	})

}

func getConversationID(userID, leadID uint64) (uint64, core.LeadUserRole, error) {

	info, err := getLeadInfo(userID, leadID)
	if err != nil {
		return 0, core.LeadUserRole_UNKNOWN, err
	}

	return info.ConversationId, info.UserRole, nil
}

func paymentDirection(role core.LeadUserRole, create bool) (payment.Direction, error) {

	var isModerator = role == core.LeadUserRole_SELLER || role == core.LeadUserRole_SUPER_SELLER || role == core.LeadUserRole_SUPPLIER
	var isClient = role == core.LeadUserRole_CUSTOMER

	switch {
	case create && isModerator, !create && isClient:
		return payment.Direction_CLIENT_PAYS, nil
	case create && isClient, !create && isModerator:
		return payment.Direction_CLIENT_RECV, nil
	}

	return payment.Direction(0), errors.New("payments.view: Bad user role in the chat")
}

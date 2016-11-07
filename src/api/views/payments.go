package views

import (
	"encoding/json"
	"errors"
	"fmt"
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
		soso.Route{"cancel", "order", CancelOrder},
		soso.Route{"create", "payment", CreatePayment},
	)
}

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
	currencyName, currencyOK := payment.Currency_name[int32(currency)]

	// retrieve card number from payments service
	shopCardID, _ := req["card"].(float64)
	shopCardNumber, err := getCardNumber(c.Token.UID, uint64(shopCardID))

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	if amount <= 0 || leadID <= 0 || !currencyOK || shopCardNumber == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, fmt.Errorf("Incorrect parameter"))
		return
	}

	leadInfo, err := getLeadInfo(c.Token.UID, uint64(leadID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	if leadInfo.Shop.Suspended {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("shop is suspended"))
		return
	}

	direction, err := paymentDirection(leadInfo.UserRole, true)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	data, err := json.Marshal(&payment.UsualData{
		Direction:      direction,
		ConversationId: leadInfo.ConversationId,
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	request := &payment.CreateOrderRequest{
		Data: &payment.OrderData{
			Amount:   uint64(amount),
			Currency: payment.Currency(currency),

			LeadId:         uint64(leadID),
			UserId:         c.Token.UID,
			ShopCardNumber: shopCardNumber,

			Gateway:     "payture",
			ServiceName: "api",
			ServiceData: string(data),
		},
	}

	if direction == payment.Direction_CLIENT_PAYS {
		plan, err := getMonetizationPlan(leadInfo.Shop.PlanId)
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}
		if plan.TransactionCommission != 0 && plan.CoinsExchangeRate != 0 {
			if plan.PrimaryCurrency != currencyName {
				c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Unexpected currency"))
			}
			request.Data.CommissionSource = uint64(leadInfo.Shop.SupplierId)
			fee := uint64(amount*plan.TransactionCommission*plan.CoinsExchangeRate + 0.5)
			if fee == 0 {
				fee = 1
			}
			request.Data.CommissionFee = fee
		}
	}

	// now -- create the order
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := paymentServiceClient.CreateOrder(ctx, request)

	if err != nil { // RPC errors
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if resp.Error > 0 { // service errors
		c.Response.ResponseMap = map[string]interface{}{
			"ErrorCode":    resp.Error,
			"ErrorMessage": resp.ErrorMessage,
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(resp.ErrorMessage))
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

	payID, _ := req["id"].(float64)
	leadID, _ := req["lead_id"].(float64)

	if leadID < 0 || payID <= 0 {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("Incorrect parameter"))
		return
	}

	if leadID != 0 {
		_, role, err := getConversationID(c.Token.UID, uint64(leadID))
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}

		// must be owner in his chat
		direction, err := paymentDirection(role, false)
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}

		orderData, paymentData, err := retrieveOrder(uint64(payID))
		if err != nil {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
			return
		}

		if !canBuy(paymentData.Direction, direction) {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, fmt.Errorf("This side of order can not pay it"))
			return
		}
		if orderData.LeadId != uint64(leadID) {
			c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, fmt.Errorf("Parameters mangled"))
			return
		}

	}

	// now -- create the order
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := paymentServiceClient.BuyOrder(ctx, &payment.BuyOrderRequest{
		PayId: uint64(payID),
		Ip:    c.RemoteIP,
	})

	if err != nil { // RPC errors
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if resp.Error > 0 { // service errors
		c.Response.ResponseMap = map[string]interface{}{
			"ErrorCode":    resp.Error,
			"ErrorMessage": resp.ErrorMessage,
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(resp.ErrorMessage))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"redirect_url": resp.RedirectUrl,
	})

}

// CancelOrder cancels given order
func CancelOrder(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

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

	// must have correct direction; IS creator
	direction, err := paymentDirection(role, true)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	orderData, paymentData, err := retrieveOrder(uint64(payID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	if !canCancelPay(paymentData.Direction, direction) {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("No access to cancel this pay"))
		return
	}
	if orderData.LeadId != uint64(leadID) {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("Parameters mangled"))
		return
	}

	// now -- create the order
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := paymentServiceClient.CancelOrder(ctx, &payment.CancelOrderRequest{
		PayId:  uint64(payID),
		UserId: c.Token.UID,
	})

	if err != nil { // RPC errors
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if resp.Error > 0 { // service errors
		c.Response.ResponseMap = map[string]interface{}{
			"ErrorCode":    resp.Error,
			"ErrorMessage": resp.ErrorMessage,
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(resp.ErrorMessage))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"cancelled": resp.Cancelled,
	})

}

// get chat ID by leadID
func getConversationID(userID, leadID uint64) (uint64, core.LeadUserRole, error) {

	info, err := getLeadInfo(userID, leadID)
	if err != nil {
		return 0, core.LeadUserRole_UNKNOWN, err
	}

	return info.ConversationId, info.UserRole, nil
}

// get payment direction
// create == if we want to create order, or use it
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

func canCancelPay(payDirection, userDirection payment.Direction) bool {

	// allow both sides cancel
	// we have only 2 direction and this may seem useless; but it is planned to have other people in conversation
	// so I will disable checks only for new payment sides
	switch payDirection {
	case payment.Direction_CLIENT_PAYS, payment.Direction_CLIENT_RECV:
		if userDirection == payment.Direction_CLIENT_PAYS || userDirection == payment.Direction_CLIENT_RECV {
			return true
		}
	}

	return false
}

func canBuy(payDirection, userDirection payment.Direction) bool {

	switch payDirection {
	case payment.Direction_CLIENT_PAYS, payment.Direction_CLIENT_RECV:
		if payDirection != userDirection { //only other side can pay
			if userDirection == payment.Direction_CLIENT_PAYS || userDirection == payment.Direction_CLIENT_RECV {
				return true
			}
		}
	}

	return false
}

// get orderData and decode serviceData to our struct
func retrieveOrder(id uint64) (*payment.OrderData, *payment.UsualData, error) {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := paymentServiceClient.GetOrder(ctx, &payment.GetOrderRequest{
		Id: id,
	})

	if err != nil {
		return nil, nil, err
	}

	var decoded payment.UsualData
	err = json.Unmarshal([]byte(resp.Order.ServiceData), &decoded)
	if err != nil {
		return nil, nil, err
	}

	return resp.Order, &decoded, nil
}

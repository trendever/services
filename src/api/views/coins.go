package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/trendcoin"
	"utils/rpc"
)

var coinsServiceClient = trendcoin.NewTrendcoinServiceClient(api.CoinsConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"coins", "balance", CoinsBalance},
		soso.Route{"coins", "log", CoinsLog},
	)
}

// balance of current user
func CoinsBalance(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	balance, err := coinsBalance(c.Token.UID)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"balance": balance,
	})
}

func coinsBalance(userID uint64) (int64, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := coinsServiceClient.Balance(ctx, &trendcoin.BalanceRequest{UserId: userID})
	if err != nil {
		return 0, err
	}
	if res.Error != "" {
		return 0, errors.New(res.Error)
	}
	return res.Balance, nil
}

func CoinsLog(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap

	limit_f, _ := req["limit"].(float64)
	if limit_f < 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("negative limit"))
		return
	}
	offset_f, _ := req["offset"].(float64)
	if offset_f < 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("negative offset"))
		return
	}
	before_f, _ := req["before"].(float64)
	after_f, _ := req["after"].(float64)
	asc, _ := req["asc"].(bool)

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := coinsServiceClient.TransactionLog(ctx, &trendcoin.TransactionLogRequest{
		UserId: c.Token.UID,
		Limit:  uint64(limit_f),
		Offset: uint64(offset_f),
		Before: int64(before_f),
		After:  int64(after_f),
		Asc:    asc,
	})

	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"transactions": res.Transactions,
		"has_more":     res.HasMore,
	})
}

package views

import (
	"errors"
	auth_protocol "proto/auth"
	"utils/rpc"
	"net/http"
	"api/auth"
	"api/soso"
)

var authClient = auth.Client

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"register", "auth", RegisterNewUser},
		soso.Route{"login", "auth", Login},
		soso.Route{"send_password", "auth", SendSmsPassword},
	)
}

//RegisterNewUser creates new user
func RegisterNewUser(c *soso.Context) {
	req := c.RequestMap
	request := &auth_protocol.NewUserRequest{}

	if value, ok := req["phone"].(string); ok {
		request.PhoneNumber = value
	}

	if value, ok := req["instagram_username"].(string); ok {
		request.InstagramUsername = value
	}

	if value, ok := req["username"].(string); ok {
		request.Username = value
	}

	if request.PhoneNumber == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("User phone number is required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := authClient.RegisterNewUser(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"ErrorCode":    resp.ErrorCode,
		"ErrorMessage": resp.ErrorMessage,
	})
}

//Login requests a token by user phone and password
func Login(c *soso.Context) {
	req := c.RequestMap
	request := &auth_protocol.LoginRequest{}

	if value, ok := req["phone"].(string); ok {
		request.PhoneNumber = value
	}

	if value, ok := req["password"].(string); ok {
		request.Password = value
	}

	if request.Password == "" || request.PhoneNumber == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Phone number and password is required"))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := authClient.Login(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	if resp.ErrorCode != 0 {
		c.Response.ResponseMap = map[string]interface{}{
			"ErrorCode":    resp.ErrorCode,
			"ErrorMessage": resp.ErrorMessage,
		}
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New(resp.ErrorMessage))
		return
	}

	tokenData, err := auth.GetTokenData(resp.Token)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	user, err := GetUser(tokenData.UID)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"token": resp.Token,
		"user":  user,
	})
}

//SendSmsPassword sends new password to user phone
func SendSmsPassword(c *soso.Context) {
	req := c.RequestMap
	request := &auth_protocol.SmsPasswordRequest{}

	if value, ok := req["phone"].(string); ok {
		request.PhoneNumber = value
	}
	if request.PhoneNumber == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("User phone number is required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := authClient.SendNewSmsPassword(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"ErrorCode":    resp.ErrorCode,
		"ErrorMessage": resp.ErrorMessage,
	})
}

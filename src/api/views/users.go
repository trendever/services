package views

import (
	"api/api"
	"api/soso"
	"errors"
	"github.com/asaskevich/govalidator"
	"net/http"
	"proto/auth"
	"proto/core"
	"strings"
	p "utils/phone"
	"utils/rpc"
)

type User struct {
	*core.User
}

var userServiceClient = core.NewUserServiceClient(api.CoreConn)
var authServiceClient = auth.NewAuthServiceClient(api.AuthConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"user", "retrieve", GetUserProfile},
		soso.Route{"user", "set_email", SetEmail},
		soso.Route{"user", "set_data", SetData},
		soso.Route{"user", "list_telegrams", ListTelegrams},
		soso.Route{"user", "confirm_telegram", ConfirmTelegram},
		soso.Route{"user", "del_telegram", DelTelegram},
	)
}

func ListTelegrams(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := userServiceClient.ListTelegrams(ctx, &core.ListTelegramsRequest{
		UserId: c.Token.UID,
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
		"list": resp.Telegrams,
	})
}

func ConfirmTelegram(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	chat_id, _ := c.RequestMap["chat_id"].(float64)
	if chat_id <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid chat_id"))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := userServiceClient.ConfirmTelegram(ctx, &core.ConfirmTelegramRequest{
		UserId: c.Token.UID,
		ChatId: uint64(chat_id),
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
		"status": "success",
	})
}

func DelTelegram(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	chat_id, _ := c.RequestMap["chat_id"].(float64)
	if chat_id <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid chat_id"))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := userServiceClient.DelTelegram(ctx, &core.DelTelegramRequest{
		UserId: c.Token.UID,
		ChatId: uint64(chat_id),
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
		"status": "success",
	})
}

func GetUserProfile(c *soso.Context) {
	req := c.RequestMap
	request := &core.ReadUserRequest{
		Public:   true,
		GetShops: true,
	}
	valid := false
	name_blank := false

	if instaname, ok := req["instagram_name"].(string); ok {
		request.InstagramUsername = instaname
		valid = true
	} else {
		name_blank = true
	}

	if value, ok := req["user_id"].(float64); ok {
		request.Id = uint64(value)
		valid = true
	}

	if request.Id == 0 && c.Token != nil && name_blank {
		request.Id = c.Token.UID
		valid = true
	}

	if !valid {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("instagram_name or user_id are required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := userServiceClient.ReadUser(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}
	if resp.Id == 0 {
		c.ErrorResponse(http.StatusNotFound, soso.LevelError, errors.New("user not found"))
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"profile": resp.User,
	})
}

func GetUser(user_id uint64, getShops bool) (*User, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := userServiceClient.ReadUser(ctx, &core.ReadUserRequest{Id: user_id, GetShops: getShops})

	if err != nil {
		return nil, err
	}

	return &User{User: resp.User}, nil
}

func (u *User) GetName() string {
	switch {
	case u.InstagramUsername != "":
		return u.InstagramUsername
	case u.Name != "":
		return u.Name
	}
	return "User"
}

func SetEmail(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	email, _ := c.RequestMap["email"].(string)
	if !govalidator.IsEmail(email) {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid email"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := userServiceClient.SetEmail(ctx, &core.SetEmailRequest{
		UserId: c.Token.UID,
		Email:  email,
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
		"status": "success",
	})
}

func SetData(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	request := &core.SetDataRequest{}

	request.UserId = c.Token.UID

	if value, ok := c.RequestMap["name"].(string); ok {
		value = strings.Trim(value, " \r\n\t")
		if !nameValidator.MatchString(value) {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Invalid user name"))
			return
		}
		request.Name = value
	}

	if phone, ok := c.RequestMap["phone"].(string); ok {
		phoneNumber, err := p.CheckNumber(phone, "")

		if err != nil {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
			return
		}

		request.Phone = phoneNumber
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := userServiceClient.SetData(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	smsRequest := &auth.SmsPasswordRequest{PhoneNumber: request.Phone}
	_, err = authServiceClient.SendNewSmsPassword(ctx, smsRequest)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}

package views

import (
	"api/api"
	"api/soso"
	"errors"
	"github.com/asaskevich/govalidator"
	"net/http"
	"proto/core"
	"utils/rpc"
	"strings"
)

type User struct {
	*core.User
}

var userServiceClient = core.NewUserServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "user", GetUserProfile},
		soso.Route{"set_email", "user", SetEmail},
		soso.Route{"set_data","user",SetData},
	)
}

func GetUserProfile(c *soso.Context) {
	req := c.RequestMap
	request := &core.ReadUserRequest{
		Public:   true,
		GetShops: true,
	}
	valid := false

	if value, ok := req["instagram_name"].(string); ok {
		request.InstagramUsername = value
		valid = true
	}

	if value, ok := req["user_id"].(float64); ok {
		request.Id = uint64(value)
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
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}

func SetData(c *soso.Context){
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

	if (request.Name == ""){
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Enter user name"))
	}

	if value, ok := c.RequestMap["phone"].(string); ok {
		request.Phone = value
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := userServiceClient.SetData(ctx, request)

	if (err != nil){
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}
package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/core"
	"utils/rpc"
)

type User struct {
	*core.User
}

var userServiceClient = core.NewUserServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "user", GetUserProfile},
	)
}

func GetUserProfile(c *soso.Context) {
	req := c.RequestMap
	request := &core.UserProfileRequest{}

	if value, ok := req["instagram_name"].(string); ok {
		request.SearchBy = &core.UserProfileRequest_InstagramName{InstagramName: value}
	}

	if value, ok := req["user_id"].(float64); ok {
		request.SearchBy = &core.UserProfileRequest_Id{Id: uint64(value)}
	}

	if request.SearchBy == nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("instagram_name or user_id are required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := userServiceClient.GetUserProfile(ctx, request)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"profile": resp.Profile,
	})
}

func GetUser(user_id uint64) (*User, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := userServiceClient.ReadUser(ctx, &core.ReadUserRequest{Id: user_id})

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

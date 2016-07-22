package views

import (
	"core/api"
	"core/db"
	"core/models"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"strings"
	"utils/log"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterUserServiceServer(s, userServer{})
	})
}

type userServer struct{}

func (s userServer) FindOrCreateUser(ctx context.Context, request *core.CreateUserRequest) (*core.ReadUserReply, error) {
	user := models.User{}.Decode(request.User)
	searchUser := models.User{}

	query := db.New().Model(&models.User{})
	ok := false

	if request.User.Email != "" {
		query = query.Or("email = ?", request.User.Email)
		ok = true
	}

	if request.User.InstagramUsername != "" {
		query = query.Or("instagram_username = ?", strings.ToLower(request.User.InstagramUsername))
		ok = true
	}

	if request.User.InstagramId != 0 {
		query = query.Or("instagram_id = ?", request.User.InstagramId)
		ok = true
	}

	if request.User.Phone != "" {
		query = query.Or("phone = ?", request.User.Phone)
		ok = true
	}

	if !ok {
		errorf := fmt.Errorf("Incorrect request")
		log.Error(errorf)
		return nil, errorf
	}

	// we don't want to return RecordNotFound as an rpc error
	req := query.Find(&searchUser)
	err := req.Error

	if req.RecordNotFound() {
		err = db.New().Create(&user).Error
		//telegram notification, moved to user callbacks
		log.Error(err)
	} else {
		//update user phone if it not exists
		if searchUser.Phone == "" {
			searchUser.Phone = request.User.Phone
			db.New().Model(&searchUser).Update("phone", searchUser.Phone)
		}
		user = searchUser
	}

	return &core.ReadUserReply{
		Id:   int64(user.ID),
		User: user.PrivateEncode(),
	}, err
}

func (s userServer) ReadUser(ctx context.Context, request *core.ReadUserRequest) (*core.ReadUserReply, error) {
	user := models.User{}
	scope := db.New()

	if request.Id > 0 {
		scope = scope.Or("id = ?", request.Id)
	}

	if request.InstagramId > 0 {
		scope = scope.Or("instagram_id = ?", request.InstagramId)
	}
	if request.Phone != "" {
		scope = scope.Or("phone = ?", request.Phone)
	}
	if request.InstagramUsername != "" {
		scope = scope.Or("instagram_username = ?", strings.ToLower(request.InstagramUsername))
	}

	query := scope.Find(&user)
	if query.Error != nil && !query.RecordNotFound() {
		log.Error(query.Error)
		return nil, query.Error
	}

	return &core.ReadUserReply{
		Id:   int64(user.ID),
		User: user.PrivateEncode(),
	}, nil
}

//GetUserProfile returns user's public profile
func (s userServer) GetUserProfile(_ context.Context, req *core.UserProfileRequest) (reply *core.UserProfileReply, err error) {
	reply = &core.UserProfileReply{}
	var user *models.User
	var shop *models.Shop

	switch {
	case req.GetId() > 0:
		user, err = models.GetUserByID(uint(req.GetId()))
	case req.GetInstagramName() != "":
		shop, err = models.GetShopByInstagramName(req.GetInstagramName())
		if err != nil || shop == nil {
			user, err = models.GetUserByInstagramName(req.GetInstagramName())
		}
	}

	if err != nil {
		log.Error(err)
		return
	}

	switch {
	case user != nil:
		reply.Profile = &core.UserProfileReply_User{User: user.PublicEncode()}
	case shop != nil:
		reply.Profile = &core.UserProfileReply_Shop{Shop: shop.Encode()}
	}

	return
}

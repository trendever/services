package views

import (
	"core/api"
	"core/db"
	"core/models"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
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
		errorf := errors.New("Incorrect request")
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

	ok := false
	if request.Id > 0 {
		scope = scope.Where("id = ?", request.Id)
		ok = true
	} else {
		if request.InstagramId > 0 {
			scope = scope.Or("instagram_id = ?", request.InstagramId)
			ok = true
		}
		if request.Phone != "" {
			scope = scope.Or("phone = ?", request.Phone)
			ok = true
		}
		if request.InstagramUsername != "" {
			scope = scope.Or("instagram_username = ?", strings.ToLower(request.InstagramUsername))
			ok = true
		}
	}
	if !ok {
		return nil, errors.New("empty conditions")
	}

	query := scope.Find(&user)
	if query.Error != nil && !query.RecordNotFound() {
		log.Error(query.Error)
		return nil, query.Error
	}

	var cUser *core.User
	if request.Public {
		cUser = user.PublicEncode()
	} else {
		cUser = user.PrivateEncode()
	}
	var err error
	if request.GetShops {
		cUser.SupplierOf, err = models.GetShopsIDWhereUserIsSupplier(user.ID)
		if err != nil {
			return nil, err
		}
		cUser.SellerOf, err = models.GetShopsIDWhereUserIsSeller(user.ID)
		if err != nil {
			return nil, err
		}

	}
	return &core.ReadUserReply{
		Id:   int64(user.ID),
		User: cUser,
	}, err
}

func (s userServer) SetEmail(_ context.Context, req *core.SetEmailRequest) (*core.SetEmailReply, error) {
	if !govalidator.IsEmail(req.Email) {
		return &core.SetEmailReply{Error: "invalid email"}, nil
	}
	res := db.New().Model(&models.User{}).Where("id = ?", req.UserId).Update("email", req.Email)
	if res.Error != nil {
		return &core.SetEmailReply{Error: fmt.Sprintf("failed to set email: %v", res.Error)}, nil
	}
	if res.RowsAffected == 0 {
		return &core.SetEmailReply{Error: "unknown UserId"}, nil
	}
	return &core.SetEmailReply{}, nil
}

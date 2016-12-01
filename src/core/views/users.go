package views

import (
	"core/api"
	"core/models"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/db"
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
	searchUser, found, err := models.FindUserMatchAny(
		uint64(user.ID), user.InstagramID,
		user.Name, user.InstagramUsername,
		user.Email, user.Phone,
	)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if !found {
		err = db.New().Create(&user).Error
		log.Error(err)
	} else {
		//update user phone if it not exists
		if !searchUser.Confirmed {
			searchUser.Phone = request.User.Phone
			db.New().Model(&searchUser).Update("phone", searchUser.Phone)
		}
		user = *searchUser
	}

	return &core.ReadUserReply{
		Id:   int64(user.ID),
		User: user.PrivateEncode(),
	}, err
}

func (s userServer) CreateFakeUser(ctx context.Context, request *core.CreateUserRequest) (*core.ReadUserReply, error) {
	user := models.User{}.Decode(request.User)
	err := db.New().Create(&user).Error

	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = db.New().Model(&models.User{}).Where("id = ?", user.ID).Update("name", fmt.Sprintf("Client%v", user.ID)).Error

	return &core.ReadUserReply{
		Id:   int64(user.ID),
		User: user.PrivateEncode(),
	}, err
}

func (s userServer) ReadUser(ctx context.Context, request *core.ReadUserRequest) (*core.ReadUserReply, error) {
	user, found, err := models.FindUserMatchAny(
		request.Id, request.InstagramId,
		request.Name, request.InstagramUsername,
		"", request.Phone,
	)

	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !found {
		// @TODO fix code which can't handle nil User in response
		return &core.ReadUserReply{User: &core.User{}}, nil
	}

	var cUser *core.User
	if request.Public {
		cUser = user.PublicEncode()
	} else {
		cUser = user.PrivateEncode()
	}
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

func (s userServer) SetData(_ context.Context, req *core.SetDataRequest) (*core.SetDataReply, error) {
<<<<<<< HEAD

=======
>>>>>>> master
	_, found, err := models.FindUserMatchAny(0, 0, req.Name, req.Name, "", req.Phone)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if found {
		return &core.SetDataReply{}, errors.New("User exists")
	}

	updateMap := map[string]interface{}{}
	updateMap["phone"] = req.Phone
<<<<<<< HEAD
	updateMap["instagram_username"] = req.Name
	updateMap["name"] = req.Name
=======
	updateMap["isFake"] = false

	if req.Name != "" {
		updateMap["instagram_username"] = req.Name
		updateMap["name"] = req.Name
	}
>>>>>>> master

	res := db.New().Model(&models.User{}).Where("id = ?", req.UserId).UpdateColumns(updateMap)

	if res.Error != nil {
		//update user error
		return &core.SetDataReply{}, errors.New("Failed to update user")
	}

	return &core.SetDataReply{}, nil
}

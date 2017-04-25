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

const ConfirmedTelegramTopic = "telegram_conformed"

func init() {
	models.RegisterNotifyTemplate(ConfirmedTelegramTopic)
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
		if !searchUser.Confirmed {
			searchUser.Phone = user.Phone
			searchUser.Source = user.Source
			err = db.New().Model(&searchUser).Updates(map[string]interface{}{
				"phone":  user.Phone,
				"source": user.Source,
			}).Error
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
	updateMap["isFake"] = false

	if req.Name != "" {
		updateMap["instagram_username"] = req.Name
		updateMap["name"] = req.Name
	}

	res := db.New().Model(&models.User{}).Where("id = ?", req.UserId).UpdateColumns(updateMap)

	if res.Error != nil {
		//update user error
		return &core.SetDataReply{}, errors.New("Failed to update user")
	}

	return &core.SetDataReply{}, nil
}

func (s userServer) ListTelegrams(_ context.Context, req *core.ListTelegramsRequest) (*core.ListTelegramsReply, error) {
	var list []*models.Telegram
	q := db.New().Where("user_id = ?", req.UserId)
	if req.ConfirmedOnly {
		q = q.Where("confirmed")
	}
	err := q.Find(&list).Error
	if err != nil {
		return &core.ListTelegramsReply{Error: err.Error()}, nil
	}

	ret := core.ListTelegramsReply{}
	for _, t := range list {
		ret.Telegrams = append(ret.Telegrams, t.Encode())
	}
	return &ret, nil
}

func (s userServer) AddTelegram(_ context.Context, req *core.AddTelegramRequest) (*core.AddTelegramReply, error) {
	var userID uint64
	if req.Username != "" {
		user, found, err := models.FindUserMatchAny(
			0, 0,
			req.Username, req.Username,
			"", "",
		)
		if err != nil {
			return &core.AddTelegramReply{Error: err.Error()}, nil
		}
		if !found {
			return &core.AddTelegramReply{Error: "user not found"}, nil
		}
		userID = uint64(user.ID)
	} else {
		userID = req.UserId
	}
	err := db.New().Assign(&models.Telegram{Username: req.SubsricberName}).FirstOrCreate(&models.Telegram{
		UserID: userID,
		ChatID: req.ChatId,
	}).Error
	if err != nil {
		return &core.AddTelegramReply{Error: err.Error()}, nil
	}
	return &core.AddTelegramReply{}, nil
}

func (s userServer) ConfirmTelegram(_ context.Context, req *core.ConfirmTelegramRequest) (*core.ConfirmTelegramReply, error) {
	tg := models.Telegram{UserID: req.UserId, ChatID: req.ChatId}
	err := db.New().First(&tg).Error
	if err != nil {
		return &core.ConfirmTelegramReply{Error: err.Error()}, nil
	}
	if tg.Confirmed {
		return &core.ConfirmTelegramReply{}, nil
	}
	tg.Confirmed = true
	err = db.New().Save(&tg).Error
	if err != nil {
		return &core.ConfirmTelegramReply{Error: err.Error()}, nil
	}
	go models.GetNotifier().NotifyUserByID(req.UserId, ConfirmedTelegramTopic, map[string]interface{}{
		"telegram": tg,
	})
	return &core.ConfirmTelegramReply{}, nil
}

func (s userServer) DelTelegram(_ context.Context, req *core.DelTelegramRequest) (*core.DelTelegramReply, error) {
	var userID uint64
	if req.Username != "" {
		user, found, err := models.FindUserMatchAny(
			0, 0,
			req.Username, req.Username,
			"", "",
		)
		if err != nil {
			return &core.DelTelegramReply{Error: err.Error()}, nil
		}
		if !found {
			return &core.DelTelegramReply{Error: "user not found"}, nil
		}
		userID = uint64(user.ID)
	} else {
		userID = req.UserId
	}
	err := db.New().Delete(&models.Telegram{UserID: userID, ChatID: req.ChatId}).Error
	if err != nil {
		return &core.DelTelegramReply{Error: err.Error()}, nil
	}
	return &core.DelTelegramReply{}, nil
}

package server

import (
	"auth/bitly"
	"auth/config"
	"auth/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dvsekhvalnov/jose2go"
	"github.com/ttacon/libphonenumber"
	"golang.org/x/net/context"
	auth_protocol "proto/auth"
	core_protocol "proto/core"
	sms_protocol "proto/sms"
	"text/template"
	"time"
	"utils/log"
)

//DefaultTokenExp is a default token ttl
const DefaultTokenExp = time.Hour * 24 * 365

type authServer struct {
	core      core_protocol.UserServiceClient
	sms       sms_protocol.SmsServiceClient
	passwords models.UserPasswords
	sharedKey interface{}
	alg       string
}

//NewAuthServer makes new server
func NewAuthServer(core core_protocol.UserServiceClient, sms sms_protocol.SmsServiceClient, passwords models.UserPasswords, alg string, sharedKey interface{}) auth_protocol.AuthServiceServer {
	return &authServer{
		core:      core,
		sms:       sms,
		passwords: passwords,
		alg:       alg,
		sharedKey: sharedKey,
	}
}

// @CHECK background contexts everywhere? why?

//RegisterNewUser creates new user
func (s *authServer) RegisterNewUser(ctx context.Context, request *auth_protocol.NewUserRequest) (*auth_protocol.UserReply, error) {
	//todo: add country to request
	phoneNumber, err := checkPhoneNumber(request.PhoneNumber, "")

	if err != nil {
		log.Debug("invalid phone number %v", phoneNumber)
		return &auth_protocol.UserReply{
			ErrorCode:    auth_protocol.ErrorCodes_INCORRECT_PHONE_FORMAT,
			ErrorMessage: err.Error(),
		}, nil
	}

	userRequest := &core_protocol.ReadUserRequest{
		Phone:             phoneNumber,
		Name:              request.Username,
		InstagramUsername: request.InstagramUsername,
	}
	userExists, err := s.core.ReadUser(context.Background(), userRequest)
	if err != nil {
		log.Error(fmt.Errorf("failed to read user with phone %v: %v", phoneNumber, err))
		return nil, err
	}
	//That's mean we found confirmed user
	if userExists.Id > 0 && userExists.User.Confirmed {
		log.Warn("User already exists: %v", userExists)
		return &auth_protocol.UserReply{ErrorCode: auth_protocol.ErrorCodes_USER_ALREADY_EXISTS}, nil
	}

	newUser := &core_protocol.CreateUserRequest{
		User: &core_protocol.User{
			Phone:             phoneNumber,
			InstagramUsername: request.InstagramUsername,
			Name:              request.Username,
		},
	}
	resp, err := s.core.FindOrCreateUser(context.Background(), newUser)
	if err != nil {
		log.Error(fmt.Errorf("failed to create user with phone %v: %v", phoneNumber, err))
		return nil, err
	}

	go (func() {
		if err := s.sendSMSWithPassword(uint(resp.Id), phoneNumber); err != nil {
			log.Error(fmt.Errorf("failed to send password sms to %v: %v", phoneNumber, err))
		}
	})()

	return &auth_protocol.UserReply{
		PhoneNumber:       phoneNumber,
		InstagramUsername: request.InstagramUsername,
		Username:          request.Username,
		Id:                uint64(resp.Id),
	}, nil

}

//Login returns JWT token for user
func (s *authServer) Login(ctx context.Context, request *auth_protocol.LoginRequest) (*auth_protocol.LoginReply, error) {

	//todo: add country to request
	phoneNumber, err := checkPhoneNumber(request.PhoneNumber, "")

	if err != nil {
		log.Debug("invalid phone number %v", phoneNumber)
		return &auth_protocol.LoginReply{
			ErrorCode:    auth_protocol.ErrorCodes_INCORRECT_PHONE_FORMAT,
			ErrorMessage: err.Error(),
		}, nil
	}

	userRequest := &core_protocol.ReadUserRequest{
		Phone: phoneNumber,
	}
	resp, err := s.core.ReadUser(context.Background(), userRequest)

	if err != nil {
		log.Error(fmt.Errorf("failed to read user with phone %v: %v", phoneNumber, err))
		return nil, err
	}

	if resp.Id == 0 {
		log.Error(fmt.Errorf("user with phone %v not found", phoneNumber))
		return s.wrongCredentialsReply(), nil
	}

	pass, err := s.passwords.FindByUserID(uint(resp.Id))
	if err != nil {
		log.Error(fmt.Errorf("failed to find password for user %v: %v", resp.Id, err))
		return nil, err
	}

	if pass == nil {
		log.Debug("password not found for user %v", resp.Id)
		return s.wrongCredentialsReply(), nil
	}

	if pass.SmsPassword != request.Password {
		log.Debug("wrong password for user %v: '%v' != '%v'", resp.Id, pass.SmsPassword, request.Password)
		return s.wrongCredentialsReply(), nil
	}

	tokenPayload, err := json.Marshal(&auth_protocol.Token{UID: uint64(resp.Id), Exp: time.Now().Add(DefaultTokenExp).Unix()})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	token, err := jose.Sign(string(tokenPayload), jose.HS256, s.sharedKey)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	s.passwords.Delete(pass)
	if !resp.User.Confirmed {
		go s.core.ConfirmUser(context.Background(), &core_protocol.ConfirmUserRequest{UserId: uint64(resp.Id)})
	}

	return &auth_protocol.LoginReply{Token: token}, nil
}

//SendNewSmsPassword sends a password to the user phone number
func (s *authServer) SendNewSmsPassword(ctx context.Context, request *auth_protocol.SmsPasswordRequest) (*auth_protocol.SmsPasswordReply, error) {

	//todo: add country to request
	phoneNumber, err := checkPhoneNumber(request.PhoneNumber, "")

	if err != nil {
		return &auth_protocol.SmsPasswordReply{
			ErrorCode:    auth_protocol.ErrorCodes_INCORRECT_PHONE_FORMAT,
			ErrorMessage: err.Error(),
		}, nil
	}

	userRequest := &core_protocol.ReadUserRequest{
		Phone: phoneNumber,
	}
	resp, err := s.core.ReadUser(context.Background(), userRequest)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if resp.Id == 0 {
		return &auth_protocol.SmsPasswordReply{ErrorCode: auth_protocol.ErrorCodes_USER_NOT_EXISTS}, nil
	}

	go (func() {
		log.Error(s.sendSMSWithPassword(uint(resp.Id), phoneNumber))
	})()

	return &auth_protocol.SmsPasswordReply{Status: "ok"}, nil
}

//GetTokenData returns parsed token data or error
func (s *authServer) GetTokenData(ctx context.Context, request *auth_protocol.TokenDataRequest) (*auth_protocol.TokenDataReply, error) {
	token := &auth_protocol.Token{}

	data, _, err := jose.Decode(request.Token, s.sharedKey)
	if err != nil {
		log.Warn("Wrong token data: %v, token: %v", err, request.Token)
		return &auth_protocol.TokenDataReply{ErrorCode: auth_protocol.ErrorCodes_WRONG_TOKEN_DATA}, err
	}

	err = json.Unmarshal([]byte(data), token)
	if err != nil {
		log.Warn("Wrong token data: %v, token: %v", err, request.Token)
		return &auth_protocol.TokenDataReply{ErrorCode: auth_protocol.ErrorCodes_WRONG_TOKEN_DATA}, err
	}
	if token.Exp < time.Now().Unix() {
		return &auth_protocol.TokenDataReply{
			Token:     token,
			ErrorCode: auth_protocol.ErrorCodes_EXPIRED_TOKEN_DATA,
		}, nil
	}
	return &auth_protocol.TokenDataReply{Token: token}, nil
}

//GetNewToken returns new valid token for user
func (s *authServer) GetNewToken(ctx context.Context, req *auth_protocol.NewTokenRequest) (*auth_protocol.NewTokenReply, error) {

	var userID = req.UserId

	// @TODO @BUG: There ARE mutiple users with the same phone number
	// I am not sure if it's intended or misuse
	// but, the scheme with using phone is faulty just because this generates unneeded requests:
	//  * core: auth.GetNewToken(phone)
	//    * auth: core.GetIdByPhone(phone)
	// Nothing else seems not be using this method:
	// https://github.com/search?l=go&q=org%3Atrendever+GetNewToken&type=Code
	if userID == 0 {
		//todo: add country to request
		phoneNumber, err := checkPhoneNumber(req.PhoneNumber, "")

		userRequest := &core_protocol.ReadUserRequest{
			Phone: phoneNumber,
		}
		if err != nil {
			return nil, err
		}

		resp, err := s.core.ReadUser(context.Background(), userRequest)

		if err != nil {
			log.Error(err)
			return nil, err
		}

		if resp.Id == 0 {
			return nil, errors.New("User not exists")
		}

		userID = uint64(resp.Id)
	}

	tokenPayload, err := json.Marshal(&auth_protocol.Token{
		UID: userID,
		Exp: time.Now().Add(DefaultTokenExp).Unix(),
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	token, err := jose.Sign(string(tokenPayload), jose.HS256, s.sharedKey)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &auth_protocol.NewTokenReply{Token: token}, nil
}

func (s *authServer) sendSMSWithPassword(uid uint, phone string) error {
	u := &models.UserPassword{UserID: uid}
	if exits, _ := s.passwords.FindByUserID(uid); exits != nil {
		s.passwords.Delete(exits)
	}
	if err := s.passwords.Create(u); err != nil {
		return err
	}

	token, err := s.getToken(uint64(uid))

	if err != nil {
		return err
	}

	url, err := bitly.GetShortUrl(bitly.GetSiteUrl(token))

	if err != nil {
		return err
	}

	tpl, err := template.New("sms_template").Parse(config.Get().SmsTemplate)

	if err != nil {
		return err
	}

	wr := &bytes.Buffer{}

	err = tpl.Execute(wr, struct {
		Password string
		Url      string
	}{
		Password: u.SmsPassword,
		Url:      url.URL,
	})

	if err != nil {
		return err
	}
	log.Debug("%s", wr.String())
	_, err = s.sms.SendSMS(
		context.Background(),
		&sms_protocol.SendSMSRequest{
			Phone: phone,
			Msg:   wr.String(),
		})
	return err
}

func (s *authServer) wrongCredentialsReply() *auth_protocol.LoginReply {
	return &auth_protocol.LoginReply{
		ErrorCode:    auth_protocol.ErrorCodes_WRONG_CREDENTIALS,
		ErrorMessage: "Wrong credentials",
	}
}

func checkPhoneNumber(phoneNumber, country string) (string, error) {
	if country == "" {
		country = "RU"
	}

	number, err := libphonenumber.Parse(phoneNumber, country)
	if err != nil {
		return "", err
	}

	if !libphonenumber.IsValidNumber(number) {
		return "", errors.New("Phone number isn't valid")
	}

	return libphonenumber.Format(number, libphonenumber.E164), nil
}

func (s *authServer) getToken(uid uint64) (string, error) {
	tokenPayload, err := json.Marshal(&auth_protocol.Token{UID: uid, Exp: time.Now().Add(DefaultTokenExp).Unix()})

	if err != nil {
		return "", err
	}

	token, err := jose.Sign(string(tokenPayload), jose.HS256, s.sharedKey)
	return token, err

}

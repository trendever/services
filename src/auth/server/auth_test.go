package server

import (
	"auth/models"
	"encoding/json"
	"errors"
	"github.com/dvsekhvalnov/jose2go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"proto/auth"
	core_protocol "proto/core"
	sms_protocol "proto/sms"
	"testing"
	"time"
)

var key = []byte("secret_key")

func TestCreateUser(t *testing.T) {
	var data = []struct {
		core          *mockCore
		sms           *mockSms
		pass          *mockPasswords
		request       *auth.NewUserRequest
		alg           string
		expectedResp  *auth.UserReply
		expectedError error
	}{
		//New user
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 0}, &core_protocol.ReadUserReply{Id: 1}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(nil),
			},
			request:       &auth.NewUserRequest{PhoneNumber: "+77775793511"},
			alg:           jose.HS256,
			expectedResp:  &auth.UserReply{},
			expectedError: nil,
		},
		//User already exists
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 1, User: &core_protocol.User{Phone: "+77775793511"}}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(),
			},
			request:       &auth.NewUserRequest{PhoneNumber: "+77775793511"},
			alg:           jose.HS256,
			expectedResp:  &auth.UserReply{ErrorCode: auth.ErrorCodes_USER_ALREADY_EXISTS},
			expectedError: nil,
		},
		//incorrect user number
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 0}, &core_protocol.ReadUserReply{Id: 1}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(nil),
			},
			request:       &auth.NewUserRequest{PhoneNumber: "+777757935zz"},
			alg:           jose.HS256,
			expectedResp:  &auth.UserReply{ErrorCode: auth.ErrorCodes_INCORRECT_PHONE_FORMAT},
			expectedError: nil,
		},
	}

	for _, test := range data {
		server := NewAuthServer(test.core, test.sms, test.pass, jose.HS256, key)
		actualResp, actualError := server.RegisterNewUser(context.Background(), test.request)
		if test.expectedError == nil {
			if actualError != nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		} else {
			if actualError == nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		}

		if test.expectedResp == nil {
			if actualResp != nil {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		} else {
			if actualResp == nil || actualResp.ErrorCode != test.expectedResp.ErrorCode {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		}

	}
}

func TestLogin(t *testing.T) {
	var data = []struct {
		core          *mockCore
		sms           *mockSms
		pass          *mockPasswords
		request       *auth.LoginRequest
		alg           string
		expectedResp  *auth.LoginReply
		expectedError error
	}{
		//Success Login
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 1}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(&models.UserPassword{SmsPassword: "123456"}),
			},
			request:       &auth.LoginRequest{PhoneNumber: "87775553311", Password: "123456"},
			alg:           jose.HS256,
			expectedResp:  &auth.LoginReply{},
			expectedError: nil,
		},
		// User not exists
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 0}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(),
			},
			request:       &auth.LoginRequest{PhoneNumber: "87775553311", Password: "123456"},
			alg:           jose.HS256,
			expectedResp:  &auth.LoginReply{ErrorCode: auth.ErrorCodes_WRONG_CREDENTIALS},
			expectedError: nil,
		},
		//Wrong Pass
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 1}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(&models.UserPassword{SmsPassword: "654321"}),
			},
			request:       &auth.LoginRequest{PhoneNumber: "87775553311", Password: "123456"},
			alg:           jose.HS256,
			expectedResp:  &auth.LoginReply{ErrorCode: auth.ErrorCodes_WRONG_CREDENTIALS},
			expectedError: nil,
		},
	}

	for _, test := range data {
		server := NewAuthServer(test.core, test.sms, test.pass, jose.HS256, key)
		actualResp, actualError := server.Login(context.Background(), test.request)
		if test.expectedError == nil {
			if actualError != nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		} else {
			if actualError == nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		}

		if test.expectedResp == nil {
			if actualResp != nil {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		} else {
			if actualResp == nil || actualResp.ErrorCode != test.expectedResp.ErrorCode {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		}

	}
}

func TestSendNewSmsPassword(t *testing.T) {
	var data = []struct {
		core          *mockCore
		sms           *mockSms
		pass          *mockPasswords
		request       *auth.SmsPasswordRequest
		alg           string
		expectedResp  *auth.SmsPasswordReply
		expectedError error
	}{
		//Success send
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 1}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(nil),
			},
			request:       &auth.SmsPasswordRequest{PhoneNumber: "87775553311"},
			alg:           jose.HS256,
			expectedResp:  &auth.SmsPasswordReply{},
			expectedError: nil,
		},
		// User not exists
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 0}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(),
			},
			request:       &auth.SmsPasswordRequest{PhoneNumber: "87775553311"},
			alg:           jose.HS256,
			expectedResp:  &auth.SmsPasswordReply{ErrorCode: auth.ErrorCodes_USER_NOT_EXISTS},
			expectedError: nil,
		},
		// Password already exist
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 0}),
				Errors:          newGenerator(),
			},
			sms: &mockSms{},
			pass: &mockPasswords{
				Passwords: newGenerator(&models.UserPassword{}),
			},
			request:       &auth.SmsPasswordRequest{PhoneNumber: "87775553311"},
			alg:           jose.HS256,
			expectedResp:  &auth.SmsPasswordReply{ErrorCode: auth.ErrorCodes_USER_NOT_EXISTS},
			expectedError: nil,
		},
	}

	for _, test := range data {
		server := NewAuthServer(test.core, test.sms, test.pass, jose.HS256, key)
		actualResp, actualError := server.SendNewSmsPassword(context.Background(), test.request)
		if test.expectedError == nil {
			if actualError != nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		} else {
			if actualError == nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		}

		if test.expectedResp == nil {
			if actualResp != nil {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		} else {
			if actualResp == nil || actualResp.ErrorCode != test.expectedResp.ErrorCode {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		}

	}
}

func getTokenData(payload, enc string, key interface{}) string {

	token, err := jose.Sign(payload, enc, key)
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func TestGetTokenData(t *testing.T) {
	validToken, _ := json.Marshal(&auth.Token{UID: 1, Exp: time.Now().Add(time.Hour).Unix()})
	invalidToken, _ := json.Marshal(&auth.Token{UID: 1, Exp: time.Now().Add(-time.Hour).Unix()})

	var data = []struct {
		core          *mockCore
		sms           *mockSms
		pass          *mockPasswords
		request       *auth.TokenDataRequest
		alg           string
		expectedResp  *auth.TokenDataReply
		expectedError error
	}{
		//Success
		{
			core:          &mockCore{},
			sms:           &mockSms{},
			pass:          &mockPasswords{},
			request:       &auth.TokenDataRequest{Token: getTokenData(string(validToken), jose.HS256, key)},
			alg:           jose.HS256,
			expectedResp:  &auth.TokenDataReply{},
			expectedError: nil,
		},
		//Expired
		{
			core:          &mockCore{},
			sms:           &mockSms{},
			pass:          &mockPasswords{},
			request:       &auth.TokenDataRequest{Token: getTokenData(string(invalidToken), jose.HS256, key)},
			alg:           jose.HS256,
			expectedResp:  &auth.TokenDataReply{ErrorCode: auth.ErrorCodes_EXPIRED_TOKEN_DATA},
			expectedError: nil,
		},
		//Wrong alg
		{
			core:          &mockCore{},
			sms:           &mockSms{},
			pass:          &mockPasswords{},
			request:       &auth.TokenDataRequest{Token: getTokenData(string(validToken), jose.NONE, nil)},
			alg:           jose.HS256,
			expectedResp:  &auth.TokenDataReply{ErrorCode: auth.ErrorCodes_WRONG_TOKEN_DATA},
			expectedError: errors.New("some error"),
		},
		//Wrong payload
		{
			core:          &mockCore{},
			sms:           &mockSms{},
			pass:          &mockPasswords{},
			request:       &auth.TokenDataRequest{Token: getTokenData("wrong data", jose.HS256, key)},
			alg:           jose.HS256,
			expectedResp:  &auth.TokenDataReply{ErrorCode: auth.ErrorCodes_WRONG_TOKEN_DATA},
			expectedError: errors.New("some error"),
		},
	}

	for _, test := range data {
		server := NewAuthServer(test.core, test.sms, test.pass, jose.HS256, key)
		actualResp, actualError := server.GetTokenData(context.Background(), test.request)
		if test.expectedError == nil {
			if actualError != nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		} else {
			if actualError == nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		}

		if test.expectedResp == nil {
			if actualResp != nil {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		} else {
			if actualResp == nil || actualResp.ErrorCode != test.expectedResp.ErrorCode {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		}

	}
}

func TestGetNewToken(t *testing.T) {
	var data = []struct {
		core          *mockCore
		sms           *mockSms
		pass          *mockPasswords
		request       *auth.NewTokenRequest
		alg           string
		expectedResp  *auth.NewTokenReply
		expectedError error
	}{
		//Success get
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 1}),
				Errors:          newGenerator(),
			},
			sms:           &mockSms{},
			pass:          &mockPasswords{},
			request:       &auth.NewTokenRequest{PhoneNumber: "87775553311"},
			alg:           jose.HS256,
			expectedResp:  &auth.NewTokenReply{},
			expectedError: nil,
		},
		//Success get
		{
			core: &mockCore{
				ReadUserReplies: newGenerator(&core_protocol.ReadUserReply{Id: 0}),
				Errors:          newGenerator(),
			},
			sms:           &mockSms{},
			pass:          &mockPasswords{},
			request:       &auth.NewTokenRequest{PhoneNumber: "87775553311"},
			alg:           jose.HS256,
			expectedResp:  nil,
			expectedError: errors.New("some err"),
		},
	}

	for _, test := range data {
		server := NewAuthServer(test.core, test.sms, test.pass, jose.HS256, key)
		actualResp, actualError := server.GetNewToken(context.Background(), test.request)
		if test.expectedError == nil {
			if actualError != nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		} else {
			if actualError == nil {
				t.Fatalf("Expected error to be %v, but go %v", test.expectedError, actualError)
			}
		}

		if test.expectedResp == nil {
			if actualResp != nil {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		} else {
			if actualResp == nil {
				t.Fatalf("Expected response to be %v, but go %v", test.expectedResp, actualResp)
			}
		}

	}
}

type mockSms struct {
}

func (m *mockSms) SendSMS(ctx context.Context, in *sms_protocol.SendSMSRequest, opts ...grpc.CallOption) (*sms_protocol.SendSMSResult, error) {
	return &sms_protocol.SendSMSResult{}, nil
}
func (m *mockSms) RetrieveSmsStatus(ctx context.Context, in *sms_protocol.RetrieveSmsStatusRequest, opts ...grpc.CallOption) (*sms_protocol.RetrieveSmsStatusResult, error) {
	return &sms_protocol.RetrieveSmsStatusResult{}, nil
}

type mockCore struct {
	ReadUserReplies *generator
	Errors          *generator
}

func (m *mockCore) FindOrCreateUser(ctx context.Context, in *core_protocol.CreateUserRequest, opts ...grpc.CallOption) (*core_protocol.ReadUserReply, error) {
	reply, ok := m.ReadUserReplies.Next()
	err, _ := m.Errors.Next()
	if !ok {
		log.Fatalf("Unexpected FindOrCreateUser(%q) call", in)
	}
	errResp, ok := err.(error)
	replyResp, ok := reply.(*core_protocol.ReadUserReply)

	return replyResp, errResp
}
func (m *mockCore) ReadUser(ctx context.Context, in *core_protocol.ReadUserRequest, opts ...grpc.CallOption) (*core_protocol.ReadUserReply, error) {
	reply, ok := m.ReadUserReplies.Next()
	err, _ := m.Errors.Next()
	if !ok {
		log.Fatalf("Unexpected ReadUser(%q) call", in)
	}
	errResp, ok := err.(error)
	replyResp, ok := reply.(*core_protocol.ReadUserReply)

	return replyResp, errResp
}

type mockPasswords struct {
	Passwords *generator
}

func (m *mockPasswords) Create(*models.UserPassword) error {
	return nil
}
func (m *mockPasswords) Delete(*models.UserPassword) error {
	return nil
}
func (m *mockPasswords) FindByUserID(userID uint) (*models.UserPassword, error) {
	pass, ok := m.Passwords.Next()
	if !ok {
		log.Fatalf("Unexpected FindByUserID(%q) call", userID)
	}
	passResp, _ := pass.(*models.UserPassword)
	return passResp, nil
}

type generator struct {
	Items []interface{}
	next  int
}

func newGenerator(item ...interface{}) *generator {
	return &generator{Items: item}
}

func (g *generator) NextIsOk() bool {
	return len(g.Items) > g.next
}

func (g *generator) Next() (interface{}, bool) {
	if g.NextIsOk() {
		g.next++
		return g.Items[g.next-1], true
	}
	return nil, false
}

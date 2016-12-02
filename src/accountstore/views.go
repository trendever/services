package main

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"proto/accountstore"
	"utils/rpc"
)

// StartServer inits grpc server
func (s *svc) StartServer() {
	server := rpc.Serve(settings.Listen)
	accountstore.RegisterAccountStoreServiceServer(server, s)
}

func (s *svc) Add(_ context.Context, in *accountstore.AddRequest) (*accountstore.AddReply, error) {

	account, err := s.ig.Login(in.InstagramUsername, in.Password)
	if err != nil {
		return nil, err
	}
	account.Role = in.Role

	// save Creates if not exists
	err = Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.AddReply{
		NeedCode: !account.Valid,
	}, nil
}

func (s *svc) Confirm(_ context.Context, in *accountstore.ConfirmRequest) (*accountstore.ConfirmReply, error) {

	account, err := FindAccount(&Account{
		InstagramUsername: in.InstagramUsername,
		InstagramID:       in.InstagramId,
	})
	if err != nil {
		return nil, err
	}

	err = s.ig.VerifyCode(account, in.Code)

	return &accountstore.ConfirmReply{}, err
}

func (s *svc) MarkInvalid(_ context.Context, in *accountstore.MarkInvalidRequest) (*accountstore.MarkInvalidReply, error) {

	account, err := FindAccount(&Account{
		InstagramUsername: in.InstagramUsername,
		InstagramID:       in.InstagramId,
	})
	if err != nil {
		return nil, err
	}

	account.Valid = false

	err = Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.MarkInvalidReply{}, nil
}

func (s *svc) Search(_ context.Context, in *accountstore.SearchRequest) (*accountstore.SearchReply, error) {

	accounts, err := Find(!in.IncludeInvalids, in.Roles)
	if err != nil {
		return nil, err
	}

	return &accountstore.SearchReply{
		Accounts: EncodeAll(accounts),
	}, nil
}

func (s *svc) Get(_ context.Context, in *accountstore.GetRequest) (*accountstore.GetReply, error) {

	account, err := FindAccount(&Account{
		InstagramUsername: in.InstagramUsername,
		InstagramID:       in.InstagramId,
	})
	if err == gorm.ErrRecordNotFound {
		return &accountstore.GetReply{
			Found: false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	if in.HidePrivate {
		return &accountstore.GetReply{
			Account: account.EncodePrivate(),
		}, nil
	}

	return &accountstore.GetReply{
		Found:   true,
		Account: account.Encode(),
	}, nil
}

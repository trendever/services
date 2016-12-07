package main

import (
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

	account, err := s.ig.Login(in.InstagramUsername, in.Password, in.PreferEmail)
	if err != nil {
		return nil, err
	}

	account.Role = in.Role
	account.OwnerID = in.OwnerId

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

	accounts, err := Find(in)
	if err != nil {
		return nil, err
	}

	return &accountstore.SearchReply{
		Accounts: EncodeAll(accounts, in.HidePrivate),
	}, nil
}

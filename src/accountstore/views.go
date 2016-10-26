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

	account, err := s.ig.Login(in.InstagramUsername, in.Password)
	if err != nil {
		return nil, err
	}

	// save Creates if not exists
	err = s.repo.Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.AddReply{
		NeedCode: !account.Valid,
	}, nil
}

func (s *svc) Confirm(_ context.Context, in *accountstore.ConfirmRequest) (*accountstore.ConfirmReply, error) {

	return nil, nil
}

func (s *svc) MarkInvalid(_ context.Context, in *accountstore.MarkInvalidRequest) (*accountstore.MarkInvalidReply, error) {

	account, err := s.repo.FindByName(in.InstagramUsername)
	if err != nil {
		return nil, err
	}

	account.Valid = false

	err = s.repo.Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.MarkInvalidReply{}, nil
}

func (s *svc) GetValid(_ context.Context, in *accountstore.GetValidRequest) (*accountstore.GetValidReply, error) {

	accounts, err := s.repo.FindValid()
	if err != nil {
		return nil, err
	}

	return &accountstore.GetValidReply{
		Accounts: EncodeAll(accounts),
	}, nil
}

func (s *svc) GetByName(_ context.Context, in *accountstore.GetByNameRequest) (*accountstore.GetByNameReply, error) {

	account, err := s.repo.FindByName(in.InstagramUsername)
	if err != nil {
		return nil, err
	}

	if in.HidePrivate {

		return &accountstore.GetByNameReply{
			Account: account.EncodePrivate(),
		}, nil
	}

	return &accountstore.GetByNameReply{
		Account: account.Encode(),
	}, nil
}

package main

import (
	"fmt"
	"golang.org/x/net/context"
	"instagram"
	"math/rand"
	"proto/accountstore"
	"utils/log"
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
	err = Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.AddReply{
		NeedCode: !account.Valid,
	}, nil
}

func (s *svc) Confirm(_ context.Context, in *accountstore.ConfirmRequest) (*accountstore.ConfirmReply, error) {

	return &accountstore.ConfirmReply{}, nil
}

func (s *svc) MarkInvalid(_ context.Context, in *accountstore.MarkInvalidRequest) (*accountstore.MarkInvalidReply, error) {

	account, err := FindByName(in.InstagramUsername)
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

func (s *svc) Get(_ context.Context, in *accountstore.GetRequest) (*accountstore.GetReply, error) {

	accounts, err := Find(!in.IncludeInvalids, in.Roles)
	if err != nil {
		return nil, err
	}

	return &accountstore.GetReply{
		Accounts: EncodeAll(accounts),
	}, nil
}

func (s *svc) GetByName(_ context.Context, in *accountstore.GetByNameRequest) (*accountstore.GetByNameReply, error) {

	account, err := FindByName(in.InstagramUsername)
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

func (s *svc) DebugTryInvalidate(_ context.Context, in *accountstore.DebugTryInvalidateRequest) (*accountstore.DebugTryInvalidateReply, error) {

	acc, err := FindByName(in.InstagramUsername)
	if err != nil {
		return nil, err
	}

	api, err := instagram.Restore(acc.Cookie)
	if err != nil {
		return nil, err
	}

	for i := uint64(0); i < in.Requests; i++ {
		go func() {
			if _, err := api.SearchUsers(fmt.Sprintf("%v", rand.Intn(32))); err != nil {
				log.Error(err)
			}
		}()
	}

	return &accountstore.DebugTryInvalidateReply{}, nil
}

package main

import (
	"common/log"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"instagram"
	"proto/accountstore"
	"proto/core"
	"utils/rpc"
)

// StartServer inits grpc server
func (s *svc) StartServer() {
	// connect to RPCs
	s.shopClient = core.NewShopServiceClient(rpc.Connect(settings.RPC.Core))

	server := rpc.MakeServer(settings.Listen)
	accountstore.RegisterAccountStoreServiceServer(server.Server, s)
	server.StartServe()
}

func (s *svc) Add(_ context.Context, in *accountstore.AddRequest) (*accountstore.AddReply, error) {

	account, err := s.ig.Login(in.InstagramUsername, in.Password, in.Proxy, in.PreferEmail, in.OwnerId)
	if err != nil {
		log.Errorf("failed to login as '%v': %v", in.InstagramUsername, err)
		notifyTelegram(fmt.Sprintf("failed to add bot '%v': %v", in.InstagramUsername, err))
		return nil, err
	}

	account.Role = in.Role
	account.OwnerID = in.OwnerId

	// attach shop if needed
	if in.Role == accountstore.Role_User {

		ctx, cancel := rpc.DefaultContext()
		defer cancel()

		res, err := s.shopClient.FindOrCreateAttachedShop(
			ctx, &core.FindOrCreateAttachedShopRequest{
				SupplierId:        in.OwnerId,
				InstagramUsername: in.InstagramUsername,
			},
		)
		if err != nil {
			log.Errorf("failed to get attached shop for %v: %v", in.InstagramUsername, err)
			return nil, fmt.Errorf("RPC error: %v", err)
		}
		if res.Error != "" {
			log.Errorf("failed to get attached shop for %v: %v", in.InstagramUsername, res.Error)
			return nil, errors.New(res.Error)
		}

	}

	// save Creates if not exists
	err = Save(account)
	if err != nil {
		log.Errorf("failed to save account %v: %v", in.InstagramUsername, err)
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

	err = s.ig.VerifyCode(account, in.Password, in.Code)

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

	log.Debug("Invalidating account %v; reason: %v", account.InstagramUsername, in.Reason)

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

func (s *svc) SetProxy(_ context.Context, in *accountstore.SetProxyRequest) (*accountstore.SetProxyReply, error) {
	account, err := FindAccount(&Account{
		InstagramUsername: in.InstagramUsername,
	})
	if err != nil {
		return nil, err
	}

	err = s.ig.SetProxy(account, in.Proxy)
	if err != nil {
		return nil, err
	}

	err = Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.SetProxyReply{}, nil
}

func (s *svc) SetDebug(_ context.Context, in *accountstore.SetDebugRequest) (*accountstore.SetDebugReply, error) {
	account, err := FindAccount(&Account{
		InstagramUsername: in.InstagramUsername,
	})
	if err != nil {
		return nil, err
	}

	ig, err := instagram.Restore(account.Cookie, "", false)
	if err != nil {
		return nil, err
	}

	ig.Debug = in.Debug
	account.Cookie, err = ig.Save()
	if err != nil {
		return nil, err
	}

	err = Save(account)
	if err != nil {
		return nil, err
	}

	return &accountstore.SetDebugReply{}, nil
}

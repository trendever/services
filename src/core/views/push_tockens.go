package views

import (
	"core/api"
	"core/models"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	proto "proto/core"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		proto.RegisterPushTokensServiceServer(
			s,
			&pushTokensServer{
				repo: models.GetPushTokensRepository(),
			},
		)
	})
}

type pushTokensServer struct {
	repo models.PushTokensRepository
}

func (s *pushTokensServer) AddToken(ctx context.Context, in *proto.AddTokenRequest) (*proto.ErrorResult, error) {
	token := models.PushToken{}.Decode(in.Token)
	err := s.repo.AddToken(token)
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return &proto.ErrorResult{errString}, nil
}

func (s *pushTokensServer) DelToken(ctx context.Context, in *proto.DelTokenRequest) (*proto.ErrorResult, error) {
	err := s.repo.DelToken(uint(in.TokenId), uint(in.UserId))
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return &proto.ErrorResult{errString}, nil
}

func (s *pushTokensServer) GetTokens(ctx context.Context, in *proto.GetTokensRequest) (*proto.GetTokensResult, error) {
	tokens, err := s.repo.GetTokens(uint(in.UserId))
	if err != nil {
		return &proto.GetTokensResult{Error: err.Error()}, nil
	}
	ret := proto.GetTokensResult{
		Tokens: make([]*proto.TokenInfo, 0),
	}
	for _, t := range tokens {
		ret.Tokens = append(ret.Tokens, t.Encode())
	}
	return &ret, nil
}

func (s *pushTokensServer) InvalidateTokens(_ context.Context, in *proto.InvalidateTokensRequest) (*proto.InvalidateTokensResult, error) {
	s.repo.InvalidateTokens(in.Type, in.Tokens)
	return &proto.InvalidateTokensResult{}, nil
}

func (s *pushTokensServer) UpdateToken(_ context.Context, in *proto.UpdateTokenRequest) (*proto.UpdateTokenResult, error) {
	s.repo.UpdateToken(in.Type, in.OldToken, in.NewToken)
	return &proto.UpdateTokenResult{}, nil
}

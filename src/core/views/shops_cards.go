package views

import (
	"core/api"
	"core/db"
	"core/models"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
)

type shopCardServer struct {
	repo models.CardRepository
}

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterShopCardServiceServer(s, shopCardServer{
			repo: models.CardRepositoryImpl{DB: db.New()},
		})
	})
}

func (s shopCardServer) GetCards(_ context.Context, req *core.GetCardsRequest) (*core.GetCardsReply, error) {

	cards, err := models.GetCardsFor(s.repo, uint(req.UserId), uint(req.ShopId))
	if err != nil {
		return &core.GetCardsReply{}, err
	}

	return &core.GetCardsReply{
		Cards: models.ShopCards(cards).Hide().Encode(),
	}, nil
}

func (s shopCardServer) CreateCard(_ context.Context, req *core.CreateCardRequest) (*core.CreateCardReply, error) {

	id, err := models.CreateCard(
		s.repo,
		models.ShopCard{}.Decode(req.Card),
	)

	return &core.CreateCardReply{
		Id: uint64(id),
	}, err
}

func (s shopCardServer) DeleteCard(_ context.Context, req *core.DeleteCardRequest) (*core.DeleteCardReply, error) {

	err := models.DeleteCard(
		s.repo,
		uint(req.UserId),
		uint(req.Id),
	)

	return &core.DeleteCardReply{}, err
}

func (s shopCardServer) GetCardByID(_ context.Context, req *core.GetCardByIDRequest) (*core.GetCardReply, error) {

	card, err := models.GetCardByID(
		s.repo,
		uint(req.UserId),
		uint(req.Id),
	)

	if err != nil {
		return nil, err
	}

	return &core.GetCardReply{
		Card: card.Encode(),
	}, nil
}

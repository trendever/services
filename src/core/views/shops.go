package views

import (
	"core/api"
	"core/db"
	"core/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/log"
)

type shopServer struct {
}

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterShopServiceServer(s, shopServer{})
	})
}

// getItemStubs returns an empty stubs of []models.ProductItem filled with ids.
//  only ID field is filled in, so it's suitable for gorm.DB.Save()
func getItemStubs(ids []int64) []models.ProductItem {

	out := make([]models.ProductItem, len(ids))

	for i, id := range ids {
		out[i] = models.ProductItem{Model: gorm.Model{ID: uint(id)}}
	}

	return out
}

func (s shopServer) ReadShop(ctx context.Context, request *core.ReadShopRequest) (*core.ReadShopReply, error) {
	shop := models.Shop{}

	switch {
	case request.GetInstagramId() > 0:
		shop.InstagramID = request.GetInstagramId()
	}

	query := db.New().Where(&shop)

	if request.WithDeleted {
		query = query.Unscoped().Order("deleted_at desc").Limit(1)
	}

	query = query.Find(&shop)

	if query.Error != nil && !query.RecordNotFound() {
		return &core.ReadShopReply{}, query.Error
	}

	return &core.ReadShopReply{
		Id:        int64(shop.ID),
		IsDeleted: shop.DeletedAt != nil,
	}, nil
}

func (s shopServer) CreateShop(ctx context.Context, request *core.CreateShopRequest) (*core.CreateShopReply, error) {

	shop := decodeShop(request.Shop)

	err := models.CreateNewShop(&shop)

	return &core.CreateShopReply{
		Id: int64(shop.ID),
	}, err
}

func decodeShop(s *core.Shop) models.Shop {

	// @CHECK: why was that necessary?
	if s == nil {
		log.Error(fmt.Errorf("Got nil ptr in decodeShop()"))
		return models.Shop{}
	}

	return models.Shop{
		Model: gorm.Model{
			ID: uint(s.Id),
		},

		InstagramID:        s.InstagramId,
		InstagramUsername:  s.InstagramUsername,
		InstagramFullname:  s.InstagramFullname,
		InstagramAvatarURL: s.InstagramAvatarUrl,
		InstagramCaption:   s.InstagramCaption,
		InstagramWebsite:   s.InstagramWebsite,
		SupplierID:         uint(s.SupplierId),
	}
}

// returns a public profile of the shop
func (s shopServer) GetShopProfile(_ context.Context, req *core.ShopProfileRequest) (reply *core.ShopProfileReply, err error) {
	reply = &core.ShopProfileReply{}
	var shop *models.Shop

	switch {
	case req.GetId() > 0:
		shop, err = models.GetShopByID(uint(req.GetId()))
	case req.GetInstagramName() != "":
		shop, err = models.GetShopByInstagramName(req.GetInstagramName())
	}

	if err != nil {
		return
	}

	reply.Shop = shop.Encode()

	return
}

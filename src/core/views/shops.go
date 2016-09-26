package views

import (
	"core/api"
	"core/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/db"
)

type shopServer struct {
}

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterShopServiceServer(s, shopServer{})
	})
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

func (s shopServer) FindOrCreateShopForSupplier(
	_ context.Context, in *core.FindOrCreateShopForSupplierRequest,
) (reply *core.FindOrCreateShopForSupplierReply, _ error) {
	reply = &core.FindOrCreateShopForSupplierReply{}

	supplier := models.User{Model: gorm.Model{ID: uint(in.SupplierId)}}
	err := db.New().First(supplier).Error
	if err != nil {
		return &core.FindOrCreateShopForSupplierReply{
			Error: fmt.Sprintf("failed to load supplier: %v", err),
		}, nil
	}

	shopID, deleted, err := models.FindOrCreateShopForSupplier(&supplier, in.RecreateDeleted)
	return &core.FindOrCreateShopForSupplierReply{
		ShopId:  shopID,
		Deleted: deleted,
		Error:   err.Error(),
	}, nil
}

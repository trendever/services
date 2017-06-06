package wantit

import (
	"proto/bot"
	"proto/core"
	"utils/log"
	"utils/rpc"
	"wantit/api"
)

func saveProduct(mention *bot.Activity) (id int64, retry bool, err error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Activity: %v", mention)
	log.Debug("Saving unknown product (activityId=%v)", mention.Id)

	res, err := api.SaveTrendClient.SaveProduct(ctx, mention)
	if err != nil {
		return -1, true, err
	}
	return res.Id, res.Retry, nil
}

func saveHelpProduct(mention *bot.Activity, code string, supplierId int64) (id int64, retry bool, err error) {
	shopID, err := shopID(uint64(supplierId))
	if err == errorShopIsDeleted {
		// ignore deleted shops
		return -1, false, err
	} else if err != nil {
		return -1, true, err
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	log.Debug("Creating producto from wantit! Shop=%v User=%v", shopID, supplierId)

	request := &core.CreateProductRequest{Product: &core.Product{
		Code:             code,
		SupplierId:       int64(shopID),
		MentionedId:      int64(supplierId),
		InstagramImageId: "0",
		Title:            "Help",
	}}

	res, err := api.ProductClient.CreateProduct(ctx, request)
	if err != nil {
		return -1, true, err
	}
	return res.Id, false, nil
}

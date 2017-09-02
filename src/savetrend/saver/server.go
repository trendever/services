package saver

import (
	"golang.org/x/net/context"
	"proto/bot"
	"utils/log"
)

// Implementation of SaveTrendService
type SaveServer struct{}

func NewSaveServer() *SaveServer {
	return &SaveServer{}
}

func (*SaveServer) SaveProduct(cxt context.Context, mention *bot.Activity) (*bot.SaveProductResult, error) {
	log.Debug("processing activity %+v from rpc request")
	id, retry, err := processProductMedia(mention.MediaId, mention)
	if retry {
		log.Debug("SaveProduct: temporarily unable to save: %v", err)
		return &bot.SaveProductResult{-1, true}, nil
	}
	if err != nil && err != errorAlreadyAdded {
		log.Debug("SaveProduct: unable to save: %v", err)
		return &bot.SaveProductResult{-1, false}, nil
	}
	return &bot.SaveProductResult{id, false}, nil
}

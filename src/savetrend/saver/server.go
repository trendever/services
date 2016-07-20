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
	for {
		id, retry, err := processProductMedia(mention.MediaId, mention)
		if retry {
			log.Debug("SaveProduct: retying after error: %v", err)
			continue
		}
		if err != nil && err != errorAlreadyAdded {
			log.Debug("SaveProduct: unable to save: %v", err)
			return &bot.SaveProductResult{-1}, nil
		}
		return &bot.SaveProductResult{id}, nil
	}
}

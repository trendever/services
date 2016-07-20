package saver

import (
	"golang.org/x/net/context"
	"proto/bot"
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
			continue
		}
		if err != nil && err != errorAlreadyAdded {
			return &bot.SaveProductResult{-1}, nil
		}
		return &bot.SaveProductResult{id}, nil
	}
}

package cmd

import (
	"common/db"
	"common/log"
	"payments/config"
	"payments/views"

	"utils/nats"
)

func (s *Service) Start() {
	log.Info("Starting payment service on %q", config.Get().RPC)
	config.Init()
	//	api.Init()
	db.Init(&config.Get().DB)
	nats.Init(&config.Get().Nats, true)
	views.Init()
}

func (s *Service) Cleanup() {

}

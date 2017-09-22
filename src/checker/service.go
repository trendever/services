package main

import (
	"accountstore/client"
	"fmt"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"os/signal"
	"proto/accountstore"
	"proto/checker"
	"syscall"
	"time"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/mandible"
	"utils/nats"
	"utils/rpc"
)

var settings struct {
	Debug       bool
	RPC         string
	MandibleURL string

	LastCheckedFile string

	MinimalTickLen  string
	RequestsPerTick uint64

	DB db.Settings

	StoreAddr       string
	client.Settings `mapstructure:",squash"`
	ResponseLogging bool

	Nats      nats.Config
	SentryDSN string
}

var (
	ImageUploader *mandible.Uploader
	storeCli      accountstore.AccountStoreServiceClient
	pool          *client.AccountsPool
)

func init() {
	err := config.LoadStruct("checker", &settings)
	log.Init(settings.Debug, "checker", settings.SentryDSN)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load config: %v", err))
	}
}

// @TODO use accountstore client
func main() {
	var cmd = cobra.Command{
		Use:   "service",
		Short: "instagram checker service",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Migration stub",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Migration stub(nothing to do)")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts service",
		Run:   RunService,
	})
	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}

func RunService(cmd *cobra.Command, args []string) {
	log.Info("Starting service...")
	rand.Seed(time.Now().Unix())

	ImageUploader = mandible.New(settings.MandibleURL)
	nats.Init(&settings.Nats, true)

	storeCli = accountstore.NewAccountStoreServiceClient(rpc.Connect(settings.StoreAddr))

	var err error
	pool, err = client.InitPoll(
		accountstore.Role_AuxPrivate, storeCli,
		nil, nil,
		&settings.Settings,
		settings.ResponseLogging,
	)
	if err != nil {
		log.Fatalf("failed to init acoounts pool: %v", err)
	}

	rpc := rpc.Serve(settings.RPC)
	db.Init(&settings.DB)
	server := NewCheckerServer()
	log.Info("Registering server...")
	checker.RegisterCheckerServiceServer(rpc, server)

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-interrupt
	rpc.Stop()
	server.Stop()
}

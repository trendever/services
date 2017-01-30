package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"instagram"
	"math/rand"
	"os"
	"os/signal"
	"proto/checker"
	"syscall"
	"time"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/mandible"
	"utils/rpc"
)

var settings struct {
	Debug       bool
	RPC         string
	MandibleURL string

	LastCheckedFile string

	DB db.Settings

	RequestsPerTick uint64
	MinimalTickLen  uint64

	Pool            instagram.PoolSettings
	ResponseLogging bool

	Users []struct {
		Name string
		Pass string
	}
	SentryDSN string
}

var (
	ImageUploader *mandible.Uploader
	Instagram     *instagram.Pool
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
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting service...")

			initInstagramPool()
			ImageUploader = mandible.New(settings.MandibleURL)
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
		},
	})
	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}

func initInstagramPool() {
	rand.Seed(time.Now().Unix())
	instagram.DoResponseLogging = settings.ResponseLogging
	Instagram = instagram.NewPool(&settings.Pool)
	if len(settings.Users) == 0 {
		log.Fatal(errors.New("Instagram users are undefined"))
	}
	for {
		activeCount := 0
		for _, user := range settings.Users {
			item, err := instagram.NewInstagram(user.Name, user.Pass, nil)
			if err != nil {
				log.Errorf("failed to add instagram user %v: %v", user.Name, err)
				continue
			}
			activeCount++
			Instagram.Add(item)
		}
		if activeCount != 0 {
			return
		}
		log.Error(errors.New("we don't have any active instagram accaounts"))
		time.Sleep(5 * time.Second)
	}
}

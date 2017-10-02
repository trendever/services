package cli

import (
	"common/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

// Service defines microservice entrypoints
type Service interface {
	Start()
	Migrate(drop bool)
	Cleanup()
}

// Main registers service and runs its functions
func Main(svc Service) {

	var cmd = cobra.Command{
		Use:   "service",
		Short: "Microservice binary",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting service...")
			svc.Start()
			interrupt := make(chan os.Signal)
			signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

			<-interrupt
			log.Info("Stopping service...")
			svc.Cleanup()
		},
	})

	var drop bool
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Runs database migration",
		Run: func(cmd *cobra.Command, args []string) {
			if drop {
				log.Warn("Starting database cleanup...")
			} else {
				log.Warn("Starting database migration...")
			}
			svc.Migrate(drop)
			log.Info("Migration done")
		},
	}
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	cmd.AddCommand(migrateCmd)

	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}

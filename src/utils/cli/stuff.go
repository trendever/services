package cli

import (
	"os"
	"os/signal"
	"syscall"
	"utils/log"
)

//Terminate waits os.Signal and exit
func Terminate(cb func(os.Signal)) {
	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// wait for terminating
	for {
		select {
		case s := <-interrupt:
			if cb != nil {
				cb(s)
			}
			log.Info("Cleanup and terminating...")
			os.Exit(0)
		}
	}
}

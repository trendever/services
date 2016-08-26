package elastic

import (
	"errors"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"time"
	"utils/log"
)

var cli *elastic.Client

type Settings struct {
	Addr string
	// basic request and es cluster status log
	Debug bool
	// full http request/answer log
	Trace bool
}

type eDebugLogger struct{}

func (*eDebugLogger) Printf(format string, values ...interface{}) {
	log.Debug(format, values...)
}

type eErrorLogger struct{}

func (*eErrorLogger) Printf(format string, values ...interface{}) {
	log.Error(fmt.Errorf(format, values...))
}

func Init(settings *Settings) {
	if cli != nil {
		log.Warn("Repeated call to elastic.Init()")
		return

	}
	opts := []elastic.ClientOptionFunc{
		elastic.SetErrorLog(&eErrorLogger{}),
	}
	if settings.Addr != "" {
		opts = append(opts, elastic.SetURL(settings.Addr))
	}
	if settings.Debug {
		opts = append(opts, elastic.SetInfoLog(&eDebugLogger{}))
	}
	if settings.Trace {
		opts = append(opts, elastic.SetTraceLog(&eDebugLogger{}))
	}

	var err error
	for {
		cli, err = elastic.NewClient(opts...)
		if err == nil {
			break
		}
		log.Warn("Failed to connect to elastic node: %v! Retrying in 1 second", err)
		<-time.After(time.Second)
	}
}

func Cli() *elastic.Client {
	if cli == nil {
		log.Fatal(errors.New("elastic: Cli() was called before Init() finished"))
	}
	return cli
}

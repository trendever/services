package elastic

import (
	"errors"
	"gopkg.in/olivere/elastic.v3"
	"time"
	"utils/log"
)

var cli *elastic.Client

type Settings struct {
	Addr  string
	Debug bool
}

type eLogger struct{}

func (*eLogger) Printf(format string, values ...interface{}) {
	log.Debug(format, values...)
}

func Init(settings *Settings) {
	if cli != nil {
		log.Warn("Repeated call to elastic.Init()")
		return

	}
	opts := []elastic.ClientOptionFunc{}
	if settings.Addr != "" {
		opts = append(opts, elastic.SetURL(settings.Addr))
	}
	if settings.Debug {
		opts = append(opts, elastic.SetInfoLog(&eLogger{}))
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

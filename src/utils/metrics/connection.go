package metrics

import (
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"utils/log"
)

var c client.Client
var db string
var points = make(chan *client.Point, 100)
var batches = make(chan client.BatchPoints, 10)
var wantEat = make(chan bool, 2)

//Init initializes influxdb client
func Init(addr, username, password, dbname string) {
	if addr == "" {
		return
	}
	db = dbname
	var err error
	c, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	go fillBatch()
	go batchFlusher()
	log.Fatal(err)
}

//Add adds metric to influxdb.
func Add(name string, tags map[string]string, fields map[string]interface{}) {
	if c == nil {
		return
	}

	pt, err := client.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		log.Error(err)
		return
	}
	points <- pt

}
func newBatch() client.BatchPoints {
	batch, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: db,
	})
	log.Error(err)
	return batch
}

func fillBatch() {
	batch := newBatch()
	for {
		select {
		case pt := <-points:
			batch.AddPoint(pt)
			if len(batch.Points()) >= 100 {
				batches <- batch
				batch = newBatch()
			}
		case b := <-batches:
			flushBatch(b)
		case <-wantEat:
			if len(batch.Points()) > 0 {
				batches <- batch
				batch = newBatch()
			}
		}

	}
}

func batchFlusher() {
	for {
		select {
		case batch := <-batches:
			flushBatch(batch)
		case <-time.After(time.Second):
			wantEat <- true
		}
	}
}

func flushBatch(batch client.BatchPoints) {
	log.Error(c.Write(batch))
}

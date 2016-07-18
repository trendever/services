package publish

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/worker"
)

type workerJobLogger struct {
	job worker.QorJobInterface
}

func (job workerJobLogger) Print(results ...interface{}) {
	job.job.AddLog(fmt.Sprint(results...))
}

type QorWorkerArgument struct {
	IDs []string
	worker.Schedule
}

func (publish *Publish) SetWorker(w *worker.Worker) {
	publish.WorkerScheduler = w
	publish.registerWorkerJob()
}

func (publish *Publish) registerWorkerJob() {
	if w := publish.WorkerScheduler; w != nil {
		if w.Admin == nil {
			fmt.Println("Need to add worker to admin first before set worker")
			return
		}

		qorWorkerArgumentResource := w.Admin.NewResource(&QorWorkerArgument{})
		qorWorkerArgumentResource.Meta(&admin.Meta{Name: "IDs", Type: "publish_job_argument", Valuer: func(record interface{}, context *qor.Context) interface{} {
			var values = map[*admin.Resource][]string{}

			if workerArgument, ok := record.(*QorWorkerArgument); ok {
				for _, id := range workerArgument.IDs {
					if keys := strings.Split(id, "__"); len(keys) == 2 {
						name, id := keys[0], keys[1]
						recordRes := w.Admin.GetResource(name)
						values[recordRes] = append(values[recordRes], id)
					}
				}
			}

			return values
		}})

		w.RegisterJob(&worker.Job{
			Name:  "Publish",
			Group: "Publish",
			Handler: func(argument interface{}, job worker.QorJobInterface) error {
				if argu, ok := argument.(*QorWorkerArgument); ok {
					var records = []interface{}{}
					var values = map[string][]string{}

					for _, id := range argu.IDs {
						if keys := strings.Split(id, "__"); len(keys) == 2 {
							name, id := keys[0], keys[1]
							values[name] = append(values[name], id)
						}
					}

					draftDB := publish.DraftDB().Unscoped()
					for name, value := range values {
						recordRes := w.Admin.GetResource(name)
						results := recordRes.NewSlice()
						if draftDB.Find(results, fmt.Sprintf("%v IN (?)", recordRes.PrimaryDBName()), value).Error == nil {
							resultValues := reflect.Indirect(reflect.ValueOf(results))
							for i := 0; i < resultValues.Len(); i++ {
								records = append(records, resultValues.Index(i).Interface())
							}
						}
					}

					publish.Logger(&workerJobLogger{job: job}).Publish(records...)
				}
				return nil
			},
			Resource: qorWorkerArgumentResource,
		})

		w.RegisterJob(&worker.Job{
			Name:  "DiscardPublish",
			Group: "Publish",
			Handler: func(argument interface{}, job worker.QorJobInterface) error {
				if argu, ok := argument.(*QorWorkerArgument); ok {
					var records = []interface{}{}
					var values = map[string][]string{}

					for _, id := range argu.IDs {
						if keys := strings.Split(id, "__"); len(keys) == 2 {
							name, id := keys[0], keys[1]
							values[name] = append(values[name], id)
						}
					}

					draftDB := publish.DraftDB().Unscoped()
					for name, value := range values {
						recordRes := w.Admin.GetResource(name)
						results := recordRes.NewSlice()
						if draftDB.Find(results, fmt.Sprintf("%v IN (?)", recordRes.PrimaryDBName()), value).Error == nil {
							resultValues := reflect.Indirect(reflect.ValueOf(results))
							for i := 0; i < resultValues.Len(); i++ {
								records = append(records, resultValues.Index(i).Interface())
							}
						}
					}

					publish.Logger(&workerJobLogger{job: job}).Discard(records...)
				}
				return nil
			},
			Resource: qorWorkerArgumentResource,
		})
	}
}

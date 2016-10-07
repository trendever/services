package worker

import (
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

// Job is a struct that hold Qor Job definations
type Job struct {
	Name       string
	Group      string
	Handler    func(interface{}, QorJobInterface) error
	Permission *roles.Permission
	Queue      Queue
	Resource   *admin.Resource
	Worker     *Worker
}

// NewStruct initialize job struct
func (job *Job) NewStruct() interface{} {
	qorJobInterface := job.Worker.JobResource.NewStruct().(QorJobInterface)
	qorJobInterface.SetJob(job)
	return qorJobInterface
}

// GetQueue get defined job's queue
func (job *Job) GetQueue() Queue {
	if job.Queue != nil {
		return job.Queue
	}
	return job.Worker.Queue
}

func (job Job) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if job.Permission == nil {
		return true
	}
	return job.Permission.HasPermission(mode, context.Roles...)
}

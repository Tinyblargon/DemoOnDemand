package backends

import (
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
)

type Task struct {
	ID     string
	Job    *job.Job
	Error  error
	Status *taskstatus.Status
}

var Main Backend

type Backend interface {
	Add(payload *job.Job, executionTimeout time.Duration) (taskID string)
	MoveToWorkQeueu(taskID string) (err error)
	GetTaskStatus(taskID string) []byte
	ListAllTasks() (tasks []string)
}

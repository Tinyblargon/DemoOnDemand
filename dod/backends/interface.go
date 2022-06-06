package backends

import (
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
)

type Task struct {
	ID     string
	Job    *job.Job
	Status *taskstatus.Status
	UserID string
}

var Main Backend

type Backend interface {
	Add(payload *job.Job, executionTimeout time.Duration, userID string) (taskID string)
	MoveToWorkQeueu(taskID string) (err error)
	GetTaskStatus(taskID string) []byte
	ListAllTasks() (tasks []*Task)
}

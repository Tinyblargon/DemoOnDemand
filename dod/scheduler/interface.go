package scheduler

import (
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
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
	GetTaskStatus(taskID string) (info []byte, userID string)
	ListAllTasks() (tasks []*Task)
}

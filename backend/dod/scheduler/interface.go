package scheduler

import (
	"time"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/scheduler/job"
)

type Task struct {
	ID     uint               `json:"id"`
	Job    *job.Job           `json:"-"`
	Status *taskstatus.Status `json:"-"`
	Info   *Info              `json:"info"`
}

type Info struct {
	UserID string  `json:"user"`
	Status *string `json:"status"`
	Time   *Time   `json:"time"`
}

type Time struct {
	Start int64 `json:"start"`
	End   int64 `json:"end,omitempty"`
}

var Main Backend

type Backend interface {
	Add(payload *job.Job, executionTimeout time.Duration, userID string) (taskID uint)
	MoveToWorkQueue(taskID uint) (err error)
	GetTaskStatus(taskID uint) (info *[]*taskstatus.Output, userID string)
	ListAllTasks() (tasks []*Task)
}

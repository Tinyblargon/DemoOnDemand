package backends

import "time"

type Task struct {
	ID      string
	Payload []byte
	Error   error
	Status  []byte
}

type Backend interface {
	Add(payload []byte, executionTimeout time.Duration) (taskID string)
	MoveToWorkQeueu(taskID string) (err error)
	GetTaskStatus(taskID string) []byte
	ListAllTasks() (tasks []string)
}

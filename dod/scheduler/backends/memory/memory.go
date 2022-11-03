package memory

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/backends/memory/demolock"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
)

const Concurrency uint = 5

type Queue struct {
	Tasks *[]*scheduler.Task
	Mutex sync.Mutex
}

type Memory struct {
	Wait         *Queue
	Work         *Queue
	Done         *Queue
	InputChannel chan *scheduler.Task
	DemoLock     *demolock.DemoLock

	taskIDCounter uint64
}

func New(concurrency uint) (memory *Memory) {
	workInputChannel := make(chan *scheduler.Task)
	tasks := make([]*scheduler.Task, 50)
	waitQueue := &Queue{}
	workQueue := &Queue{}
	doneQueue := &Queue{
		Tasks: &tasks,
	}
	demoLock := &demolock.DemoLock{}
	memory = &Memory{
		Wait:         waitQueue,
		Work:         workQueue,
		Done:         doneQueue,
		InputChannel: workInputChannel,
		DemoLock:     demoLock,
	}
	go memory.watchdogWaitQueue()
	memory.spawnWorkers(concurrency)
	return
}

// Adds task to queue (wait)
func (m *Memory) Add(payload *job.Job, executionTimeout time.Duration, userID string) (taskID string) {
	id := atomic.AddUint64(&m.taskIDCounter, 1)
	taskID = strconv.FormatUint(id, 10)
	status := taskstatus.NewStatus()
	task := &scheduler.Task{
		ID:     taskID,
		Job:    payload,
		Status: status,
		UserID: userID,
	}
	addTaskToQueue(m.Wait, task)
	return taskID
}

// Move Item from the Wait to the Work queue
func (m *Memory) MoveToWorkQueue(taskID string) (err error) {
	moveTaskToQueue(m.Wait, m.Work, taskID)
	return
}

func (m *Memory) moveToDoneQueue(taskID string) {
	task := removeTaskFromQueue(m.Work, taskID)
	tmpTasks := make([]*scheduler.Task, global.TaskHistoryDepth)
	tmpTasks[0] = task
	loopLimit := int(global.TaskHistoryDepth) - 1
	m.Done.Mutex.Lock()
	for i, e := range *m.Done.Tasks {
		if i < loopLimit {
			tmpTasks[i+1] = e
		}
	}
	m.Done.Tasks = &tmpTasks
	m.Done.Mutex.Unlock()
}

func (m *Memory) GetTaskStatus(taskID string) (info []byte, userID string) {
	task := getTaskFromQueue(m.Wait, taskID)
	if task != nil {
		return task.Status.Info, task.UserID
	}
	task = getTaskFromQueue(m.Work, taskID)
	if task != nil {
		return task.Status.Info, task.UserID
	}
	task = getTaskFromQueue(m.Done, taskID)
	if task != nil {
		return task.Status.Info, task.UserID
	}
	return nil, ""
}

func (m *Memory) ListAllTasks() (tasks []*scheduler.Task) {
	tasks = listTasksFromQueue(m.Done)
	if m.Work.Tasks != nil {
		tasks = append(tasks, (*m.Work.Tasks)[:]...)
	}
	if m.Wait.Tasks != nil {
		tasks = append(tasks, (*m.Wait.Tasks)[:]...)
	}
	return
}

func (m *Memory) watchdogWaitQueue() {
	for {
		if m.Wait.Tasks != nil {
			m.Wait.Mutex.Lock()
			var task *scheduler.Task
			if m.Wait.Tasks != nil {
				task = unsafeRemoveFirstItemOfQueue(m.Wait)
			}
			m.Wait.Mutex.Unlock()
			if task != nil {
				addTaskToQueue(m.Work, task)
				m.InputChannel <- task
			}
		}
		time.Sleep(time.Microsecond)
	}
}

func moveTaskToQueue(from, to *Queue, taskID string) {
	movedTask := removeTaskFromQueue(from, taskID)
	addTaskToQueue(to, movedTask)
}

func checkTaskExistence(queue *Queue, taskID string) bool {
	for _, e := range *queue.Tasks {
		if e.ID == taskID {
			return true
		}
	}
	return false
}

func listTasksFromQueue(queue *Queue) (tasks []*scheduler.Task) {
	tasks = make([]*scheduler.Task, 0)
	if queue.Tasks != nil {
		for _, e := range *queue.Tasks {
			if e != nil {
				tasks = append(tasks, e)
			}
		}
	}
	return
}

func getTaskFromQueue(queue *Queue, taskID string) (task *scheduler.Task) {
	if queue.Tasks != nil {
		for _, e := range *queue.Tasks {
			if e != nil {
				if e.ID == taskID {
					return e
				}
			}
		}
	}
	return nil
}

func addTaskToQueue(queue *Queue, task *scheduler.Task) {
	var newToTask []*scheduler.Task
	queue.Mutex.Lock()
	if queue.Tasks == nil {
		newToTask = make([]*scheduler.Task, 1)
		newToTask[0] = task
	} else {
		newToTask = append(*queue.Tasks, task)
	}
	queue.Tasks = &newToTask
	queue.Mutex.Unlock()
}

func removeTaskFromQueue(queue *Queue, taskID string) (movedTask *scheduler.Task) {
	var counter uint
	queue.Mutex.Lock()
	numberOfTasks := len(*queue.Tasks)
	if numberOfTasks > 1 {
		tmpTasks := make([]*scheduler.Task, numberOfTasks-1)
		for _, e := range *queue.Tasks {
			if e.ID != taskID {
				tmpTasks[counter] = e
				counter++
			} else {
				movedTask = e
			}
		}
		queue.Tasks = &tmpTasks
	} else {
		movedTask = (*queue.Tasks)[0]
		queue.Tasks = nil
	}
	queue.Mutex.Unlock()
	return
}

// this function does not Lock and Unlock, the function calling this is responsible for Lock and Unlock
func unsafeRemoveFirstItemOfQueue(queue *Queue) (task *scheduler.Task) {
	numberOfTasks := len(*queue.Tasks)
	task = (*queue.Tasks)[0]
	if numberOfTasks > 1 {
		tmpTasks := make([]*scheduler.Task, numberOfTasks-1)
		for i, e := range (*queue.Tasks)[1:] {
			tmpTasks[i] = e
		}
		queue.Tasks = &tmpTasks
	} else {
		queue.Tasks = nil
	}
	return
}

func (m *Memory) spawnWorkers(concurrency uint) {
	if concurrency == 0 {
		concurrency = 1
	}
	for x := 0; x < int(concurrency); x++ {
		go m.worker()
	}
}

func (m *Memory) worker() {
	for e := range m.InputChannel {
		// This unsafe function is possible because this is the only thread using this variable
		e.Status.UnsafeSetStarted()
		// e.Job.Execute will spawn more threads
		e.Job.Execute(e.Status, m.DemoLock)
		m.moveToDoneQueue(e.ID)
	}
}

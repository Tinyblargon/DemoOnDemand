package memory

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends"
	"github.com/Tinyblargon/DemoOnDemand/dod/backends/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/backends/memory/demolock"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
)

const Concurency uint = 5

var AddedToQueue = []byte("Task added to queue.")
var TaskStarted = []byte("Task Started.")

type Queue struct {
	Tasks *[]*backends.Task
	Mutex sync.Mutex
}

type Memory struct {
	Wait         *Queue
	Work         *Queue
	Done         *Queue
	InputChannel chan *backends.Task
	DemoLock     *demolock.DemoLock

	taskIDCounter uint64
}

func New(concurrency uint) (memory *Memory) {
	workInputChannel := make(chan *backends.Task)
	tasks := make([]*backends.Task, 50)
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
func (m *Memory) Add(payload *job.Job, executionTimeout time.Duration) (taskID string) {
	id := atomic.AddUint64(&m.taskIDCounter, 1)
	taskID = strconv.FormatUint(id, 10)
	status := &taskstatus.Status{
		Info: AddedToQueue,
	}
	task := &backends.Task{
		ID:     taskID,
		Job:    payload,
		Status: status,
	}
	addTaskToQueue(m.Wait, task)
	return taskID
}

// Move Item from the Wait to the Work queue
func (m *Memory) MoveToWorkQeueu(taskID string) (err error) {
	moveTaskToQueue(m.Wait, m.Work, taskID)
	return
}

func (m *Memory) moveToDoneQeueu(taskID string) {
	task := removeTaskFromQueue(m.Work, taskID)
	tmpTasks := make([]*backends.Task, global.TaskHistoryDepth)
	tmpTasks[0] = task
	looplimit := int(global.TaskHistoryDepth) - 1
	m.Done.Mutex.Lock()
	for i, e := range *m.Done.Tasks {
		if i < looplimit {
			tmpTasks[i+1] = e
		}
	}
	m.Done.Tasks = &tmpTasks
	m.Done.Mutex.Unlock()
}

func (m *Memory) GetTaskStatus(taskID string) []byte {
	task := getTaskFromQueue(m.Wait, taskID)
	if task != nil {
		return task.Status.Info
	}
	task = getTaskFromQueue(m.Work, taskID)
	if task != nil {
		return task.Status.Info
	}
	task = getTaskFromQueue(m.Done, taskID)
	if task != nil {
		return task.Status.Info
	}
	return nil
}

func (m *Memory) ListAllTasks() (tasks []string) {
	tasks = make([]string, 0)
	doneTasks := listTasksFromQueue(m.Done)
	tasks = append(tasks, doneTasks[:]...)
	workTasks := listTasksFromQueue(m.Work)
	tasks = append(tasks, workTasks[:]...)
	waitTasks := listTasksFromQueue(m.Wait)
	return append(tasks, waitTasks[:]...)
}

func (m *Memory) watchdogWaitQueue() {
	for {
		if m.Wait.Tasks != nil {
			m.Wait.Mutex.Lock()
			var task *backends.Task
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

func checkTaskExistance(queue *Queue, taskID string) bool {
	for _, e := range *queue.Tasks {
		if e.ID == taskID {
			return true
		}
	}
	return false
}

func listTasksFromQueue(queue *Queue) (tasks []string) {
	tasks = make([]string, 0)
	if queue.Tasks != nil {
		for _, e := range *queue.Tasks {
			if e != nil {
				tasks = append(tasks, e.ID)
			}
		}
	}
	return
}

func getTaskFromQueue(queue *Queue, taskID string) (task *backends.Task) {
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

func addTaskToQueue(queue *Queue, task *backends.Task) {
	var newToTask []*backends.Task
	queue.Mutex.Lock()
	if queue.Tasks == nil {
		newToTask = make([]*backends.Task, 1)
		newToTask[0] = task
	} else {
		newToTask = append(*queue.Tasks, task)
	}
	queue.Tasks = &newToTask
	queue.Mutex.Unlock()
}

func removeTaskFromQueue(queue *Queue, taskID string) (movedTask *backends.Task) {
	var counter uint
	queue.Mutex.Lock()
	numberOftasks := len(*queue.Tasks)
	if numberOftasks > 1 {
		tmpTasks := make([]*backends.Task, numberOftasks-1)
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
func unsafeRemoveFirstItemOfQueue(queue *Queue) (task *backends.Task) {
	numberOftasks := len(*queue.Tasks)
	task = (*queue.Tasks)[0]
	if numberOftasks > 1 {
		tmpTasks := make([]*backends.Task, numberOftasks-1)
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
		e.Status.Info = TaskStarted
		e.Job.Execute(e.Status, m.DemoLock)
		m.moveToDoneQeueu(e.ID)
	}
}

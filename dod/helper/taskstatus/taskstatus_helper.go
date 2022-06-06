package taskstatus

import "sync"

type Status struct {
	Info   []byte
	Status string
	Mutex  sync.Mutex
}

func (s *Status) AddError(err error) {
	s.Mutex.Lock()
	s.unsafeAddToInfo(err.Error())
	s.unsafeAddToInfo("Task Failed!")
	s.unsafeSetStatus("error")
	s.Mutex.Unlock()
}

func (s *Status) UnsafeSetStarted() {
	s.Info = []byte("Task Started.")
	s.unsafeSetStatus("started")
}

func (s *Status) AddCompleted() {
	s.Mutex.Lock()
	s.unsafeAddToInfo("OK")
	s.unsafeSetStatus("ok")
	s.Mutex.Unlock()
}

func NewStatus() (status *Status) {
	return &Status{
		Info:   []byte("Task Added to Queue."),
		Status: "queued",
	}
}

func (s *Status) AddToInfo(newLine string) {
	s.Mutex.Lock()
	s.unsafeAddToInfo(newLine)
	s.Mutex.Unlock()
}

func (s *Status) unsafeAddToInfo(newLine string) {
	s.Info = append(s.Info, []byte("\n"+newLine)...)
}

func (s *Status) unsafeSetStatus(statusCode string) {
	s.Status = statusCode
}

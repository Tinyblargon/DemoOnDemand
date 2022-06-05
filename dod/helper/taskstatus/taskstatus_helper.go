package taskstatus

import "sync"

type Status struct {
	Info  []byte
	Mutex sync.Mutex
}

func (s *Status) AddError(err error) {
	s.AddToStatus(err.Error())
	s.AddToStatus("Task Failed!")
}

func (s *Status) AddToStatus(newLine string) {
	s.Mutex.Lock()
	s.Info = append(s.Info, []byte("\n"+newLine)...)
	s.Mutex.Unlock()
}

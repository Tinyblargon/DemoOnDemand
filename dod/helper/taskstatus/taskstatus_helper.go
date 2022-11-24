package taskstatus

import "sync"

const prefixError string = "[ERROR] "
const prefixInfo string = "[INFO] "
const prefixWarning string = "[WARN] "
const prefixSuccess string = "[SUCCESS] "

type Status struct {
	Info   []byte
	Status string
	Mutex  sync.Mutex
}

func (s *Status) AddError(err error) {
	s.Mutex.Lock()
	s.unsafeAddToInfo(prefixError, err.Error())
	s.unsafeAddToInfo(prefixError, "Task Failed!")
	s.unsafeSetStatus("error")
	s.Mutex.Unlock()
}

func (s *Status) UnsafeSetStarted() {
	s.Info = []byte(prefixInfo + "Task Started.")
	s.unsafeSetStatus("started")
}

func (s *Status) AddCompleted() {
	s.Mutex.Lock()
	s.unsafeAddToInfo(prefixSuccess, "OK")
	s.unsafeSetStatus("ok")
	s.Mutex.Unlock()
}

func NewStatus() (status *Status) {
	return &Status{
		Info:   []byte(prefixInfo + "Task Added to Queue."),
		Status: "queued",
	}
}

func (s *Status) AddToInfo(text string) {
	if s != nil {
		s.Mutex.Lock()
		s.unsafeAddToInfo(prefixInfo, text)
		s.Mutex.Unlock()
	}
}

func (s *Status) AddWarning(text string) {
	if s != nil {
		s.Mutex.Lock()
		s.unsafeAddToInfo(prefixWarning, text)
		s.Mutex.Unlock()
	}
}

func (s *Status) unsafeAddToInfo(prefix, newLine string) {
	s.Info = append(s.Info, []byte("\n"+prefix+newLine)...)
}

func (s *Status) unsafeSetStatus(statusCode string) {
	s.Status = statusCode
}

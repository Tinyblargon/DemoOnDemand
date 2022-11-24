package taskstatus

import "sync"

const prefixError string = "[ERROR] "
const prefixInfo string = "[INFO] "
const prefixWarning string = "[WARN] "
const prefixSuccess string = "[SUCCESS] "

type Status struct {
	Output *[]*Output
	Status string
	Mutex  sync.Mutex
}

type Output struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

func (s *Status) AddError(err error) {
	s.Mutex.Lock()
	s.unsafeAddToInfo(prefixError, err.Error())
	s.unsafeAddToInfo(prefixError, "Task Failed!")
	s.unsafeSetStatus("error")
	s.Mutex.Unlock()
}

func (s *Status) UnsafeSetStarted() {
	output := make([]*Output, 1)
	output[0] = &Output{
		Kind: prefixInfo,
		Text: "Task Started.",
	}
	s.Output = &output
	s.unsafeSetStatus("started")
}

func (s *Status) AddCompleted() {
	s.Mutex.Lock()
	s.unsafeAddToInfo(prefixSuccess, "OK")
	s.unsafeSetStatus("ok")
	s.Mutex.Unlock()
}

func NewStatus() (status *Status) {
	output := make([]*Output, 1)
	output[0] = &Output{
		Kind: prefixInfo,
		Text: "Task Added to Queue.",
	}
	return &Status{
		Output: &output,
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

func (s *Status) unsafeAddToInfo(prefix, text string) {
	*(s.Output) = append(*s.Output, &Output{
		Kind: prefix,
		Text: text,
	})
}

func (s *Status) unsafeSetStatus(statusCode string) {
	s.Status = statusCode
}

// Converts the Output struct into text with line breaks
func ToString(output *[]*Output) (text string) {
	for _, e := range *output {
		text = text + e.Kind + " " + e.Text + "\n"
	}
	return
}

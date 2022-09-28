package concurrency

import (
	"fmt"
	"sync"
	"time"
)

var threads uint

type Object struct {
	Mutex           sync.Mutex
	Err             error
	Cycles          uint
	CompletedCycles uint
	Threads         uint
}

func Initialize(Threads uint) error {
	if threads > 0 {
		return fmt.Errorf("threads can only be set once")
	}
	if Threads == 0 {
		return fmt.Errorf("threads per task may not be 0")
	}
	threads = Threads
	return nil
}

func Threads() uint {
	return threads
}

func New(numberOfObjects, requestedThreads uint) *Object {
	threads := decideMinimumTreads(numberOfObjects, requestedThreads)
	return &Object{
		Cycles:  numberOfObjects,
		Threads: threads,
	}
}

func decideMinimumTreads(numberOfObjects, concurrency uint) uint {
	if concurrency == 0 {
		concurrency = 1
	} else if numberOfObjects < concurrency {
		concurrency = numberOfObjects
	}
	return concurrency
}

func (o *Object) ChannelLooperError() error {
	for {
		time.Sleep(time.Microsecond * 100)
		if o.Cycles == o.CompletedCycles || o.Err != nil {
			break
		}
	}
	return o.Err
}

func (o *Object) Cycle(err error) {
	if err != nil {
		o.Mutex.Lock()
		if o.Err != nil {
			o.Err = err
		}
	} else {
		o.Mutex.Lock()
	}
	o.CompletedCycles++
	o.Mutex.Unlock()
}

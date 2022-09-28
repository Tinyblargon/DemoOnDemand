package concurrency

import (
	"sync"
	"time"
)

type Object struct {
	Mutex           sync.Mutex
	Err             error
	Cycles          uint
	CompletedCycles uint
	Threads         uint
}

func Initialize(numberOfObjects, requestedThreads uint) *Object {
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
	o.Mutex.Lock()
	if o.Err != nil {
		o.Err = err
	}
	o.CompletedCycles++
	o.Mutex.Unlock()
}

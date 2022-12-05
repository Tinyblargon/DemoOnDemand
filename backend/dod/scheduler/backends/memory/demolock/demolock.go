package demolock

import (
	"sync"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
)

type DemoLock struct {
	DemoID []string
	Mutex  sync.Mutex
}

func (d *DemoLock) Lock(ID string, status *taskstatus.Status) {
	var notFirstRun bool
	for {
		d.Mutex.Lock()
		if d.DemoID != nil {
			if util.IsStringUnique(&d.DemoID, ID) {
				d.DemoID = append(d.DemoID, ID)
				d.Mutex.Unlock()
				break
			}
		} else {
			d.DemoID = make([]string, 1)
			d.DemoID[0] = ID
			d.Mutex.Unlock()
			break
		}
		d.Mutex.Unlock()
		if !notFirstRun {
			status.AddToInfo("Waiting: Trying to get Lock")
		}
		time.Sleep(time.Millisecond)
		notFirstRun = true
	}
}

func (d *DemoLock) Unlock(ID string) {
	var count uint
	d.Mutex.Lock()
	length := len(d.DemoID)
	if length > 1 {
		tmpDemoID := make([]string, length-1)
		for _, e := range d.DemoID {
			if e != ID {
				tmpDemoID[count] = e
				count++
			}
		}
		d.DemoID = tmpDemoID
	} else {
		d.DemoID = nil
	}
	d.Mutex.Unlock()
}

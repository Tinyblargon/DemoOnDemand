package demolock

import (
	"sync"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
)

type DemoLock struct {
	DemoID []string
	Mutex  sync.Mutex
}

func (d *DemoLock) Lock(ID string, status *taskstatus.Status) {
	var notFirstRun bool
	for {
		IdExists := false
		d.Mutex.Lock()
		if d.DemoID != nil {
			for _, e := range d.DemoID {
				if e == ID {
					IdExists = true
					break
				}
				d.DemoID = append(d.DemoID, ID)
			}
		} else {
			d.DemoID = make([]string, 1)
			d.DemoID[0] = ID
		}
		d.Mutex.Unlock()
		if !IdExists {
			break
		}
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
	} else {
		d.DemoID = nil
	}
	d.Mutex.Unlock()
}

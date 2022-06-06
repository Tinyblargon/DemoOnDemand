package job

import (
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends/memory/demolock"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
)

type Job struct {
	Demo *Demo
}

type Demo struct {
	Template string `json:"template"`
	UserName string `json:"username"`
	Number   uint   `json:"number"`
	Create   bool
	Start    bool
	Stop     bool
	Destroy  bool
}

func (j *Job) Execute(status *taskstatus.Status, demoLock *demolock.DemoLock) {
	if j.Demo != nil {
		ID := j.Demo.UserName + "_" + j.Demo.Template + "_" + strconv.Itoa(int(j.Demo.Number))
		demoLock.Lock(ID, status)
		c, err := newSession(status, global.VMwareConfig)
		if err != nil {
			return
		}
		if j.Demo.Create {
			err = demo.New(c.VimClient, global.DB, global.VMwareConfig.DataCenter, j.Demo.Template, j.Demo.UserName, global.VMwareConfig.Pool, j.Demo.Number, 5, status)
		}
		if j.Demo.Destroy {
			err = demo.Delete(c.VimClient, global.DB, global.VMwareConfig.DataCenter, j.Demo.Template, j.Demo.UserName, j.Demo.Number, status)
		}
		if j.Demo.Stop {
			err = demo.Stop(c.VimClient, global.DB, global.VMwareConfig.DataCenter, j.Demo.Template, j.Demo.UserName, j.Demo.Number, status)
			if err != nil {
				status.AddError(err)
				demoLock.Unlock(ID)
				return
			}
		}
		if j.Demo.Start {
			err = demo.Start(c.VimClient, global.DB, global.VMwareConfig.DataCenter, j.Demo.Template, j.Demo.UserName, j.Demo.Number, status)
		}
		demoLock.Unlock(ID)
		if err != nil {
			status.AddError(err)
			return
		}
	}
	status.AddCompleted()
}

func newSession(status *taskstatus.Status, config *programconfig.VMwareConfiguration) (c *session.Client, err error) {
	c, err = session.New(*config)
	if err != nil {
		status.AddError(err)
	}
	return
}

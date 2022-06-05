package job

import (
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
	Restart  bool
	Destroy  bool
}

func (j *Job) Execute(status *taskstatus.Status) {
	if j.Demo != nil {
		if j.Demo.Create {
			c, err := newSession(status, global.VMwareConfig)
			if err != nil {
				return
			}
			err = demo.New(c.VimClient, global.DB, global.VMwareConfig.DataCenter, j.Demo.Template, j.Demo.UserName, global.VMwareConfig.Pool, j.Demo.Number, 5, status)
			if err != nil {
				status.AddError(err)
				return
			}
		}
		if j.Demo.Destroy {
			c, err := newSession(status, global.VMwareConfig)
			if err != nil {
				return
			}
			err = demo.Delete(c.VimClient, global.DB, global.VMwareConfig.DataCenter, j.Demo.Template, j.Demo.UserName, j.Demo.Number, status)
			if err != nil {
				status.AddError(err)
				return
			}
		}
	}
	status.AddToStatus("OK")
}

func newSession(status *taskstatus.Status, config *programconfig.VMwareConfiguration) (c *session.Client, err error) {
	c, err = session.New(*config)
	if err != nil {
		status.AddError(err)
	}
	return
}

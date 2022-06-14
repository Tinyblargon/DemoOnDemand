package job

import (
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/backends/memory/demolock"
)

type Job struct {
	Demo     *Demo
	Template *Template
}

type Template struct {
	Config       template.Config
	Import       bool
	Destroy      bool
	ChildDestroy bool
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
	var err error
	var c *session.Client
	if j.Demo != nil {
		ID := createID(j.Demo.UserName, j.Demo.Template, j.Demo.Number)
		demoLock.Lock(ID, status)
		c, err = newSession(status, global.VMwareConfig)
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
	}
	if j.Template != nil {
		c, err = newSession(status, global.VMwareConfig)
		if err != nil {
			return
		}
		if j.Template.Import {
			err = j.Template.Config.Import(c.VimClient, global.VMwareConfig.DataCenter, global.VMwareConfig.Pool, status)
		}
		if j.Template.Destroy {
			err = template.Destroy(c.VimClient, global.VMwareConfig.DataCenter, j.Template.Config.Name, status)
		}
		if j.Template.ChildDestroy {
			err = deleteTemplateChilds(status, demoLock, c, j.Template.Config.Name)
		}
	}
	if err != nil {
		status.AddError(err)
		return
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

func deleteTemplateChilds(status *taskstatus.Status, demoLock *demolock.DemoLock, c *session.Client, templateName string) (err error) {
	demos, err := database.ListDemosOfTemplate(global.DB, templateName)
	if err != nil {
		return
	}
	for _, e := range *demos {
		ID := createID(e.UserName, e.DemoName, e.DemoNumber)
		demoLock.Lock(ID, status)
		err = demo.Delete(c.VimClient, global.DB, global.VMwareConfig.DataCenter, e.DemoName, e.UserName, e.DemoNumber, status)
		demoLock.Unlock(ID)
		if err != nil {
			return
		}
	}
	return
}

func createID(userName, template string, number uint) string {
	return userName + "_" + template + "_" + strconv.Itoa(int(number))
}

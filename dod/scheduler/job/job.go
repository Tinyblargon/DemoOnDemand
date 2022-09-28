package job

import (
	demoactions "github.com/Tinyblargon/DemoOnDemand/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/backends/memory/demolock"
	"github.com/Tinyblargon/DemoOnDemand/dod/template"
)

type Job struct {
	Demo     *Demo
	Template *Template
}

type Template struct {
	Config       *template.Config
	Import       bool
	Destroy      bool
	ChildDestroy bool
}

type Demo struct {
	Template string `json:"template"`
	UserName string `json:"username,omitempty"`
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
		demoObj := demo.Demo{
			Name: j.Demo.Template,
			User: j.Demo.UserName,
			ID:   j.Demo.Number,
		}
		ID := demoObj.CreateID()
		demoLock.Lock(ID, status)
		c, err = newSession(status, global.VMwareConfig)
		if err != nil {
			return
		}
		if j.Demo.Create {
			err = demoactions.New(c.VimClient, global.DB, datacenter.GetObject(), global.VMwareConfig.Pool, &demoObj, 5, status)
		}
		if j.Demo.Destroy {
			err = demoactions.Delete(c.VimClient, global.DB, datacenter.GetObject(), &demoObj, status)
		}
		if j.Demo.Stop {
			err = demoactions.Stop(c.VimClient, global.DB, datacenter.GetObject(), &demoObj, status)
			if err != nil {
				status.AddError(err)
				demoLock.Unlock(ID)
				return
			}
		}
		if j.Demo.Start {
			err = demoactions.Start(c.VimClient, global.DB, datacenter.GetObject(), &demoObj, status)
		}
		demoLock.Unlock(ID)
	}
	if j.Template != nil {
		c, err = newSession(status, global.VMwareConfig)
		if err != nil {
			return
		}
		if j.Template.Import {
			err = j.Template.Config.Import(c.VimClient, datacenter.GetObject(), global.VMwareConfig.Pool, status)
		}
		if j.Template.Destroy {
			err = template.Destroy(c.VimClient, datacenter.GetObject(), j.Template.Config.Name, status)
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
		demoObj := demo.Demo{
			Name: e.DemoName,
			User: e.UserName,
			ID:   e.DemoNumber,
		}
		ID := demoObj.CreateID()
		demoLock.Lock(ID, status)
		err = demoactions.Delete(c.VimClient, global.DB, datacenter.GetObject(), &demoObj, status)
		demoLock.Unlock(ID)
		if err != nil {
			return
		}
	}
	return
}

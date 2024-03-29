package job

import (
	"context"

	demoactions "github.com/Tinyblargon/DemoOnDemand/backend/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/session"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/scheduler/backends/memory/demolock"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/template"
	"github.com/vmware/govmomi/object"
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
	Number   uint   `json:"number,omitempty"` //only empty during creation
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
		c, err = newSession(status, vsphere.GetConfig())
		ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
		defer cancel()
		defer c.VimClient.Logout(ctx)
		if err != nil {
			return
		}
		var dataCenter *object.Datacenter
		dataCenter, err = datacenter.Get(c.VimClient, datacenter.GetName())
		if err != nil {
			return
		}
		if j.Demo.Create {
			err = demoactions.New(c.VimClient, global.DB, dataCenter, vsphere.GetConfig().Pool, &demoObj, 5, status)
		}
		if j.Demo.Destroy {
			err = demoactions.Delete(c.VimClient, global.DB, dataCenter, &demoObj, status)
		}
		if j.Demo.Stop {
			err = demoactions.Stop(c.VimClient, global.DB, dataCenter, &demoObj, status)
			if err != nil {
				status.AddError(err)
				demoLock.Unlock(ID)
				return
			}
		}
		if j.Demo.Start {
			err = demoactions.Start(c.VimClient, global.DB, dataCenter, &demoObj, status)
		}
		demoLock.Unlock(ID)
	}
	if j.Template != nil {
		c, err = newSession(status, vsphere.GetConfig())
		ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
		defer cancel()
		defer c.VimClient.Logout(ctx)
		if err != nil {
			return
		}
		var dataCenter *object.Datacenter
		dataCenter, err = datacenter.Get(c.VimClient, datacenter.GetName())
		if err != nil {
			return
		}
		if j.Template.Import {
			err = j.Template.Config.Import(c.VimClient, dataCenter, vsphere.GetConfig().Pool, status)
		}
		if j.Template.Destroy {
			err = template.Destroy(c.VimClient, dataCenter, j.Template.Config.Name, status)
		}
		if j.Template.ChildDestroy {
			err = deleteTemplateChildren(c, dataCenter, j.Template.Config.Name, demoLock, status)
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

func deleteTemplateChildren(c *session.Client, dataCenter *object.Datacenter, templateName string, demoLock *demolock.DemoLock, status *taskstatus.Status) (err error) {
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
		err = demoactions.Delete(c.VimClient, global.DB, dataCenter, &demoObj, status)
		demoLock.Unlock(ID)
		if err != nil {
			return
		}
	}
	return
}

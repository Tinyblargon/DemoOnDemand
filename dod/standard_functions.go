package dod

import (
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

func Intialize(client *govmomi.Client, dc *object.Datacenter) (err error) {
	_, err = folder.Create(client, dc, global.TemplateFodler)
	if err != nil {
		return
	}
	_, err = folder.Create(client, dc, global.RouterFodler)
	if err != nil {
		return
	}
	_, err = folder.Create(client, dc, global.DemoFodler)
	return
}

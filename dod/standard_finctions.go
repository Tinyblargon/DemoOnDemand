package dod

import (
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/folder"
	"github.com/vmware/govmomi"
)

func Intialize(client *govmomi.Client, dataCenter string) (err error) {
	_, err = folder.Create(client, dataCenter, global.TemplateFodler)
	if err != nil {
		return
	}
	_, err = folder.Create(client, dataCenter, global.RouterFodler)
	if err != nil {
		return
	}
	_, err = folder.Create(client, dataCenter, global.DemoFodler)
	return
}

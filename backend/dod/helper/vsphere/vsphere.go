package vsphere

import (
	"context"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/folder"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/host"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/session"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

var globalConfig programconfig.VMwareConfiguration

func Initialize(config *programconfig.VMwareConfiguration, vlanPrefix string) (err error) {
	globalConfig = *config
	err = provider.Initialize(globalConfig.APITimeout)
	if err != nil {
		return
	}
	c, err := session.New(globalConfig)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	defer c.VimClient.Logout(ctx)
	err = datacenter.Initialize(c.VimClient, globalConfig.DataCenter)
	if err != nil {
		return
	}
	dataCenter, err := datacenter.Get(c.VimClient, datacenter.GetName())
	if err != nil {
		return
	}
	err = host.Initialize(c.VimClient, dataCenter, globalConfig.Hosts)
	if err != nil {
		return
	}
	return setupFolderStructure(c.VimClient, dataCenter, vlanPrefix)
}

func GetConfig() *programconfig.VMwareConfiguration {
	configCopy := globalConfig
	return &configCopy
}

func setupFolderStructure(c *govmomi.Client, dc *object.Datacenter, vlanPrefix string) (err error) {
	_, err = folder.Create(c, dc, folder.VSphereFolderTypeVM, global.TemplateFolder)
	if err != nil {
		return
	}
	_, err = folder.Create(c, dc, folder.VSphereFolderTypeVM, global.RouterFolder)
	if err != nil {
		return
	}
	_, err = folder.Create(c, dc, folder.VSphereFolderTypeVM, global.DemoFolder)
	return
}

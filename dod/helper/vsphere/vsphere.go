package vsphere

import (
	"context"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/host"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/session"
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
	err = datacenter.Initialize(c.VimClient, globalConfig.DataCenter)
	if err != nil {
		return
	}
	err = host.Initialize(c.VimClient, datacenter.GetObject(), globalConfig.Hosts)
	if err != nil {
		return
	}
	err = setupFolderStructure(c.VimClient, datacenter.GetObject(), vlanPrefix)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	err = c.VimClient.Logout(ctx)
	return
}

func GetConfig() *programconfig.VMwareConfiguration {
	configCopy := globalConfig
	return &configCopy
}

func setupFolderStructure(c *govmomi.Client, dc *object.Datacenter, vlanPrefix string) (err error) {
	_, err = folder.Create(c, dc, folder.VSphereFolderTypeVM, global.TemplateFodler)
	if err != nil {
		return
	}
	_, err = folder.Create(c, dc, folder.VSphereFolderTypeVM, global.RouterFodler)
	if err != nil {
		return
	}
	_, err = folder.Create(c, dc, folder.VSphereFolderTypeVM, global.DemoFodler)
	if err != nil {
		return
	}
	_, err = folder.CreateSingleFolder(c, dc, folder.VSphereFolderTypeNetwork, vlanPrefix)
	return
}

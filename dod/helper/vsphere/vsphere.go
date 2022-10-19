package vsphere

import (
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
	// Never close this session!
	// or it will make the Datacenter and host objects invalid.
	c, err := session.New(globalConfig)
	if err != nil {
		return
	}
	// TODO these items can get incalidated get thembefore every call
	err = datacenter.Initialize(c.VimClient, globalConfig.DataCenter)
	if err != nil {
		return
	}
	// TODO these items can get incalidated get thembefore every call
	err = host.Initialize(c.VimClient, datacenter.GetObject(), globalConfig.Hosts)
	if err != nil {
		return
	}
	return setupFolderStructure(c.VimClient, datacenter.GetObject(), vlanPrefix)
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
	return
}

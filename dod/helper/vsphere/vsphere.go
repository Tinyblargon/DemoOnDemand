package vsphere

import (
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/host"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/session"
)

var globalConfig programconfig.VMwareConfiguration

func Initialize(config *programconfig.VMwareConfiguration) (err error) {
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
	return
}

func GetConfig() *programconfig.VMwareConfiguration {
	configCopy := globalConfig
	return &configCopy
}

package virtualhost

import (
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/vmware/govmomi/vim25/mo"
)

func GetInterfaceSettings(vmProperties *mo.VirtualMachine, template *[]template.Network) {
	net := virtualmachine.GetMac(vmProperties)
	_ = net
	return
}

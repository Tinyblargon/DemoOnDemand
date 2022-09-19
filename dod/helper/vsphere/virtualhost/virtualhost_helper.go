package virtualhost

import (
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/vmware/govmomi/vim25/mo"
)

func GetInterfaceSettings(vmProperties *mo.VirtualMachine, vlans []*vlan.LocalList) []*vlan.LocalList {
	return virtualmachine.GetMac(vmProperties, vlans)
}

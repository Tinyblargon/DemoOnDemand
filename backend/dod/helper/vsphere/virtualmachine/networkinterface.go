package virtualmachine

import (
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vlan"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func GetMac(vmProperties *mo.VirtualMachine, networkList []*vlan.LocalList) []*vlan.LocalList {
	networkInterfaces := getNetworkInterfaces(vmProperties)
	for _, e := range *networkInterfaces {
		for _, ee := range networkList {
			if e.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().DeviceInfo.GetDescription().Summary == (*ee.BackingInfo).(types.BaseVirtualDeviceDeviceBackingInfo).GetVirtualDeviceDeviceBackingInfo().DeviceName {
				if ee.Mac == "" {
					ee.Mac = getMac(e)
					break
				}
			}
		}
	}
	return networkList
}

// Get the MAC address of the first network interface
func GetFirstMac(vmProperties *mo.VirtualMachine) string {
	networkInterfaces := getNetworkInterfaces(vmProperties)
	return getMac((*networkInterfaces)[0])
}

func AddNetworkInterface(vmProperties *mo.VirtualMachine, spec *types.VirtualMachineCloneSpec, backing *types.BaseVirtualDeviceBackingInfo) (*types.VirtualMachineCloneSpec, error) {
	spec = addVmSpec(spec)
	devices := object.VirtualDeviceList(vmProperties.Config.Hardware.Device)

	device, err := devices.CreateEthernetCard("vmxnet3", *backing)
	if err != nil {
		return nil, err
	}

	dSpec, err := object.VirtualDeviceList{device}.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}
	spec.Config.DeviceChange = append(spec.Config.DeviceChange, dSpec...)
	return spec, err
}

func ChangeNetworkInterface(vmProperties *mo.VirtualMachine, spec *types.VirtualMachineCloneSpec, networks []*vlan.LocalList, status *taskstatus.Status) *types.VirtualMachineCloneSpec {
	spec = addVmSpec(spec)
	networkInterfaces := ReadNetworkInterfaces(object.VirtualDeviceList(vmProperties.Config.Hardware.Device), status)

	baseVDevices := []types.BaseVirtualDeviceConfigSpec{}
	for _, e := range *networkInterfaces {
		e = staticMac(e)
		e = changeConnectedNetwork(e, networks)
		baseVDevices = append(baseVDevices, &types.VirtualDeviceConfigSpec{
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
			Device:    e,
		})
	}
	spec.Config.DeviceChange = baseVDevices
	return spec
}

func getNetworkInterfaces(vmProperties *mo.VirtualMachine) *object.VirtualDeviceList {
	return ReadNetworkInterfaces(object.VirtualDeviceList(vmProperties.Config.Hardware.Device), nil)
}

func getMac(nic types.BaseVirtualDevice) string {
	return nic.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().MacAddress
}

// converts the mac address of the network adapter to a static address
func staticMac(baseVDevice types.BaseVirtualDevice) types.BaseVirtualDevice {
	baseVDevice.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().AddressType = "manual"
	return baseVDevice
}

// Changes the network the network interface is connected to
func changeConnectedNetwork(baseVDevice types.BaseVirtualDevice, networks []*vlan.LocalList) types.BaseVirtualDevice {
	for _, e := range networks {
		if e.OriginalNetwork == baseVDevice.GetVirtualDevice().DeviceInfo.GetDescription().Summary {
			baseVDevice.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().Backing = *(e.BackingInfo)
			break
		}
	}
	return baseVDevice
}

// Code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"

// ReadNetworkInterfaces returns a list of network interfaces. This is used
// in the VM data source to discover the properties of the network interfaces on the
// virtual machine. The list is sorted by the order that they would be added in
// if a clone were to be done.
func ReadNetworkInterfaces(l object.VirtualDeviceList, status *taskstatus.Status) *object.VirtualDeviceList {
	status.AddToInfo("ReadNetworkInterfaces: Fetching network interfaces")
	devices := l.Select(func(device types.BaseVirtualDevice) bool {
		if _, ok := device.(types.BaseVirtualEthernetCard); ok {
			return true
		}
		return false
	})
	return &devices
}

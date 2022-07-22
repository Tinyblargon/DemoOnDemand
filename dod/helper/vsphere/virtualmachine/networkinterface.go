package virtualmachine

import (
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func AddNetworkInterface(vmProperties *mo.VirtualMachine, spec *types.VirtualMachineCloneSpec, backing *types.BaseVirtualDeviceBackingInfo) (*types.VirtualMachineCloneSpec, error) {
	spec = addVmSpec(spec)
	devices := object.VirtualDeviceList(vmProperties.Config.Hardware.Device)

	device, err := devices.CreateEthernetCard("vmxnet3", *backing)
	if err != nil {
		return nil, err
	}

	dspec, err := object.VirtualDeviceList{device}.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}
	spec.Config.DeviceChange = append(spec.Config.DeviceChange, dspec...)
	return spec, err
}

func ChangeNetworkInterface(vmProperties *mo.VirtualMachine, spec *types.VirtualMachineCloneSpec, networks *vlan.LocalList, status *taskstatus.Status) *types.VirtualMachineCloneSpec {
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

// converts the mac address of the network adapter to a static address
func staticMac(baseVDevice types.BaseVirtualDevice) types.BaseVirtualDevice {
	baseVDevice.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().AddressType = "manual"
	return baseVDevice
}

// Changes the network the network interface is connected to
func changeConnectedNetwork(baseVDevice types.BaseVirtualDevice, networks *vlan.LocalList) types.BaseVirtualDevice {
	if networks != nil {
		for i, e := range *networks.Original {
			if e.Name == baseVDevice.GetVirtualDevice().DeviceInfo.GetDescription().Summary {
				baseVDevice.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().Backing = *(*(networks.Remapped))[i]
				break
			}
		}
	}
	return baseVDevice
}

// Code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"

const (
	networkInterfaceSubresourceTypeE1000   = "e1000"
	networkInterfaceSubresourceTypeE1000e  = "e1000e"
	networkInterfaceSubresourceTypePCNet32 = "pcnet32"
	networkInterfaceSubresourceTypeSriov   = "sriov"
	networkInterfaceSubresourceTypeVmxnet2 = "vmxnet2"
	networkInterfaceSubresourceTypeVmxnet3 = "vmxnet3"
	networkInterfaceSubresourceTypeUnknown = "unknown"
)

var networkInterfaceSubresourceTypeAllowedValues = []string{
	networkInterfaceSubresourceTypeE1000,
	networkInterfaceSubresourceTypeE1000e,
	networkInterfaceSubresourceTypeVmxnet3,
}

// // virtualDeviceListSorter is an internal type to facilitate sorting of a BaseVirtualDeviceList.
// type virtualDeviceListSorter struct {
// 	Sort       object.VirtualDeviceList
// 	DeviceList object.VirtualDeviceList
// }

// // Len implements sort.Interface for virtualDeviceListSorter.
// func (l virtualDeviceListSorter) Len() int {
// 	return len(l.Sort)
// }

// // Less helps implement sort.Interface for virtualDeviceListSorter. A
// // BaseVirtualDevice is "less" than another device if its controller's bus
// // number and unit number combination are earlier in the order than the other.
// func (l virtualDeviceListSorter) Less(i, j int) bool {
// 	li := l.Sort[i]
// 	lj := l.Sort[j]
// 	liCtlr := l.DeviceList.FindByKey(li.GetVirtualDevice().ControllerKey)
// 	ljCtlr := l.DeviceList.FindByKey(lj.GetVirtualDevice().ControllerKey)
// 	if liCtlr == nil || ljCtlr == nil {
// 		panic(errors.New("virtualDeviceListSorter cannot be used with devices that are not assigned to a controller"))
// 	}
// 	liCtlrBus := liCtlr.(types.BaseVirtualController).GetVirtualController().BusNumber
// 	ljCtlrBus := ljCtlr.(types.BaseVirtualController).GetVirtualController().BusNumber
// 	if liCtlrBus != ljCtlrBus {
// 		return liCtlrBus < ljCtlrBus
// 	}
// 	liUnit := li.GetVirtualDevice().UnitNumber
// 	ljUnit := lj.GetVirtualDevice().UnitNumber
// 	if liUnit == nil || ljUnit == nil {
// 		panic(errors.New("virtualDeviceListSorter cannot be used with devices that do not have unit numbers set"))
// 	}
// 	return *liUnit < *ljUnit
// }

// // Swap helps implement sort.Interface for virtualDeviceListSorter.
// func (l virtualDeviceListSorter) Swap(i, j int) {
// 	l.Sort[i], l.Sort[j] = l.Sort[j], l.Sort[i]
// }

// ReadNetworkInterfaces returns a list of network interfaces. This is used
// in the VM data source to discover the properties of the network interfaces on the
// virtual machine. The list is sorted by the order that they would be added in
// if a clone were to be done.
func ReadNetworkInterfaces(l object.VirtualDeviceList, status *taskstatus.Status) *object.VirtualDeviceList {
	if status != nil {
		status.AddToInfo("[DEBUG] ReadNetworkInterfaces: Fetching network interfaces")
	}
	devices := l.Select(func(device types.BaseVirtualDevice) bool {
		if _, ok := device.(types.BaseVirtualEthernetCard); ok {
			return true
		}
		return false
	})
	return &devices
}

// DeviceListString pretty-prints each device in a virtual device list, used
// for logging purposes and what not.
func DeviceListString(l object.VirtualDeviceList) string {
	var names []string
	for _, d := range l {
		if d == nil {
			names = append(names, "<nil>")
		} else {
			names = append(names, l.Name(d))
		}
	}
	return strings.Join(names, ",")
}

// virtualEthernetCardString prints a string representation of the ethernet device passed in.
func virtualEthernetCardString(d types.BaseVirtualEthernetCard) string {
	switch d.(type) {
	case *types.VirtualE1000:
		return networkInterfaceSubresourceTypeE1000
	case *types.VirtualE1000e:
		return networkInterfaceSubresourceTypeE1000e
	case *types.VirtualPCNet32:
		return networkInterfaceSubresourceTypePCNet32
	case *types.VirtualSriovEthernetCard:
		return networkInterfaceSubresourceTypeSriov
	case *types.VirtualVmxnet2:
		return networkInterfaceSubresourceTypeVmxnet2
	case *types.VirtualVmxnet3:
		return networkInterfaceSubresourceTypeVmxnet3
	}
	return networkInterfaceSubresourceTypeUnknown
}

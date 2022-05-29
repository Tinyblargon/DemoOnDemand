package virtualmachine

// Code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"

import (
	"log"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

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
func ReadNetworkInterfaces(l object.VirtualDeviceList) *object.VirtualDeviceList {
	log.Printf("[DEBUG] ReadNetworkInterfaces: Fetching network interfaces")
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

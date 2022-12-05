package virtualmachine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/concurrency"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/generic"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func Properties(vm *object.VirtualMachine, status *taskstatus.Status) (*mo.VirtualMachine, error) {
	status.AddToInfo(fmt.Sprintf("Fetching properties for VM %q", vm.InventoryPath))
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	var props mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

func Get(client *govmomi.Client, dc *object.Datacenter, Path string) (*object.VirtualMachine, error) {
	ctx, cancel, finder, checkPath := generic.NewFinder(client, dc, Path)
	defer cancel()
	vm, err := finder.VirtualMachine(ctx, checkPath)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// Clone wraps the creation of a virtual machine and the subsequent waiting of
// the task. A higher-level virtual machine object is returned.
func Clone(c *govmomi.Client, src *object.VirtualMachine, f *object.Folder, name string, spec types.VirtualMachineCloneSpec, timeout int, status *taskstatus.Status) (*object.VirtualMachine, error) {
	status.AddToInfo(fmt.Sprintf("Cloning virtual machine %q", fmt.Sprintf("%s/%s", f.InventoryPath, name)))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(timeout))
	defer cancel()
	task, err := src.Clone(ctx, f, name, spec)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = errors.New("timeout waiting for clone to complete")
		}
		return nil, err
	}
	result, err := task.WaitForResult(ctx, nil)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = errors.New("timeout waiting for clone to complete")
		}
		return nil, err
	}
	status.AddToInfo(fmt.Sprintf("Virtual machine %q: clone complete (MOID: %q)", fmt.Sprintf("%s/%s", f.InventoryPath, name), result.Result.(types.ManagedObjectReference).Value))
	return FromID(c, result.Result.(types.ManagedObjectReference).Value)
}

// FromID locates a VirtualMachine by its managed object reference ID.
func FromID(client *govmomi.Client, id string) (*object.VirtualMachine, error) {
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "VirtualMachine",
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	vm, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}
	return vm.(*object.VirtualMachine), nil
}

func CreateSnapshots(vmObjects []*object.VirtualMachine, SnapshotName string, memory bool) (err error) {
	for _, e := range vmObjects {
		err = CreateSnapshot(e, SnapshotName, memory)
		if err != nil {
			return
		}
	}
	return
}

func CreateSnapshot(vm *object.VirtualMachine, SnapshotName string, memory bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	task, err := vm.CreateSnapshot(ctx, SnapshotName, "", memory, true)

	if err != nil {
		return fmt.Errorf("cannot create snapshot of virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task, "create snapshot of virtualmachine")
}

func StartObjects(vmObjects []*object.VirtualMachine, concurrency uint, status *taskstatus.Status) (err error) {
	in, conObject := channelInitialize(uint(len(vmObjects)), concurrency)
	// spawn "conObject.Threads" amount of threads
	for x := 0; x < int(conObject.Threads); x++ {
		go func() {
			for x := range in {
				conObject.Cycle(Start(x, status))
				if conObject.Err != nil {
					break
				}
			}
		}()
	}
	err = channelLooper(in, conObject, &vmObjects, uint(len(vmObjects)))
	return
}

func Start(vm *object.VirtualMachine, status *taskstatus.Status) error {
	powerState, err := GetPowerState(vm)
	if err != nil {
		return err
	}
	if powerState == "poweredOn" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()

	status.AddToInfo(fmt.Sprintf("Starting virtual machine %s", vm.InventoryPath))
	task, err := vm.PowerOn(ctx)
	if err != nil {
		return fmt.Errorf("cannot start virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task, "start virtualmachine")
}

func StopObjects(vmObjects []*object.VirtualMachine, concurrency uint, status *taskstatus.Status) (err error) {
	in, conObject := channelInitialize(uint(len(vmObjects)), concurrency)
	// spawn "conObject.Threads" amount of threads
	for x := 0; x < int(conObject.Threads); x++ {
		go func() {
			for x := range in {
				conObject.Cycle(Stop(x, status))
				if conObject.Err != nil {
					break
				}
			}
		}()
	}
	err = channelLooper(in, conObject, &vmObjects, uint(len(vmObjects)))
	return
}

func Stop(vm *object.VirtualMachine, status *taskstatus.Status) error {
	powerState, err := GetPowerState(vm)
	if err != nil {
		return err
	}
	if powerState == "poweredOff" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()

	status.AddToInfo(fmt.Sprintf("Stopping virtual machine %s", vm.InventoryPath))
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return fmt.Errorf("cannot stop virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task, "stop virtualmachine")
}

func Delete(vm *object.VirtualMachine, status *taskstatus.Status) error {
	err := Stop(vm, status)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()

	status.AddToInfo(fmt.Sprintf("Removing virtual machine %s", vm.InventoryPath))
	task, err := vm.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("cannot delete virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task, "delete virtualmachine")
}

func DeleteObjects(vmObjects []*object.VirtualMachine, concurrency uint, status *taskstatus.Status) (err error) {
	in, conObject := channelInitialize(uint(len(vmObjects)), concurrency)
	// spawn "conObject.Threads" amount of threads
	for x := 0; x < int(conObject.Threads); x++ {
		go func() {
			for x := range in {
				conObject.Cycle(Delete(x, status))
				if conObject.Err != nil {
					break
				}
			}
		}()
	}
	err = channelLooper(in, conObject, &vmObjects, uint(len(vmObjects)))
	return
}

// Returns the powerState of the given VirtualMachine object
func GetPowerState(vm *object.VirtualMachine) (types.VirtualMachinePowerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	return vm.PowerState(ctx)
}

func GetGuestIP(client *govmomi.Client, path, name string, networks []string, dc *object.Datacenter, status *taskstatus.Status) (guestIP string, vmProperties *mo.VirtualMachine, err error) {
	status.AddToInfo(fmt.Sprintf("Fetching IP of guest %s", name))
	// try until the guest ip is readable from vmware tools
	for {
		var vmObject *object.VirtualMachine
		vmObject, err = Get(client, dc, path+"/"+name)
		if err != nil {
			return
		}
		vmProperties, err = Properties(vmObject, nil)
		if err != nil {
			return
		}
		if vmProperties.Guest.IpAddress != "" {
			guestIP, err = filterIP(vmProperties, networks)
			if err != nil {
				return
			}
			status.AddToInfo(fmt.Sprintf("Obtained IP (%s) of guest %s", guestIP, vmObject.Name()))
			break
		}
		time.Sleep(time.Second * 2)
	}
	return
}

func addVmSpec(cloneSpec *types.VirtualMachineCloneSpec) *types.VirtualMachineCloneSpec {
	if cloneSpec.Config != nil {
		return cloneSpec
	}
	vmSpec := new(types.VirtualMachineConfigSpec)
	cloneSpec.Config = vmSpec
	return cloneSpec
}

func channelInitialize(numberOfObjects, concurrencyNumber uint) (chan *object.VirtualMachine, *concurrency.Object) {
	in := make(chan *object.VirtualMachine)
	return in, concurrency.New(numberOfObjects, concurrencyNumber)
}

// Loops over the in and ret channels
func channelLooper(in chan *object.VirtualMachine, conObject *concurrency.Object, vmObjects *[]*object.VirtualMachine, cycles uint) error {
	go func() {
		for _, e := range *vmObjects {
			// loop over all items
			in <- e
		}
		close(in)
	}()
	return conObject.ChannelLooperError()
}

// Returns the networks of the vmObject
func GetNetworks(vmObject *object.VirtualMachine, status *taskstatus.Status) (networks []string, err error) {
	networks = make([]string, 0)
	var vmProperties *mo.VirtualMachine
	vmProperties, err = Properties(vmObject, status)
	if err != nil {
		return
	}
	networkInterfaces := ReadNetworkInterfaces(object.VirtualDeviceList(vmProperties.Config.Hardware.Device), status)
	for _, e := range *networkInterfaces {
		networks = append(networks, e.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().VirtualDevice.DeviceInfo.GetDescription().Summary)
	}
	return
}

// gets the ip of the first network that is not in the list
func filterIP(vmProperties *mo.VirtualMachine, networks []string) (string, error) {
	for _, e := range vmProperties.Guest.Net {
		if util.IsStringUnique(&networks, e.Network) {
			if len(e.IpAddress) != 0 {
				return e.IpAddress[0], nil
			} else {
				break
			}
		}
	}
	return "", fmt.Errorf("no valid ip address found")
}

package virtualmachine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/generic"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func Properties(vm *object.VirtualMachine) (*mo.VirtualMachine, error) {
	log.Printf("[DEBUG] Fetching properties for VM %q", vm.InventoryPath)
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	var props mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

func Get(client *govmomi.Client, DataCenter, Path string) (*object.VirtualMachine, error) {
	ctx, cancel, finder, checkPath := generic.NewFinder(client, DataCenter, Path)
	defer cancel()
	vm, err := finder.VirtualMachine(ctx, checkPath)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// Clone wraps the creation of a virtual machine and the subsequent waiting of
// the task. A higher-level virtual machine object is returned.
func Clone(c *govmomi.Client, src *object.VirtualMachine, f *object.Folder, name string, spec types.VirtualMachineCloneSpec, timeout int) (*object.VirtualMachine, error) {

	log.Printf("[DEBUG] Cloning virtual machine %q", fmt.Sprintf("%s/%s", f.InventoryPath, name))
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
	log.Printf("[DEBUG] Virtual machine %q: clone complete (MOID: %q)", fmt.Sprintf("%s/%s", f.InventoryPath, name), result.Result.(types.ManagedObjectReference).Value)
	return FromID(c, result.Result.(types.ManagedObjectReference).Value)
}

// FromID locates a VirtualMachine by its managed object reference ID.
func FromID(client *govmomi.Client, id string) (*object.VirtualMachine, error) {
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "VirtualMachine",
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	task, err := vm.CreateSnapshot(ctx, SnapshotName, "", memory, true)

	if err != nil {
		return fmt.Errorf("cannot create snapshot of virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task)
}

func StartOjects(vmObjects []*object.VirtualMachine) (err error) {
	for _, e := range vmObjects {
		err = Start(e)
		if err != nil {
			return
		}
	}
	return
}

func Start(vm *object.VirtualMachine) error {
	powerState, err := GetPowerState(vm)
	if err != nil {
		return err
	}
	if powerState == "poweredOn" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	log.Printf("[DEBUG] Starting virtual machine %q", fmt.Sprintf("%s", vm.InventoryPath))
	task, err := vm.PowerOn(ctx)
	if err != nil {
		return fmt.Errorf("cannot start virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task)
}

func StopOjects(vmObjects []*object.VirtualMachine) (err error) {
	for _, e := range vmObjects {
		err = Stop(e)
		if err != nil {
			return
		}
	}
	return
}

func Stop(vm *object.VirtualMachine) error {
	powerState, err := GetPowerState(vm)
	if err != nil {
		return err
	}
	if powerState == "poweredOff" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	log.Printf("[DEBUG] Stopping virtual machine %q", fmt.Sprintf("%s", vm.InventoryPath))
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return fmt.Errorf("cannot stop virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task)
}

func Delete(vm *object.VirtualMachine) error {
	err := Stop(vm)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	log.Printf("[DEBUG] Removing virtual machine %q", fmt.Sprintf("%s", vm.InventoryPath))
	task, err := vm.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("cannot delete virtualmachine: %s", err)
	}
	return generic.RunTaskWait(task)
}

func DeleteOjects(vmObjects []*object.VirtualMachine) (err error) {
	for _, e := range vmObjects {
		err = Delete(e)
		if err != nil {
			return
		}
	}
	return
}

func GetPowerState(vm *object.VirtualMachine) (types.VirtualMachinePowerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return vm.PowerState(ctx)
}

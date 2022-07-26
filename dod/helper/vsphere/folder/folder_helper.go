package folder

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/generic"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/clustercomputeresource"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type FileSystemItem struct {
	Name           string
	Subitems       []*FileSystemItem
	Folder         *object.Folder
	VirtualMachine *object.VirtualMachine
}

// Clone Wil clone all items in the speciefied folder and all it's subfolders
func Clone(client *govmomi.Client, dc *object.Datacenter, vlans []*vlan.LocalList, Path, newPath, pool string, vmTemplate bool, status *taskstatus.Status) (err error) {
	fileSystem, err := ReadFileSystem(client, dc, Path)
	if err != nil {
		return
	}
	err = fileSystem.Create(client, dc, vlans, newPath, pool, vmTemplate, status)
	return
}

func ReadFileSystem(client *govmomi.Client, dc *object.Datacenter, Path string) (*FileSystemItem, error) {
	var err error
	fileSystem := new(FileSystemItem)
	fileSystem.Folder, err = Get(client, dc, Path)
	if err != nil {
		return nil, err
	}
	fileSystem.Subitems, err = fileSystem.recursiveRead(client)
	if err != nil {
		return nil, err
	}
	return fileSystem, nil
}

func (fileSystem *FileSystemItem) Create(client *govmomi.Client, dc *object.Datacenter, vlans []*vlan.LocalList, basefolder, pool string, vmTemplate bool, status *taskstatus.Status) (err error) {
	_, err = Create(client, dc, basefolder)
	if err != nil {
		return
	}
	var clusterProp *mo.ClusterComputeResource
	if !vmTemplate {
		clusterProp, err = clustercomputeresource.PropertiesFromPath(client, dc, pool, status)
		if err != nil {
			return
		}
	}
	err = fileSystem.recursiveCreate(client, dc, basefolder, vmTemplate, vlans, clusterProp, status)
	return
}

// this function recursivly calls itself to create all items found in parent at a the location of basefolder
func (parent *FileSystemItem) recursiveCreate(client *govmomi.Client, dc *object.Datacenter, basefolder string, vmTemplate bool, vlans []*vlan.LocalList, clusterProp *mo.ClusterComputeResource, status *taskstatus.Status) (err error) {
	// this can be more parallelized but would require rewriting, look at the DeleteObjects function
	if parent.Subitems == nil {
		return
	}
	vmCounter := 0
	for _, e := range parent.Subitems {
		if e.Folder != nil {
			newBaseFolder := basefolder + "/" + e.Name
			_, err = CreateSingleFolder(client, dc, newBaseFolder)
			if err != nil {
				break
			}
			err = e.recursiveCreate(client, dc, newBaseFolder, vmTemplate, vlans, clusterProp, status)
		}
		if e.VirtualMachine != nil {
			vmCounter += 1
		}
	}
	if vmCounter == 0 {
		return
	}
	var wg sync.WaitGroup
	vmArray := make([]*FileSystemItem, vmCounter)
	vmCounter = 0
	for _, e := range parent.Subitems {
		if e.VirtualMachine != nil {
			vmArray[vmCounter] = e
			vmCounter += 1
		}
	}
	var ob *object.Folder
	ob, err = Get(client, dc, basefolder)
	if err != nil {
		return
	}
	wg.Add(len(vmArray))
	for _, e := range vmArray {
		go func(client *govmomi.Client, e *FileSystemItem, vmTemplate bool, clusterProp *mo.ClusterComputeResource) {

			var properties *mo.VirtualMachine
			properties, err = virtualmachine.Properties(e.VirtualMachine, status)
			if err != nil {
				wg.Done()
				return
			}
			spec := &types.VirtualMachineCloneSpec{}
			virtualmachine.ChangeNetworkInterface(properties, spec, vlans, status)
			spec.Template = vmTemplate
			if clusterProp != nil {
				spec.Location.Pool = &types.ManagedObjectReference{
					Value: clusterProp.ComputeResource.ResourcePool.Value,
					Type:  clusterProp.ComputeResource.ResourcePool.Type,
				}
			}
			_, err = virtualmachine.Clone(client, e.VirtualMachine, ob, e.Name, *spec, 1000, status)
			wg.Done()
		}(client, e, vmTemplate, clusterProp)
	}
	wg.Wait()
	return
}

// ReadFileSystem Wil recursivly get all items in the subfolder
func (parent *FileSystemItem) recursiveRead(client *govmomi.Client) ([]*FileSystemItem, error) {
	ob, err := parent.Folder.Children(context.Background())
	SubItems := len(ob)
	if SubItems == 0 {
		return nil, nil
	}
	array := make([]*FileSystemItem, SubItems)
	for i, e := range ob {
		array[i] = new(FileSystemItem)
		test := e.Reference().Type
		switch test {
		case "Folder":
			subOb, err := FromID(client, e.Reference().Value)
			if err != nil {
				return nil, err
			}
			array[i].Folder = subOb
			array[i].Name = path.Base(subOb.InventoryPath)
			obList, err := array[i].recursiveRead(client)
			array[i].Subitems = obList
		case "VirtualMachine":
			subOb, err := virtualmachine.FromID(client, e.Reference().Value)
			if err != nil {
				return nil, err
			}
			array[i].VirtualMachine = subOb
			array[i].Name = path.Base(subOb.InventoryPath)
		}
	}
	return array, err
}

func (fileSystem *FileSystemItem) GetVmObjects() []*object.VirtualMachine {
	amount := fileSystem.RecursiveCountVmObjects(0)
	virtualMachineList, _ := fileSystem.RecursiveGetVmObjects(make([]*object.VirtualMachine, amount), 0)
	return virtualMachineList
}

// this function recursivly calls itself to count all VirtualMachines in the filesystem
func (fileSystem *FileSystemItem) RecursiveCountVmObjects(NumberOfVms uint) uint {
	if fileSystem.Subitems == nil {
		return NumberOfVms
	}
	for _, e := range fileSystem.Subitems {
		if e.Folder != nil {
			NumberOfVms = e.RecursiveCountVmObjects(NumberOfVms)
		}
		if e.VirtualMachine != nil {
			NumberOfVms += 1
		}
	}
	return NumberOfVms
}

func (fileSystem *FileSystemItem) RecursiveGetVmObjects(vmArray []*object.VirtualMachine, vmCounter int) ([]*object.VirtualMachine, int) {
	if fileSystem.Subitems == nil {
		return vmArray, vmCounter
	}
	for _, e := range fileSystem.Subitems {
		if e.Folder != nil {
			vmArray, vmCounter = e.RecursiveGetVmObjects(vmArray, vmCounter)
		}
		if e.VirtualMachine != nil {
			vmArray[vmCounter] = e.VirtualMachine
			vmCounter += 1
		}
	}
	return vmArray, vmCounter
}

func CreateSnapshot(client *govmomi.Client, dc *object.Datacenter, Path, SnapshotName string, memory bool) error {
	folder, err := Get(client, dc, Path)
	if err != nil {
		return err
	}
	childrenExist, err := HasChildren(folder)
	if err != nil {
		return err
	}
	if childrenExist {
		fileSystem, err := ReadFileSystem(client, dc, Path)
		if err != nil {
			return err
		}
		err = virtualmachine.CreateSnapshots(fileSystem.GetVmObjects(), SnapshotName, memory)
		if err != nil {
			return err
		}
	}
	return nil
}

func Delete(client *govmomi.Client, dc *object.Datacenter, Path string, status *taskstatus.Status) error {

	folder, err := Get(client, dc, Path)
	if err != nil {
		return err
	}

	childrenExist, err := HasChildren(folder)
	if err != nil {
		return err
	}
	if childrenExist {
		fileSystem, err := ReadFileSystem(client, dc, Path)
		if err != nil {
			return err
		}
		err = virtualmachine.DeleteObjects(fileSystem.GetVmObjects(), global.Concurency, status)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	task, err := folder.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("cannot delete folder: %s", err)
	}

	return generic.RunTaskWait(task, "delete folder")
}

// Restarts all the virtualmachines in the folder and subfolders
func ReStart(client *govmomi.Client, dc *object.Datacenter, Path string, status *taskstatus.Status) (err error) {
	vmObjects, err := GetVmObjectsFromPath(client, dc, Path)
	if err != nil {
		return
	}
	if len(vmObjects) != 0 {
		err = virtualmachine.StopObjects(vmObjects, global.Concurency, status)
		if err != nil {
			return
		}
		err = virtualmachine.StartObjects(vmObjects, global.Concurency, status)
	}
	return
}

// Starts all the virtualmachines in the folder and subfolders
func Start(client *govmomi.Client, dc *object.Datacenter, Path string, status *taskstatus.Status) (err error) {
	vmObjects, err := GetVmObjectsFromPath(client, dc, Path)
	if err != nil {
		return
	}
	if len(vmObjects) != 0 {
		err = virtualmachine.StartObjects(vmObjects, global.Concurency, status)
	}
	return
}

// Stops all the virtualmachines in the folder and subfolders
func Stop(client *govmomi.Client, dc *object.Datacenter, Path string, status *taskstatus.Status) (err error) {
	vmObjects, err := GetVmObjectsFromPath(client, dc, Path)
	if err != nil {
		return
	}
	if len(vmObjects) != 0 {
		err = virtualmachine.StopObjects(vmObjects, global.Concurency, status)
	}
	return
}

// CreateFolder Creates the full folder path spaeciefied
func Create(client *govmomi.Client, dc *object.Datacenter, Path string) (folderObject *object.Folder, err error) {
	folders := strings.Split(strings.Trim(Path, "/"), "/")
	var CurrentPath string
	for _, e := range folders {
		CurrentPath += "/" + e
		folderObject, err = CreateSingleFolder(client, dc, CurrentPath)
	}
	return
}

// CreateFolder only creates last subfolder, it fails if the path doesnt exist
func CreateSingleFolder(client *govmomi.Client, dc *object.Datacenter, Path string) (*object.Folder, error) {
	var folderObject *object.Folder
	parent, err := Get(client, dc, path.Dir(Path))
	if err != nil {
		return nil, fmt.Errorf("error trying to determine parent targetFolder: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if !Exists(client, dc, Path) {
		folderObject, err = parent.CreateFolder(ctx, path.Base(Path))
		if err != nil {
			return nil, fmt.Errorf("error creating targetFolder: %s", err)
		}
	}

	return folderObject, nil
}

func Exists(client *govmomi.Client, dc *object.Datacenter, Path string) bool {
	_, err := Get(client, dc, Path)
	if err != nil {
		return false
	}
	return true
}

func ListSubFolders(client *govmomi.Client, dc *object.Datacenter, Path string) (*[]string, error) {
	return ListFolderItems(client, dc, Path, "Folder")
}

func ListVirtualMachinesInFolder(client *govmomi.Client, dc *object.Datacenter, Path string) (*[]string, error) {
	return ListFolderItems(client, dc, Path, "VirtualMachine")
}

func ListFolderItems(client *govmomi.Client, dc *object.Datacenter, Path, Type string) (*[]string, error) {
	parentFolder, err := Get(client, dc, Path)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	children, err := parentFolder.Children(ctx)

	var amountOfTypeItems int
	for _, e := range children {
		if e.Reference().Type == Type {
			amountOfTypeItems += 1
		}
	}

	subItems := make([]string, amountOfTypeItems)

	var counter int
	for _, e := range children {
		if e.Reference().Type == Type {
			switch Type {
			case "Folder":
				ob, err := FromID(client, e.Reference().Value)
				if err != nil {
					return nil, err
				}
				subItems[counter] = path.Base(ob.InventoryPath)
			case "VirtualMachine":
				ob, err := virtualmachine.FromID(client, e.Reference().Value)
				if err != nil {
					return nil, err
				}
				subItems[counter] = path.Base(ob.InventoryPath)
			}
			counter += 1
		}
	}
	return &subItems, nil
}

func HasChildren(f *object.Folder) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	children, err := f.Children(ctx)
	if err != nil {
		return false, fmt.Errorf("error checking for folder contents: %s", err)
	}
	return len(children) > 0, nil
}

// GetFolder returns an *object.Folder from a given absolute path.
// If no such folder is found, an appropriate error will be returned.
func Get(client *govmomi.Client, dc *object.Datacenter, Path string) (*object.Folder, error) {
	ctx, cancel, finder, checkPath := generic.NewFinder(client, dc, Path)
	defer cancel()
	folder, err := finder.Folder(ctx, checkPath)
	if err != nil {
		return nil, fmt.Errorf("cannot locate folder: %s", err)
	}
	return folder, nil
}

func GetChildrenFromPath(client *govmomi.Client, dc *object.Datacenter, Path string) ([]*object.Folder, error) {
	ctx, cancel, finder, checkPath := generic.NewFinder(client, dc, Path)
	defer cancel()
	folder, err := finder.FolderList(ctx, checkPath)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

// FromID locates a Folder by its managed object reference ID.
func FromID(client *govmomi.Client, id string) (*object.Folder, error) {
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "Folder",
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	folder, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}
	return folder.(*object.Folder), nil
}

func GetVmObjectsFromPath(client *govmomi.Client, dc *object.Datacenter, Path string) (vmObjects []*object.VirtualMachine, err error) {
	folder, err := Get(client, dc, Path)
	if err != nil {
		return
	}

	childrenExist, err := HasChildren(folder)
	if err != nil {
		return
	}
	if childrenExist {
		var fileSystem *FileSystemItem
		fileSystem, err = ReadFileSystem(client, dc, Path)
		if err != nil {
			return
		}
		vmObjects = fileSystem.GetVmObjects()
	}
	return
}

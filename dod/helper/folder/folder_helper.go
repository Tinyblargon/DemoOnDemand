package folder

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/generic"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/virtualmachine"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type FileSystemItem struct {
	Name           string
	Subitems       *[]FileSystemItem
	Folder         *object.Folder
	VirtualMachine *object.VirtualMachine
}

// Clone Wil clone all items in the speciefied folder and all it's subfolders
func Clone(client *govmomi.Client, DataCenter, Path, newPath string) (err error) {
	fileSystem, err := ReadFileSystem(client, DataCenter, Path)
	if err != nil {
		return
	}
	err = fileSystem.Create(client, DataCenter, newPath)
	return
}

func ReadFileSystem(client *govmomi.Client, DataCenter, Path string) (*FileSystemItem, error) {
	var err error
	fileSystem := new(FileSystemItem)
	fileSystem.Folder, err = Get(client, DataCenter, Path)
	if err != nil {
		return nil, err
	}
	fileSystem.Subitems, err = fileSystem.RecursiveRead(client)
	if err != nil {
		return nil, err
	}
	return fileSystem, nil
}

func (fileSystem *FileSystemItem) Create(client *govmomi.Client, DataCenter, basefolder string) (err error) {
	_, err = Create(client, DataCenter, basefolder)
	if err != nil {
		return
	}
	err = fileSystem.RecursiveCreate(client, DataCenter, basefolder)
	return
}

// this function recursivly calls itself to create all items found in parent at a the location of basefolder
func (parent *FileSystemItem) RecursiveCreate(client *govmomi.Client, DataCenter, basefolder string) (err error) {
	if parent.Subitems == nil {
		return
	}
	for _, e := range *parent.Subitems {
		if e.Folder != nil {
			newBaseFolder := basefolder + "/" + e.Name
			_, err = CreateSingleFolder(client, DataCenter, newBaseFolder)
			if err != nil {
				return
			}
			err = e.RecursiveCreate(client, DataCenter, newBaseFolder)
		}
		if e.VirtualMachine != nil {
			var ob *object.Folder
			ob, err = Get(client, DataCenter, basefolder)
			if err != nil {
				return
			}
			spec := new(types.VirtualMachineCloneSpec)
			_, err = virtualmachine.Clone(client, e.VirtualMachine, ob, e.Name, *spec, 1000)
		}
		if err != nil {
			return
		}
	}
	return
}

// ReadFileSystem Wil recursivly get all items in the subfolder
func (parent *FileSystemItem) RecursiveRead(client *govmomi.Client) (*[]FileSystemItem, error) {
	ob, err := parent.Folder.Children(context.Background())
	SubItems := len(ob)
	if SubItems == 0 {
		return nil, nil
	}
	array := make([]FileSystemItem, SubItems)
	for i, e := range ob {
		test := e.Reference().Type
		switch test {
		case "Folder":
			subOb, err := FromID(client, e.Reference().Value)
			if err != nil {
				return nil, err
			}
			array[i].Folder = subOb
			array[i].Name = path.Base(subOb.InventoryPath)
			obList, err := array[i].RecursiveRead(client)
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
	return &array, err
}

func (fileSystem *FileSystemItem) GetVmOjects() []*object.VirtualMachine {
	amount := fileSystem.RecursiveCountVmOjects(0)
	virtualMachineList, _ := fileSystem.RecursiveGetVmOjects(make([]*object.VirtualMachine, amount), 0)
	return virtualMachineList
}

// this function recursivly calls itself to count all VirtualMachines in the filesystem
func (fileSystem *FileSystemItem) RecursiveCountVmOjects(NumberOfVms int) int {
	if fileSystem.Subitems == nil {
		return NumberOfVms
	}
	for _, e := range *fileSystem.Subitems {
		if e.Folder != nil {
			NumberOfVms = e.RecursiveCountVmOjects(NumberOfVms)
		}
		if e.VirtualMachine != nil {
			NumberOfVms += 1
		}
	}
	return NumberOfVms
}

func (fileSystem *FileSystemItem) RecursiveGetVmOjects(vmArray []*object.VirtualMachine, vmCounter int) ([]*object.VirtualMachine, int) {
	if fileSystem.Subitems == nil {
		return vmArray, vmCounter
	}
	for _, e := range *fileSystem.Subitems {
		if e.Folder != nil {
			vmArray, vmCounter = e.RecursiveGetVmOjects(vmArray, vmCounter)
		}
		if e.VirtualMachine != nil {
			vmArray[vmCounter] = e.VirtualMachine
			vmCounter += 1
		}
	}
	return vmArray, vmCounter
}

func Delete(client *govmomi.Client, DataCenter, Path string) error {

	folder, err := Get(client, DataCenter, Path)
	if err != nil {
		return fmt.Errorf("cannot locate folder: %s", err)
	}

	childrenExist, err := HasChildren(folder)
	if err != nil {
		return fmt.Errorf("error checking for folder contents: %s", err)
	}
	if childrenExist {
		fileSystem, err := ReadFileSystem(client, DataCenter, Path)
		if err != nil {
			return err
		}
		err = virtualmachine.DeleteOjects(fileSystem.GetVmOjects())
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

	return generic.RunTaskWait(task)
}

// Restarts all the virtualmachines in the folder and subfolders
func ReStart(client *govmomi.Client, DataCenter, Path string) error {

	folder, err := Get(client, DataCenter, Path)
	if err != nil {
		return fmt.Errorf("cannot locate folder: %s", err)
	}

	childrenExist, err := HasChildren(folder)
	if err != nil {
		return fmt.Errorf("error checking for folder contents: %s", err)
	}
	if childrenExist {
		fileSystem, err := ReadFileSystem(client, DataCenter, Path)
		if err != nil {
			return err
		}
		vmObjects := fileSystem.GetVmOjects()
		virtualmachine.StopOjects(vmObjects)
		virtualmachine.StartOjects(vmObjects)
	}
	return nil
}

// Starts all the virtualmachines in the folder and subfolders
func Start(client *govmomi.Client, DataCenter, Path string) error {

	folder, err := Get(client, DataCenter, Path)
	if err != nil {
		return fmt.Errorf("cannot locate folder: %s", err)
	}

	childrenExist, err := HasChildren(folder)
	if err != nil {
		return fmt.Errorf("error checking for folder contents: %s", err)
	}
	if childrenExist {
		fileSystem, err := ReadFileSystem(client, DataCenter, Path)
		if err != nil {
			return err
		}
		virtualmachine.StartOjects(fileSystem.GetVmOjects())
	}
	return nil
}

// Stops all the virtualmachines in the folder and subfolders
func Stop(client *govmomi.Client, DataCenter, Path string) error {

	folder, err := Get(client, DataCenter, Path)
	if err != nil {
		return fmt.Errorf("cannot locate folder: %s", err)
	}

	childrenExist, err := HasChildren(folder)
	if err != nil {
		return fmt.Errorf("error checking for folder contents: %s", err)
	}
	if childrenExist {
		fileSystem, err := ReadFileSystem(client, DataCenter, Path)
		if err != nil {
			return err
		}
		virtualmachine.StopOjects(fileSystem.GetVmOjects())
	}
	return nil
}

// CreateFolder Creates the full folder path spaeciefied
func Create(client *govmomi.Client, DataCenter, Path string) (folderObject *object.Folder, err error) {
	folders := strings.Split(strings.Trim(Path, "/"), "/")
	var CurrentPath string
	for _, e := range folders {
		CurrentPath += "/" + e
		folderObject, err = CreateSingleFolder(client, "DemoLab-Son-DC", CurrentPath)
	}
	return
}

// CreateFolder only creates last subfolder, it fails if the path doesnt exist
func CreateSingleFolder(client *govmomi.Client, DataCenter, Path string) (*object.Folder, error) {
	var folderObject *object.Folder
	parent, err := Get(client, DataCenter, path.Dir(Path))
	if err != nil {
		return nil, fmt.Errorf("error trying to determine parent targetFolder: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if !FolderExists(client, DataCenter, Path) {
		folderObject, err = parent.CreateFolder(ctx, path.Base(Path))
		if err != nil {
			return nil, fmt.Errorf("error creating targetFolder: %s", err)
		}
	}

	return folderObject, nil
}

func FolderExists(client *govmomi.Client, DataCenter, Path string) bool {
	_, err := Get(client, DataCenter, Path)
	if err != nil {
		return false
	}
	return true
}

func ListSubFolders(client *govmomi.Client, DataCenter, Path string) (*[]string, error) {
	return ListFolderItems(client, DataCenter, Path, "Folder")
}

func ListVirtualMachinesInFolder(client *govmomi.Client, DataCenter, Path string) (*[]string, error) {
	return ListFolderItems(client, DataCenter, Path, "VirtualMachine")
}

func ListFolderItems(client *govmomi.Client, DataCenter, Path, Type string) (*[]string, error) {
	parentFolder, err := Get(client, DataCenter, Path)
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
		return false, err
	}
	return len(children) > 0, nil
}

// GetFolder returns an *object.Folder from a given absolute path.
// If no such folder is found, an appropriate error will be returned.
func Get(client *govmomi.Client, DataCenter, Path string) (*object.Folder, error) {
	ctx, cancel, finder, checkPath := generic.NewFinder(client, DataCenter, Path)
	defer cancel()
	folder, err := finder.Folder(ctx, checkPath)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func GetChildrenFromPath(client *govmomi.Client, DataCenter, Path string) ([]*object.Folder, error) {
	ctx, cancel, finder, checkPath := generic.NewFinder(client, DataCenter, Path)
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

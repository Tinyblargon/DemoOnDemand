package demo

import (
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/file"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/folder"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/virtualmachine"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
	"gopkg.in/yaml.v3"
)

type PortForward struct {
	SourcePort      int
	DestinationPort int
	DestinationIP   string
}

type DemoConfig struct {
	PortForwards []*PortForward
}

func Start(client *govmomi.Client, dataCenter, demoName, userName string, number int) error {
	return folder.Start(client, dataCenter, CreateDemoURl(demoName, userName, number)+"/Demo")
}

func Stop(client *govmomi.Client, dataCenter, demoName, userName string, number int) error {
	return folder.Stop(client, dataCenter, CreateDemoURl(demoName, userName, number)+"/Demo")
}

// Imports a new demo from the speciefid folder
func Import(client *govmomi.Client, dataCenter, path, name string, config *DemoConfig) (err error) {
	filePath := global.ConfigFolder + "/" + name
	data, err := yaml.Marshal(&config)
	if err != nil {
		return
	}
	err = file.Write(filePath, data)
	if err != nil {
		return
	}
	err = folder.Clone(client, dataCenter, path, global.TemplateFodler+"/"+name)
	return
}

func New(client *govmomi.Client, dataCenter, demoName, userName string, number int) (err error) {
	basePath := CreateDemoURl(demoName, userName, number)
	folderObject, err := folder.Create(client, dataCenter, basePath)
	if err != nil {
		return
	}
	vmObject, err := virtualmachine.Get(client, dataCenter, global.RouterFodler+"/"+global.IngressVM)
	if err != nil {
		return
	}
	spec := new(types.VirtualMachineCloneSpec)
	newVmObject, err := virtualmachine.Clone(client, vmObject, folderObject, vmObject.Name(), *spec, 999)
	if err != nil {
		return err
	}
	err = virtualmachine.Start(newVmObject)
	if err != nil {
		return err
	}

	err = folder.Clone(client, dataCenter, global.TemplateFodler+"/"+demoName, basePath+"/Demo")
	return
}

func ListAll(client *govmomi.Client, dataCenter string) (*[]string, error) {
	return folder.ListSubFolders(client, dataCenter, global.DemoFodler)
}

func CreateDemoURl(demoName, userName string, number int) string {
	return global.DemoFodler + "/" + userName + "_" + strconv.Itoa(number) + "_" + demoName
}

func Delete(client *govmomi.Client, dataCenter, demoName, userName string, number int) error {
	return folder.Delete(client, dataCenter, CreateDemoURl(demoName, userName, number))
}

func DestroyTemplate(client *govmomi.Client, dataCenter, TempalateName string) error {
	err := folder.Delete(client, dataCenter, global.TemplateFodler+"/"+TempalateName)
	if err != nil {
		return err
	}
	return file.Delete(global.ConfigFolder + "/" + TempalateName)
}

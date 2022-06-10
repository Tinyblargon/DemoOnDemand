package demo

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/file"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
	"gopkg.in/yaml.v3"
)

type PortForward struct {
	SourcePort      uint
	DestinationPort uint
	DestinationIP   string
}

type DemoConfig struct {
	Description  string
	PortForwards []*PortForward
}

func Start(client *govmomi.Client, db *sql.DB, dataCenter, demoName, userName string, demoNumber uint, status *taskstatus.Status) (err error) {
	err = folder.Start(client, dataCenter, CreateDemoURl(demoName, userName, demoNumber)+"/Demo", status)
	if err != nil {
		return
	}
	return database.UpdateDemoOfUser(db, userName, demoName, demoNumber, true)
}

func Stop(client *govmomi.Client, db *sql.DB, dataCenter, demoName, userName string, demoNumber uint, status *taskstatus.Status) (err error) {
	err = folder.Stop(client, dataCenter, CreateDemoURl(demoName, userName, demoNumber)+"/Demo", status)
	if err != nil {
		return
	}
	return database.UpdateDemoOfUser(db, userName, demoName, demoNumber, false)
}

// Imports a new demo from the speciefid folder
func Import(client *govmomi.Client, dataCenter, path, name, pool string, config *DemoConfig, status *taskstatus.Status) (err error) {
	filePath := global.ConfigFolder + "/" + name
	err = folder.Clone(client, dataCenter, path, global.TemplateFodler+"/"+name, pool, true, status)
	if err != nil {
		return
	}
	data, _ := yaml.Marshal(&config)
	return file.Write(filePath, data)
}

func New(client *govmomi.Client, db *sql.DB, dataCenter, demoName, userName, pool string, demoNumber, demoLimit uint, status *taskstatus.Status) (err error) {
	numberOfDemos, err := database.NumberOfDomosOfUser(db, userName)
	if err != nil {
		return
	}
	if numberOfDemos > demoLimit {
		return fmt.Errorf("max number of concurrent demos reached")
	}
	err = database.AddDemoOfUser(db, userName, demoName, demoNumber)
	if err != nil {
		return
	}
	err = New_Subroutine(client, dataCenter, demoName, userName, pool, demoNumber, status)
	if err != nil {
		_ = database.DeleteDemoOfUser(db, userName, demoName, demoNumber)
	}
	return
}

func New_Subroutine(client *govmomi.Client, dataCenter, demoName, userName, pool string, demoNumber uint, status *taskstatus.Status) (err error) {
	basePath := CreateDemoURl(demoName, userName, demoNumber)
	folderObject, err := folder.Create(client, dataCenter, basePath)
	if err != nil {
		return
	}
	vmObject, err := virtualmachine.Get(client, dataCenter, global.RouterFodler+"/"+global.IngressVM)
	if err != nil {
		return
	}
	spec := new(types.VirtualMachineCloneSpec)
	newVmObject, err := virtualmachine.Clone(client, vmObject, folderObject, vmObject.Name(), *spec, 999, status)
	if err != nil {
		return
	}
	err = virtualmachine.Start(newVmObject, status)
	if err != nil {
		return
	}
	return folder.Clone(client, dataCenter, global.TemplateFodler+"/"+demoName, basePath+"/Demo", pool, false, status)
}

func ListAll(client *govmomi.Client, dataCenter string) (*[]string, error) {
	return folder.ListSubFolders(client, dataCenter, global.DemoFodler)
}

func CreateDemoURl(demoName, userName string, number uint) string {
	return global.DemoFodler + "/" + userName + "_" + strconv.Itoa(int(number)) + "_" + demoName
}

func Delete(client *govmomi.Client, db *sql.DB, dataCenter, demoName, userName string, demoNumber uint, status *taskstatus.Status) (err error) {
	demoURL := CreateDemoURl(demoName, userName, demoNumber)
	if !folder.Exists(client, dataCenter, demoURL) {
		err = database.DeleteDemoOfUser(db, userName, demoName, demoNumber)
		if err != nil {
			return
		}
	}
	err = folder.Delete(client, dataCenter, demoURL, status)
	if err != nil {
		return
	}
	return database.DeleteDemoOfUser(db, userName, demoName, demoNumber)
}

func DestroyTemplate(client *govmomi.Client, dataCenter, TempalateName string, status *taskstatus.Status) error {
	err := folder.Delete(client, dataCenter, global.TemplateFodler+"/"+TempalateName, status)
	if err != nil {
		return err
	}
	return file.Delete(global.ConfigFolder + "/" + TempalateName)
}

func ListTemplates() (files []string, err error) {
	return file.ReadDir(global.ConfigFolder)
}

func GetTemplate(templateName string) (templateConfig *DemoConfig, err error) {
	contents, err := file.Read(global.ConfigFolder + "/" + templateName)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(contents, &templateConfig)
	return
}

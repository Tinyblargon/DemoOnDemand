package demoactions

// demo actions

import (
	"database/sql"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/host"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/portgroup"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const demoDoesNotExist string = "demo does not exist"

func Start(client *govmomi.Client, db *sql.DB, dataCenter string, demo *demo.Demo, status *taskstatus.Status) (err error) {
	err = folder.Start(client, dataCenter, demo.CreateDemoURl()+"/Demo", status)
	if err != nil {
		return
	}
	return database.UpdateDemoOfUser(db, demo, true)
}

func Stop(client *govmomi.Client, db *sql.DB, dataCenter string, demo *demo.Demo, status *taskstatus.Status) (err error) {
	err = folder.Stop(client, dataCenter, demo.CreateDemoURl()+"/Demo", status)
	if err != nil {
		return
	}
	return database.UpdateDemoOfUser(db, demo, false)
}

// Creates a new demo of the the speciefied template
func New(client *govmomi.Client, db *sql.DB, dataCenter, pool string, demo *demo.Demo, demoLimit uint, status *taskstatus.Status) (err error) {
	numberOfDemos, err := database.NumberOfDomosOfUser(db, demo.User)
	if err != nil {
		return
	}
	if numberOfDemos > demoLimit {
		return fmt.Errorf("max number of concurrent demos reached")
	}
	err = database.AddDemoOfUser(db, demo)
	if err != nil {
		return
	}
	templateConf, err := template.Get(demo.Name)
	if err != nil {
		return
	}
	err = createAndSetupDemo(client, dataCenter, pool, demo, templateConf, status)
	if err != nil {
		_ = database.DeleteDemoOfUser(db, demo)
	}
	return
}

func createAndSetupDemo(client *govmomi.Client, dataCenter, pool string, demo *demo.Demo, templateConf *template.Config, status *taskstatus.Status) (err error) {
	basePath := demo.CreateDemoURl()
	folderObject, err := folder.Create(client, dataCenter, basePath)
	if err != nil {
		return
	}
	vlans, err := createAndSetupVlans(client, demo, templateConf, status)
	if err != nil {
		return
	}
	err = cloneRouterVM(client, dataCenter, folderObject, vlans, status)
	if err != nil {
		return
	}
	return folder.Clone(client, dataCenter, global.TemplateFodler+"/"+demo.Name, basePath+"/Demo", pool, false, status)
}

func createAndSetupVlans(client *govmomi.Client, demo *demo.Demo, templateConf *template.Config, status *taskstatus.Status) (vlans *vlan.LocalList, err error) {
	reservedVlans, err := vlan.ReserveAmount(demo, uint(len(templateConf.Networks)))
	if err != nil {
		return
	}
	vlans = &vlan.LocalList{
		Original: &templateConf.Networks,
		Remapped: &reservedVlans,
	}
	err = portgroup.Create(client, host.List, &reservedVlans, vlan.List.NewPrefix, global.VMwareConfig.Vswitch, global.Concurency, status)
	return
}

// setup the vm responsible for making all the routing work
func cloneRouterVM(client *govmomi.Client, dataCenter string, folderObject *object.Folder, vlans *vlan.LocalList, status *taskstatus.Status) (err error) {
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
	return
}

func ListAll(client *govmomi.Client, dataCenter string) (*[]string, error) {
	return folder.ListSubFolders(client, dataCenter, global.DemoFodler)
}

func Delete(client *govmomi.Client, db *sql.DB, dataCenter string, demo *demo.Demo, status *taskstatus.Status) (err error) {
	demoURL := demo.CreateDemoURl()
	existance, err := CheckExistance(db, *demo)
	if err != nil {
		return
	}
	if !existance {
		return fmt.Errorf(demoDoesNotExist)
	}
	if folder.Exists(client, dataCenter, demoURL) {
		err = folder.Delete(client, dataCenter, demoURL, status)
		if err != nil {
			return
		}
		// err = database.DeleteDemoOfUser(db, demo)
		// if err != nil {
		// 	return
		// }
	}
	err = deleteAndReleaseNetworks(client, db, demo, status)
	if err != nil {
		return
	}
	return database.DeleteDemoOfUser(db, demo)
}

func deleteAndReleaseNetworks(client *govmomi.Client, db *sql.DB, demo *demo.Demo, status *taskstatus.Status) (err error) {
	demoID := demo.CreateID()
	vlanObjList, err := database.ListUsedVlansOfDemo(db, demoID)
	if err != nil {
		return
	}
	vlanIdList := make([]uint, len(*vlanObjList))
	for i, e := range *vlanObjList {
		vlanIdList[i] = e.ID
	}
	if len(*vlanObjList) != 0 {
		err = portgroup.Delete(client, host.List, (*vlanObjList)[0].Prefix, &vlanIdList, global.Concurency, status)
		if err != nil {
			return
		}
		err = vlan.Release(demoID)
	}
	return
}

// Get the current properties like VLANS of a new demo you would like to import.
func GetImportProperties(client *govmomi.Client, dataCenter, folderContainingNewTemplate string) (networks []string, err error) {
	networks = make([]string, 0)
	status := new(taskstatus.Status)
	vmObjects, err := folder.GetVmObjectsFromPath(client, dataCenter, folderContainingNewTemplate)
	if err != nil {
		return
	}
	if len(vmObjects) == 0 {
		return
	}
	for _, e := range vmObjects {
		var vmNetworks []string
		vmNetworks, err = virtualmachine.GetNetworks(e, status)
		if err != nil {
			return
		}
		for _, networkID := range vmNetworks {
			if util.IsStringUnique(&networks, networkID) {
				networks = append(networks, networkID)
			}
		}
	}
	return
}

func CheckExistance(db *sql.DB, demo demo.Demo) (existance bool, err error) {
	userDemos, err := database.ListDemosOfUser(db, demo.User)
	if err != nil {
		return
	}
	for _, e := range *userDemos {
		if e.DemoName == demo.Name && e.DemoNumber == demo.ID {
			existance = true
			break
		}
	}
	return
}

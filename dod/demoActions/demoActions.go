package demoactions

// demo actions

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/host"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/network"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/portgroup"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualhost"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/Tinyblargon/DemoOnDemand/dod/template"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const demoDoesNotExist string = "demo does not exist"

func Start(client *govmomi.Client, db *sql.DB, dc *object.Datacenter, demo *demo.Demo, status *taskstatus.Status) (err error) {
	err = folder.Start(client, dc, demo.CreateDemoURl()+"/Demo", status)
	if err != nil {
		return
	}
	return database.UpdateDemoOfUser(db, demo, true)
}

func Stop(client *govmomi.Client, db *sql.DB, dc *object.Datacenter, demo *demo.Demo, status *taskstatus.Status) (err error) {
	err = folder.Stop(client, dc, demo.CreateDemoURl()+"/Demo", status)
	if err != nil {
		return
	}
	return database.UpdateDemoOfUser(db, demo, false)
}

// Creates a new demo of the the speciefied template
func New(client *govmomi.Client, db *sql.DB, dc *object.Datacenter, pool string, demo *demo.Demo, demoLimit uint, status *taskstatus.Status) (err error) {
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
	networkList := vlan.CreateLocalList(templateConf.Networks)
	err = createAndSetupDemo(client, dc, pool, demo, networkList, status)
	// if err != nil {
	// 	_ = database.DeleteDemoOfUser(db, demo)
	// }
	return
}

func createAndSetupDemo(client *govmomi.Client, dc *object.Datacenter, pool string, demo *demo.Demo, networkList []*vlan.LocalList, status *taskstatus.Status) (err error) {
	basePath := demo.CreateDemoURl()
	folderObject, err := folder.Create(client, dc, basePath)
	if err != nil {
		return
	}
	vlans, err := createAndSetupVlans(client, dc, demo, networkList, status)
	if err != nil {
		return
	}
	vmProperties, guestIP, err := cloneRouterVM(client, dc, folderObject, basePath, vlans, status)
	if err != nil {
		return
	}
	err = configureRouterVM(vmProperties, vlans, guestIP)
	if err != nil {
		return
	}
	return folder.Clone(client, dc, vlans, global.TemplateFodler+"/"+demo.Name, basePath+"/Demo", pool, false, status)
}

func createAndSetupVlans(client *govmomi.Client, dc *object.Datacenter, demo *demo.Demo, networkList []*vlan.LocalList, status *taskstatus.Status) (vlans []*vlan.LocalList, err error) {

	reservedVlans, err := vlan.ReserveVlans(demo, networkList)
	// reservedVlans, err := vlan.ReserveAmount(demo, uint(len(*templateConf.Networks)))
	if err != nil {
		return
	}
	err = portgroup.Create(client, host.List, &reservedVlans, vlan.List.NewPrefix, global.VMwareConfig.Vswitch, global.Concurency, status)
	if err != nil {
		return
	}
	time.Sleep(10 * time.Second)
	backingList, err := getAllbackingInfo(client, dc, reservedVlans)
	if err != nil {
		return
	}

	for _, e := range backingList {
		var networkName string
		for _, ee := range networkList {
			if ee.BackingInfo == nil {
				if networkName == "" {
					networkName = ee.OriginalNetwork
				}
				if networkName == ee.OriginalNetwork {
					ee.BackingInfo = e
				}
			}
		}
	}
	vlans = networkList
	return
}

// get backing info of the provided vlans
func getAllbackingInfo(client *govmomi.Client, dc *object.Datacenter, vlanList []uint) (backingList []*types.BaseVirtualDeviceBackingInfo, err error) {
	backingList = make([]*types.BaseVirtualDeviceBackingInfo, len(vlanList))
	for i, e := range vlanList {
		var networkObj *object.NetworkReference
		networkObj, err = network.FromName(client, dc, vlan.List.NewPrefix+strconv.Itoa(int(e)))
		if err != nil {
			return
		}
		var backing *types.BaseVirtualDeviceBackingInfo
		backing, err = network.GetBackingInfo(networkObj)
		if err != nil {
			return
		}
		backingList[i] = backing
	}
	return
}

// setup the vm responsible for making all the routing work
func cloneRouterVM(client *govmomi.Client, dc *object.Datacenter, folderObject *object.Folder, basePath string, vlans []*vlan.LocalList, status *taskstatus.Status) (vmProperties *mo.VirtualMachine, guestIP string, err error) {
	vmObject, err := virtualmachine.Get(client, dc, global.RouterFodler+"/"+global.IngressVM)
	if err != nil {
		return
	}
	vmProperties, err = virtualmachine.Properties(vmObject, status)
	if err != nil {
		return
	}
	spec := new(types.VirtualMachineCloneSpec)
	for _, e := range vlans {
		spec, err = virtualmachine.AddNetworkInterface(vmProperties, spec, e.BackingInfo)
		if err != nil {
			return
		}
	}
	newVmObject, err := virtualmachine.Clone(client, vmObject, folderObject, vmObject.Name(), *spec, 999, status)
	if err != nil {
		return
	}
	err = virtualmachine.Start(newVmObject, status)
	if err != nil {
		return
	}
	guestIP, vmProperties, err = virtualmachine.GetGuestIP(client, basePath, global.IngressVM, dc, status)
	if err != nil {
		return
	}
	return
}

func configureRouterVM(vmProperties *mo.VirtualMachine, vlan []*vlan.LocalList, ip string) (err error) {
	virtualhost.GetInterfaceSettings(vmProperties, vlan)
	return
}

func Delete(client *govmomi.Client, db *sql.DB, dc *object.Datacenter, demo *demo.Demo, status *taskstatus.Status) (err error) {
	demoURL := demo.CreateDemoURl()
	existance, err := CheckExistance(db, *demo)
	if err != nil {
		return
	}
	if !existance {
		return fmt.Errorf(demoDoesNotExist)
	}
	if folder.Exists(client, dc, demoURL) {
		err = folder.Delete(client, dc, demoURL, status)
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
func GetImportProperties(client *govmomi.Client, dc *object.Datacenter, folderContainingNewTemplate string) (networks []string, err error) {
	networks = make([]string, 0)
	status := new(taskstatus.Status)
	vmObjects, err := folder.GetVmObjectsFromPath(client, dc, folderContainingNewTemplate)
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
		networks = *(util.FilterUniqueStrings(&vmNetworks))
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
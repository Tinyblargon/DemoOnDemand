package demoactions

// demo actions

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/concurrency"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/os/firewallconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/os/networkinterfaceconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/ssh"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere"
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
	"github.com/yahoo/vssh"
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
	templateConf, err := template.Get(demo.Name)
	if err != nil {
		return
	}
	demo.ID, err = database.AddDemoOfUser(db, demo)
	if err != nil {
		return
	}
	networkList, err := vlan.CreateLocalList(templateConf.Networks)
	if err != nil {
		return
	}
	err = createAndSetupDemo(client, dc, pool, demo, templateConf, networkList, status)
	return
}

func createAndSetupDemo(client *govmomi.Client, dc *object.Datacenter, pool string, demo *demo.Demo, config *template.Config, networkList []*vlan.LocalList, status *taskstatus.Status) (err error) {
	basePath := demo.CreateDemoURl()
	folderObject, err := folder.Create(client, dc, folder.VSphereFolderTypeVM, basePath)
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
	err = configureRouterVM(vmProperties, vlans, config, guestIP, global.RouterConfiguration.User, global.RouterConfiguration.Password, global.RouterConfiguration.Port, status)
	if err != nil {
		return
	}
	return folder.Clone(client, dc, vlans, global.TemplateFolder+"/"+demo.Name, basePath+"/Demo", pool, false, status)
}

func createAndSetupVlans(c *govmomi.Client, dc *object.Datacenter, demo *demo.Demo, networkList []*vlan.LocalList, status *taskstatus.Status) (vlans []*vlan.LocalList, err error) {
	reservedVlans, err := vlan.ReserveVlans(demo, networkList)
	if err != nil {
		return
	}
	hosts, err := host.ListAll(c, dc, host.GetArray())
	if err != nil {
		return
	}
	err = portgroup.Create(c, hosts, &reservedVlans, vlan.GetPrefix(), vsphere.GetConfig().Vswitch, concurrency.Threads(), status)
	if err != nil {
		return
	}
	time.Sleep(10 * time.Second)
	backingList, err := getAllBackingInfo(c, dc, reservedVlans)
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
func getAllBackingInfo(client *govmomi.Client, dc *object.Datacenter, vlanList []uint) (backingList []*types.BaseVirtualDeviceBackingInfo, err error) {
	backingList = make([]*types.BaseVirtualDeviceBackingInfo, len(vlanList))
	for i, e := range vlanList {
		var networkObj *object.NetworkReference
		networkObj, err = network.FromName(client, dc, vlan.GetPrefix()+strconv.Itoa(int(e)))
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
	vmObject, err := virtualmachine.Get(client, dc, global.RouterFolder+"/"+global.IngressVM)
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
	guestIP, vmProperties, err = virtualmachine.GetGuestIP(client, basePath, global.IngressVM, vlan.GetNetworkList(vlans), dc, status)
	if err != nil {
		return
	}
	return
}

func configureRouterVM(vmProperties *mo.VirtualMachine, vlan []*vlan.LocalList, config *template.Config, ip, username, password string, sshPort uint16, status *taskstatus.Status) (err error) {
	vs, err := ssh.New(username, password, ip, sshPort)
	if err != nil {
		return
	}
	networks, interfaces, firstInterface, err := getInterfaces(vs, vmProperties, vlan, status)
	if err != nil {
		return
	}
	err = writeNetConfig(vs, networks, interfaces, firstInterface, status)
	if err != nil {
		return
	}
	err = writeFirewallConfig(vs, config.PortForwards, firstInterface, sshPort, status)
	if err != nil {
		return
	}
	return restartRouterVM(vs, status)
}

func restartRouterVM(vs *vssh.VSSH, status *taskstatus.Status) error {
	status.AddToInfo("Restarting routerVM")
	return ssh.RestartOS(vs)
}

func writeNetConfig(vs *vssh.VSSH, networks []*vlan.LocalList, interfaces *[]ssh.NetworkInterfaces, firstInterface string, status *taskstatus.Status) error {
	status.AddToInfo("Writing network config")
	return ssh.WriteToFile(vs, "/etc/network/interfaces", buildNetConfig(networks, interfaces, firstInterface))
}

func buildNetConfig(networks []*vlan.LocalList, interfaces *[]ssh.NetworkInterfaces, firstInterface string) (netConfig *[]string) {
	netConfig = networkinterfaceconfig.Base()
	*netConfig = append(*netConfig, networkinterfaceconfig.New(firstInterface, nil, true)...)
	for _, e := range networks {
		for _, ee := range *interfaces {
			if e.Mac == ee.Mac {
				cidr := net.IPNet{
					IP:   e.RouterIP,
					Mask: e.Net.Mask,
				}
				*netConfig = append(*netConfig, networkinterfaceconfig.New(ee.Name, &cidr, false)...)
			}
		}
	}
	return
}

func writeFirewallConfig(vs *vssh.VSSH, portForwards []*template.PortForward, firstInterface string, sshPort uint16, status *taskstatus.Status) (err error) {
	status.AddToInfo("Writing firewall config")
	err = ssh.WriteToFile(vs, firewallconfig.FirewallFile, buildFirewallConfig(portForwards, firstInterface, sshPort))
	if err != nil {
		return
	}
	return ssh.ChangeModifiers(vs, "755", firewallconfig.FirewallFile)
}

func buildFirewallConfig(portForwards []*template.PortForward, firstInterface string, sshPort uint16) (firewallConfig *[]string) {
	firewallConfig = firewallconfig.Base()
	*firewallConfig = append(*firewallConfig, firewallconfig.New("TCP", sshPort))
	for _, e := range portForwards {
		*firewallConfig = append(*firewallConfig, firewallconfig.NewPreRouting(uint16(e.SourcePort), uint16(e.DestinationPort), e.DestinationIP, e.Protocol, firstInterface))
	}
	return
}

func getFirstNetworkInterface(interfaces *[]ssh.NetworkInterfaces, firstMac string) (firstInterface string) {
	for _, e := range *interfaces {
		if e.Mac == firstMac {
			firstInterface = e.Name
			break
		}
	}
	return
}

func getInterfaces(vs *vssh.VSSH, vmProperties *mo.VirtualMachine, vlan []*vlan.LocalList, status *taskstatus.Status) (networks []*vlan.LocalList, interfaces *[]ssh.NetworkInterfaces, firstInterface string, err error) {
	status.AddToInfo("Obtaining network interfaces of routerVM")
	networks = virtualhost.GetInterfaceSettings(vmProperties, vlan)
	interfaces, err = ssh.ListNetworkInterfaces(vs)
	if err != nil {
		return
	}
	err = ssh.GetMacAddresses(vs, interfaces)
	if err != nil {
		return
	}
	firstInterface = getFirstNetworkInterface(interfaces, virtualmachine.GetFirstMac(vmProperties))
	return
}

func Delete(client *govmomi.Client, db *sql.DB, dc *object.Datacenter, demoObj *demo.Demo, status *taskstatus.Status) (err error) {
	demoURL := demoObj.CreateDemoURl()
	existence, err := CheckExistence(db, *demoObj)
	if err != nil {
		return
	}
	if !existence {
		return fmt.Errorf(demoDoesNotExist)
	}
	if folder.Exists(client, dc, folder.VSphereFolderTypeVM, demoURL) {
		err = folder.Delete(client, dc, demoURL, status)
		if err != nil {
			return
		}
	}
	err = deleteAndReleaseNetworks(client, dc, db, demoObj, status)
	if err != nil {
		return
	}
	return database.DeleteDemoOfUser(db, demoObj)
}

func deleteAndReleaseNetworks(c *govmomi.Client, dc *object.Datacenter, db *sql.DB, demoObj *demo.Demo, status *taskstatus.Status) (err error) {
	vlanObjList, err := database.ListUsedVlansOfDemo(db, demoObj)
	if err != nil {
		return
	}
	vlanIdList := make([]uint, len(*vlanObjList))
	for i, e := range *vlanObjList {
		vlanIdList[i] = e.ID
	}
	if len(*vlanObjList) != 0 {
		var hosts []*object.HostSystem
		hosts, err = host.ListAll(c, dc, host.GetArray())
		if err != nil {
			return
		}
		err = portgroup.Delete(c, hosts, (*vlanObjList)[0].Prefix, &vlanIdList, concurrency.Threads(), status)
		if err != nil {
			return
		}
		err = vlan.Release(demoObj)
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
		vmNetworks = *util.FilterUniqueStrings(&vmNetworks)
		networks = append(networks, vmNetworks...)
	}
	networks = *util.FilterUniqueStrings(&networks)
	return
}

func CheckExistence(db *sql.DB, demo demo.Demo) (existence bool, err error) {
	userDemos, err := database.ListDemosOfUser(db, demo.User)
	if err != nil {
		return
	}
	for _, e := range *userDemos {
		if e.DemoName == demo.Name && e.DemoNumber == demo.ID {
			existence = true
			break
		}
	}
	return
}

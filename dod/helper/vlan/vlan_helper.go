package vlan

import (
	"fmt"
	"net"
	"sync"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/vmware/govmomi/vim25/types"
)

const OutOfVlans string = "no more vlans availible"

// This is used to keep track of wich vlans have been reserved. it should be synced to the database after every change.
var list *VlanData

type VlanData struct {
	Vlans     *[]*database.Vlan
	NewPrefix string
	Mutex     sync.Mutex
}

type LocalList struct {
	OriginalNetwork string
	RouterIP        net.IP
	Net             *net.IPNet
	Mac             string
	BackingInfo     *types.BaseVirtualDeviceBackingInfo
}

// Create a list containg all the localy needed vlan/network information
func CreateLocalList(configList *[]template.Network) (List []*LocalList, err error) {
	List = make([]*LocalList, len(*configList))
	for i, e := range *configList {
		routerIP, net, err := net.ParseCIDR(e.RouterSubnet)
		if err != nil {
			return nil, err
		}
		List[i] = &LocalList{
			OriginalNetwork: e.Name,
			RouterIP:        routerIP,
			Net:             net,
		}
	}
	return
}

func Initialize(vlanIDs *[]uint, prefix string) (err error) {
	if list != nil {
		return fmt.Errorf("list can only be initialized once")
	}
	err = validatedSettigns(prefix)
	if err != nil {
		return
	}
	dbVlans, err := database.ListUsedVlans(global.DB)
	if err != nil {
		return
	}
	vlans := make([]*database.Vlan, len(*(vlanIDs)))
	for i, e := range *vlanIDs {
		vlan := database.Vlan{
			ID:     e,
			Prefix: prefix,
		}
		for _, ee := range *dbVlans {
			if e == ee.ID {
				vlan.Demo = ee.Demo
				vlan.Prefix = ee.Prefix
				break
			}
		}
		vlans[i] = &vlan
	}
	list = &VlanData{
		Vlans:     &vlans,
		NewPrefix: prefix,
	}
	return
}

// Reserves a vlan for each unique itm in the list
func ReserveVlans(demo *demo.Demo, list []*LocalList) (idList []uint, err error) {
	tmpList := make([]string, len(list))
	for i, e := range list {
		tmpList[i] = e.OriginalNetwork
	}
	return reserveAmount(demo, uint(len(*util.FilterUniqueStrings(&tmpList))))
}

// Reserves x amount of vlans from the list of availible vlans
func reserveAmount(demo *demo.Demo, numberOfVlans uint) (idList []uint, err error) {
	idList = make([]uint, numberOfVlans)
	for i := range idList {
		var id uint
		id, err = reserve(demo)
		if err != nil {
			return
		}
		idList[i] = id
	}
	return
}

// Reseves a vlan from the list of availible vlans
func reserve(demo *demo.Demo) (id uint, err error) {
	var counter int
	list.Mutex.Lock()
	for _, e := range *list.Vlans {
		if e.Demo == "" {
			e.Demo = demo.Name
			e.Prefix = list.NewPrefix
			id = e.ID
			break
		}
		counter++
	}
	list.Mutex.Unlock()
	if counter == len(*list.Vlans) {
		err = fmt.Errorf(OutOfVlans)
	} else {
		err = database.SetVlanInUse(global.DB, id, list.NewPrefix, demo)
	}
	return
}

// Releases all vlans associated with the speciefied demo from the list of availible vlans
func Release(demoObj *demo.Demo) (err error) {
	err = database.DeleteVlanInUse(global.DB, demoObj)
	if err != nil {
		return
	}
	demoId := demoObj.CreateID()
	for _, e := range *list.Vlans {
		if e.Demo == demoId {
			list.Mutex.Lock()
			e.Demo = ""
			list.Mutex.Unlock()
		}
	}
	return
}

func GetPrefix() string {
	return list.NewPrefix
}

func validatedSettigns(prefix string) error {
	if prefix == "" {
		return fmt.Errorf("vlan prefix may not be empty")
	}
	return nil
}

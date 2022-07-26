package vlan

import (
	"fmt"
	"sync"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/vmware/govmomi/vim25/types"
)

const OutOfVlans string = "no more vlans availible"

// This is used to keep track of wich vlans have been reserved. it should be synced to the database after every change.
var List *VlanData

type VlanData struct {
	Vlans     *[]*database.Vlan
	NewPrefix string
	Mutex     sync.Mutex
}

// This Stuct is for keeping track of localized vlan usage on translation/remapping of vlans/networks.
type LocalList struct {
	Original *[]template.Network
	Remapped *[]*types.BaseVirtualDeviceBackingInfo
}

func Initialize(vlanIDs *[]uint, prefix string) (err error) {
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
	vlanData := VlanData{
		Vlans:     &vlans,
		NewPrefix: prefix,
	}
	List = &vlanData
	return
}

// Reserves x amount of vlans from the list of availible vlans
func ReserveAmount(demo *demo.Demo, numberOfVlans uint) (idList []uint, err error) {
	idList = make([]uint, numberOfVlans)
	for i := range idList {
		var id uint
		id, err = reserve(demo)
		if err != nil {
			return
		}
		idList[i] = id
		// idList[i] = List.Prefix + strconv.Itoa(int(id))
	}
	return
}

// Reseves a vlan from the list of availible vlans
func reserve(demo *demo.Demo) (id uint, err error) {
	var counter int
	test := List
	_ = test
	List.Mutex.Lock()
	for _, e := range *List.Vlans {
		if e.Demo == "" {
			e.Demo = demo.Name
			e.Prefix = List.NewPrefix
			id = e.ID
			break
		}
		counter++
	}
	List.Mutex.Unlock()
	if counter == len(*List.Vlans) {
		err = fmt.Errorf(OutOfVlans)
	} else {
		err = database.SetVlanInUse(global.DB, id, List.NewPrefix, demo)
	}
	return
}

// Releases all vlans associated with the speciefied demo from the list of availible vlans
func Release(demo string) (err error) {
	err = database.DeleteVlanInUse(global.DB, demo)
	if err != nil {
		return
	}
	for _, e := range *List.Vlans {
		if e.Demo == demo {
			List.Mutex.Lock()
			e.Demo = ""
			List.Mutex.Unlock()
		}
	}
	return
}

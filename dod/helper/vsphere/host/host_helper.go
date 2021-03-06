package host

import (
	"context"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

var List []*object.HostSystem

func SetGlobal(hostList []*object.HostSystem) {
	List = hostList
}

func ListAll(client *govmomi.Client, datacenter *object.Datacenter, hostsArray []string) (hostList []*object.HostSystem, err error) {
	hostList = make([]*object.HostSystem, len(hostsArray))
	for i, e := range hostsArray {
		hostOBJ, err := objectFromName(client, e, datacenter)
		if err != nil {
			return nil, err
		}
		hostList[i] = hostOBJ
	}
	return hostList, nil
}

func objectFromName(c *govmomi.Client, hostName string, datacenter *object.Datacenter) (hostOBJ *object.HostSystem, err error) {
	hostOBJ, err = systemOrDefault(c, hostName, datacenter)
	if err != nil {
		return nil, fmt.Errorf("error fetching host: %s", err)
	}
	return
}

// SystemOrDefault returns a HostSystem from a specific host name and
// datacenter. If the user is connecting over ESXi, the default host system is
// used.
func systemOrDefault(client *govmomi.Client, name string, dc *object.Datacenter) (*object.HostSystem, error) {
	finder := find.NewFinder(client.Client, false)
	finder.SetDatacenter(dc)

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	t := client.ServiceContent.About.ApiType
	switch t {
	case "HostAgent":
		return finder.DefaultHostSystem(ctx)
	case "VirtualCenter":
		if name != "" {
			return finder.HostSystem(ctx, name)
		}
		return finder.DefaultHostSystem(ctx)
	}
	return nil, fmt.Errorf("unsupported ApiType: %s", t)
}

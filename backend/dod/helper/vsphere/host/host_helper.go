package host

import (
	"context"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

var hosts []string

func Initialize(client *govmomi.Client, dataCenter *object.Datacenter, hostsArray []string) (err error) {
	_, err = ListAll(client, dataCenter, hostsArray)
	if err == nil {
		hosts = hostsArray
	}
	return
}

func GetArray() []string {
	return hosts
}

func ListAll(client *govmomi.Client, dataCenter *object.Datacenter, hostsArray []string) (hostList []*object.HostSystem, err error) {
	hostList = make([]*object.HostSystem, len(hostsArray))
	for i, e := range hostsArray {
		hostOBJ, err := objectFromName(client, e, dataCenter)
		if err != nil {
			return nil, err
		}
		hostList[i] = hostOBJ
	}
	return hostList, nil
}

func objectFromName(c *govmomi.Client, hostName string, dataCenter *object.Datacenter) (hostOBJ *object.HostSystem, err error) {
	hostOBJ, err = systemOrDefault(c, hostName, dataCenter)
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

	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
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

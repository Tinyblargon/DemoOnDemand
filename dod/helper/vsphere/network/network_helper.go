package network

import (
	"context"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func GetBackingInfo(net *object.NetworkReference) (*types.BaseVirtualDeviceBackingInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	backing, err := (*net).EthernetCardBackingInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &backing, err
}

func IdFromName(client *govmomi.Client, name string, dc *object.Datacenter) (netID string, err error) {
	net, err := FromName(client, dc, name)
	if err != nil {
		return "", fmt.Errorf("error fetching network: %s", err)
	}

	netID = (*net).Reference().Value
	return
}

func FromName(client *govmomi.Client, dc *object.Datacenter, name string) (*object.NetworkReference, error) {
	finder := find.NewFinder(client.Client, false)
	if dc != nil {
		finder.SetDatacenter(dc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	networks, err := finder.NetworkList(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(networks) == 0 {
		return nil, fmt.Errorf("%s %s not found", "Network", name)
	}

	switch {
	case len(networks) == 1:
		return &networks[0], nil
	case len(networks) > 1:
		return nil, fmt.Errorf("path '%s' resolves to multiple %ss, Please specify", name, "network")
	}
	return nil, fmt.Errorf("%s %s not found", "Network", name)
}

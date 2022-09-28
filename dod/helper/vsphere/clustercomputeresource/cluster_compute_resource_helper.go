package clustercomputeresource

import (
	"context"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

func PropertiesFromPath(client *govmomi.Client, dc *object.Datacenter, pool string, status *taskstatus.Status) (clusterProp *mo.ClusterComputeResource, err error) {
	clusteOB, err := FromPath(client, pool, dc, status)
	if err != nil {
		return
	}
	clusterProp, err = Properties(clusteOB)
	return
}

// code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/clustercomputeresource"

// FromPath loads a ClusterComputeResource from its path. The datacenter is
// optional if the path is specific enough to not require it.
func FromPath(client *govmomi.Client, name string, dc *object.Datacenter, status *taskstatus.Status) (*object.ClusterComputeResource, error) {
	finder := find.NewFinder(client.Client, false)
	if dc != nil {
		status.AddToInfo(fmt.Sprintf("Attempting to locate compute cluster %q in datacenter %q", name, dc.InventoryPath))
		finder.SetDatacenter(dc)
	} else {
		status.AddToInfo(fmt.Sprintf("Attempting to locate compute cluster at absolute path %q", name))
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return finder.ClusterComputeResource(ctx, name)
}

// Properties is a convenience method that wraps fetching the
// ClusterComputeResource MO from its higher-level object.
func Properties(cluster *object.ClusterComputeResource) (*mo.ClusterComputeResource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	var props mo.ClusterComputeResource
	if err := cluster.Properties(ctx, cluster.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

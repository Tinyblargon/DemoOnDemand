package clustercomputeresource

import (
	"context"
	"log"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

func PropertiesFromPath(client *govmomi.Client, DataCenter, pool string) (clusterProp *mo.ClusterComputeResource, err error) {
	dataCenterOB, err := datacenter.FromPath(client, DataCenter)
	if err != nil {
		return
	}
	clusteOB, err := FromPath(client, pool, dataCenterOB)
	if err != nil {
		return
	}
	clusterProp, err = Properties(clusteOB)
	return
}

// code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/clustercomputeresource"

// FromPath loads a ClusterComputeResource from its path. The datacenter is
// optional if the path is specific enough to not require it.
func FromPath(client *govmomi.Client, name string, dc *object.Datacenter) (*object.ClusterComputeResource, error) {
	finder := find.NewFinder(client.Client, false)
	if dc != nil {
		log.Printf("[DEBUG] Attempting to locate compute cluster %q in datacenter %q", name, dc.InventoryPath)
		finder.SetDatacenter(dc)
	} else {
		log.Printf("[DEBUG] Attempting to locate compute cluster at absolute path %q", name)
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

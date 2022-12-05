package datacenter

import (
	"context"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

var obj string

func Initialize(client *govmomi.Client, dataCenter string) (err error) {
	_, err = fromName(client, dataCenter)
	if err == nil {
		obj = dataCenter
	}
	return
}

func Get(client *govmomi.Client, dataCenter string) (*object.Datacenter, error) {
	return fromName(client, dataCenter)
}

func GetName() string {
	return obj
}

// code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"

// FromName returns a Datacenter via its supplied name.
func fromName(client *govmomi.Client, datacenter string) (*object.Datacenter, error) {
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	return finder.Datacenter(ctx, datacenter)
}

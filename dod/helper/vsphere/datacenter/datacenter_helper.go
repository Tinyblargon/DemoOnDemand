package datacenter

import (
	"context"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

var obj *object.Datacenter

func Initialize(client *govmomi.Client, datacenter string) (err error) {
	obj, err = fromName(client, datacenter)
	return
}

func GetObject() *object.Datacenter {
	return obj
}

// code borrowed from "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/virtualdevice"

// FromName returns a Datacenter via its supplied name.
func fromName(client *govmomi.Client, datacenter string) (*object.Datacenter, error) {
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return finder.Datacenter(ctx, datacenter)
}

package folder

import (
	"github.com/vmware/govmomi/object"
)

// Code burrowed from https://github.com/hashicorp/terraform-provider-vsphere/blob/main/vsphere/internal/helper/folder/folder_helper.go

// VSphereFolderType is an enumeration type for vSphere folder types.
type VSphereFolderType string

// The following are constants for the 5 vSphere folder types - these are used
// to help determine base paths and also to validate folder types in the
// vsphere_folder resource.
const (
	VSphereFolderTypeVM      = VSphereFolderType("vm")
	VSphereFolderTypeNetwork = VSphereFolderType("network")
	// VSphereFolderTypeHost      = VSphereFolderType("host")
	// VSphereFolderTypeDatastore = VSphereFolderType("datastore")

	// VSphereFolderTypeDatacenter is a special folder type - it does not get a
	// root path particle generated for it as it is an integral part of the path
	// generation process, but is defined so that it can be properly referenced
	// and used in validation.
	// VSphereFolderTypeDatacenter = VSphereFolderType("datacenter")
)

// RootPathParticle is the section of a vSphere inventory path that denotes a
// specific kind of inventory item.
type rootPathParticle VSphereFolderType

// RootFromDatacenter returns the root path for the particle from the given
// datacenter's inventory path.
func (p rootPathParticle) rootFromDatacenter(dc *object.Datacenter) string {
	return dc.InventoryPath + "/" + string(p)
}

// PathFromDatacenter returns the combined result of RootFromDatacenter plus a
// relative path for a given particle and datacenter object.
func (p rootPathParticle) pathFromDatacenter(dc *object.Datacenter, relative string) string {
	return p.rootFromDatacenter(dc) + "/" + relative
}

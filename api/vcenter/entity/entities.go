// Package entity contains vCenter object types referenceable via govmomi.
package entity

// Entity represents a vCenter entity.
type Entity int

// nolint:revive
const (
	Cluster Entity = iota
	Datacenter
	Datastore
	DistributedVirtualPortgroup
	DistributedVirtualSwitch
	Folder
	Host
	Network
	ResourcePool
	VCenterRoot
	VirtualApp
	VirtualMachine
)

// String converts an Entity to a string.
func (e Entity) String() string {
	names := []string{
		"cluster",
		"datacenter",
		"datastore",
		"dvp", // uncertain if meaningful for govmomi
		"dvs", // uncertain if meaningful for govmomi
		"folder",
		"host",
		"network",
		"resourcepool",
		"root", // not meaningful for govmomi - cosmetic only
		"vapp",
		"vm",
	}
	if e > VirtualMachine || e < Cluster {
		return "Unknown"
	}
	return names[e]
}

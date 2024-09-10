// Package entity contains vCenter objects/entities.
package entity

import "slices"

var (
	// Labels contains the pretty names of all vCenter entities.
	Labels []string

	// Map maps entity labels to entities.
	Map map[string]Entity
)

func init() {
	Map = make(map[string]Entity, len(LabelMap))
	for entity, label := range LabelMap {
		Labels = append(Labels, label)
		Map[label] = entity
	}
	slices.Sort(Labels)
}

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

// LabelMap maps entities to their pretty names.
var LabelMap = map[Entity]string{
	Cluster:                     "Cluster",
	Datacenter:                  "Datacenter",
	Datastore:                   "Datastore",
	DistributedVirtualPortgroup: "Distributed Port Group",
	DistributedVirtualSwitch:    "Distributed Switch",
	Folder:                      "Folder",
	Host:                        "ESXi Host",
	Network:                     "Network",
	ResourcePool:                "Resource Pool",
	VCenterRoot:                 "vCenter Root",
	VirtualApp:                  "Virtual App",
	VirtualMachine:              "Virtual Machine",
}

// nolint:revive
var ComputeResourceScopes = []string{
	Cluster.String(),
	Host.String(),
	ResourcePool.String(),
}

// String converts an Entity to a string.
func (e Entity) String() string {
	if e > VirtualMachine || e < Cluster {
		return "Unknown"
	}
	return LabelMap[e]
}

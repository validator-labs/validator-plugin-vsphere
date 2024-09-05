// Package constants contains the constants used in validator-plugin-vsphere
package constants

const (
	// PluginCode is the code of the plugin
	PluginCode string = "vSphere"

	// ValidationTypePrivileges is the validation type for privileges
	ValidationTypePrivileges string = "vsphere-privileges"

	// ValidationTypeTag is the validation type for tags
	ValidationTypeTag string = "vsphere-tags"

	// ValidationTypeComputeResources is the validation type for compute resources
	ValidationTypeComputeResources string = "vsphere-compute-resources"

	// ValidationTypeNTP is the validation type for NTP
	ValidationTypeNTP string = "vsphere-ntp"

	// ClusterInventoryPath is the path for cluster inventory
	ClusterInventoryPath = "/%s/host/%s"

	// HostSystemInventoryPath is the path for host system inventory
	HostSystemInventoryPath = "/%s/host/%s/%s"

	// ResourcePoolInventoryPath is the path for resource pool inventory
	ResourcePoolInventoryPath = "/%s/host/%s/Resources/%s"

	// ClusterDefaultResourcePoolName is the default resource pool name for cluster
	ClusterDefaultResourcePoolName = "Resources"
)

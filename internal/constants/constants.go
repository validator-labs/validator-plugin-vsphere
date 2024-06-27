package constants

const (
	PluginCode                     string = "vSphere"
	ValidationTypeRolePrivileges   string = "vsphere-role-privileges"
	ValidationTypeEntityPrivileges string = "vsphere-entity-privileges"
	ValidationTypeTag              string = "vsphere-tags"
	ValidationTypeComputeResources string = "vsphere-compute-resources"
	ValidationTypeNTP              string = "vsphere-ntp"

	ClusterInventoryPath      = "/%s/host/%s"
	HostSystemInventoryPath   = "/%s/host/%s/%s"
	ResourcePoolInventoryPath = "/%s/host/%s/Resources/%s"

	ClusterDefaultResourcePoolName = "Resources"
)

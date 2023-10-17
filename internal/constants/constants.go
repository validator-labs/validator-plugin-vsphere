package constants

const (
	PluginCode                     string = "VSPHERE"
	ValidationTypeRolePrivileges   string = "vsphere-role-privileges"
	ValidationTypeEntityPrivileges string = "vsphere-entity-privileges"
	ValidationTypeTag              string = "vsphere-tags"
	ValidationTypeComputeResources string = "vsphere-compute-resources"

	DatacenterInventoryPath     = "%s"
	ClusterInventoryPath        = "/%s/host/%s"
	HostSystemInventoryPath     = "/%s/host/%s/%s"
	VirtualMachineInventoryPath = "%s"
	FolderInventoryPath         = "%s"
	ResourcePoolInventoryPath   = "/%s/host/%s/Resources/%s"

	ClusterDefaultResourcePoolName = "Resources"
)

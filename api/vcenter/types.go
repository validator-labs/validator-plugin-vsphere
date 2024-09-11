// Package vcenter contains vCenter object types.
package vcenter

import (
	"net/url"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

const (
	// ClusterDefaultResourcePoolName is the default resource pool name for a cluster.
	ClusterDefaultResourcePoolName = "Resources"

	// DefaultDomain is the default vCenter domain.
	DefaultDomain = "VSPHERE.LOCAL"

	// Inventory Paths

	// ClusterInventoryPath is the path for cluster inventory.
	// Replacements: datacenter name, cluster name.
	ClusterInventoryPath = "/%s/host/%s"

	// HostSystemInventoryPath is the path for host system inventory.
	// Replacements: datacenter name, cluster name, host system name.
	HostSystemInventoryPath = "/%s/host/%s/%s"

	// NetworkInventoryPath is the path for network inventory.
	// Replacements: datacenter name, network name.
	NetworkInventoryPath = "/%s/network/%s"

	// ResourcePoolInventoryPath is the path for resource pool inventory.
	// Replacements: datacenter name, cluster name, resource pool name.
	ResourcePoolInventoryPath = "/%s/host/%s/Resources/%s"

	// VMFolderInventoryPath is the path for VM folder inventory.
	// Replacements: datacenter name, vm folder name.
	VMFolderInventoryPath = "/%s/vm/%s"
)

// Account contains vCenter account details.
type Account struct {
	// Insecure controls whether to validate the vCenter server's certificate.
	Insecure bool `json:"insecure" yaml:"insecure"`

	// Password is the vCenter password.
	Password string `json:"password" yaml:"password"`

	// Username is the vCenter username.
	Username string `json:"username" yaml:"username"`

	// Host is the vCenter URL.
	Host string `json:"host" yaml:"host"`
}

// Userinfo returns a vCenter account's credentials in Userinfo format.
func (a Account) Userinfo() *url.Userinfo {
	return url.UserPassword(a.Username, a.Password)
}

// Datastore defines a datastore
type Datastore struct {
	Name string
	ID   string
}

// HostSystem defines a vCenter host system.
type HostSystem struct {
	Name      string
	Reference string
}

// HostDateInfo defines date information for a vCenter host system.
type HostDateInfo struct {
	types.HostDateTimeInfo
	HostName      string
	NTPServers    []string
	Service       *types.HostService
	Current       *time.Time
	ClientStatus  string
	ServiceStatus string
}

// Servers returns a slice of NTP servers for a vCenter host system.
func (i *HostDateInfo) Servers() []string {
	return i.NtpConfig.Server
}

// Network defines a vCenter network.
type Network struct {
	Type      string
	IP        string
	Interface string
}

// SSHInfo defines the SSH information.
type SSHInfo struct {
	Username   string
	Password   string
	PublicKey  []string
	PrivateKey []string
}

// VMInfo defines scope information for a vCenter VM.
type VMInfo struct {
	Folder    string
	Cluster   string
	Datastore string
	Network   string
}

// AdditionalDisk defines an additional disk.
type AdditionalDisk struct {
	Name      string
	Device    string
	Capacity  string
	Used      string
	Available string
	Usage     string
}

// Metrics defines the VM metrics.
type Metrics struct {
	CPUCores        string
	CPUUsage        string
	MemoryBytes     string
	MemoryUsage     string
	DiskUsage       string
	DiskProvisioned string
}

// VM defines a vCenter virtual machine.
type VM struct {
	Name           string
	Type           string
	Status         string
	IPAddress      string
	Host           string
	CPU            int32
	Memory         int32
	RootDiskSize   int32
	Network        []Network
	VMInfo         VMInfo
	SSHInfo        SSHInfo
	AdditionalDisk []AdditionalDisk
	Metrics        Metrics
	Storage        []Datastore
}

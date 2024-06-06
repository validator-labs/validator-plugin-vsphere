package vsphere

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
)

type VSphereCloudDriver struct {
	VCenterServer   string
	VCenterUsername string
	VCenterPassword string
	Datacenter      string
	Client          *govmomi.Client
	RestClient      *rest.Client
}

type VsphereCloudAccount struct {
	// Insecure is a flag that controls whether to validate the vSphere server's certificate.
	Insecure bool `json:"insecure"`

	// password
	// Required: true
	Password string `json:"password"`

	// username
	// Required: true
	Username string `json:"username"`

	// VcenterServer is the address of the vSphere endpoint
	// Required: true
	VcenterServer string `json:"vcenterServer"`
}

type Session struct {
	GovmomiClient *govmomi.Client
	RestClient    *rest.Client
}

type VSphereVM struct {
	Name           string
	Type           string
	Status         string
	IpAddress      string
	Host           string
	Cpu            int32
	Memory         int32
	RootDiskSize   int32
	Network        []VSphereNetwork
	LibvirtVmInfo  LibvirtVmInfo
	VSphereVmInfo  VSphereVmInfo
	SshInfo        SshInfo
	AdditionalDisk []AdditionalDisk
	Metrics        Metrics
	Storage        []Datastore
}

type VSphereNetwork struct {
	Type      string
	Ip        string
	Interface string
}

type LibvirtVmInfo struct {
	ImagePool string
	DataPool  string
}

type VSphereVmInfo struct {
	Folder    string
	Cluster   string
	Datastore string
	Network   string
}

type SshInfo struct {
	Username   string
	Password   string
	PublicKey  []string
	Privatekey []string
}

type AdditionalDisk struct {
	Name      string
	Device    string
	Capacity  string
	Used      string
	Available string
	Usage     string
}

type Metrics struct {
	CpuCores        string
	CpuUsage        string
	MemoryBytes     string
	MemoryUsage     string
	DiskUsage       string
	DiskProvisioned string
}

type VSphereHostSystem struct {
	Name      string
	Reference string
}

type Datastore struct {
	Name string
	Id   string
}

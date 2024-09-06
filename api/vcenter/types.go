// Package vcenter contains vCenter object types.
package vcenter

import "net/url"

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

// Entity represents a vCenter entity, referenceable via govmomi.
type Entity int

// nolint:revive
const (
	Cluster Entity = iota
	Datacenter
	Datastore
	Folder
	Host
	Network
	ResourcePool
	VApp
	VCenterRoot
	VDS
	VM
)

// String converts an Entity to a string.
func (e Entity) String() string {
	names := []string{
		"cluster",
		"datacenter",
		"datastore",
		"folder",
		"host",
		"network",
		"resourcepool",
		"vapp",
		"",
		"vds",
		"vm",
	}
	if e > VM || e < Cluster {
		return "Unknown"
	}
	return names[e]
}

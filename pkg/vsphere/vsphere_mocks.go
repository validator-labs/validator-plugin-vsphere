package vsphere

import (
	"context"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
)

// MockVsphereDriver is a mock implementation of the Driver interface
type MockVsphereDriver struct {
	Datacenters        []string
	Clusters           []string
	VMs                []VM
	VMFolders          []string
	HostSystems        map[string][]HostSystem
	VApps              []mo.VirtualApp
	ResourcePools      []*object.ResourcePool
	HostClusterMapping map[string]string
	ResourceTags       map[string]tags.AttachedTags
}

// ensure that MockVsphereDriver implements the Driver interface
var _ Driver = &MockVsphereDriver{}

// GetVSphereVMFolders returns a mocked response
func (d MockVsphereDriver) GetVSphereVMFolders(_ context.Context, _ string) ([]string, error) {
	return d.VMFolders, nil
}

// GetVSphereDatacenters returns a mocked response
func (d MockVsphereDriver) GetVSphereDatacenters(_ context.Context) ([]string, error) {
	return d.Datacenters, nil
}

// GetVSphereClusters returns a mocked response
func (d MockVsphereDriver) GetVSphereClusters(_ context.Context, _ string) ([]string, error) {
	return d.Clusters, nil
}

// GetVSphereHostSystems returns a mocked response
func (d MockVsphereDriver) GetVSphereHostSystems(_ context.Context, datacenter, cluster string) ([]HostSystem, error) {
	return d.HostSystems[concat(datacenter, cluster)], nil
}

// IsValidVSphereCredentials returns a mocked response
func (d MockVsphereDriver) IsValidVSphereCredentials() (bool, error) {
	return true, nil
}

// ValidateVsphereVersion returns a mocked response
func (d MockVsphereDriver) ValidateVsphereVersion(_ string) error {
	return nil
}

// GetHostClusterMapping returns a mocked response
func (d MockVsphereDriver) GetHostClusterMapping(_ context.Context) (map[string]string, error) {
	return d.HostClusterMapping, nil
}

// GetVSphereVms returns a mocked response
func (d MockVsphereDriver) GetVSphereVms(_ context.Context, _ string) ([]VM, error) {
	return d.VMs, nil
}

// GetResourcePools returns a mocked response
func (d MockVsphereDriver) GetResourcePools(_ context.Context, _ string, _ string) ([]*object.ResourcePool, error) {
	return d.ResourcePools, nil
}

// GetVapps returns a mocked response
func (d MockVsphereDriver) GetVapps(_ context.Context) ([]mo.VirtualApp, error) {
	return d.VApps, nil
}

// GetResourceTags returns a mocked response
func (d MockVsphereDriver) GetResourceTags(_ context.Context, _ string) (map[string]tags.AttachedTags, error) {
	return d.ResourceTags, nil
}

func concat(ss ...string) string {
	return strings.Join(ss, "_")
}

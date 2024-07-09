package vsphere

import (
	"context"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
)

// MockVsphereDriver is a mock implementation of the VsphereDriver interface
type MockVsphereDriver struct {
	Datacenters        []string
	Clusters           []string
	VMs                []VSphereVM
	VMFolders          []string
	HostSystems        map[string][]VSphereHostSystem
	VApps              []mo.VirtualApp
	ResourcePools      []*object.ResourcePool
	HostClusterMapping map[string]string
	ResourceTags       map[string]tags.AttachedTags
}

// ensure that MockVsphereDriver implements the VsphereDriver interface
var _ VsphereDriver = &MockVsphereDriver{}

// GetVSphereVMFolders returns a mocked response
func (d MockVsphereDriver) GetVSphereVMFolders(ctx context.Context, datacenter string) ([]string, error) {
	return d.VMFolders, nil
}

// GetVSphereDatacenters returns a mocked response
func (d MockVsphereDriver) GetVSphereDatacenters(ctx context.Context) ([]string, error) {
	return d.Datacenters, nil
}

// GetVSphereClusters returns a mocked response
func (d MockVsphereDriver) GetVSphereClusters(ctx context.Context, datacenter string) ([]string, error) {
	return d.Clusters, nil
}

// GetVSphereHostSystems returns a mocked response
func (d MockVsphereDriver) GetVSphereHostSystems(ctx context.Context, datacenter, cluster string) ([]VSphereHostSystem, error) {
	return d.HostSystems[concat(datacenter, cluster)], nil
}

// IsValidVSphereCredentials returns a mocked response
func (d MockVsphereDriver) IsValidVSphereCredentials(ctx context.Context) (bool, error) {
	return true, nil
}

// ValidateVsphereVersion returns a mocked response
func (d MockVsphereDriver) ValidateVsphereVersion(constraint string) error {
	return nil
}

// GetHostClusterMapping returns a mocked response
func (d MockVsphereDriver) GetHostClusterMapping(ctx context.Context) (map[string]string, error) {
	return d.HostClusterMapping, nil
}

// GetVSphereVms returns a mocked response
func (d MockVsphereDriver) GetVSphereVms(ctx context.Context, dcName string) ([]VSphereVM, error) {
	return d.VMs, nil
}

// GetResourcePools returns a mocked response
func (d MockVsphereDriver) GetResourcePools(ctx context.Context, datacenter string, cluster string) ([]*object.ResourcePool, error) {
	return d.ResourcePools, nil
}

// GetVapps returns a mocked response
func (d MockVsphereDriver) GetVapps(ctx context.Context) ([]mo.VirtualApp, error) {
	return d.VApps, nil
}

// GetResourceTags returns a mocked response
func (d MockVsphereDriver) GetResourceTags(ctx context.Context, resourceType string) (map[string]tags.AttachedTags, error) {
	return d.ResourceTags, nil
}

// IsAdminAccount returns a mocked response
func (d MockVsphereDriver) IsAdminAccount(ctx context.Context) (bool, error) {
	return true, nil
}

func concat(ss ...string) string {
	return strings.Join(ss, "_")
}

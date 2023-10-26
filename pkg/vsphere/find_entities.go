package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"

	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/constants"
)

func (v *VSphereCloudDriver) GetFolderIfExists(ctx context.Context, finder *find.Finder, datacenter, folderName string) (bool, *object.Folder, error) {
	folder, err := finder.Folder(ctx, folderName)
	if err != nil {
		return false, nil, err
	}
	return true, folder, nil
}

func (v *VSphereCloudDriver) GetClusterIfExists(ctx context.Context, finder *find.Finder, datacenter, clusterName string) (bool, *object.ClusterComputeResource, error) {
	path := fmt.Sprintf("/%s/host/%s", datacenter, clusterName)
	cluster, err := finder.ClusterComputeResource(ctx, path)
	if err != nil {
		return false, nil, err
	}
	return true, cluster, nil
}

func (v *VSphereCloudDriver) GetHostIfExists(ctx context.Context, finder *find.Finder, datacenter, clusterName, hostName string) (bool, *object.HostSystem, error) {
	path := fmt.Sprintf("/%s/host/%s/%s", datacenter, clusterName, hostName)
	// Handle datacenter level hosts
	if clusterName == "" {
		path = fmt.Sprintf("/%s/host/%s", datacenter, hostName)
	}
	host, err := finder.HostSystem(ctx, path)
	if err != nil {
		return false, nil, err
	}
	return true, host, nil
}

func (v *VSphereCloudDriver) GetResourcePoolIfExists(ctx context.Context, finder *find.Finder, datacenter, cluster, resourcePoolName string) (bool, *object.ResourcePool, error) {
	path := fmt.Sprintf("/%s/host/%s/Resources/%s", datacenter, cluster, resourcePoolName)
	// Handle Cluster level defaut resource pool called 'Resources'
	if resourcePoolName == constants.ClusterDefaultResourcePoolName {
		path = fmt.Sprintf("/%s/host/%s/%s", datacenter, cluster, resourcePoolName)
	}
	rp, err := finder.ResourcePool(ctx, path)
	if err != nil {
		return false, nil, err
	}
	return true, rp, nil
}

func (v *VSphereCloudDriver) GetVAppIfExists(ctx context.Context, finder *find.Finder, datacenter, vAppName string) (bool, *object.VirtualApp, error) {
	vapp, err := finder.VirtualApp(ctx, vAppName)
	if err != nil {
		return false, nil, err
	}
	return true, vapp, nil
}

func (v *VSphereCloudDriver) GetVMIfExists(ctx context.Context, finder *find.Finder, datacenter, cluster, vmName string) (bool, *object.VirtualMachine, error) {
	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		return false, nil, err
	}
	return true, vm, nil
}

func (v *VSphereCloudDriver) GetDatacenterIfExists(ctx context.Context, finder *find.Finder, datacenter string) (bool, *object.Datacenter, error) {
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return false, nil, err
	}
	return true, dc, nil
}

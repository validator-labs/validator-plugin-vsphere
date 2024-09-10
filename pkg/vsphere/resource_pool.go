package vsphere

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"

	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
)

// GetResourcePool returns the resource pool if it exists
func (v *VCenterDriver) GetResourcePool(ctx context.Context, finder *find.Finder, datacenter, cluster, resourcePoolName string) (*object.ResourcePool, error) {
	path := fmt.Sprintf("/%s/host/%s/Resources/%s", datacenter, cluster, resourcePoolName)

	// Handle the cluster-level default resource pool, 'Resources'
	if resourcePoolName == vcenter.ClusterDefaultResourcePoolName {
		path = fmt.Sprintf("/%s/host/%s/%s", datacenter, cluster, resourcePoolName)
	}

	rp, err := finder.ResourcePool(ctx, path)
	if err != nil {
		return nil, err
	}
	return rp, nil
}

// GetResourcePools returns a list of resource pools
func (v *VCenterDriver) GetResourcePools(ctx context.Context, datacenter string, cluster string) ([]*object.ResourcePool, error) {
	path := fmt.Sprintf("/%s/host/%s/Resources/*", datacenter, cluster)

	if cluster == "" {
		path = fmt.Sprintf("/%s/host/*", datacenter)
	}

	rps, err := v.getResourcePools(ctx, datacenter, path)
	if err != nil {
		return nil, err
	}

	return rps, nil
}

// GetVSphereResourcePools returns a sorted list of resource pool paths
func (v *VCenterDriver) GetVSphereResourcePools(ctx context.Context, datacenter string, cluster string) (resourcePools []string, err error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	searchPath := fmt.Sprintf("/%s/host/%s/Resources/*", dc, cluster)
	pools, govErr := finder.ResourcePoolList(ctx, searchPath)
	if govErr != nil {
		//ignore NotFoundError, to allow selection of "Resources" as the default option for rs pool
		if _, ok := govErr.(*find.NotFoundError); !ok {
			return nil, fmt.Errorf("failed to fetch vSphere resource pools. datacenter: %s, code: %d", datacenter, http.StatusBadRequest)
		}
	}

	for i := 0; i < len(pools); i++ {
		pool := pools[i]
		prefix := fmt.Sprintf("/%s/host/%s/Resources/", dc, cluster)
		poolPath := strings.TrimPrefix(pool.InventoryPath, prefix)
		resourcePools = append(resourcePools, poolPath)
		childPoolSearchPath := fmt.Sprintf("/%s/host/%s/Resources/%s/*", dc, cluster, poolPath)
		childPools, err := finder.ResourcePoolList(ctx, childPoolSearchPath)
		if err == nil {
			pools = append(pools, childPools...)
		}
	}

	sort.Strings(resourcePools)
	return resourcePools, nil
}

func (v *VCenterDriver) getResourcePools(ctx context.Context, datacenter, path string) ([]*object.ResourcePool, error) {
	finder, _, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get finder with datacenter")
	}

	rps, err := finder.ResourcePoolList(ctx, path)
	if err != nil {
		return nil, err
	}

	return rps, nil
}

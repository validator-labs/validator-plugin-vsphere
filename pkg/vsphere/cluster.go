package vsphere

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// GetClusterIfExists returns the cluster if it exists
func (v *VCenterDriver) GetClusterIfExists(ctx context.Context, finder *find.Finder, datacenter, clusterName string) (bool, *object.ClusterComputeResource, error) {
	path := fmt.Sprintf("/%s/host/%s", datacenter, clusterName)
	cluster, err := finder.ClusterComputeResource(ctx, path)
	if err != nil {
		return false, nil, err
	}
	return true, cluster, nil
}

// GetVSphereClusters returns a sorted list of vSphere clusters
func (v *VCenterDriver) GetVSphereClusters(ctx context.Context, datacenter string) ([]string, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vSphere clusters")
	}

	if len(ccrs) == 0 {
		return nil, errors.New("No compute clusters found")
	}

	client := ccrs[0].Client()

	tags, categoryID, err := v.getTagsAndCategory(ctx, client, "ClusterComputeResource", ComputeClusterTagCategory)
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, ccr := range ccrs {
		if v.ifTagHasCategory(tags[ccr.Reference().Value].Tags, categoryID) {
			prefix := fmt.Sprintf("/%s/host/", dc)
			cluster := strings.TrimPrefix(ccr.InventoryPath, prefix)
			clusters = append(clusters, cluster)
		}
	}

	if len(clusters) == 0 {
		return nil, errors.Errorf("No compute clusters with tag category %s found", ComputeClusterTagCategory)
	}

	sort.Strings(clusters)
	return clusters, nil
}

func (v *VCenterDriver) getClusterComputeResources(ctx context.Context, finder *find.Finder) ([]*object.ClusterComputeResource, error) {
	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get compute cluster resources: %s", err.Error())
	}
	return ccrs, nil
}
